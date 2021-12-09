package main

import (
	"fmt"
	"github.com/catbugdemo/sorm/dialect"
	"github.com/catbugdemo/sorm/session"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"testing"
	"time"
)

type TestUser struct {
	Id   int `sorm:"PRIMARY KEY"`
	Name string
}

func InitDB() *session.Session {
	engine, err := NewEngine("postgres", fmt.Sprint("host=118.89.121.211 port=5432 user=postgres password=123456 dbname=mydb sslmode=disable"))
	if err != nil {
		panic(err)
	}
	return engine.NewSession()
}

func TestSession_CreateTable(t *testing.T) {
	engine, err := NewEngine("postgres", fmt.Sprint("host=118.89.121.211 port=5432 user=postgres password=123456 dbname=mydb sslmode=disable"))
	if err != nil {
		panic(err)
	}
	defer engine.Close()
	model := engine.NewSession().Model(&TestUser{})

	_ = model.DropTable()
	_ = model.CreateTable()
	if !model.HasTable() {
		t.Fatal("Failed to create table test_user")
	}
}

type UserTest struct {
	Id         int       `db:"id"`
	CreateTime time.Time `db:"create_time"`
	Name       string    `db:"name"`
	NameId     int       `db:"name_id"`
}

func TestInsert(t *testing.T) {
	db := InitDB().Model(&UserTest{})
	t.Run("reflect", func(t *testing.T) {
		if reflect.Indirect(reflect.ValueOf(&UserTest{})).Type() == reflect.Indirect(reflect.ValueOf(&UserTest{Name: "111", NameId: 111})).Type() {
			fmt.Println("success")
		} else {
			fmt.Println("fail")
		}
	})

	t.Run("Insert", func(t *testing.T) {
		test1 := UserTest{
			Name:   "222",
			NameId: 333}
		test2 := UserTest{
			Name:   "555",
			NameId: 666,
		}
		ut := make([]UserTest, 0, 2)
		ut = append(ut, test1, test2)
		err := db.Create(&test1)
		if err != nil {
			panic(err)
		}
		fmt.Println(&test1)
	})

	t.Run("Find", func(t *testing.T) {
		var users []UserTest
		err := db.Find(&users)
		if err != nil {
			panic(err)
		}
		fmt.Println(users)
	})

	t.Run("update", func(t *testing.T) {
		err := db.Model(UserTest{}).Where("id=?", 73).Update("name", "111")
		assert.Nil(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		err := db.Where("name=?", "111").Delete()
		assert.Nil(t, err)
	})

	t.Run("count", func(t *testing.T) {
		var count int
		err := db.Count(&count)
		assert.Nil(t, err)
		fmt.Println(count)
	})

	t.Run("first", func(t *testing.T) {
		var ut UserTest
		err := db.Where("id in (?)", []int{74}).Where("name=?", 222).First(&ut)
		assert.Nil(t, err)
		fmt.Println(ut)
	})

	t.Run("rollabck", func(t *testing.T) {
		d := sqlx.DB{}
		driverName := reflect.Indirect(reflect.ValueOf(d)).FieldByName("driverName").String()
		dial, ok := dialect.GetDialect(driverName)
		if !ok {
			log.Fatalf("dialect %s Not Found", driverName)
			return
		}
		_ = session.New(d.DB, dial)
	})
}

/*func (o *UserTest) AfterQuery(s *session.Session) error {
	log.Println("After query: ", o)
	o.NameId = 0
	return nil
}*/

func TestHasTable(t *testing.T) {
	db := InitDB()
	if !db.Model(UserTest{}).HasTable() {
		log.Println("fail")
	}

	/*	d := gorm.DB{}
		d.Create()*/
}
