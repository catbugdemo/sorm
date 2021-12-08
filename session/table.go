package session

import (
	"fmt"
	"github.com/catbugdemo/sorm/log"
	"github.com/catbugdemo/sorm/schema"
	"reflect"
	"strings"
)

func (s *Session) Model(value interface{}) *Session {
	// nil or different model , update refTable
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("Model is not set")
	}
	return s.refTable
}

func (s *Session) CreateTable() error {
	table := s.RefTable()
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.SqlName, field.Type, field.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s)", table.Name, desc)).Exec()
	return err
}

func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s;", s.refTable.Name)).Exec()
	return err
}

func (s *Session) HasTable() bool {
	sql, values := s.dialect.TableExistSQL(s.RefTable().Name)
	rows := s.Raw(sql, values...).QueryRow()
	var tmp string
	if err := rows.Scan(&tmp); err != nil {
		log.Error(err)
		return false
	}
	return tmp == s.RefTable().Name
}
