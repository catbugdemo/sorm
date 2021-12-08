package clause

import (
	"fmt"
	"reflect"
	"strings"
)

type generator func(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
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

func _insert(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s(%v)", tableName, fields), []interface{}{}
}

func _values(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
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

func _select(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

func _limit(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
	// LIMIT $num
	return "LIMIT ?", values
}

func _where(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	str := strings.ToLower(desc.(string))
	str, vars = checkIn(str, vars...)
	// 已经存在数据了
	if len(sqlVars) > 0 {
		sql += " and " + str
		sqlVars = append(sqlVars, vars...)
		return fmt.Sprintf("%s", sql), sqlVars
	}
	return fmt.Sprintf("WHERE %s", str), vars
}

// checkIn in 的特殊处理
func checkIn(str string, vars ...interface{}) (string, []interface{}) {
	if strings.Contains(str, " in") {
		split := strings.Split(str, " in")
		var count int
		for i, _ := range split {
			if i > 0 {
				reflectValue := reflect.Indirect(reflect.ValueOf(vars[count]))
				reflectLen := reflectValue.Len()
				switch reflectValue.Kind() {
				case reflect.Slice, reflect.Array:
					vars = append(vars[:count], vars[count+1:]...) // delete slice
					for j := 0; j < reflectLen; j++ {
						vars = append(vars[:count+j], append([]interface{}{reflectValue.Index(j).Interface()}, vars[count+j:]...)...)
					}
				}
				// 修改 ? 数量
				repeat := strings.Repeat("?,", reflectLen)
				split[i] = strings.Replace(split[i], "?", repeat[:len(repeat)-1], 1)
			}
			// 计数 ? 数量
			count += strings.Count(split[i], "?")
		}
		str = strings.Join(split, " IN")
	}
	return str, vars
}

func _orderby(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

func _update(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
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

func _delete(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

func _count(sql string, sqlVars []interface{}, values ...interface{}) (string, []interface{}) {
	return _select(sql, sqlVars, values[0], []string{"count(*)"})
}
