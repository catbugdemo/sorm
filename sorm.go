package sorm

import (
	"database/sql"
	"errors"
	"github.com/catbugdemo/sorm/dialect"
	"github.com/catbugdemo/sorm/log"
	"github.com/catbugdemo/sorm/session"
	"reflect"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func Open(driver, source string) (*session.Session, error) {
	engine, err := NewEngine(driver, source)
	if err != nil {
		return nil, err
	}
	return engine.NewSession(), nil
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// Send a ping to make sure the database connection is alive.
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	// make sure the specific dialect exists
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s Not Found", driver)
		return
	}
	e = &Engine{db: db, dialect: dial}
	log.Info("Connect database success")
	return
}

func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database")
	}
	log.Info("Close database success")
	return
}

func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

func ReplaceSqlx(values interface{}) (*session.Session, error) {
	value := reflect.ValueOf(values)
	switch value.Kind() {
	case reflect.Ptr:
		dest := reflect.Indirect(value)
		driver := dest.FieldByName("driverName").String()
		dial, ok := dialect.GetDialect(driver)
		if !ok {
			log.Errorf("dialect %s Not Found", driver)
			return nil, nil
		}
		return session.New(dest.FieldByName("DB").Interface().(*sql.DB), dial), nil
	default:
		return nil, errors.New("values not pointer sqlx.DB")
	}
}

type TxFunc func(*session.Session) (interface{}, error)

func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := engine.NewSession()
	if err = s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = s.Rollback() // err is non-nil; don't change it
		} else {
			err = s.Commit()
		}
	}()
	return f(s)
}
