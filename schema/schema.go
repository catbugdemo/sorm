package schema

import (
	"github.com/catbugdemo/sorm/dialect"
	"go/ast"
	"reflect"
	"strings"
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
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// Parse 将任意的对象解析为 Schema 实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     GetUnderlineName(modelType.Name()),
		fieldMap: make(map[string]*Field),
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
		}
	}
	return schema
}

func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
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
