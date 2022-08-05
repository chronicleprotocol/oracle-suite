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

package grafana

import (
	"context"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

var _ log.LoggerService = (*logger)(nil)

func (c *logger) Start(ctx context.Context) error {
	c.logger.Info("Starting")
	if c.ctx != nil {
		return fmt.Errorf("service can be started only once")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	c.ctx = ctx
	c.waitCh = make(chan error)
	go c.pushRoutine()
	return nil
}

func (c *logger) Wait() chan error {
	return c.waitCh
}
