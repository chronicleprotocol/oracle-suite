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

package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const MuSigLoggerTag = "MUSIG_STORE"

type SignatureProvider interface {
	// SignaturesByDataModel returns a list of signatures for the given data
	// model.
	SignaturesByDataModel(model string) []*messages.MuSigSignature
}

// Store stores MuSigSignature messages received from the transport layer.
//
// It stores only the latest signature provided by each feed for each data
// model.
type Store struct {
	ctx    context.Context
	mu     sync.Mutex
	waitCh chan error
	log    log.Logger

	transport  transport.Transport
	dataModels []string
	signatures map[storeKey]*messages.MuSigSignature
}

// Config is the configuration for Store.
type Config struct {
	// Transport is an implementation of transport used to fetch data from
	// feeds.
	Transport transport.Service

	// DataModels is the list of models for which we should collect
	// signatures.
	DataModels []string

	// Logger is a current logger interface used by the store.
	Logger log.Logger
}

// New creates a new Store instance.
func New(cfg Config) *Store {
	return &Store{
		waitCh:     make(chan error),
		log:        cfg.Logger.WithField("tag", MuSigLoggerTag),
		transport:  cfg.Transport,
		dataModels: cfg.DataModels,
		signatures: make(map[storeKey]*messages.MuSigSignature),
	}
}

// Start implements the supervisor.Service interface.
func (m *Store) Start(ctx context.Context) error {
	if m.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	m.log.Info("Starting")
	m.ctx = ctx
	go m.collectorRoutine()
	go m.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (m *Store) Wait() <-chan error {
	return m.waitCh
}

// SignaturesByDataModel implements SignatureProvider interface.
func (m *Store) SignaturesByDataModel(model string) []*messages.MuSigSignature {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Collect signatures for the given data model.
	var signatures []*messages.MuSigSignature
	for k, v := range m.signatures {
		if k.wat == model {
			signatures = append(signatures, v)
		}
	}

	// Sort signatures by newest first.
	sort.Slice(signatures, func(i, j int) bool {
		return signatures[i].ComputedAt.After(signatures[j].ComputedAt)
	})

	return signatures
}

func (m *Store) collectSignature(feed types.Address, sig *messages.MuSigSignature) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := storeKey{wat: m.signatureDataModel(sig), feed: feed}

	// If we already have a signature for the given feed and data model, we
	// should not override it with an older one.
	if _, ok := m.signatures[key]; ok && sig.ComputedAt.Before(m.signatures[key].ComputedAt) {
		return
	}

	m.signatures[key] = sig
}

func (m *Store) shouldCollectSignature(sig *messages.MuSigSignature) bool {
	model := m.signatureDataModel(sig)
	if model == "" {
		return false
	}
	for _, a := range m.dataModels {
		if a == model {
			return true
		}
	}
	return false
}

func (m *Store) handleSignatureMessage(msg transport.ReceivedMessage) {
	if msg.Error != nil {
		m.log.
			WithError(msg.Error).
			WithAdvice("Ignore if occurs occasionally, especially if it is related to temporary network issues").
			Error("Unable to receive a message from the transport layer")
		return
	}
	sig, ok := msg.Message.(*messages.MuSigSignature)
	if !ok {
		m.log.
			WithField("type", fmt.Sprintf("%T", msg.Message)).
			WithAdvice("This is a bug and must be investigated").
			Error("Unexpected value returned from the transport layer")
		return
	}
	if !m.shouldCollectSignature(sig) {
		return
	}
	m.collectSignature(msgAuthorToAddr(msg.Author), sig)
}

func (m *Store) signatureDataModel(sig *messages.MuSigSignature) string {
	msgMeta := sig.MsgMeta.TickV1()
	if msgMeta == nil {
		return ""
	}
	return msgMeta.Wat
}

func (m *Store) collectorRoutine() {
	sigCh := m.transport.Messages(messages.MuSigSignatureV1MessageName)
	for {
		select {
		case <-m.ctx.Done():
			return
		case msg := <-sigCh:
			m.handleSignatureMessage(msg)
		}
	}
}

// contextCancelHandler handles context cancellation.
func (m *Store) contextCancelHandler() {
	defer func() { close(m.waitCh) }()
	defer m.log.Info("Stopped")
	<-m.ctx.Done()
}

type storeKey struct {
	wat  string
	feed types.Address
}

func msgAuthorToAddr(author []byte) types.Address {
	addr, _ := types.AddressFromBytes(author)
	return addr
}
