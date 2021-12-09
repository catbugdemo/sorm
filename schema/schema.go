package schema

import (
	"github.com/catbugdemo/sorm/dialect"
	"go/ast"
	"reflect"
	"strings"
	"time"
)

// Field represents a column of database
type Field struct {
	Name    string
	SqlName string
	Type    string
	Tag     string
}

// Schema represents a table of database
type Schema struct {
	Model       interface{}
	Name        string
	Fields      []*Field
	FieldNames  []string
	fieldMap    map[string]*Field
	FieldSqlMap map[string]string
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// Parse 将任意的对象解析为 Schema 实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:       dest,
		Name:        GetUnderlineName(modelType.Name()),
		fieldMap:    make(map[string]*Field),
		FieldSqlMap: make(map[string]string),
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		// Anonymous 是否匿名字段， IsExported 是否大写
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("sorm"); ok { // table 关键字,如 : primary key
				field.Tag = v
			}
			if v, ok := p.Tag.Lookup("db"); ok {
				field.SqlName = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, field.SqlName)
			schema.fieldMap[p.Name] = field // fieldMap 通过名称作为键值,能够快速查找 field
			schema.FieldSqlMap[field.SqlName] = p.Name
		}
	}
	return schema
}

func (schema *Schema) RecordValues(dest interface{}) ([]string, []interface{}) {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldSqlNames []string
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		if !IsBlank(destValue.FieldByName(field.Name)) {
			fieldSqlNames = append(fieldSqlNames, field.SqlName)
			fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
		}
	}
	return fieldSqlNames, fieldValues
}

// GetUnderlineName generate table_
func GetUnderlineName(name string) string {
	var index, count int
	nameList := make([]string, 0, len(name))
	for i, str := range strings.Split(name, "") {
		if i == 0 {
			continue
		}

		if str >= "A" && str <= "Z" {
			nameList = append(nameList, name[index:i])
			index = i
			count++
		}
	}
	nameList = append(nameList, name[index:])
	return strings.ToLower(strings.Join(nameList, "_"))
}

// IsBlank check if  value == nil
func IsBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		if t, ok := value.Interface().(time.Time); ok {
			return t.IsZero()
		}
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}
