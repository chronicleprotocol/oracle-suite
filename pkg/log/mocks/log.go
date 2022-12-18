package mocks

import (
	"sync"

	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

type Logger struct {
	mock mock.Mock
	mu   sync.Mutex

	fields log.Fields
}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) Mock() *mock.Mock {
	l.mu.Lock()
	defer l.mu.Unlock()
	return &l.mock
}

func (l *Logger) Level() log.Level {
	return log.Debug
}

func (l *Logger) WithField(key string, value interface{}) log.Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	args := l.mock.Called(key, value)
	return args.Get(0).(log.Logger)
}

func (l *Logger) WithFields(fields log.Fields) log.Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	args := l.mock.Called(fields)
	return args.Get(0).(log.Logger)
}

func (l *Logger) WithError(err error) log.Logger {
	args := l.mock.Called(err)
	l.mu.Lock()
	defer l.mu.Unlock()
	return args.Get(0).(log.Logger)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(format, args)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(format, args)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(format, args)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(format, args)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(format, args)
}

func (l *Logger) Debug(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(args)
}

func (l *Logger) Info(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(args)
}

func (l *Logger) Warn(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(args)
}

func (l *Logger) Error(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(args)
}

func (l *Logger) Panic(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mock.Called(args)
}
