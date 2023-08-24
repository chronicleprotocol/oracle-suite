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

package chain

import (
	"context"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
)

// New creates a new logger that can chain multiple loggers.
//
// If the provided loggers implement the log.LoggerService interface, they must
// not be started. To start them, use the Start method of the chain logger.
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
	waitCh <-chan error
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
	// Start all chained loggers that implement the log.LoggerService interface.
	for _, lg := range l.loggers {
		if srv, ok := lg.(log.LoggerService); ok {
			if err := srv.Start(ctx); err != nil {
				return err
			}
		}
	}
	// Merge all wait channels from chained loggers to one wait channel.
	fi := chanutil.NewFanIn[error]()
	for _, t := range l.loggers {
		if s, ok := t.(log.LoggerService); ok {
			_ = fi.Add(s.Wait())
		}
	}
	// Add additional wait channel that is closed when the context is done.
	// It is important to add this channel because there is no guarantee that
	// any of chained loggers implement the log.LoggerService interface. In
	// that case, there would be no wait channel to wait for and fan-in channel
	// will be closed immediately.
	ch := make(chan error)
	_ = fi.Add(ch)
	l.waitCh = fi.Chan()
	fi.AutoClose()
	go l.contextCancelHandler(ch)

	return nil
}

// Wait implements the supervisor.Service interface.
func (l *logger) Wait() <-chan error {
	return l.waitCh
}

func (l *logger) contextCancelHandler(ch chan error) {
	<-l.ctx.Done()
	close(ch)
}

var _ log.LoggerService = (*logger)(nil)
