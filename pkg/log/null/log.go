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

package null

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

type logger struct{}

// New creates a null logger. It does nothing. May be useful for testing.
func New() log.Logger {
	return &logger{}
}

func (n *logger) Level() log.Level                     { return log.Panic }
func (n *logger) WithField(_ string, _ any) log.Logger { return n }
func (n *logger) WithFields(_ log.Fields) log.Logger   { return n }
func (n *logger) WithError(_ error) log.Logger         { return n }
func (n *logger) WithAdvice(_ string) log.Logger       { return n }
func (n *logger) Panicf(format string, args ...any)    { panic(fmt.Sprintf(format, args...)) }
func (n *logger) Debug(_ ...any)                       {}
func (n *logger) Info(_ ...any)                        {}
func (n *logger) Warn(_ ...any)                        {}
func (n *logger) Error(_ ...any)                       {}
func (n *logger) Panic(args ...any)                    { panic(fmt.Sprint(args...)) }
