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

package null

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

type Logger struct{}

func New() *Logger {
	return &Logger{}
}

func (n *Logger) Level() log.Level                             { return log.Panic }
func (n *Logger) WithField(_ string, _ interface{}) log.Logger { return n }
func (n *Logger) WithFields(_ log.Fields) log.Logger           { return n }
func (n *Logger) WithError(_ error) log.Logger                 { return n }
func (n *Logger) Debugf(_ string, _ ...interface{})            {}
func (n *Logger) Infof(_ string, _ ...interface{})             {}
func (n *Logger) Warnf(_ string, _ ...interface{})             {}
func (n *Logger) Errorf(_ string, _ ...interface{})            {}
func (n *Logger) Panicf(format string, args ...interface{})    { panic(fmt.Sprintf(format, args...)) }
func (n *Logger) Debug(_ ...interface{})                       {}
func (n *Logger) Info(_ ...interface{})                        {}
func (n *Logger) Warn(_ ...interface{})                        {}
func (n *Logger) Error(_ ...interface{})                       {}
func (n *Logger) Panic(args ...interface{})                    { panic(fmt.Sprint(args...)) }
