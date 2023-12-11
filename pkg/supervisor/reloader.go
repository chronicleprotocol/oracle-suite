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

package supervisor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
)

const ReloaderLoggerTag = "RELOADER"

// Reloader is a service that can reload another wrapped service.
type Reloader struct {
	mu     sync.Mutex
	ctx    context.Context
	waitCh chan error
	log    log.Logger

	factoryCtx    context.Context
	factoryCancel context.CancelFunc
	serviceCtx    context.Context
	serviceCancel context.CancelFunc
	serviceErr    error
	service       Service
	factory       func(ctx context.Context, service chan Service) error
	factoryCh     chan Service
	serviceWaitCh <-chan error
}

// ReloaderConfig is a configuration for the Reloader service.
type ReloaderConfig struct {
	// Factory is a function that creates a new service instance. Every time
	// the new service is sent to the channel, the service is reloaded with
	// the new instance.
	//
	// The service is stopped when the factory channel is closed or when the
	// factory function returns an error.
	Factory func(ctx context.Context, service chan Service) error

	// Logger is a logger instance.
	Logger log.Logger
}

// NewReloader returns a new Reloader instance.
func NewReloader(cfg ReloaderConfig) *Reloader {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	return &Reloader{
		waitCh:        make(chan error),
		log:           cfg.Logger.WithField("tag", ReloaderLoggerTag),
		factory:       cfg.Factory,
		factoryCh:     make(chan Service),
		serviceWaitCh: make(chan error),
	}
}

// Start implements the Service interface.
func (r *Reloader) Start(ctx context.Context) (err error) {
	if r.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	r.log.Info("Starting")
	r.ctx = ctx
	go r.serviceFactoryRoutine()
	go r.serviceReloaderRoutine()
	return nil
}

// Wait implements the Service interface.
func (r *Reloader) Wait() <-chan error {
	return r.waitCh
}

// ServiceName implements the supervisor.WithName interface.
func (r *Reloader) ServiceName() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.service == nil {
		return "Reloader(uninitialized)"
	}
	return fmt.Sprintf("Reloader(%s)", ServiceName(r.service))
}

func (r *Reloader) reloadService(service Service) (err error) {
	if r.serviceCancel != nil {
		r.log.
			WithField("service", ServiceName(r.service)).
			Info("Reloading service")

		// Cancel the service context to stop the service and wait for it to stop.
		r.serviceCancel()
		if err := <-r.service.Wait(); err != nil {
			return fmt.Errorf("service reloader: failed to stop service: %w", err)
		}
	}

	// Update the service instance.
	r.mu.Lock()
	r.serviceCtx, r.serviceCancel = context.WithCancel(r.ctx)
	r.service = service
	r.serviceWaitCh = service.Wait()
	r.mu.Unlock()

	// Start the new service.
	if err := r.service.Start(r.serviceCtx); err != nil {
		return fmt.Errorf("service reloader: failed to start service: %w", err)
	}

	r.log.
		WithField("service", ServiceName(r.service)).
		Info("Service reloaded")

	return nil
}

func (r *Reloader) serviceFactoryRoutine() {
	r.mu.Lock()
	r.factoryCtx, r.factoryCancel = context.WithCancel(r.ctx)
	r.mu.Unlock()
	if err := r.factory(r.factoryCtx, r.factoryCh); err != nil {
		r.mu.Lock()
		r.serviceErr = errutil.Append(r.serviceErr, err)
		r.factoryCancel()
		if r.serviceCancel != nil {
			r.serviceCancel()
		}
		r.mu.Unlock()
	}
}

func (r *Reloader) serviceReloaderRoutine() {
	defer func() {
		r.mu.Lock()
		if r.factoryCancel != nil {
			r.factoryCancel()
		}
		if r.serviceCancel != nil {
			r.serviceCancel()
		}
		if r.serviceErr != nil {
			r.waitCh <- r.serviceErr
		}
		close(r.waitCh)
		r.mu.Unlock()
		r.log.Info("Stopped")
	}()
	for {
		select {
		// Note, that we do not want to stop this goroutine when r.ctx is done
		// because the wrapped service should also stop when r.ctx is done, and
		// we want to wait for it to stop. Otherwise, in case of an error, the
		// wrapped service may not be able to send an error to the wait channel
		// because no one is waiting for it.
		case err := <-r.serviceWaitCh:
			if err != nil {
				r.mu.Lock()
				r.serviceErr = errutil.Append(r.serviceErr, err)
				r.mu.Unlock()
			}
			return
		case service, ok := <-r.factoryCh:
			if !ok {
				r.mu.Lock()
				r.log.
					WithField("service", ServiceName(r.service)).
					Info("Stopping service due to closing factory channel")
				r.mu.Unlock()
				return
			}
			if err := r.reloadService(service); err != nil {
				r.mu.Lock()
				r.serviceErr = errutil.Append(r.serviceErr, err)
				r.mu.Unlock()
				return
			}
		}
	}
}
