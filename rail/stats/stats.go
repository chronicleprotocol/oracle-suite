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

package stats

import (
	"errors"
	"reflect"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
	"github.com/defiweb/go-eth/types"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/chronicleprotocol/oracle-suite/rail/node"
)

type PeerEvent any

type Peer struct {
	ID   peer.ID
	Addr types.Address

	Protocols []protocol.ID
	Addresses []core.Multiaddr

	Connectedness  string
	Network        string
	Identification string
	Ping           string
	Topic          string
	Name           string
	Counter        int
	UserAgent      string
}

type Stats struct {
	Peers          map[peer.ID]Peer
	PeerOrdering   []peer.ID
	PeerModelCount map[peer.ID]int
	Nets           map[string]map[peer.ID]int
	Models         ModelsByFeeds
	ModelOrdering  []string
}

func (p *Stats) Add(np any) {
	if p.Peers == nil {
		p.Peers = make(map[peer.ID]Peer)
	}
	if p.Models == nil {
		p.Models = make(ModelsByFeeds)
	}
	if p.Nets == nil {
		p.Nets = make(map[string]map[peer.ID]int)
	}
	if p.PeerModelCount == nil {
		p.PeerModelCount = make(map[peer.ID]int)
	}
	switch t := np.(type) {
	case event.EvtPeerIdentificationCompleted:
		pid := t.Peer
		if old, ok := p.Peers[pid]; ok {
			old.Identification = "completed"

			p.Peers[pid] = old
			return
		}
		a := ethkey.PeerIDToAddress(pid)
		p.Peers[pid] = Peer{
			ID:      pid,
			Addr:    a,
			Network: node.Net(a),
			Name:    node.BootName(pid),

			Identification: "completed",
		}
		p.PeerOrdering = sliceutil.AppendUnique(p.PeerOrdering, pid)
	case node.PeerConnectedness:
		pid := t.Peer
		if old, ok := p.Peers[pid]; ok {
			old.Connectedness = t.Connectedness.String()

			p.Peers[pid] = old
			return
		}
		a := ethkey.PeerIDToAddress(pid)
		p.Peers[pid] = Peer{
			ID:      pid,
			Addr:    a,
			Network: node.Net(a),
			Name:    node.BootName(pid),

			Connectedness: t.Connectedness.String(),
		}
		p.PeerOrdering = sliceutil.AppendUnique(p.PeerOrdering, pid)
	case node.LocalReachability:
		pid := t.Peer
		if old, ok := p.Peers[pid]; ok {
			old.Connectedness = t.Reachability.String()

			p.Peers[pid] = old
			return
		}
		a := ethkey.PeerIDToAddress(pid)
		p.Peers[pid] = Peer{
			ID:      pid,
			Addr:    a,
			Network: "local",
			Name:    node.BootName(pid),

			Connectedness: t.Reachability.String(),
		}
		p.PeerOrdering = sliceutil.AppendUnique(p.PeerOrdering, pid)
	case node.PingResult:
		rtt := t.RTT.String()
		if t.Error != nil {
			rtt = t.Error.Error()
		}
		pid := t.Peer
		if old, ok := p.Peers[pid]; ok {
			old.Ping = rtt
			p.Peers[pid] = old
			return
		}
		a := ethkey.PeerIDToAddress(pid)
		p.Peers[pid] = Peer{
			ID:      pid,
			Addr:    a,
			Network: node.Net(a),
			Name:    node.BootName(pid),

			Ping: rtt,
		}
		p.PeerOrdering = sliceutil.AppendUnique(p.PeerOrdering, pid)
	case node.PeerMessage:
		_, err := p.Models.add(t.Topic, t.Msg)
		if err == nil {
			p.PeerModelCount[t.Msg.GetFrom()]++
			net := node.Net(ethkey.PeerIDToAddress(t.Msg.GetFrom()))
			if _, ok := p.Nets[net]; !ok {
				p.Nets[net] = make(map[peer.ID]int)
			}
			p.Nets[net][t.Msg.GetFrom()]++
		}

		func(pid peer.ID) {
			if old, ok := p.Peers[pid]; ok {
				old.Counter++
				old.Topic = *t.Msg.Topic

				p.Peers[pid] = old
				return
			}
			a := ethkey.PeerIDToAddress(pid)
			p.Peers[pid] = Peer{
				ID:      pid,
				Addr:    a,
				Network: node.Net(a),
				Name:    node.BootName(pid),

				Topic: *t.Msg.Topic,
			}
			p.PeerOrdering = sliceutil.AppendUnique(p.PeerOrdering, pid)
		}(t.Msg.GetFrom())

		func(pid peer.ID) {
			if old, ok := p.Peers[pid]; ok {
				old.Counter++

				p.Peers[pid] = old
				return
			}
			a := ethkey.PeerIDToAddress(pid)
			p.Peers[pid] = Peer{
				ID:      pid,
				Addr:    a,
				Network: node.Net(a),
				Name:    node.BootName(pid),
			}
			p.PeerOrdering = sliceutil.AppendUnique(p.PeerOrdering, pid)
		}(t.Msg.ReceivedFrom)
	}
}

type ModelsByFeeds map[string]map[peer.ID]int

func (m *ModelsByFeeds) add(topic string, message *pubsub.Message) (string, error) {
	typ := messages.AllMessagesMap[topic]
	typRefl := reflect.TypeOf(typ).Elem()

	msg := reflect.New(typRefl).Interface().(transport.Message)
	if msg.UnmarshallBinary(message.Data) != nil {
		return "", msg.UnmarshallBinary(message.Data)
	}

	var model string
	switch m := msg.(type) {
	case *messages.Price:
		model = m.Price.Wat
	case *messages.DataPoint:
		model = m.Model
	case *messages.MuSigSignature:
		model = m.MsgMeta.TickV1().Wat
	default:
		return "", errors.New("unknown message type")
	}

	if _, ok := (*m)[model]; !ok {
		(*m)[model] = make(map[peer.ID]int)
	}
	(*m)[model][message.GetFrom()]++

	return model, nil
}
