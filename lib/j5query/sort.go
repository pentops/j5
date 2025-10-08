package j5query

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/lib/j5schema"
)

type sortSpec struct {
	*NestedField
	desc bool
}

func (ss sortSpec) errorName() string {
	return ss.Path.JSONPathQuery()
}

func buildFallbackTieBreakerFields(dataColumn string, rootObject *j5schema.ObjectSchema, fallback []ProtoField) ([]sortSpec, error) {
	tieBreakerFields := make([]sortSpec, 0, len(fallback))
	for _, tieBreaker := range fallback {

		path, err := NewJSONPath(rootObject, tieBreaker.pathInRoot)
		if err != nil {
			return nil, fmt.Errorf("field %s in fallback sort tiebreaker for %s: %w", tieBreaker.pathInRoot, rootObject.FullName(), err)
		}

		tieBreakerFields = append(tieBreakerFields, sortSpec{
			NestedField: &NestedField{
				Path:        *path,
				RootColumn:  dataColumn,
				ValueColumn: tieBreaker.valueColumn,
			},
			desc: false,
		})
	}

	return tieBreakerFields, nil
}

func buildRequestObjectTieBreakerFields(dataColumn string, req *j5schema.ObjectSchema, rootObject *j5schema.ObjectSchema) ([]sortSpec, error) {
	tieBreakerFields := make([]sortSpec, 0, len(req.ListRequest.SortTiebreaker))
	for _, tieBreaker := range req.ListRequest.SortTiebreaker {
		spec, err := NewJSONPath(rootObject, tieBreaker)
		if err != nil {
			return nil, fmt.Errorf("field %s in annotated sort tiebreaker for %s: %w", tieBreaker, req.FullName(), err)
		}

		tieBreakerFields = append(tieBreakerFields, sortSpec{
			NestedField: &NestedField{
				RootColumn: dataColumn,
				Path:       *spec,
			},
			desc: false,
		})
	}

	return tieBreakerFields, nil
}

func IsPathSortable(path *Path) bool {
	var isPathSortable = true
	path.WalkPathNodes(func(prop *j5schema.ObjectProperty) bool {
		switch prop.Schema.(type) {
		case *j5schema.MapField, *j5schema.ArrayField:
			isPathSortable = false
		}

		return isPathSortable
	})

	return isPathSortable
}

func getFieldSorting(path *Path) *list_j5pb.SortingConstraint {
	field := path.LeafField()
	if field == nil {
		return nil // oneof or something
	}

	if !IsPathSortable(path) {
		return nil
	}

	scalar, ok := field.Schema.(*j5schema.ScalarSchema)
	if !ok {
		return nil // only scalars are sortable
	}

	schema := scalar.ToJ5Field()

	switch st := schema.Type.(type) {
	case *schema_j5pb.Field_Float:
		if st.Float.ListRules == nil {
			return nil
		}
		return st.Float.ListRules.Sorting
	case *schema_j5pb.Field_Integer:
		if st.Integer.ListRules == nil {
			return nil
		}
		return st.Integer.ListRules.Sorting
	case *schema_j5pb.Field_Timestamp:
		if st.Timestamp.ListRules == nil {
			return nil
		}
		return st.Timestamp.ListRules.Sorting
	case *schema_j5pb.Field_Date:
		if st.Date.ListRules == nil {
			return nil
		}
		return st.Date.ListRules.Sorting
	case *schema_j5pb.Field_Decimal:
		if st.Decimal.ListRules == nil {
			return nil
		}
		return st.Decimal.ListRules.Sorting
	default:
		return nil
	}
}

func buildDefaultSorts(columnName string, message *j5schema.ObjectSchema) ([]sortSpec, error) {
	var defaultSortFields []sortSpec

	err := WalkPathNodes(message, func(path Path) error {
		sortConstraint := getFieldSorting(&path)
		if sortConstraint == nil {
			return nil // not a sortable field
		}

		if sortConstraint.DefaultSort {
			defaultSortFields = append(defaultSortFields, sortSpec{
				NestedField: &NestedField{
					RootColumn: columnName,
					Path:       path,
				},
				desc: true,
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return defaultSortFields, nil
}

func validateQueryRequestSorts(message *j5schema.ObjectSchema, sorts []*list_j5pb.Sort) error {
	for _, sort := range sorts {
		pathSpec := ParseJSONPathSpec(sort.Field)
		path, err := NewJSONPath(message, pathSpec)
		if err != nil {
			return fmt.Errorf("find field %s: %w", sort.Field, err)
		}

		field := path.LeafField()
		if field == nil {
			return fmt.Errorf("node %s is not a field", path.DebugName())
		}

		sortAnnotation := getFieldSorting(path)
		if sortAnnotation == nil || !sortAnnotation.Sortable {
			return fmt.Errorf("requested sort field '%s' is not sortable", sort.Field)
		}
	}

	return nil
}
