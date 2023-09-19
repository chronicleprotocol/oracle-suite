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
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestHandleMessage(t *testing.T) {
	t.Run("handleLegacyPriceMessage", func(t *testing.T) {
		msg := &messages.Price{
			Price: &median.Price{
				Wat: "ETH/USD",
				Val: big.NewInt(3000),
				Age: time.Unix(1234567890, 0),
				Sig: types.MustSignatureFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01"),
			},
			Trace: json.RawMessage(`{"DAI/USD@exchange": "1.00001"}`),
		}
		receivedMessage := transport.ReceivedMessage{
			Message: msg,
			Meta:    createMeta(priceMessageType),
		}
		result, err := handleMessage(receivedMessage)
		assert.Nil(t, err)

		expectedJSON := `{
			"type": "price/v1",
			"data": {
				"wat": "ETH/USD",
				"val": "3000",
				"age": 1234567890,
				"vrs": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01",
				"trace": {
					"DAI/USD@exchange": "1.00001"
				}
			},
			"meta": {
				"transport": "libp2p",
				"user_agent": "spire/v0.0.0",
				"topic": "price/v1",
				"message_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"peer_id": "peer1",
				"peer_addr": "0x1234567890abcdef1234567890abcdef1234567890abcdef",
				"received_from_peer_id": "peer2",
				"received_from_peer_addr": "0x234567890abcdef1234567890abcdef123456789"
			}
		}`
		resultJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})

	t.Run("handleTickDataPointMessage", func(t *testing.T) {
		tick := value.Tick{Price: bn.DecFloatPoint(3000)}
		msg := &messages.DataPoint{
			Model: "ETH/USD",
			Point: datapoint.Point{
				Value: tick,
				Time:  time.Unix(1234567890, 0),
				Meta: map[string]any{
					"trace": map[string]any{
						"DAI/USD@exchange": "1.00001",
					},
				},
			},
			ECDSASignature: types.MustSignatureFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01"),
		}
		receivedMessage := transport.ReceivedMessage{
			Message: msg,
			Meta:    createMeta(priceMessageType),
		}
		result, err := handleMessage(receivedMessage)
		assert.Nil(t, err)

		expectedJSON := `{
			"type": "price/v1",
			"data": {
				"wat": "ETH/USD",
				"val": "3000",
				"age": 1234567890,
				"vrs": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01",
				"trace": {
					"DAI/USD@exchange": "1.00001"
				}
			},
			"meta": {
				"transport": "libp2p",
				"user_agent": "spire/v0.0.0",
				"topic": "price/v1",
				"message_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"peer_id": "peer1",
				"peer_addr": "0x1234567890abcdef1234567890abcdef1234567890abcdef",
				"received_from_peer_id": "peer2",
				"received_from_peer_addr": "0x234567890abcdef1234567890abcdef123456789"
			}
		}`
		resultJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})

	t.Run("handleMuSigInitializeMessage", func(t *testing.T) {
		msg := &messages.MuSigInitialize{
			SessionID:    types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
			StartedAt:    time.Unix(1234567890, 0),
			MuSigMessage: createMuSigMessage(),
		}
		receivedMessage := transport.ReceivedMessage{
			Message: msg,
			Meta:    createMeta(muSigInitializeMessageType),
		}
		result, err := handleMessage(receivedMessage)
		assert.Nil(t, err)

		expectedJSON := `{
			"type": "musig_initialize/v1",
			"data": {
				"session_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"started_at": 1234567890,
				"msg_type": "musig_initialize/v1",
				"msg_body": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"msg_meta": {
					"wat": "ETH/USD",
					"val": "3000",
					"age": 1234567890,
					"feed_ticks": [
						{
							"val": "3000",
							"age": 1234567890,
							"vrs": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01"
						}
					],
					"optimistic": [
						{
							"ecdsa_signature": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01",
							"signers_blob": "0x123456"
						}
					]
				},
				"signers": [
					"0x1234567890abcdef1234567890abcdef12345678"
				]
			},
			"meta": {
				"transport": "libp2p",
				"user_agent": "spire/v0.0.0",
				"topic": "musig_initialize/v1",
				"message_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"peer_id": "peer1",
				"peer_addr": "0x1234567890abcdef1234567890abcdef1234567890abcdef",
				"received_from_peer_id": "peer2",
				"received_from_peer_addr": "0x234567890abcdef1234567890abcdef123456789"
			}
		}`
		resultJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})

	t.Run("handleMuSigCommitmentMessage", func(t *testing.T) {
		msg := &messages.MuSigCommitment{
			SessionID:      types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
			CommitmentKeyX: big.NewInt(1234567890),
			CommitmentKeyY: big.NewInt(1234567891),
			PublicKeyX:     big.NewInt(1234567892),
			PublicKeyY:     big.NewInt(1234567893),
		}
		receivedMessage := transport.ReceivedMessage{
			Message: msg,
			Meta:    createMeta(muSigCommitmentMessageType),
		}
		result, err := handleMessage(receivedMessage)
		assert.Nil(t, err)

		expectedJSON := `{
		"type": "musig_commitment/v1",
		"data": {
			"session_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"commitment_key_x": "0x499602d2",
			"commitment_key_y": "0x499602d3",
			"public_key_x": "0x499602d4",
			"public_key_y": "0x499602d5"
		},
		"meta": {
			"transport": "libp2p",
			"user_agent": "spire/v0.0.0",
			"topic": "musig_commitment/v1",
			"message_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"peer_id": "peer1",
			"peer_addr": "0x1234567890abcdef1234567890abcdef1234567890abcdef",
			"received_from_peer_id": "peer2",
			"received_from_peer_addr": "0x234567890abcdef1234567890abcdef123456789"
		}
	}`
		resultJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})

	t.Run("handleMuSigPartialSignatureMessage", func(t *testing.T) {
		msg := &messages.MuSigPartialSignature{
			SessionID:        types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
			PartialSignature: big.NewInt(1234567890),
		}
		receivedMessage := transport.ReceivedMessage{
			Message: msg,
			Meta:    createMeta(muSigPartialSignatureMessageType),
		}
		result, err := handleMessage(receivedMessage)
		assert.Nil(t, err)

		expectedJSON := `{
		"type": "musig_partial_signature/v1",
		"data": {
			"session_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"partial_signature": "0x499602d2"
		},
		"meta": {
			"transport": "libp2p",
			"user_agent": "spire/v0.0.0",
			"topic": "musig_partial_signature/v1",
			"message_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"peer_id": "peer1",
			"peer_addr": "0x1234567890abcdef1234567890abcdef1234567890abcdef",
			"received_from_peer_id": "peer2",
			"received_from_peer_addr": "0x234567890abcdef1234567890abcdef123456789"
		}
	}`
		resultJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})

	t.Run("handleMuSigSignatureMessage", func(t *testing.T) {
		msg := &messages.MuSigSignature{
			SessionID:        types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
			ComputedAt:       time.Unix(1234567890, 0),
			Commitment:       types.MustAddressFromHex("0x1234567890abcdef1234567890abcdef12345678"),
			SchnorrSignature: big.NewInt(1234567890),
			MuSigMessage:     createMuSigMessage(),
		}
		receivedMessage := transport.ReceivedMessage{
			Message: msg,
			Meta:    createMeta(muSigSignatureMessageType),
		}
		result, err := handleMessage(receivedMessage)
		assert.Nil(t, err)

		expectedJSON := `{
			"type": "musig_signature/v1",
			"data": {
				"session_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"computed_at": 1234567890,
				"commitment": "0x1234567890abcdef1234567890abcdef12345678",
				"schnorr_signature": "0x499602d2",
				"msg_type": "musig_initialize/v1",
				"msg_body": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"msg_meta": {
					"wat": "ETH/USD",
					"val": "3000",
					"age": 1234567890,
					"feed_ticks": [
						{
							"val": "3000",
							"age": 1234567890,
							"vrs": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01"
						}
					],
					"optimistic": [
						{
							"ecdsa_signature": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01",
							"signers_blob": "0x123456"
						}
					]
				},
				"signers": [
					"0x1234567890abcdef1234567890abcdef12345678"
				]
			},
			"meta": {
				"transport": "libp2p",
				"user_agent": "spire/v0.0.0",
				"topic": "musig_signature/v1",
				"message_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"peer_id": "peer1",
				"peer_addr": "0x1234567890abcdef1234567890abcdef1234567890abcdef",
				"received_from_peer_id": "peer2",
				"received_from_peer_addr": "0x234567890abcdef1234567890abcdef123456789"
			}
		}`
		resultJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})

	t.Run("handleMuSigTerminateMessage", func(t *testing.T) {
		msg := &messages.MuSigTerminate{
			SessionID: types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
			Reason:    "termination reason",
		}
		receivedMessage := transport.ReceivedMessage{
			Message: msg,
			Meta:    createMeta(muSigTerminateMessageType),
		}
		result, err := handleMessage(receivedMessage)
		assert.Nil(t, err)

		expectedJSON := `{
		"type": "musig_terminate/v1",
		"data": {
			"session_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"reason": "termination reason"
		},
		"meta": {
			"transport": "libp2p",
			"user_agent": "spire/v0.0.0",
			"topic": "musig_terminate/v1",
			"message_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"peer_id": "peer1",
			"peer_addr": "0x1234567890abcdef1234567890abcdef1234567890abcdef",
			"received_from_peer_id": "peer2",
			"received_from_peer_addr": "0x234567890abcdef1234567890abcdef123456789"
		}
	}`
		resultJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})

	t.Run("handleGreetMessage", func(t *testing.T) {
		msg := &messages.Greet{
			Signature:  types.MustSignatureFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01"),
			PublicKeyX: big.NewInt(1234567892),
			PublicKeyY: big.NewInt(1234567893),
		}
		receivedMessage := transport.ReceivedMessage{
			Message: msg,
			Meta:    createMeta(greetMessageType),
		}
		result, err := handleMessage(receivedMessage)
		assert.Nil(t, err)

		expectedJSON := `{
			"type": "greet/v1",
			"data": {
				"ecdsa_signature": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01",
				"public_key_x": "0x499602d4",
				"public_key_y": "0x499602d5"
			},
			"meta": {
				"transport": "libp2p",
				"user_agent": "spire/v0.0.0",
				"topic": "greet/v1",
				"message_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"peer_id": "peer1",
				"peer_addr": "0x1234567890abcdef1234567890abcdef1234567890abcdef",
				"received_from_peer_id": "peer2",
				"received_from_peer_addr": "0x234567890abcdef1234567890abcdef123456789"
			}
		}`
		resultJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		assert.JSONEq(t, expectedJSON, string(resultJSON))
	})
}

