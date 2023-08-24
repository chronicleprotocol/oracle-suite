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

package relay

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const MuSigLoggerTag = "MUSIG_STORE"

type MuSigStore struct {
	ctx    context.Context
	mu     sync.Mutex
	waitCh chan error
	log    log.Logger

	transport          transport.Transport
	scribeDataModels   []string
	opScribeDataModels []string
	signatures         map[storeKey]*messages.MuSigSignature
	opSignatures       map[storeKey]*messages.MuSigOptimisticSignature
}

// MuSigStoreConfig is the configuration for MuSigStore.
type MuSigStoreConfig struct {
	// Transport is an implementation of transport used to fetch data from
	// feeds.
	Transport transport.Service

	// ScribeDataModels is the list of models for which we should collect
	// signatures.
	ScribeDataModels []string

	// OpScribeDataModels is the list of models for which we should collect
	// optimistic signatures.
	OpScribeDataModels []string

	// Logger is a current logger interface used by the store.
	Logger log.Logger
}

func NewMuSigStore(cfg MuSigStoreConfig) *MuSigStore {
	return &MuSigStore{
		waitCh:             make(chan error),
		log:                cfg.Logger.WithField("tag", MuSigLoggerTag),
		transport:          cfg.Transport,
		scribeDataModels:   cfg.ScribeDataModels,
		opScribeDataModels: cfg.OpScribeDataModels,
		signatures:         make(map[storeKey]*messages.MuSigSignature),
	}
}

// Start implements the supervisor.Service interface.
func (m *MuSigStore) Start(ctx context.Context) error {
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
func (m *MuSigStore) Wait() <-chan error {
	return m.waitCh
}

func (m *MuSigStore) SignaturesByDataModel(model string) []*messages.MuSigSignature {
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

func (m *MuSigStore) OptimisticSignaturesByDataModel(model string) []*messages.MuSigOptimisticSignature {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Collect signatures for the given data model.
	var signatures []*messages.MuSigOptimisticSignature
	for k, v := range m.opSignatures {
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

func (m *MuSigStore) collectSignature(feed types.Address, sig *messages.MuSigSignature) {
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

func (m *MuSigStore) collectOpSignature(feed types.Address, sig *messages.MuSigOptimisticSignature) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := storeKey{wat: m.signatureDataModel(&sig.MuSigSignature), feed: feed}

	// If we already have a signature for the given feed and data model, we
	// should not override it with an older one.
	if _, ok := m.opSignatures[key]; ok && sig.ComputedAt.Before(m.opSignatures[key].ComputedAt) {
		return
	}

	m.opSignatures[key] = sig
}

func (m *MuSigStore) shouldCollectSignature(sig *messages.MuSigSignature) bool {
	model := m.signatureDataModel(sig)
	if model == "" {
		return false
	}
	for _, a := range m.scribeDataModels {
		if a == model {
			return true
		}
	}
	return false
}

func (m *MuSigStore) shouldCollectOpSignature(sig *messages.MuSigOptimisticSignature) bool {
	model := m.signatureDataModel(&sig.MuSigSignature)
	if model == "" {
		return false
	}
	for _, a := range m.opScribeDataModels {
		if a == model {
			return true
		}
	}
	return false
}

func (m *MuSigStore) handleSignatureMessage(msg transport.ReceivedMessage) {
	if msg.Error != nil {
		m.log.WithError(msg.Error).Error("Unable to receive message")
		return
	}
	sig, ok := msg.Message.(*messages.MuSigSignature)
	if !ok {
		m.log.Error("Unexpected value returned from the transport layer")
		return
	}
	if !m.shouldCollectSignature(sig) {
		return
	}
	m.collectSignature(msgAuthorToAddr(msg.Author), sig)
}

func (m *MuSigStore) handleOptimisticSignatureMessage(msg transport.ReceivedMessage) {
	if msg.Error != nil {
		m.log.WithError(msg.Error).Error("Unable to receive message")
		return
	}
	sig, ok := msg.Message.(*messages.MuSigOptimisticSignature)
	if !ok {
		m.log.Error("Unexpected value returned from the transport layer")
		return
	}
	if !m.shouldCollectOpSignature(sig) {
		return
	}
	m.collectOpSignature(msgAuthorToAddr(msg.Author), sig)
}

func (m *MuSigStore) signatureDataModel(sig *messages.MuSigSignature) string {
	model, ok := sig.MsgMeta["wat"]
	if !ok {
		return ""
	}
	return string(model)
}

func (m *MuSigStore) collectorRoutine() {
	sigCh := m.transport.Messages(messages.MuSigSignatureV1MessageName)
	opSigCh := m.transport.Messages(messages.MuSigOptimisticSignatureV1MessageName)
	for {
		select {
		case <-m.ctx.Done():
			return
		case msg := <-sigCh:
			m.handleSignatureMessage(msg)
		case msg := <-opSigCh:
			m.handleOptimisticSignatureMessage(msg)
		}
	}
}

// contextCancelHandler handles context cancellation.
func (m *MuSigStore) contextCancelHandler() {
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
