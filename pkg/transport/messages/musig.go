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

package messages

import (
	"fmt"
	"math/big"
	"time"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages/pb"

	"google.golang.org/protobuf/proto"
)

const (
	MuSigStartV1MessageName               = "musig_initialize/v1"
	MuSigTerminateV1MessageName           = "musig_terminate/v1"
	MuSigCommitmentV1MessageName          = "musig_commitment/v1"
	MuSigPartialSignatureV1MessageName    = "musig_partial_signature/v1"
	MuSigSignatureV1MessageName           = "musig_signature/v1"
	MuSigOptimisticSignatureV1MessageName = "musig_optimistic_signature/v1"
)

type MuSigInitialize struct {
	// SessionID is the unique ID of the MuSig session.
	SessionID types.Hash `json:"session_id"`

	// CreatedAt is the time when the session was started.
	StartedAt time.Time `json:"started_at"`

	// Type of the message that will be signed.
	MsgType string `json:"msg_type"`

	// Message body that will be signed.
	MsgBody types.Hash `json:"msg_body"`

	// Meta is a map of metadata that may be necessary to verify the message.
	MsgMeta map[string][]byte `json:"msg_meta"`

	// Signers is a list of signers that will participate in the MuSig session.
	Signers []types.Address `json:"signers"`
}

func (m *MuSigInitialize) MarshallBinary() ([]byte, error) {
	msg := pb.MuSigInitializeMessage{
		SessionID:          m.SessionID.Bytes(),
		StartedAtTimestamp: m.StartedAt.Unix(),
		MsgType:            m.MsgType,
		MsgBody:            m.MsgBody.Bytes(),
		MsgMeta:            m.MsgMeta,
		Signers:            make([][]byte, len(m.Signers)),
	}
	for i, signer := range m.Signers {
		msg.Signers[i] = signer.Bytes()
	}
	return proto.Marshal(&msg)
}

func (m *MuSigInitialize) UnmarshallBinary(bytes []byte) (err error) {
	msg := pb.MuSigInitializeMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	if len(msg.MsgBody) > types.HashLength {
		return fmt.Errorf("invalid message body length")
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.StartedAt = time.Unix(msg.StartedAtTimestamp, 0)
	m.MsgType = msg.MsgType
	m.MsgBody = types.MustHashFromBytes(msg.MsgBody, types.PadLeft)
	m.MsgMeta = msg.MsgMeta
	m.Signers = make([]types.Address, len(msg.Signers))
	for i, signer := range msg.Signers {
		m.Signers[i], err = types.AddressFromBytes(signer)
		if err != nil {
			return err
		}
	}
	return nil
}

type MuSigTerminate struct {
	// Unique SessionID of the MuSig session.
	SessionID types.Hash `json:"session_id"`

	// Reason for terminating the MuSig session.
	Reason string `json:"reason"`
}

func (m *MuSigTerminate) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigTerminateMessage{
		SessionID: m.SessionID.Bytes(),
		Reason:    m.Reason,
	})
}

func (m *MuSigTerminate) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigTerminateMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.Reason = msg.Reason
	return nil
}

type MuSigCommitment struct {
	// Unique SessionID of the MuSig session.
	SessionID types.Hash `json:"session_id"`

	CommitmentKeyX *big.Int `json:"commitment_key_x"`
	CommitmentKeyY *big.Int `json:"commitment_key_y"`

	PublicKeyX *big.Int `json:"public_key_x"`
	PublicKeyY *big.Int `json:"public_key_y"`
}

func (m *MuSigCommitment) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigCommitmentMessage{
		SessionID:      m.SessionID.Bytes(),
		PubKeyX:        m.PublicKeyX.Bytes(),
		PubKeyY:        m.PublicKeyY.Bytes(),
		CommitmentKeyX: m.CommitmentKeyX.Bytes(),
		CommitmentKeyY: m.CommitmentKeyY.Bytes(),
	})
}

func (m *MuSigCommitment) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigCommitmentMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.PublicKeyX = new(big.Int).SetBytes(msg.PubKeyX)
	m.PublicKeyY = new(big.Int).SetBytes(msg.PubKeyY)
	m.CommitmentKeyX = new(big.Int).SetBytes(msg.CommitmentKeyX)
	m.CommitmentKeyY = new(big.Int).SetBytes(msg.CommitmentKeyY)
	return nil
}

