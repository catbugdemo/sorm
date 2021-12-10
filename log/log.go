package log

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var ErrRecordNotFound = errors.New("record not found")

var (
	errLog  = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags)
	infoLog = log.New(os.Stdout, "\033[32m[info]\033[0m ", log.LstdFlags)
	loggers = []*log.Logger{errLog, infoLog}
	mu      sync.Mutex
)

var (
	Error  = errLog.Println
	Errorf = errLog.Printf
	Infof  = infoLog.Printf
	Info   = infoLog.Println
)

// log levels
const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

// SetLevel controls log level
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if ErrorLevel < level {
		errLog.SetOutput(ioutil.Discard)
	}
	if InfoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}
