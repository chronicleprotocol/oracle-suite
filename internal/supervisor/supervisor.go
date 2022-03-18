package supervisor

import (
	"context"
	"reflect"
)

type Service interface {
	Start(ctx context.Context) error
	Wait() chan error
}

type Supervisor struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	waitCh    chan error
	services  []Service
}

func New(ctx context.Context) *Supervisor {
	ctx, ctxCancel := context.WithCancel(ctx)
	return &Supervisor{ctx: ctx, ctxCancel: ctxCancel, waitCh: make(chan error)}
}

func (s *Supervisor) Watch(services ...Service) {
	s.services = append(s.services, services...)
}

func (s *Supervisor) Start() error {
	for _, srv := range s.services {
		if err := srv.Start(s.ctx); err != nil {
			s.ctxCancel()
			return err
		}
	}
	go s.serviceWatcher()
	return nil
}

func (s *Supervisor) Wait() chan error {
	return s.waitCh
}

func (s *Supervisor) serviceWatcher() {
	var err error
	for len(s.services) > 0 {
		cases := make([]reflect.SelectCase, len(s.services))
		for i, srv := range s.services {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(srv.Wait())}
		}
		n, v, _ := reflect.Select(cases)
		if !v.IsNil() {
			if err != nil {
				err = v.Interface().(error)
			}
			s.ctxCancel()
		}
		s.services = append(s.services[:n], s.services[n+1:]...)
	}
	s.waitCh <- err
}
