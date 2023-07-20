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
	MuSigStartV1MessageName            = "musig_initialize/v1"
	MuSigTerminateV1MessageName        = "musig_terminate/v1"
	MuSigCommitmentV1MessageName       = "musig_commitment/v1"
	MuSigPartialSignatureV1MessageName = "musig_partial_signature/v1"
	MuSigSignatureV1MessageName        = "musig_signature/v1"
)

type MuSigInitialize struct {
	// SessionID is the unique ID of the MuSig session.
	SessionID [32]byte

	// CreatedAt is the time when the session was started.
	StartedAt time.Time

	// Type of the message that will be signed.
	MsgType string

	// Message body that will be signed.
	MsgBody types.Hash

	// Meta is a map of metadata that may be necessary to verify the message.
	MsgMeta map[string][]byte

	// Signers is a list of signers that will participate in the MuSig session.
	Signers []types.Address
}

func (m *MuSigInitialize) MarshallBinary() ([]byte, error) {
	msg := pb.MuSigInitializeMessage{
		SessionID:          m.SessionID[:],
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
	copy(m.SessionID[:], msg.SessionID)
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
	SessionID [32]byte

	// Reason for terminating the MuSig session.
	Reason string
}

func (m *MuSigTerminate) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigTerminateMessage{
		SessionID: m.SessionID[:],
		Reason:    m.Reason,
	})
}

func (m *MuSigTerminate) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigTerminateMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	copy(m.SessionID[:], msg.SessionID)
	m.Reason = msg.Reason
	return nil
}

type MuSigCommitment struct {
	// Unique SessionID of the MuSig session.
	SessionID [32]byte

	CommitmentKeyX *big.Int
	CommitmentKeyY *big.Int

	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

func (m *MuSigCommitment) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigCommitmentMessage{
		SessionID:      m.SessionID[:],
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
	copy(m.SessionID[:], msg.SessionID)
	m.PublicKeyX = new(big.Int).SetBytes(msg.PubKeyX)
	m.PublicKeyY = new(big.Int).SetBytes(msg.PubKeyY)
	m.CommitmentKeyX = new(big.Int).SetBytes(msg.CommitmentKeyX)
	m.CommitmentKeyY = new(big.Int).SetBytes(msg.CommitmentKeyY)
	return nil
}

type MuSigPartialSignature struct {
	// Unique SessionID of the MuSig session.
	SessionID [32]byte

	// Partial signature of the MuSig session.
	PartialSignature *big.Int
}

func (m *MuSigPartialSignature) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigPartialSignatureMessage{
		SessionID:        m.SessionID[:],
		PartialSignature: m.PartialSignature.Bytes(),
	})
}

func (m *MuSigPartialSignature) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigPartialSignatureMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	copy(m.SessionID[:], msg.SessionID)
	m.PartialSignature = new(big.Int).SetBytes(msg.PartialSignature)
	return nil
}

type MuSigSignature struct {
	// Unique SessionID of the MuSig session.
	SessionID [32]byte

	// Type of the data that was signed.
	MsgType string

	// Data that was signed.
	MsgBody types.Hash

	// Commitment of the MuSig session.
	Commitment types.Address

	// SchnorrSignature is a MuSig Schnorr signature calculated from the partial
	// signatures of all participants.
	SchnorrSignature *big.Int

	// ECDSASignature is a ECDSA signature calculated by the MuSig session
	// coordinator.
	ECDSASignature types.Signature
}

func (m *MuSigSignature) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigSignatureMessage{
		SessionID:        m.SessionID[:],
		MsgType:          m.MsgType,
		MsgBody:          m.MsgBody.Bytes(),
		Commitment:       m.Commitment.Bytes(),
		SchnorrSignature: m.SchnorrSignature.Bytes(),
		EcdsaSignature:   m.ECDSASignature.Bytes(),
	})
}

func (m *MuSigSignature) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigSignatureMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	if len(msg.MsgBody) > types.HashLength {
		return fmt.Errorf("invalid message body length")
	}
	com, err := types.AddressFromBytes(msg.Commitment)
	if err != nil {
		return err
	}
	sig, err := types.SignatureFromBytes(msg.EcdsaSignature)
	if err != nil {
		return err
	}
	copy(m.SessionID[:], msg.SessionID)
	m.MsgType = msg.MsgType
	m.MsgBody = types.MustHashFromBytes(msg.MsgBody, types.PadLeft)
	m.Commitment = com
	m.SchnorrSignature = new(big.Int).SetBytes(msg.SchnorrSignature)
	m.ECDSASignature = sig
	return nil
}
