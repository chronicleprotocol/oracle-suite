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
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
)

// New creates a new logger that can chain multiple loggers.
//
// If provided loggers implement the log.LoggerService interface, then
// they must not be started. To start them, use the Start method of the
// chain logger.
func New(loggers ...log.Logger) log.Logger {
	return &logger{
		shared:  &shared{waitCh: make(chan error)},
		loggers: loggers,
	}
}

type logger struct {
	*shared
	loggers []log.Logger
}

type shared struct {
	ctx    context.Context
	waitCh chan error
}

// Level implements the log.Logger interface.
func (l *logger) Level() log.Level {
	lvl := log.Panic
	for _, l := range l.loggers {
		if l.Level() > lvl {
			lvl = l.Level()
		}
	}
	return lvl
}

// WithField implements the log.Logger interface.
func (l *logger) WithField(key string, value interface{}) log.Logger {
	loggers := make([]log.Logger, len(l.loggers))
	for n, l := range l.loggers {
		loggers[n] = l.WithField(key, value)
	}
	return &logger{shared: l.shared, loggers: loggers}
}

// WithFields implements the log.Logger interface.
func (l *logger) WithFields(fields log.Fields) log.Logger {
	loggers := make([]log.Logger, len(l.loggers))
	for n, l := range l.loggers {
		loggers[n] = l.WithFields(fields)
	}
	return &logger{shared: l.shared, loggers: loggers}
}

// WithError implements the log.Logger interface.
func (l *logger) WithError(err error) log.Logger {
	loggers := make([]log.Logger, len(l.loggers))
	for n, l := range l.loggers {
		loggers[n] = l.WithError(err)
	}
	return &logger{shared: l.shared, loggers: loggers}
}

// Debugf implements the log.Logger interface.
func (l *logger) Debugf(format string, args ...interface{}) {
	for _, l := range l.loggers {
		l.Debugf(format, args...)
	}
}

// Infof implements the log.Logger interface.
func (l *logger) Infof(format string, args ...interface{}) {
	for _, l := range l.loggers {
		l.Infof(format, args...)
	}
}

// Warnf implements the log.Logger interface.
func (l *logger) Warnf(format string, args ...interface{}) {
	for _, l := range l.loggers {
		l.Warnf(format, args...)
	}
}

// Errorf implements the log.Logger interface.
func (l *logger) Errorf(format string, args ...interface{}) {
	for _, l := range l.loggers {
		l.Errorf(format, args...)
	}
}

// Panicf implements the log.Logger interface.
func (l *logger) Panicf(format string, args ...interface{}) {
	for _, l := range l.loggers {
		func() {
			defer func() { recover() }() //nolint:errcheck // same panic is thrown below
			l.Panicf(format, args...)
		}()
	}
	panic(fmt.Sprintf(format, args...))
}

// Debug implements the log.Logger interface.
func (l *logger) Debug(args ...interface{}) {
	for _, l := range l.loggers {
		l.Debug(args...)
	}
}

// Info implements the log.Logger interface.
func (l *logger) Info(args ...interface{}) {
	for _, l := range l.loggers {
		l.Info(args...)
	}
}

// Warn implements the log.Logger interface.
func (l *logger) Warn(args ...interface{}) {
	for _, l := range l.loggers {
		l.Warn(args...)
	}
}

// Error implements the log.Logger interface.
func (l *logger) Error(args ...interface{}) {
	for _, l := range l.loggers {
		l.Error(args...)
	}
}

// Panic implements the log.Logger interface.
func (l *logger) Panic(args ...interface{}) {
	for _, l := range l.loggers {
		func() {
			defer func() { recover() }() //nolint:errcheck // same panic is thrown below
			l.Panic(args...)
		}()
	}
	panic(fmt.Sprint(args...))
}

// Start implements the supervisor.Service interface.
func (l *logger) Start(ctx context.Context) error {
	if l.ctx != nil {
		return fmt.Errorf("service can be started only once")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	l.ctx = ctx
	for _, lg := range l.loggers {
		if srv, ok := lg.(log.LoggerService); ok {
			if err := srv.Start(ctx); err != nil {
				return err
			}
		}
	}
	go l.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (l *logger) Wait() chan error {
	chs := make([]chan error, 0, len(l.loggers)+1)
	chs = append(chs, l.waitCh)
	for _, lg := range l.loggers {
		if srv, ok := lg.(log.LoggerService); ok {
			chs = append(chs, srv.Wait())
		}
	}
	return chanutil.Merge(chs...)
}

func (l *logger) contextCancelHandler() {
	<-l.ctx.Done()
	close(l.waitCh)
}

var _ log.LoggerService = (*logger)(nil)
