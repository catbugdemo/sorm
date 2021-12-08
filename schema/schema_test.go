package schema

import (
	"github.com/catbugdemo/sorm/dialect"
	"github.com/stretchr/testify/assert"
	"testing"
)

type User struct {
	Name string `db:"name"`
	Age  int
}

var TestDial, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, TestDial)

	assert.Equal(t, "User", schema.Name)
	assert.Len(t, schema.Fields, 2)
	assert.Equal(t, schema.GetField("Name").Type, "text")
}
