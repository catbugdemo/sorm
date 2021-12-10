package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[TABLE] = _table
	generators[LIMIT] = _limit
	generators[OFFSET] = _offset
	generators[WHERE] = _where
	generators[ORDERBY] = _orderby
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}

func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ",")
}

func _insert(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s(%v)", tableName, fields), []interface{}{}
}

func _values(values ...interface{}) (string, []interface{}) {
	// VALUES ($v1) ,($v2) , ...
	var bindStr string
	var sqlStr strings.Builder
	var vars []interface{}
	sqlStr.WriteString("VALUES ")
	for i, value := range values[1:] {
		v := value.([]interface{})
		if bindStr == "" {
			bindStr = genBindVars(len(v))
		}
		sqlStr.WriteString(fmt.Sprintf("(%v)", bindStr))
		if i+2 != len(values) {
			sqlStr.WriteString(",")
		}
		vars = append(vars, v...)
	}
	sqlStr.WriteString(fmt.Sprintf(" RETURNING %v", strings.Join(values[0].([]string), ",")))
	return sqlStr.String(), vars
}

func _select(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	return fmt.Sprintf("SELECT %v FROM", strings.Join(values[0].([]string), ",")), []interface{}{}
}

func _table(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("%s ", values[0]), []interface{}{}
}

func _limit(values ...interface{}) (string, []interface{}) {
	// LIMIT $num
	return "LIMIT ?", values
}

func _where(values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	if strings.Contains(desc.(string), " WHERE ") {
		return fmt.Sprintf("%s", desc), vars
	}
	return fmt.Sprintf(" WHERE %s", desc), vars
}

func _orderby(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

func _offset(values ...interface{}) (string, []interface{}) {
	return "OFFSET ?", values
}

func _update(values ...interface{}) (string, []interface{}) {
	// UPDATE $tableName set $fields
	tableName := values[0]
	m := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for k, v := range m {
		keys = append(keys, k+"=?")
		vars = append(vars, v)
	}
	return fmt.Sprintf("UPDATE %v SET %v", tableName, strings.Join(keys, ",")), vars
}

func _delete(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

func _count(values ...interface{}) (string, []interface{}) {
	return _select([]string{"count(*)"})
}
