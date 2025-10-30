package j5query

import (
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
)

type ListRules struct {
	Sort   *list_j5pb.SortingConstraint
	Filter *list_j5pb.FilteringConstraint
	Search *list_j5pb.SearchingConstraint
}

func FieldListRules(path *Path) *ListRules {
	field := path.LeafField()
	if field == nil {
		return nil
	}

	out := &ListRules{}
	out.Sort = getFieldSorting(path)
	out.Filter = getFieldFiltering(field)
	out.Search = getFieldSearching(field)

	return out
}
