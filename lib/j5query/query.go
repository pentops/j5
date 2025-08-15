package j5query

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/elgris/sqrl"
	"github.com/pentops/sqrlx.go/sqrlx"
)

type aliasSet int

func (as *aliasSet) Next(name string) string {
	*as++
	return fmt.Sprintf("_%s__a%d", name, *as)
}

func newAliasSet() *aliasSet {
	return new(aliasSet)
}

type Transactor interface {
	Transact(ctx context.Context, opts *sqrlx.TxOptions, callback sqrlx.Callback) error
}

type AuthProvider interface {
	AuthFilter(ctx context.Context) (map[string]string, error)
}

type AuthProviderFunc func(ctx context.Context) (map[string]string, error)

func (f AuthProviderFunc) AuthFilter(ctx context.Context) (map[string]string, error) {
	return f(ctx)
}

// LeftJoin is a specification for joining in the form
// <TableName> ON <TableName>.<JoinKeyColumn> = <Main>.<MainKeyColumn>
// Main is defined in the outer struct holding this LeftJoin
type LeftJoin struct {
	TableName string
	On        JoinFields
}

type Query struct {
	aliasSet       *aliasSet
	rootTableAlias string
	mainDataColumn string
	*sq.SelectBuilder
	sortFields []sortSpec
}

func (ll *Query) AddRootColumn() {
	ll.Column(fmt.Sprintf("%s.%s", ll.rootTableAlias, ll.mainDataColumn))

}

func (ll *Query) newAlias(name string) string {
	return ll.aliasSet.Next(name)
}

func (ll *Query) J5LeftJoinRoot(newTable string, joinFields ...string) string {
	newAlias := ll.newAlias(newTable)
	whereClauses := make([]string, 0, len(joinFields))
	for _, field := range joinFields {
		whereClauses = append(whereClauses, fmt.Sprintf("%s.%s = %s.%s", newAlias, field, ll.rootTableAlias, field))
	}
	joinClause := fmt.Sprintf("%s AS %s ON %s", newTable, newAlias, strings.Join(whereClauses, " AND "))
	ll.LeftJoin(joinClause)
	return newAlias
}