func createMuSigMessage() *messages.MuSigMessage {
	return &messages.MuSigMessage{
		MsgType: muSigInitializeMessageType,
		MsgBody: types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
		MsgMeta: messages.MuSigMeta{
			Meta: messages.MuSigMetaTickV1{
				Wat: "ETH/USD",
				Val: bn.DecFixedPoint(3000, 18),
				Age: time.Unix(1234567890, 0),
				FeedTicks: []messages.MuSigMetaFeedTick{
					{
						Val: bn.DecFixedPoint(3000, 18),
						Age: time.Unix(1234567890, 0),
						VRS: types.MustSignatureFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01"),
					},
				},
				Optimistic: []messages.MuSigMetaOptimistic{
					{
						ECDSASignature: types.MustSignatureFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef01"),
						SignerIndexes:  []byte{0x12, 0x34, 0x56},
					},
				},
			},
		},
		Signers: []types.Address{
			types.MustAddressFromHex("0x1234567890abcdef1234567890abcdef12345678"),
		},
	}
}

func createMeta(topic string) transport.Meta {
	return transport.Meta{
		Transport:            "libp2p",
		Topic:                topic,
		MessageID:            "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		PeerID:               "peer1",
		PeerAddr:             "0x1234567890abcdef1234567890abcdef1234567890abcdef",
		UserAgent:            "spire/v0.0.0",
		ReceivedFromPeerID:   "peer2",
		ReceivedFromPeerAddr: "0x234567890abcdef1234567890abcdef123456789",
	}
}
