//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package chain

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

// Logger is a log.Logger implementation that can chain multiple loggers.
type Logger struct {
	loggers []log.Logger
}

// New creates a new chain.Logger instance.
func New(loggers ...log.Logger) log.Logger {
	return &Logger{
		loggers: loggers,
	}
}

// Level implements the log.Logger interface.
func (c *Logger) Level() log.Level {
	lvl := log.Panic
	for _, l := range c.loggers {
		if l.Level() > lvl {
			lvl = l.Level()
		}
	}
	return lvl
}

// WithField implements the log.Logger interface.
func (c *Logger) WithField(key string, value interface{}) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithField(key, value)
	}
	return &Logger{loggers: loggers}
}

// WithFields implements the log.Logger interface.
func (c *Logger) WithFields(fields log.Fields) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithFields(fields)
	}
	return &Logger{loggers: loggers}
}

// WithError implements the log.Logger interface.
func (c *Logger) WithError(err error) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithError(err)
	}
	return &Logger{loggers: loggers}
}

// Debugf implements the log.Logger interface.
func (c *Logger) Debugf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Debugf(format, args...)
	}
}

// Infof implements the log.Logger interface.
func (c *Logger) Infof(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Infof(format, args...)
	}
}

// Warnf implements the log.Logger interface.
func (c *Logger) Warnf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Warnf(format, args...)
	}
}

// Errorf implements the log.Logger interface.
func (c *Logger) Errorf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Errorf(format, args...)
	}
}

// Panicf implements the log.Logger interface.
func (c *Logger) Panicf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		func() {
			defer func() { recover() }() //nolint:errcheck // same panic is thrown below
			l.Panicf(format, args...)
		}()
	}
	panic(fmt.Sprintf(format, args...))
}

// Debug implements the log.Logger interface.
func (c *Logger) Debug(args ...interface{}) {
	for _, l := range c.loggers {
		l.Debug(args...)
	}
}

// Info implements the log.Logger interface.
func (c *Logger) Info(args ...interface{}) {
	for _, l := range c.loggers {
		l.Info(args...)
	}
}

// Warn implements the log.Logger interface.
func (c *Logger) Warn(args ...interface{}) {
	for _, l := range c.loggers {
		l.Warn(args...)
	}
}

// Error implements the log.Logger interface.
func (c *Logger) Error(args ...interface{}) {
	for _, l := range c.loggers {
		l.Error(args...)
	}
}

// Panic implements the log.Logger interface.
func (c *Logger) Panic(args ...interface{}) {
	for _, l := range c.loggers {
		func() {
			defer func() { recover() }() //nolint:errcheck // same panic is thrown below
			l.Panic(args...)
		}()
	}
	panic(fmt.Sprint(args...))
}
