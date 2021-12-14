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

package p2p

import (
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

// tdLogger is thread-safe wrapper for logger field.
type tsLogger struct {
	mu  sync.RWMutex
	log log.Logger
}

func (l *tsLogger) set(logger log.Logger) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.log = logger
}

func (l *tsLogger) get() log.Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.log
}

// Logger configures node to use given logger instance.
func Logger(logger log.Logger) Options {
	return func(n *Node) error {
		n.tsLog.set(logger)
		return nil
	}
}
