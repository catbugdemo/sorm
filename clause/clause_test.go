package clause

import "testing"

func TestSelect(t *testing.T) {
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "test_user", []string{"*"})
	clause.Set(WHERE, "Name = ?", "Tom")
	clause.Set(ORDERBY, "name Asc")

	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)
}
