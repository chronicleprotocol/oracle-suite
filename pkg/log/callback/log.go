//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package callback

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

type LogFunc func(level log.Level, fields log.Fields, log string)

// New creates a new logger that allows using a custom callback function that
// will be invoked every time a log message is created.
func New(level log.Level, callback LogFunc) log.Logger {
	return &logger{
		level:    level,
		fields:   log.Fields{},
		callback: callback,
	}
}

type logger struct {
	level    log.Level
	fields   log.Fields
	callback LogFunc
}

// Level implements the log.Logger interface.
func (c *logger) Level() log.Level {
	return c.level
}

// WithField implements the log.Logger interface.
func (c *logger) WithField(key string, value any) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	f[key] = value
	return &logger{
		level:    c.level,
		fields:   f,
		callback: c.callback,
	}
}

// WithFields implements the log.Logger interface.
func (c *logger) WithFields(fields log.Fields) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	for k, v := range fields {
		f[k] = v
	}
	return &logger{
		level:    c.level,
		fields:   f,
		callback: c.callback,
	}
}

// WithError implements the log.Logger interface.
func (c *logger) WithError(err error) log.Logger {
	return c.WithField("err", err.Error())
}

// WithAdvice implements the log.Logger interface.
func (c *logger) WithAdvice(advice string) log.Logger {
	return c.WithField("advice", advice)
}

// Debug implements the log.Logger interface.
func (c *logger) Debug(args ...any) {
	if c.level >= log.Debug {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

// Info implements the log.Logger interface.
func (c *logger) Info(args ...any) {
	if c.level >= log.Info {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

// Warn implements the log.Logger interface.
func (c *logger) Warn(args ...any) {
	if c.level >= log.Warn {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

// Error implements the log.Logger interface.
func (c *logger) Error(args ...any) {
	if c.level >= log.Error {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

// Panic implements the log.Logger interface.
func (c *logger) Panic(args ...any) {
	msg := fmt.Sprint(args...)
	c.callback(c.level, c.fields, msg)
	panic(msg)
}