type MuSigPartialSignature struct {
	// Unique SessionID of the MuSig session.
	SessionID types.Hash `json:"session_id"`

	// Partial signature of the MuSig session.
	PartialSignature *big.Int `json:"partial_signature"`
}

func (m *MuSigPartialSignature) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigPartialSignatureMessage{
		SessionID:        m.SessionID.Bytes(),
		PartialSignature: m.PartialSignature.Bytes(),
	})
}

func (m *MuSigPartialSignature) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigPartialSignatureMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.PartialSignature = new(big.Int).SetBytes(msg.PartialSignature)
	return nil
}

type MuSigSignature struct {
	// Unique SessionID of the MuSig session.
	SessionID types.Hash `json:"sessionID"`

	// ComputedAt is the time at which the signature was computed.
	ComputedAt time.Time `json:"computedAt"`

	// Type of the data that was signed.
	MsgType string `json:"msgType"`

	// Data that was signed.
	MsgBody types.Hash `json:"msgBody"`

	// Meta is a map of metadata associated with the message.
	MsgMeta map[string][]byte

	// Commitment of the MuSig session.
	Commitment types.Address `json:"commitment"`

	// Signers is a list of addresses of the signers that will participate in the MuSig session.
	Signers []types.Address `json:"signers"`

	// SchnorrSignature is a MuSig Schnorr signature calculated from the partial
	// signatures of all participants.
	SchnorrSignature *big.Int `json:"schnorrSignature"`
}

func (m *MuSigSignature) toProtobuf() *pb.MuSigSignatureMessage {
	msg := &pb.MuSigSignatureMessage{
		SessionID:           m.SessionID[:],
		ComputedAtTimestamp: m.ComputedAt.Unix(),
		MsgType:             m.MsgType,
		MsgBody:             m.MsgBody.Bytes(),
		Commitment:          m.Commitment.Bytes(),
		Signers:             make([][]byte, len(m.Signers)),
		SchnorrSignature:    m.SchnorrSignature.Bytes(),
	}
	m.Signers = make([]types.Address, len(msg.Signers))
	for i, signer := range m.Signers {
		msg.Signers[i] = signer.Bytes()
	}
	return msg
}

func (m *MuSigSignature) fromProtobuf(msg *pb.MuSigSignatureMessage) error {
	if len(msg.MsgBody) > types.HashLength {
		return fmt.Errorf("invalid message body length")
	}
	com, err := types.AddressFromBytes(msg.Commitment)
	if err != nil {
		return err
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.ComputedAt = time.Unix(msg.ComputedAtTimestamp, 0)
	m.MsgType = msg.MsgType
	m.MsgBody = types.MustHashFromBytes(msg.MsgBody, types.PadLeft)
	m.MsgMeta = msg.MsgMeta
	m.Commitment = com
	m.Signers = make([]types.Address, len(msg.Signers))
	for i, signer := range msg.Signers {
		m.Signers[i], err = types.AddressFromBytes(signer)
		if err != nil {
			return err
		}
	}
	m.SchnorrSignature = new(big.Int).SetBytes(msg.SchnorrSignature)
	return nil
}

func (m *MuSigSignature) MarshallBinary() ([]byte, error) {
	return proto.Marshal(m.toProtobuf())
}

func (m *MuSigSignature) UnmarshallBinary(bytes []byte) error {
	msg := &pb.MuSigSignatureMessage{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return err
	}
	return m.fromProtobuf(msg)
}

type MuSigOptimisticSignature struct {
	MuSigSignature

	// ECDSASignature is a ECDSA signature calculated by the MuSig session
	// coordinator.
	ECDSASignature types.Signature `json:"ecdsa_signature"`
}

func (m *MuSigOptimisticSignature) MarshallBinary() ([]byte, error) {
	msg := &pb.MuSigOptimisticSignatureMessage{
		EcdsaSignature: m.ECDSASignature.Bytes(),
	}
	msg.Signature = m.MuSigSignature.toProtobuf()
	return proto.Marshal(msg)
}

func (m *MuSigOptimisticSignature) UnmarshallBinary(bytes []byte) error {
	var err error
	msg := pb.MuSigOptimisticSignatureMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	if err := m.MuSigSignature.fromProtobuf(msg.Signature); err != nil {
		return err
	}
	m.ECDSASignature, err = types.SignatureFromBytes(msg.EcdsaSignature)
	if err != nil {
		return err
	}
	return nil
}
