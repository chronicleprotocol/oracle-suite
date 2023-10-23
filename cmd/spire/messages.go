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

package main

import (
	"github.com/defiweb/go-eth/hexutil"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

// This file contains functions that normalize messages received through the
// transport layer, converting them into a standardized format. The
// normalization process is crucial for maintaining consistency in the message
// structure output by the `spire stream`, regardless of any internal changes
// in the oracle-suite over time. This approach allows for smoother
// integrations and updates.

const (
	priceMessageType = "price"
	greetMessageType = "greet"
)

type streamType struct {
	Type       string           `json:"type,omitempty"`
	Version    string           `json:"version,omitempty"`
	Data       any              `json:"data,omitempty"`
	Signature  string           `json:"signature,omitempty"`
	Signatures []map[string]any `json:"signatures,omitempty"`
	Meta       map[string]any   `json:"meta,omitempty"`
}

var removeEmptyFields = func(v any) bool {
	if v == nil {
		return false
	}
	if s, ok := v.(string); ok {
		return s != ""
	}
	return true
}

func handleMessage(msg transport.ReceivedMessage) streamType {
	var v streamType

	switch msgType := msg.Message.(type) {
	case *messages.Price:
		v = handleLegacyPriceMessage(msgType)
	case *messages.DataPoint:
		switch msgType.Point.Value.(type) { //nolint:gocritic
		case value.Tick:
			v = handleTickDataPointMessage(msgType)
		}
	case *messages.MuSigInitialize:
		v = handleMuSigInitializeMessage(msgType)
	case *messages.MuSigCommitment:
		v = handleMuSigCommitmentMessage(msgType)
	case *messages.MuSigPartialSignature:
		v = handleMuSigPartialSignatureMessage(msgType)
	case *messages.MuSigSignature:
		v = handleMuSigSignatureMessage(msgType)
	case *messages.MuSigTerminate:
		v = handleMuSigTerminateMessage(msgType)
	case *messages.Greet:
		v = handleGreetMessage(msgType)
	default:
		v = streamType{
			Data: msg.Message,
		}
	}

	v.Meta = maputil.Merge(v.Meta, maputil.Filter(map[string]any{
		"transport":               msg.Meta.Transport,
		"user_agent":              msg.Meta.UserAgent,
		"topic":                   msg.Meta.Topic,
		"message_id":              msg.Meta.MessageID,
		"peer_id":                 msg.Meta.PeerID,
		"peer_addr":               msg.Meta.PeerAddr,
		"received_from_peer_id":   msg.Meta.ReceivedFromPeerID,
		"received_from_peer_addr": msg.Meta.ReceivedFromPeerAddr,
	}, removeEmptyFields))

	return v
}

func handleLegacyPriceMessage(msg *messages.Price) streamType {

	return streamType{
		Type:    priceMessageType,
		Version: "1.0",
		Data: map[string]any{
			"wat": msg.Price.Wat,
			"val": msg.Price.Val.String(),
			"age": msg.Price.Age.Unix(),
		},
		Meta: map[string]any{
			"trace":      msg.Trace,
			"user_agent": "omnia/" + msg.Version,
		},
		Signature: msg.Price.Sig.String(),
		Signatures: []map[string]any{{
			"type":      "median/v1",
			"signature": msg.Price.Sig.String(),
		}},
	}
}

func handleTickDataPointMessage(msg *messages.DataPoint) streamType {
	tick := msg.Point.Value.(value.Tick)
	return streamType{
		Type:    priceMessageType,
		Version: "1.1",
		Data: map[string]any{
			"wat": msg.Model,
			"val": tick.Price.DecFixedPoint(chronicle.MedianPricePrecision).RawBigInt().String(),
			"age": msg.Point.Time.Unix(),
		},
		Meta: map[string]any{
			"trace": msg.Point.Meta["trace"],
		},
		Signature: msg.ECDSASignature.String(),
		Signatures: []map[string]any{{
			"type":      "median/v1",
			"signature": msg.ECDSASignature.String(),
		}},
	}
}

func handleMuSigSignatureMessage(msg *messages.MuSigSignature) streamType {
	msm := handleMuSigMessage(msg.MuSigMessage)

	msm.Type = priceMessageType // from the front-end perspective, this is a price message
	msm.Version = "2.0"         // this indicates that the message is signed with Schnorr
	msm.Meta = maputil.Merge(map[string]any{
		"session_id":  msg.SessionID.String(),
		"computed_at": msg.ComputedAt.Unix(),
	}, maputil.Filter(msm.Meta, removeEmptyFields))
	msm.Data = maputil.Merge(msm.Data.(map[string]any), map[string]any{
		"commitment": msg.Commitment.String(),
	})
	msm.Signature = hexutil.BigIntToHex(msg.SchnorrSignature)
	msm.Signatures = append(msm.Signatures, map[string]any{
		"type":      "scribe/v1",
		"signature": msm.Signature,
	})

	return msm
}

func handleMuSigInitializeMessage(msg *messages.MuSigInitialize) streamType {
	msm := handleMuSigMessage(msg.MuSigMessage)

	msm.Type = "musig_initialize"
	msm.Version = "0.1"
	msm.Meta = maputil.Merge(map[string]any{
		"session_id": msg.SessionID.String(),
		"started_at": msg.StartedAt.Unix(),
	}, maputil.Filter(msm.Meta, removeEmptyFields))

	return msm
}

func handleMuSigCommitmentMessage(msg *messages.MuSigCommitment) streamType {
	return streamType{
		Type:    "musig_commitment",
		Version: "0.1", // this means that the message is a WIP
		Data: map[string]any{
			"commitment_key_x": hexutil.BigIntToHex(msg.CommitmentKeyX),
			"commitment_key_y": hexutil.BigIntToHex(msg.CommitmentKeyY),
			"public_key_x":     hexutil.BigIntToHex(msg.PublicKeyX),
			"public_key_y":     hexutil.BigIntToHex(msg.PublicKeyY),
		},
		Meta: map[string]any{
			"session_id": msg.SessionID.String(),
		},
	}
}

func handleMuSigPartialSignatureMessage(msg *messages.MuSigPartialSignature) streamType {
	return streamType{
		Type:    "musig_partial_signature",
		Version: "0.1", // this means that the message is a WIP
		Meta: map[string]any{
			"session_id": msg.SessionID.String(),
		},
		Signature: hexutil.BigIntToHex(msg.PartialSignature),
	}
}

func handleMuSigTerminateMessage(msg *messages.MuSigTerminate) streamType {
	return streamType{
		Type:    "musig_terminate",
		Version: "0.1", // this means that the message is a WIP
		Meta: map[string]any{
			"session_id": msg.SessionID.String(),
			"reason":     msg.Reason,
		},
	}
}

func handleGreetMessage(msg *messages.Greet) streamType {
	return streamType{
		Type:    greetMessageType,
		Version: "0.1", // this means that the message is a WIP
		Data: map[string]any{
			"public_key_x": hexutil.BigIntToHex(msg.PublicKeyX),
			"public_key_y": hexutil.BigIntToHex(msg.PublicKeyY),
			"web_url":      msg.WebURL,
			"greet":        msg.Signature.String(),
		},
	}
}

func handleMuSigMessage(msg *messages.MuSigMessage) streamType {
	data := map[string]any{}
	meta := map[string]any{
		"type": msg.MsgType,
	}

	var ticks []map[string]any
	var signatures []map[string]any

	switch { //nolint:gocritic
	case msg.MsgMeta.TickV1() != nil:
		msgTickMeta := msg.MsgMeta.TickV1()

		for _, tick := range msgTickMeta.FeedTicks {
			ticks = append(ticks, map[string]any{
				"val": tick.Val.SetPrec(chronicle.ScribePricePrecision).RawBigInt().String(),
				"age": tick.Age.Unix(),
				"sig": tick.VRS.String(),
			})
		}

		for _, optimistic := range msgTickMeta.Optimistic {
			signatures = append(signatures, map[string]any{
				"type":         "scribe-optimistic/v1",
				"signature":    optimistic.ECDSASignature.String(),
				"signers_blob": hexutil.BytesToHex(optimistic.SignerIndexes),
			})
		}

		data = map[string]any{
			"body": msg.MsgBody.String(),
			"wat":  msgTickMeta.Wat,
			"val":  msgTickMeta.Val.SetPrec(chronicle.ScribePricePrecision).RawBigInt().String(),
			"age":  msgTickMeta.Age.Unix(),
		}
	}

	var signers []string
	for _, signer := range msg.Signers {
		signers = append(signers, signer.String())
	}
	if signers != nil {
		meta["trace_signers"] = signers
	}

	if ticks != nil {
		meta["trace"] = ticks
	}

	return streamType{
		Data:       data,
		Meta:       meta,
		Signatures: signatures,
	}
}
