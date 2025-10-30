package j5query

import (
	"context"
	"fmt"

	sq "github.com/elgris/sqrl"
	"github.com/elgris/sqrl/pg"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TableReflectionSet struct {
	arrayObject       *j5schema.ObjectSchema
	defaultSortFields []sortSpec
	tieBreakerFields  []sortSpec

	defaultFilterFields []filterSpec

	tsvColumnMap map[string]string // map[JSON Path]PGColumName

	tableName string

	// TODO: This should be an array/map of columns to data types, allowing
	// multiple JSONB values, as well as cached field values direcrly on the
	// table
	dataColumn string
}

func NewTableReflectionSet(table TableSpec) (*TableReflectionSet, error) {
	err := table.Validate()
	if err != nil {
		return nil, fmt.Errorf("new table reflection: validate table spec: %w", err)
	}

	ll := &TableReflectionSet{
		dataColumn:  table.DataColumn,
		arrayObject: table.RootObject,
		tableName:   table.TableName,
	}

	ll.defaultSortFields, err = buildDefaultSorts(ll.dataColumn, ll.arrayObject)
	if err != nil {
		return nil, fmt.Errorf("default sorts: %w", err)
	}

	ll.tieBreakerFields, err = buildFallbackTieBreakerFields(ll.dataColumn, ll.arrayObject, table.FallbackSortColumns)
	if err != nil {
		return nil, fmt.Errorf("tie breaker fields: %w", err)
	}

	f, err := buildDefaultFilters(ll.dataColumn, ll.arrayObject)
	if err != nil {
		return nil, fmt.Errorf("default filters: %w", err)
	}

	ll.defaultFilterFields = f

	ll.tsvColumnMap, err = buildTsvColumnMap(ll.arrayObject)
	if err != nil {
		return nil, fmt.Errorf("build tsv column map: %w", err)
	}
	return ll, nil
}

func (ll *TableReflectionSet) ArrayObject() *j5schema.ObjectSchema {
	return ll.arrayObject
}

func (ll *TableReflectionSet) BuildQuery(ctx context.Context, reqQuery *list_j5pb.QueryRequest) (*Query, error) {
	as := newAliasSet()
	tableAlias := as.Next(ll.tableName)

	selectQuery := sq.Select().
		From(fmt.Sprintf("%s AS %s", ll.tableName, tableAlias))

	query := &Query{
		aliasSet:       as,
		rootTableAlias: tableAlias,
		mainDataColumn: ll.dataColumn,
		SelectBuilder:  selectQuery,
		sortFields:     append(ll.defaultSortFields, ll.tieBreakerFields...),
	}

	filterFields := []sq.Sqlizer{}
	if reqQuery != nil {
		if err := ll.validateQueryRequest(reqQuery); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "query validation: %s", err)
		}

		querySorts := reqQuery.GetSorts()
		if len(querySorts) > 0 {
			dynSorts, err := ll.buildDynamicSortSpec(querySorts)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "build sorts: %s", err)
			}

			query.sortFields = dynSorts
		}

		queryFilters := reqQuery.GetFilters()
		if len(queryFilters) > 0 {
			dynFilters, err := ll.buildDynamicFilter(tableAlias, queryFilters)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "build filters: %s", err)
			}

			filterFields = append(filterFields, dynFilters...)
		}

		querySearches := reqQuery.GetSearches()
		if len(querySearches) > 0 {
			searchFilters, err := ll.buildDynamicSearches(tableAlias, querySearches)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "build searches: %s", err)
			}

			filterFields = append(filterFields, searchFilters...)
		}
	}

	for i := range filterFields {
		selectQuery.Where(filterFields[i])
	}

	// apply default filters if no filters have been requested
	if ll.defaultFilterFields != nil && len(filterFields) == 0 {
		and := sq.And{}
		for _, spec := range ll.defaultFilterFields {
			or := sq.Or{}
			for _, val := range spec.filterVals {
				or = append(or, sq.Expr(fmt.Sprintf("jsonb_path_query_array(%s.%s, '%s') @> ?", tableAlias, ll.dataColumn, spec.Path.JSONPathQuery()), pg.JSONB(val)))
			}

			and = append(and, or)
		}

		if len(and) > 0 {
			selectQuery.Where(and)
		}
	}

	for _, sortField := range query.sortFields {
		direction := "ASC"
		if sortField.desc {
			direction = "DESC"
		}
		selectQuery.OrderBy(fmt.Sprintf("%s %s", sortField.Selector(tableAlias), direction))
	}

	return query, nil
}

func (ll *TableReflectionSet) validateQueryRequest(query *list_j5pb.QueryRequest) error {
	err := validateQueryRequestSorts(ll.arrayObject, query.GetSorts())
	if err != nil {
		return fmt.Errorf("sort validation: %w", err)
	}

	err = validateQueryRequestFilters(ll.arrayObject, query.GetFilters())
	if err != nil {
		return fmt.Errorf("filter validation: %w", err)
	}

	/*
		err = validateQueryRequestSearches(ll.arrayObject, query.GetSearches())
		if err != nil {
			return fmt.Errorf("search validation: %w", err)
		}*/

	return nil
}

func (ll *TableReflectionSet) buildDynamicSortSpec(sorts []*list_j5pb.Sort) ([]sortSpec, error) {
	results := []sortSpec{}
	direction := ""
	for _, sort := range sorts {
		pathSpec := ParseJSONPathSpec(sort.Field)
		spec, err := NewJSONPath(ll.arrayObject, pathSpec)
		if err != nil {
			return nil, fmt.Errorf("dynamic filter: find field: %w", err)
		}

		biggerSpec := &NestedField{
			Path:       *spec,
			RootColumn: ll.dataColumn,
		}

		results = append(results, sortSpec{
			NestedField: biggerSpec,
			desc:        sort.Descending,
		})

		// TODO: Remove this constraint, we can sort by different directions once we have the reversal logic in place
		// validate direction of all the fields is the same
		if direction == "" {
			direction = "ASC"
			if sort.Descending {
				direction = "DESC"
			}
		} else {
			if (direction == "DESC" && !sort.Descending) || (direction == "ASC" && sort.Descending) {
				return nil, fmt.Errorf("requested sorts have conflicting directions, they must all be the same")
			}
		}
	}

	return results, nil
}
