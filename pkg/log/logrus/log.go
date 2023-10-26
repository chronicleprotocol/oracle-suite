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

package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

// New creates a new logger that uses Logrus for logging.
func New(logrusLogger logrus.FieldLogger) log.Logger {
	lvl := log.Debug
	if l, ok := logrusLogger.(*logrus.Logger); ok {
		switch l.Level {
		case logrus.PanicLevel:
			lvl = log.Panic
		case logrus.FatalLevel:
			lvl = log.Panic
		case logrus.ErrorLevel:
			lvl = log.Error
		case logrus.WarnLevel:
			lvl = log.Warn
		case logrus.InfoLevel:
			lvl = log.Info
		case logrus.DebugLevel:
			lvl = log.Debug
		case logrus.TraceLevel:
			lvl = log.Debug
		}
	}
	return &logger{log: logrusLogger, lvl: lvl}
}

type logger struct {
	log logrus.FieldLogger
	lvl log.Level
}

// Level implements new log.Logger interface.
func (l *logger) Level() log.Level {
	return l.lvl
}

// WithField implements new log.Logger interface.
func (l *logger) WithField(key string, value any) log.Logger {
	return &logger{log: l.log.WithField(key, value), lvl: l.lvl}
}

// WithFields implements new log.Logger interface.
func (l *logger) WithFields(fields log.Fields) log.Logger {
	return &logger{log: l.log.WithFields(fields), lvl: l.lvl}
}

// WithError implements new log.Logger interface.
func (l *logger) WithError(err error) log.Logger {
	if fErr, ok := err.(log.ErrorWithFields); ok {
		return &logger{log: l.log.WithFields(fErr.LogFields()).WithError(err), lvl: l.lvl}
	}
	return &logger{log: l.log.WithError(err), lvl: l.lvl}
}

// WithAdvice implements the log.Logger interface.
func (l *logger) WithAdvice(advice string) log.Logger {
	return l.WithField("advice", advice)
}

// Debug implements new log.Logger interface.
func (l *logger) Debug(args ...any) {
	l.log.Debug(args...)
}

// Info implements new log.Logger interface.
func (l *logger) Info(args ...any) {
	l.log.Info(args...)
}

// Warn implements new log.Logger interface.
func (l *logger) Warn(args ...any) {
	l.log.Warn(args...)
}

// Error implements new log.Logger interface.
func (l *logger) Error(args ...any) {
	l.log.Error(args...)
}

// Panic implements new log.Logger interface.
func (l *logger) Panic(args ...any) {
	l.log.Panic(args...)
}
