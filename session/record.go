package session

import (
	"errors"
	"github.com/catbugdemo/sorm/clause"
	"github.com/catbugdemo/sorm/log"
	"reflect"
)

func (s *Session) Create(values interface{}) error {
	switch reflect.ValueOf(values).Kind() {
	case reflect.Slice, reflect.Array:
		return errors.New("Create can only insert one date")
	case reflect.Ptr:
		dest := reflect.Indirect(reflect.ValueOf(values))
		destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
		destSlice.Set(reflect.Append(destSlice, dest))
		if err := s.Insert(destSlice.Addr().Interface()); err != nil {
			return err
		}
		dest.Set(destSlice.Index(0))
	default:
		return errors.New("Model is not pointer")
	}
	return nil
}

func (s *Session) Insert(values interface{}) error {
	recordValues := make([]interface{}, 0)
	for _, value := range insertInBatches(values) {
		table := s.Model(value).RefTable()
		s.CallMethod(BeforeInsert, value)
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
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()
	s.CallMethod(BeforeQuery, nil)
	if get, _ := s.clause.Get(clause.SELECT); len(get) == 0 {
		s.clause.Set(clause.SELECT, table.FieldNames)
	}
	if get, _ := s.clause.Get(clause.TABLE); len(get) == 0 {
		s.clause.Set(clause.TABLE, table.Name)
	}
	sql, vars := s.clause.Build(clause.SELECT, clause.TABLE, clause.WHERE, clause.ORDERBY, clause.LIMIT, clause.OFFSET)
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
func (s *Session) Update(kv ...interface{}) error {
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
		return err
	}
	s.CallMethod(AfterUpdate, nil)
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	log.Info("Update affects rows:", affected)
	return nil
}

func (s *Session) Delete() error {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return err
	}
	s.CallMethod(AfterDelete, nil)
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	log.Info("DELETE affects rows:", affected)
	return nil
}

func (s *Session) Count(values interface{}) error {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	if err := row.Scan(values); err != nil {
		return err
	}
	return nil
}

func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Select(query interface{}, values ...interface{}) *Session {
	switch v := query.(type) {
	case []string:
		s.clause.Set(clause.SELECT, v)
	default:
		list := make([]string, 0, 1+len(values))
		list = append(list, query.(string))
		for _, value := range values {
			list = append(list, value.(string))
		}
		s.clause.Set(clause.SELECT, list)
	}
	return s
}

func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	sql, sqlVars := s.clause.Get(clause.WHERE)
	if len(sql) > 0 {
		desc = sql + " AND " + desc
		args = append(sqlVars, args...)
	}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

func (s *Session) Offset(num int) *Session {
	s.clause.Set(clause.OFFSET, num)
	return s
}

func (s *Session) First(values interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(values))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return log.ErrRecordNotFound
	}
	dest.Set(destSlice.Index(0))
	return nil
}
