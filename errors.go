package sorm

import (
	"errors"
	"github.com/catbugdemo/sorm/log"
)

var (
	// ErrRecordNotFound
	ErrRecordNotFound = log.ErrRecordNotFound
	//
	ErrValuesNotPointer = errors.New("values not pointer")
)
