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
