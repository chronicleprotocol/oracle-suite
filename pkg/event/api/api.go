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

package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/internal/httpserver"
	"github.com/chronicleprotocol/oracle-suite/pkg/event/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "EVENT_API"

const defaultTimeout = 3 * time.Second

// EventAPI provides an HTTP API for EventStore.
type EventAPI struct {
	ctx    context.Context
	waitCh chan error

	srv *httpserver.HTTPServer
	es  *store.EventStore
	log log.Logger
}

type Config struct {
	EventStore *store.EventStore
	Address    string
	Logger     log.Logger
}

type jsonEvent struct {
	Timestamp  int64                    `json:"timestamp"`
	Data       map[string]string        `json:"data"`
	Signatures map[string]jsonSignature `json:"signatures"`
}

type jsonSignature struct {
	Signer    string `json:"signer"`
	Signature string `json:"signature"`
}

// New returns a new instance of the EventAPI struct.
func New(ctx context.Context, cfg Config) (*EventAPI, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	api := &EventAPI{
		ctx:    ctx,
		waitCh: make(chan error),
		es:     cfg.EventStore,
		log:    cfg.Logger.WithField("tag", LoggerTag),
	}
	api.srv = httpserver.New(ctx, &http.Server{
		Addr:         cfg.Address,
		Handler:      http.HandlerFunc(api.handler),
		IdleTimeout:  defaultTimeout,
		ReadTimeout:  defaultTimeout,
		WriteTimeout: defaultTimeout,
	})
	return api, nil
}

// Start starts HTTP server.
func (e *EventAPI) Start() error {
	e.log.Infof("Starting")
	err := e.srv.Start()
	if err != nil {
		return fmt.Errorf("unable to start the HTTP server: %w", err)
	}
	go e.contextCancelHandler()
	return nil
}

// Wait waits until the context is canceled or until an error occurs.
func (e *EventAPI) Wait() chan error {
	return e.waitCh
}

func (e *EventAPI) handler(res http.ResponseWriter, req *http.Request) {
	typ, ok := req.URL.Query()["type"]
	if !ok || len(typ) != 1 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	idxHex, ok := req.URL.Query()["index"]
	if !ok || len(idxHex) != 1 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	idx, err := decodeHex(idxHex[0])
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx, ctxCancel := context.WithTimeout(e.ctx, defaultTimeout)
	defer ctxCancel()
	events, err := e.es.Events(ctx, typ[0], idx)
	if err != nil {
		e.log.WithError(err).Error("Event store error")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(res).Encode(mapEvents(events))
}

func mapEvents(es []*messages.Event) (r []*jsonEvent) {
	for _, e := range es {
		j := &jsonEvent{
			Timestamp:  e.EventDate.Unix(),
			Data:       map[string]string{},
			Signatures: map[string]jsonSignature{},
		}
		for k, v := range e.Data {
			j.Data[k] = hex.EncodeToString(v)
		}
		for k, v := range e.Signatures {
			j.Signatures[k] = jsonSignature{
				Signer:    hex.EncodeToString(v.Signer),
				Signature: hex.EncodeToString(v.Signature),
			}
		}
		r = append(r, j)
	}
	return r
}

func (e *EventAPI) contextCancelHandler() {
	var err error
	defer func() { e.waitCh <- err }()
	defer e.log.Info("Stopped")
	<-e.ctx.Done()
	err = <-e.srv.Wait()
}

func decodeHex(h string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(h, "0x"))
}
