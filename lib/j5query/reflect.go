package j5query

import (
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/lib/j5schema"
)

type ListRules struct {
	Sort   *list_j5pb.SortingConstraint
	Filter *list_j5pb.FilteringConstraint
	Search *list_j5pb.SearchingConstraint
}

func FieldListRules(field *j5schema.ObjectProperty) *ListRules {
	out := &ListRules{}
	out.Sort = getFieldSorting(field)
	out.Filter = getFieldFiltering(field)
	out.Search = getFieldSearching(field)

	return out
}
