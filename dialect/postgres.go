package dialect

import (
	"fmt"
	"reflect"
	"time"
)

type postgres struct{}

var _ Dialect = (*postgres)(nil)

func init() {
	RegisterDialect("postgres", &postgres{})
}

func (p *postgres) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "int"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "varchar"
	case reflect.Array, reflect.Slice:
		return "bytea"
	case reflect.Struct: // 关于时间的处理方法
		if _, ok := typ.Interface().(time.Time); ok {
			return "timestamp"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

func (p postgres) TableExistSQL(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT tablename FROM pg_tables WHERE schemaname='public' and tablename=$1", args
}
