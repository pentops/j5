package j5client

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/lib/j5query"
	"github.com/pentops/j5/lib/j5schema"
)

func buildListRequest(response j5schema.RootSchema) (*client_j5pb.ListRequest, error) {
	responseObj, ok := response.(*j5schema.ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("expected object schema, got %T", response)
	}

	var foundArray *j5schema.ArrayField

	for _, field := range responseObj.Properties {
		asArray, ok := field.Schema.(*j5schema.ArrayField)
		if !ok {
			continue
		}

		if foundArray != nil {
			return nil, fmt.Errorf("found multiple arrays in response")
		}

		foundArray = asArray
	}

	if foundArray == nil {
		return nil, fmt.Errorf("no array found in response")
	}

	rootSchema, ok := foundArray.ItemSchema.(*j5schema.ObjectField)
	if !ok {
		return nil, fmt.Errorf("expected object schema, got %T", foundArray.ItemSchema)
	}

	out := &client_j5pb.ListRequest{}

	addSearch := func(field j5query.Path, searching *list_j5pb.SearchingConstraint) {
		if searching == nil {
			return
		}

		out.SearchableFields = append(out.SearchableFields, &client_j5pb.ListRequest_SearchField{
			Name: field.ClientPath(),
		})
	}

	addFilter := func(field j5query.Path, filtering *list_j5pb.FilteringConstraint) {
		if filtering == nil {
			return
		}

		filter := &client_j5pb.ListRequest_FilterField{
			Name:           field.ClientPath(),
			DefaultFilters: filtering.DefaultFilters,
		}

		out.FilterableFields = append(out.FilterableFields, filter)
	}

	addSort := func(field j5query.Path, sorting *list_j5pb.SortingConstraint) {
		if sorting == nil {
			return
		}

		var ds *client_j5pb.ListRequest_SortField_Direction
		if sorting.DefaultSort {
			direction := client_j5pb.ListRequest_SortField_DIRECTION_ASC
			typeName := field.LeafField().Schema.TypeName()
			if typeName == "timestamp" {
				direction = client_j5pb.ListRequest_SortField_DIRECTION_DESC
			}

			ds = &direction
		}

		out.SortableFields = append(out.SortableFields, &client_j5pb.ListRequest_SortField{
			Name:        field.ClientPath(),
			DefaultSort: ds,
		})
	}

	err := j5query.WalkPathNodes(rootSchema.ObjectSchema(), func(path j5query.Path) error {
		field := path.LeafField()

		rules := j5query.FieldListRules(field)
		if rules == nil {
			return nil
		}

		addFilter(path, rules.Filter)
		addSort(path, rules.Sort)
		addSearch(path, rules.Search)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk path nodes: %w", err)
	}

	return out, nil
}
