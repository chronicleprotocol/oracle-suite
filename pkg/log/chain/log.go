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

type Logger struct {
	loggers []log.Logger
}

func New(loggers ...log.Logger) *Logger {
	return &Logger{
		loggers: loggers,
	}
}

func (c *Logger) Level() log.Level {
	lvl := log.Panic
	for _, l := range c.loggers {
		if l.Level() > lvl {
			lvl = l.Level()
		}
	}
	return lvl
}

func (c *Logger) WithField(key string, value interface{}) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithField(key, value)
	}
	return &Logger{loggers: loggers}
}

func (c *Logger) WithFields(fields log.Fields) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithFields(fields)
	}
	return &Logger{loggers: loggers}
}

func (c *Logger) WithError(err error) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithError(err)
	}
	return &Logger{loggers: loggers}
}

func (c *Logger) Debugf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Debugf(format, args...)
	}
}

func (c *Logger) Infof(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Infof(format, args...)
	}
}

func (c *Logger) Warnf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Warnf(format, args...)
	}
}

func (c *Logger) Errorf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Errorf(format, args...)
	}
}

func (c *Logger) Panicf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		func() {
			defer func() { recover() }() //nolint:errcheck
			l.Panicf(format, args...)
		}()
	}
	panic(fmt.Sprintf(format, args...))
}

func (c *Logger) Debug(args ...interface{}) {
	for _, l := range c.loggers {
		l.Debug(args...)
	}
}

func (c *Logger) Info(args ...interface{}) {
	for _, l := range c.loggers {
		l.Info(args...)
	}
}

func (c *Logger) Warn(args ...interface{}) {
	for _, l := range c.loggers {
		l.Warn(args...)
	}
}

func (c *Logger) Error(args ...interface{}) {
	for _, l := range c.loggers {
		l.Error(args...)
	}
}

func (c *Logger) Panic(args ...interface{}) {
	for _, l := range c.loggers {
		func() {
			defer func() { recover() }() //nolint:errcheck
			l.Panic(args...)
		}()
	}
	panic(fmt.Sprint(args...))
}
