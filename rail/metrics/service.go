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

package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	ctx context.Context
	wg  sync.WaitGroup

	server *http.Server
}

func (s *Prometheus) Start(ctx context.Context) error {
	if s.ctx != nil {
		return fmt.Errorf("already started %T", s)
	}
	if ctx == nil {
		return fmt.Errorf("nil context for %T", s)
	}
	{
		sm := http.NewServeMux()
		sm.Handle("/metrics", promhttp.Handler())
		s.server = &http.Server{Addr: ":8080", Handler: sm}
	}
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()

		if err := s.server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Debug(err)
				return
			}
			log.Error(err)
		}
	}()
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()

		<-ctx.Done()
		if err := s.server.Shutdown(ctx); err != nil {
			log.Error(err)
		}
	}()
	return nil
}

func (s *Prometheus) Wait() {
	s.wg.Wait()
}
