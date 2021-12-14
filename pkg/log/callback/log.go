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

package callback

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

type LogFunc func(level log.Level, fields log.Fields, log string)

// Logger implements the log.Logger interface. It allows using a custom
// callback function that will be invoked every time a log message is created.
type Logger struct {
	level    log.Level
	fields   log.Fields
	callback LogFunc
}

func New(level log.Level, callback LogFunc) *Logger {
	return &Logger{
		level:    level,
		fields:   log.Fields{},
		callback: callback,
	}
}

func (c *Logger) Level() log.Level {
	return c.level
}

func (c *Logger) WithField(key string, value interface{}) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	f[key] = value
	return &Logger{
		level:    c.level,
		fields:   f,
		callback: c.callback,
	}
}

func (c *Logger) WithFields(fields log.Fields) log.Logger {
	f := log.Fields{}
	for k, v := range c.fields {
		f[k] = v
	}
	for k, v := range fields {
		f[k] = v
	}
	return &Logger{
		level:    c.level,
		fields:   f,
		callback: c.callback,
	}
}

func (c *Logger) WithError(err error) log.Logger {
	return c.WithField("err", err.Error())
}

func (c *Logger) Debugf(format string, args ...interface{}) {
	if c.level >= log.Debug {
		c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
	}
}

func (c *Logger) Infof(format string, args ...interface{}) {
	if c.level >= log.Info {
		c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
	}
}

func (c *Logger) Warnf(format string, args ...interface{}) {
	if c.level >= log.Warn {
		c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
	}
}

func (c *Logger) Errorf(format string, args ...interface{}) {
	if c.level >= log.Error {
		c.callback(c.level, c.fields, fmt.Sprintf(format, args...))
	}
}

func (c *Logger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	c.callback(c.level, c.fields, msg)
	panic(msg)
}

func (c *Logger) Debug(args ...interface{}) {
	if c.level >= log.Debug {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

func (c *Logger) Info(args ...interface{}) {
	if c.level >= log.Info {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

func (c *Logger) Warn(args ...interface{}) {
	if c.level >= log.Warn {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

func (c *Logger) Error(args ...interface{}) {
	if c.level >= log.Error {
		c.callback(c.level, c.fields, fmt.Sprint(args...))
	}
}

func (c *Logger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	c.callback(c.level, c.fields, msg)
	panic(msg)
}
