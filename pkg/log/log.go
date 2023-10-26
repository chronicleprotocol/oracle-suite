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

package log

import (
	"context"
	"fmt"
	"strings"
)

type Level uint8

const (
	Panic Level = iota
	Error
	Warn
	Info
	Debug
)

// ParseLevel takes a string level and returns the Logrus log level constant.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return Panic, nil
	case "error", "err":
		return Error, nil
	case "warning", "warn":
		return Warn, nil
	case "info":
		return Info, nil
	case "debug":
		return Debug, nil
	}
	return Level(0), fmt.Errorf("not a valid log level: %q", lvl)
}

func (l Level) String() string {
	switch l {
	case Panic:
		return "panic"
	case Error:
		return "error"
	case Warn:
		return "warning"
	case Info:
		return "info"
	case Debug:
		return "debug"
	}
	return "unknown"
}

// IsLevel reports whether current logger shows logs for the given log level.
func IsLevel(logger Logger, level Level) bool {
	return logger.Level() >= level
}

type Fields = map[string]any

type ErrorWithFields interface {
	error
	LogFields() Fields
}

// Logger defines an interface for structured logging with varying levels of
// severity.
type Logger interface {
	// Level returns the current logging level.
	Level() Level

	// WithField returns a new logger instance that includes the specified
	// key-value pair.
	WithField(key string, value any) Logger

	// WithFields returns a new logger instance that includes the specified
	// key-value pairs.
	WithFields(fields Fields) Logger

	// WithError associates an error with the logger, returning a new logger
	// instance. The associated error will be included in subsequent log
	// messages.
	WithError(err error) Logger

	// WithAdvice associates a recommended action or advice with the logger,
	// detailing what steps should be taken in response to the event being
	// logged. This can be helpful to guide system administrators or developers
	// on the appropriate course of action when reading the logs.
	//
	// Examples:
	// - "Ignore if happens occasionally"
	// - "This is a known issue; a fix is in progress"
	// - "Report immediately to the development team"
	WithAdvice(advice string) Logger

	// Debug logs detailed system-level diagnostic messages useful during
	// development and troubleshooting. It should contain information that's
	// typically too verbose for regular operation.
	Debug(args ...any)

	// Info logs informational messages that highlight the progress of the
	// application's normal operation, such as startup and significant runtime
	// events. These messages should be concise but informative for system
	// administrators and should not occur at a high rate.
	Info(args ...any)

	// Warn logs potentially harmful situations, unexpected events, or minor
	// issues. This might include things like deprecations or approaching
	// resource limits. These aren't immediate errors but can lead to them if
	// unaddressed.
	Warn(args ...any)

	// Error logs failures that prevent an operation from completing
	// successfully. While the application might continue running, these issues
	// typically require intervention to resolve, either as system
	// administration or code changes.
	Error(args ...any)

	// Panic logs severe errors that might cause the application to crash or be
	// in an unstable state. Logging at this level should be rare and often
	// followed by program termination. These messages should provide enough
	// context to diagnose and rectify catastrophic failures.
	Panic(args ...any)
}

// LoggerService is a logger that needs to be started to be used.
type LoggerService interface {
	Logger
	Start(ctx context.Context) error
	Wait() <-chan error
}
