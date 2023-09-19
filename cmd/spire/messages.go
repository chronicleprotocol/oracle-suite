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

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
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

// Message types supported by the `spire stream` command.
//
// Please note that although there are similarities, these types are distinct
// from the ones defined in the transport layer.
//
// The "type" field in each message delineates the specific type and version of
// that message. Meanwhile, the "meta.topic" field identifies the topic under
// which the message was received, providing context for the message's origin.
const (
	priceMessageType                 = "price/v1"
	muSigInitializeMessageType       = "musig_initialize/v1"
	muSigCommitmentMessageType       = "musig_commitment/v1"
	muSigPartialSignatureMessageType = "musig_partial_signature/v1"
	muSigSignatureMessageType        = "musig_signature/v1"
	muSigTerminateMessageType        = "musig_terminate/v1"
	greetMessageType                 = "greet/v1"
)

func handleMessage(msg transport.ReceivedMessage) map[string]any {
	var (
		typ  string
		data map[string]any
	)
	switch msgType := msg.Message.(type) {
	case *messages.Price:
		typ, data = handleLegacyPriceMessage(msgType)
	case *messages.DataPoint:
		switch msgType.Point.Value.(type) { //nolint:gocritic
		case value.Tick:
			typ, data = handleTickDataPointMessage(msgType)
		}
	case *messages.MuSigInitialize:
		typ, data = handleMuSigInitializeMessage(msgType)
	case *messages.MuSigCommitment:
		typ, data = handleMuSigCommitmentMessage(msgType)
	case *messages.MuSigPartialSignature:
		typ, data = handleMuSigPartialSignatureMessage(msgType)
	case *messages.MuSigSignature:
		typ, data = handleMuSigSignatureMessage(msgType)
	case *messages.MuSigTerminate:
		typ, data = handleMuSigTerminateMessage(msgType)
	case *messages.Greet:
		typ, data = handleGreetMessage(msgType)
	}
	if data == nil {
		return nil
	}
	meta := map[string]any{
		"transport":               msg.Meta.Transport,
		"user_agent":              msg.Meta.UserAgent,
		"topic":                   msg.Meta.Topic,
		"message_id":              msg.Meta.MessageID,
		"peer_id":                 msg.Meta.PeerID,
		"peer_addr":               msg.Meta.PeerAddr,
		"received_from_peer_id":   msg.Meta.ReceivedFromPeerID,
		"received_from_peer_addr": msg.Meta.ReceivedFromPeerAddr,
	}
	return map[string]any{
		"type": typ,
		"data": data,
		"meta": meta,
	}
}

func handleLegacyPriceMessage(msg *messages.Price) (string, map[string]any) {
	return priceMessageType, map[string]any{
		"wat":   msg.Price.Wat,
		"val":   msg.Price.Val.String(),
		"age":   msg.Price.Age.Unix(),
		"vrs":   msg.Price.Sig.String(),
		"trace": msg.Trace,
	}
}

func handleTickDataPointMessage(msg *messages.DataPoint) (string, map[string]any) {
	tick := msg.Point.Value.(value.Tick)
	return priceMessageType, map[string]any{
		"wat":   msg.Model,
		"val":   tick.Price.DecFixedPoint(contract.MedianPricePrecision).String(),
		"age":   msg.Point.Time.Unix(),
		"vrs":   msg.ECDSASignature.String(),
		"trace": msg.Point.Meta["trace"],
	}
}

func handleMuSigInitializeMessage(msg *messages.MuSigInitialize) (string, map[string]any) {
	return muSigInitializeMessageType, maputil.Merge(map[string]any{
		"session_id": msg.SessionID.String(),
		"started_at": msg.StartedAt.Unix(),
	}, handleMuSigMessage(msg.MuSigMessage))
}

func handleMuSigCommitmentMessage(msg *messages.MuSigCommitment) (string, map[string]any) {
	return muSigCommitmentMessageType, map[string]any{
		"session_id":       msg.SessionID.String(),
		"commitment_key_x": hexutil.BigIntToHex(msg.CommitmentKeyX),
		"commitment_key_y": hexutil.BigIntToHex(msg.CommitmentKeyY),
		"public_key_x":     hexutil.BigIntToHex(msg.PublicKeyX),
		"public_key_y":     hexutil.BigIntToHex(msg.PublicKeyY),
	}
}

func handleMuSigPartialSignatureMessage(msg *messages.MuSigPartialSignature) (string, map[string]any) {
	return muSigPartialSignatureMessageType, map[string]any{
		"session_id":        msg.SessionID.String(),
		"partial_signature": hexutil.BigIntToHex(msg.PartialSignature),
	}
}

func handleMuSigSignatureMessage(msg *messages.MuSigSignature) (string, map[string]any) {
	return muSigSignatureMessageType, maputil.Merge(map[string]any{
		"session_id":        msg.SessionID.String(),
		"computed_at":       msg.ComputedAt.Unix(),
		"commitment":        msg.Commitment.String(),
		"schnorr_signature": hexutil.BigIntToHex(msg.SchnorrSignature),
	}, handleMuSigMessage(msg.MuSigMessage))
}

func handleMuSigTerminateMessage(msg *messages.MuSigTerminate) (string, map[string]any) {
	return muSigTerminateMessageType, map[string]any{
		"session_id": msg.SessionID.String(),
		"reason":     msg.Reason,
	}
}

func handleGreetMessage(msg *messages.Greet) (string, map[string]any) {
	return greetMessageType, map[string]any{
		"ecdsa_signature": msg.Signature.String(),
		"public_key_x":    hexutil.BigIntToHex(msg.PublicKeyX),
		"public_key_y":    hexutil.BigIntToHex(msg.PublicKeyY),
	}
}

func handleMuSigMessage(msg *messages.MuSigMessage) map[string]any {
	meta := map[string]any{}
	switch { //nolint:gocritic
	case msg.MsgMeta.TickV1() != nil:
		msgTickMeta := msg.MsgMeta.TickV1()
		var tickMeta []map[string]any
		var optimisticTickMeta []map[string]any
		for _, tick := range msgTickMeta.FeedTicks {
			tickMeta = append(tickMeta, map[string]any{
				"val": tick.Val.SetPrec(contract.MedianPricePrecision).String(),
				"age": tick.Age.Unix(),
				"vrs": tick.VRS.String(),
			})
		}
		for _, optimistic := range msgTickMeta.Optimistic {
			optimisticTickMeta = append(optimisticTickMeta, map[string]any{
				"ecdsa_signature": optimistic.ECDSASignature.String(),
				"signers_blob":    hexutil.BytesToHex(optimistic.SignerIndexes),
			})
		}
		meta = map[string]any{
			"wat":        msgTickMeta.Wat,
			"val":        msgTickMeta.Val.SetPrec(contract.MedianPricePrecision).String(),
			"age":        msgTickMeta.Age.Unix(),
			"feed_ticks": tickMeta,
			"optimistic": optimisticTickMeta,
		}
	}
	var singers []string
	for _, signer := range msg.Signers {
		singers = append(singers, signer.String())
	}
	return map[string]any{
		"msg_type": msg.MsgType,
		"msg_body": msg.MsgBody.String(),
		"msg_meta": meta,
		"signers":  singers,
	}
}
