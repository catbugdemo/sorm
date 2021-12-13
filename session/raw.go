package session

import (
	"database/sql"
	"github.com/catbugdemo/sorm/clause"
	"github.com/catbugdemo/sorm/dialect"
	"github.com/catbugdemo/sorm/log"
	"github.com/catbugdemo/sorm/schema"
	"reflect"
	"strconv"
	"strings"
)

type Session struct {
	db       *sql.DB
	dialect  dialect.Dialect
	tx       *sql.Tx
	refTable *schema.Schema
	clause   clause.Clause
	sql      strings.Builder
	sqlVars  []interface{}
	content  Content
}

// CommonDB is a minimal function set of db
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(qury string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// Raw 原生查询
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec raw sql with sqlVars
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	// 将 ？ 改成 $num
	sql, sqlVars, logs := QueToDoller(s.sql.String(), s.sqlVars)
	log.Info(": " + logs)
	if result, err = s.DB().Exec(sql, sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// QueryRow gets a record from db
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	sql, sqlVars, logs := QueToDoller(s.sql.String(), s.sqlVars)
	log.Info(": " + logs)
	return s.DB().QueryRow(sql, sqlVars...)
}

// QueryRows gets a list of records from db
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	sql, sqlVars, logs := QueToDoller(s.sql.String(), s.sqlVars)
	log.Info(": " + logs)
	if rows, err = s.DB().Query(sql, sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// QueToDoller  ? to $num
func QueToDoller(sql string, vars []interface{}) (string, []interface{}, string) {
	sql = strings.ReplaceAll(sql, " in ", " IN ")
	if strings.Contains(sql, " IN ") {
		split := strings.Split(sql, " IN ")
		count := strings.Count(split[0], "?")
		for i := 1; i < len(split); i++ {
			reflectValue := reflect.Indirect(reflect.ValueOf(vars[count]))
			switch reflectValue.Kind() {
			case reflect.Slice, reflect.Array:
				reflectLen := reflectValue.Len()
				vars = append(vars[:count], vars[count+1:]...) // delete slice
				for j := 0; j < reflectLen; j++ {
					vars = append(vars[:count+j], append([]interface{}{reflectValue.Index(j).Interface()}, vars[count+j:]...)...)
				}
				// 修改 ? 数量
				repeat := strings.Repeat("?,", reflectLen)
				split[i] = strings.Replace(split[i], "?", repeat[:len(repeat)-1], 1)
			}
			count += strings.Count(split[i], "?")
		}
		// 计数 ? 数量
		sql = strings.Join(split, " IN ")
	}
	// log
	queCount := strings.Count(sql, "?")
	logs := sql
	for i := 0; i < queCount; i++ {
		logs = strings.Replace(sql, "?", "'"+vars[i].(string)+"'", 1)
	}

	// ? to $num
	for i := 0; i < queCount; i++ {
		sql = strings.Replace(sql, "?", "$"+strconv.Itoa(i+1), 1)
	}
	return sql, vars, logs
}
