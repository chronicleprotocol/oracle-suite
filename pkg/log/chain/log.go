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
	"context"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

// New creates a new logger that can chain multiple loggers.
func New(loggers ...log.Logger) log.Logger {
	return &logger{
		loggers: loggers,
	}
}

type logger struct {
	ctx    context.Context
	waitCh chan error

	loggers []log.Logger
}

// Level implements the log.Logger interface.
func (c *logger) Level() log.Level {
	lvl := log.Panic
	for _, l := range c.loggers {
		if l.Level() > lvl {
			lvl = l.Level()
		}
	}
	return lvl
}

// WithField implements the log.Logger interface.
func (c *logger) WithField(key string, value interface{}) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithField(key, value)
	}
	return &logger{loggers: loggers}
}

// WithFields implements the log.Logger interface.
func (c *logger) WithFields(fields log.Fields) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithFields(fields)
	}
	return &logger{loggers: loggers}
}

// WithError implements the log.Logger interface.
func (c *logger) WithError(err error) log.Logger {
	loggers := make([]log.Logger, len(c.loggers))
	for n, l := range c.loggers {
		loggers[n] = l.WithError(err)
	}
	return &logger{loggers: loggers}
}

// Debugf implements the log.Logger interface.
func (c *logger) Debugf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Debugf(format, args...)
	}
}

// Infof implements the log.Logger interface.
func (c *logger) Infof(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Infof(format, args...)
	}
}

// Warnf implements the log.Logger interface.
func (c *logger) Warnf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Warnf(format, args...)
	}
}

// Errorf implements the log.Logger interface.
func (c *logger) Errorf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		l.Errorf(format, args...)
	}
}

// Panicf implements the log.Logger interface.
func (c *logger) Panicf(format string, args ...interface{}) {
	for _, l := range c.loggers {
		func() {
			defer func() { recover() }() //nolint:errcheck // same panic is thrown below
			l.Panicf(format, args...)
		}()
	}
	panic(fmt.Sprintf(format, args...))
}

// Debug implements the log.Logger interface.
func (c *logger) Debug(args ...interface{}) {
	for _, l := range c.loggers {
		l.Debug(args...)
	}
}

// Info implements the log.Logger interface.
func (c *logger) Info(args ...interface{}) {
	for _, l := range c.loggers {
		l.Info(args...)
	}
}

// Warn implements the log.Logger interface.
func (c *logger) Warn(args ...interface{}) {
	for _, l := range c.loggers {
		l.Warn(args...)
	}
}

// Error implements the log.Logger interface.
func (c *logger) Error(args ...interface{}) {
	for _, l := range c.loggers {
		l.Error(args...)
	}
}

// Panic implements the log.Logger interface.
func (c *logger) Panic(args ...interface{}) {
	for _, l := range c.loggers {
		func() {
			defer func() { recover() }() //nolint:errcheck // same panic is thrown below
			l.Panic(args...)
		}()
	}
	panic(fmt.Sprint(args...))
}
