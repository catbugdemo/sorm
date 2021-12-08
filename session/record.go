package session

import (
	"errors"
	"github.com/catbugdemo/sorm/clause"
	"github.com/catbugdemo/sorm/log"
	"reflect"
)

func (s *Session) Insert(values interface{}) error {
	if reflect.ValueOf(values).Kind() != reflect.Ptr {
		return errors.New("values not pointer")
	}
	recordValues := make([]interface{}, 0)
	for _, value := range insertInBatches(values) {
		s.CallMethod(BeforeInsert, value)
		table := s.Model(value).RefTable()
		if len(recordValues) == 0 { // 添加返回数据
			recordValues = append(recordValues, table.FieldNames)
		}
		fieldSqlNames, fieldValues := table.RecordValues(value)
		s.clause.Set(clause.INSERT, table.Name, fieldSqlNames)
		recordValues = append(recordValues, fieldValues)
	}
	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	// binding returning
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	var index int
	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var result []interface{}
		for _, field := range s.RefTable().Fields {
			result = append(result, dest.FieldByName(field.Name).Addr().Interface())
		}
		if err = rows.Scan(result...); err != nil {
			return err
		}
		s.CallMethod(AfterInsert, nil)
		destSlice.Index(index).Set(dest)
		index++
	}
	log.Info("INSERT affects rows:", index)
	return nil
}

func (s *Session) Find(values interface{}) error {
	s.CallMethod(BeforeQuery, nil)
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var result []interface{}
		for _, field := range table.Fields {
			result = append(result, dest.FieldByName(field.Name).Addr().Interface())
		}
		if err = rows.Scan(result...); err != nil {
			return err
		}
		s.CallMethod(AfterQuery, dest.Addr().Interface())
		destSlice.Set(reflect.Append(destSlice, dest))
	}

	return rows.Close()
}

// insertInBatches imitate gorm CreateInBatches
func insertInBatches(value interface{}) []interface{} {
	values := make([]interface{}, 0)
	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	switch reflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		reflectLen := reflectValue.Len()
		for i := 0; i < reflectLen; i++ {
			values = append(values, reflectValue.Index(i).Interface())
		}
	default:
		values = append(values, value)
	}
	return values
}

// support map[string]interface{}
// also support kv list: "Name", "Tom", "Age", 18, ....
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.CallMethod(BeforeUpdate, nil)
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterUpdate, nil)
	return result.RowsAffected()
}

/*func (s *Session) Updates(kv ...interface{}) (int64, error) {

}*/

func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return result.RowsAffected()
}

func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

func (s *Session) First(values interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(values))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("record not found")
	}
	dest.Set(destSlice.Index(0))
	return nil
}
