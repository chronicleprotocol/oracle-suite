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

package com

import (
	"context"
	"sync"
)

type Service interface {
	Start(context.Context) error
	Wait()
}

func RunServicesAndWait(ctx context.Context, services ...Service) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, s := range services {
		log.Debugf("start %T", s)
		if err := s.Start(ctx); err != nil {
			log.Fatal(err)
		}
		log.Debugf("started %T", s)
	}

	var wg sync.WaitGroup
	wg.Add(len(services))
	for _, s := range services {
		go func(s Service) {
			defer wg.Done()
			log.Debugf("wait %T", s)
			s.Wait()
			cancel()
			log.Debugf("done %T", s)
		}(s)
	}

	wg.Wait()
	log.Debug("all services finished")
}
