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
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"

	"google.golang.org/protobuf/proto"
)

const (
	MuSigTickV1DataType = "tick/v1"
)

const (
	MuSigStartV1MessageName            = "musig_initialize/v1.1"
	MuSigTerminateV1MessageName        = "musig_terminate/v1.1"
	MuSigCommitmentV1MessageName       = "musig_commitment/v1.1"
	MuSigPartialSignatureV1MessageName = "musig_partial_signature/v1.1"
	MuSigSignatureV1MessageName        = "musig_signature/v1.1"
)

type muSigMeta interface {
	muSigMeta()
}

type MuSigMeta struct {
	// Meta must be one of the following types:
	// * MuSigMetaTickV1
	Meta muSigMeta
}

func (m *MuSigMeta) TickV1() *MuSigMetaTickV1 {
	if tick, ok := m.Meta.(MuSigMetaTickV1); ok {
		return &tick
	}
	return nil
}

type MuSigMetaTickV1 struct {
	Wat       string                  // Asset name.
	Val       *bn.DecFixedPointNumber // Median price.
	Age       time.Time               // Oldest tick timestamp.
	ECDSAData *types.Signature        // Optional ECDSA signature data required for the optimistic poke.
	FeedTicks []MuSigMetaFeedTick     // All ticks used to calculate the median price.
}

func (m MuSigMetaTickV1) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"wat":        m.Wat,
		"val":        m.Val.String(),
		"age":        m.Age.In(time.UTC).Format(time.RFC3339Nano),
		"ecdsa_data": m.ECDSAData,
		"ticks":      m.FeedTicks,
	})
}

func (m MuSigMetaTickV1) muSigMeta() {}

type MuSigMetaFeedTick struct {
	Val *bn.DecFixedPointNumber // Price.
	Age time.Time               // Price timestamp.
	VRS types.Signature         // Signature.
}

func (m MuSigMetaFeedTick) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"val": m.Val.String(),
		"age": m.Age.In(time.UTC).Format(time.RFC3339Nano),
		"vrs": m.VRS.String(),
	})
}

func (m *MuSigMeta) toProtobuf() (*pb.MuSigMeta, error) {
	var err error
	meta := &pb.MuSigMeta{}
	switch t := m.Meta.(type) { //nolint:gocritic
	case MuSigMetaTickV1:
		tickV1 := &pb.MuSigMetaTickV1{
			Wat: t.Wat,
			Age: t.Age.Unix(),
		}
		if t.Val != nil {
			tickV1.Val, err = t.Val.MarshalBinary()
			if err != nil {
				return nil, err
			}
		}
		for _, tick := range t.FeedTicks {
			var valBin []byte
			if t.Val != nil {
				valBin, err = tick.Val.MarshalBinary()
				if err != nil {
					return nil, err
				}
			}
			tickV1.Ticks = append(tickV1.Ticks, &pb.MuSigMetaTickV1_FeedTick{
				Val: valBin,
				Age: tick.Age.Unix(),
				Vrs: tick.VRS.Bytes(),
			})
		}
		if t.ECDSAData != nil {
			tickV1.EcdsaData = t.ECDSAData.Bytes()
		}
		meta.MsgMeta = &pb.MuSigMeta_Ticks{
			Ticks: tickV1,
		}
	}
	return meta, nil
}

func (m *MuSigMeta) fromProtobuf(msg *pb.MuSigMeta) error {
	switch { //nolint:gocritic
	case msg.GetTicks() != nil:
		msg := msg.GetTicks()
		val := &bn.DecFixedPointNumber{}
		if len(msg.Val) > 0 {
			if err := val.UnmarshalBinary(msg.Val); err != nil {
				return err
			}
		}
		tick := MuSigMetaTickV1{
			Wat:       msg.Wat,
			Val:       val,
			Age:       time.Unix(msg.Age, 0),
			FeedTicks: nil,
		}
		for _, feedTick := range msg.Ticks {
			val := &bn.DecFixedPointNumber{}
			if err := val.UnmarshalBinary(feedTick.Val); err != nil {
				return err
			}
			vrs, err := types.SignatureFromBytes(feedTick.Vrs)
			if err != nil {
				return err
			}
			tick.FeedTicks = append(tick.FeedTicks, MuSigMetaFeedTick{
				Val: val,
				Age: time.Unix(feedTick.Age, 0),
				VRS: vrs,
			})
		}
		if msg.EcdsaData != nil {
			ecdsaData, err := types.SignatureFromBytes(msg.EcdsaData)
			if err != nil {
				return err
			}
			tick.ECDSAData = &ecdsaData
		}
		m.Meta = tick
	}
	return nil
}

type MuSigMessage struct {
	// Type of the message.
	MsgType string

	// Message body that will be signed.
	MsgBody types.Hash

	// Meta is a message-specific metadata.
	MsgMeta MuSigMeta

	// Signers is a list of signers that will participate in the MuSig session.
	Signers []types.Address
}

type MuSigInitialize struct {
	transport.AppInfo

	*MuSigMessage

	// SessionID is the unique ID of the MuSig session.
	SessionID types.Hash

	// CreatedAt is the time when the session was started.
	StartedAt time.Time
}

func (m MuSigInitialize) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"session_id": m.SessionID.String(),
		"started_at": m.StartedAt.In(time.UTC).Format(time.RFC3339Nano),
		"msg_type":   m.MsgType,
		"msg_body":   m.MsgBody.String(),
		"msg_meta":   m.MsgMeta.Meta,
		"signers":    m.Signers,
	})
}

// MarshallBinary implements the transport.Message interface.
func (m MuSigInitialize) MarshallBinary() ([]byte, error) {
	if m.MuSigMessage == nil {
		return nil, fmt.Errorf("empty message")
	}
	meta, err := m.MsgMeta.toProtobuf()
	if err != nil {
		return nil, err
	}
	msg := pb.MuSigInitializeMessage{
		SessionID:          m.SessionID.Bytes(),
		StartedAtTimestamp: m.StartedAt.Unix(),
		MsgType:            m.MsgType,
		MsgBody:            m.MsgBody.Bytes(),
		MsgMeta:            meta,
		Signers:            make([][]byte, len(m.Signers)),
		AppInfo:            appInfoToProtobuf(m.AppInfo),
	}
	for i, signer := range m.Signers {
		msg.Signers[i] = signer.Bytes()
	}
	return proto.Marshal(&msg)
}

// UnmarshallBinary implements the transport.Message interface.
func (m *MuSigInitialize) UnmarshallBinary(bytes []byte) (err error) {
	if len(bytes) == 0 {
		return fmt.Errorf("empty data")
	}
	msg := pb.MuSigInitializeMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	m.MuSigMessage = &MuSigMessage{}
	if len(msg.MsgBody) != types.HashLength {
		return fmt.Errorf("invalid message body length")
	}
	if len(msg.SessionID) != types.HashLength {
		return fmt.Errorf("invalid session ID length")
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.StartedAt = time.Unix(msg.StartedAtTimestamp, 0)
	m.MsgType = msg.MsgType
	m.MsgBody = types.MustHashFromBytes(msg.MsgBody, types.PadLeft)
	if err := m.MsgMeta.fromProtobuf(msg.MsgMeta); err != nil {
		return err
	}
	m.Signers = make([]types.Address, len(msg.Signers))
	for i, signer := range msg.Signers {
		m.Signers[i], err = types.AddressFromBytes(signer)
		if err != nil {
			return err
		}
	}
	m.AppInfo = appInfoFromProtobuf(msg.AppInfo)
	return nil
}

type MuSigTerminate struct {
	transport.AppInfo

	// Unique SessionID of the MuSig session.
	SessionID types.Hash

	// Reason for terminating the MuSig session.
	Reason string
}

func (m MuSigTerminate) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"session_id": m.SessionID.String(),
		"reason":     m.Reason,
	})
}

// MarshallBinary implements the transport.Message interface.
func (m MuSigTerminate) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigTerminateMessage{
		SessionID: m.SessionID.Bytes(),
		Reason:    m.Reason,
		AppInfo:   appInfoToProtobuf(m.AppInfo),
	})
}

// UnmarshallBinary implements the transport.Message interface.
func (m *MuSigTerminate) UnmarshallBinary(bytes []byte) error {
	if len(bytes) == 0 {
		return fmt.Errorf("empty data")
	}
	msg := pb.MuSigTerminateMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.Reason = msg.Reason
	m.AppInfo = appInfoFromProtobuf(msg.AppInfo)
	return nil
}

type MuSigCommitment struct {
	transport.AppInfo

	// Unique SessionID of the MuSig session.
	SessionID types.Hash

	CommitmentKeyX *big.Int
	CommitmentKeyY *big.Int

	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

func (m MuSigCommitment) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"session_id":       m.SessionID.String(),
		"commitment_key_x": hexutil.BigIntToHex(m.CommitmentKeyX),
		"commitment_key_y": hexutil.BigIntToHex(m.CommitmentKeyY),
		"public_key_x":     hexutil.BigIntToHex(m.PublicKeyX),
		"public_key_y":     hexutil.BigIntToHex(m.PublicKeyY),
	})
}

// MarshallBinary implements the transport.Message interface.
func (m MuSigCommitment) MarshallBinary() ([]byte, error) {
	var (
		pubKeyX []byte
		pubKeyY []byte
		comKeyX []byte
		comKeyY []byte
	)
	if m.PublicKeyX != nil {
		pubKeyX = m.PublicKeyX.Bytes()
	}
	if m.PublicKeyY != nil {
		pubKeyY = m.PublicKeyY.Bytes()
	}
	if m.CommitmentKeyX != nil {
		comKeyX = m.CommitmentKeyX.Bytes()
	}
	if m.CommitmentKeyY != nil {
		comKeyY = m.CommitmentKeyY.Bytes()
	}
	return proto.Marshal(&pb.MuSigCommitmentMessage{
		SessionID:      m.SessionID.Bytes(),
		PubKeyX:        pubKeyX,
		PubKeyY:        pubKeyY,
		CommitmentKeyX: comKeyX,
		CommitmentKeyY: comKeyY,
		AppInfo:        appInfoToProtobuf(m.AppInfo),
	})
}

// UnmarshallBinary implements the transport.Message interface.
func (m *MuSigCommitment) UnmarshallBinary(bytes []byte) error {
	if len(bytes) == 0 {
		return fmt.Errorf("empty data")
	}
	msg := pb.MuSigCommitmentMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	if len(msg.SessionID) != types.HashLength {
		return fmt.Errorf("invalid session ID length")
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.PublicKeyX = new(big.Int).SetBytes(msg.PubKeyX)
	m.PublicKeyY = new(big.Int).SetBytes(msg.PubKeyY)
	m.CommitmentKeyX = new(big.Int).SetBytes(msg.CommitmentKeyX)
	m.CommitmentKeyY = new(big.Int).SetBytes(msg.CommitmentKeyY)
	m.AppInfo = appInfoFromProtobuf(msg.AppInfo)
	return nil
}

type MuSigPartialSignature struct {
	transport.AppInfo

	// Unique SessionID of the MuSig session.
	SessionID types.Hash

	// Partial signature of the MuSig session.
	PartialSignature *big.Int
}

func (m MuSigPartialSignature) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"session_id":        m.SessionID.String(),
		"partial_signature": hexutil.BigIntToHex(m.PartialSignature),
	})
}

// MarshallBinary implements the transport.Message interface.
func (m MuSigPartialSignature) MarshallBinary() ([]byte, error) {
	var partialSignature []byte
	if m.PartialSignature != nil {
		partialSignature = m.PartialSignature.Bytes()
	}
	return proto.Marshal(&pb.MuSigPartialSignatureMessage{
		SessionID:        m.SessionID.Bytes(),
		PartialSignature: partialSignature,
		AppInfo:          appInfoToProtobuf(m.AppInfo),
	})
}

// UnmarshallBinary implements the transport.Message interface.
func (m *MuSigPartialSignature) UnmarshallBinary(bytes []byte) error {
	if len(bytes) == 0 {
		return fmt.Errorf("empty data")
	}
	msg := pb.MuSigPartialSignatureMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	if len(msg.SessionID) != types.HashLength {
		return fmt.Errorf("invalid session ID length")
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.PartialSignature = new(big.Int).SetBytes(msg.PartialSignature)
	m.AppInfo = appInfoFromProtobuf(msg.AppInfo)
	return nil
}

type MuSigSignature struct {
	transport.AppInfo

	*MuSigMessage

	// Unique SessionID of the MuSig session.
	SessionID types.Hash `json:"sessionID"`

	// ComputedAt is the time at which the signature was computed.
	ComputedAt time.Time `json:"computedAt"`

	// Commitment of the MuSig session.
	Commitment types.Address `json:"commitment"`

	// SchnorrSignature is a MuSig Schnorr signature calculated from the partial
	// signatures of all participants.
	SchnorrSignature *big.Int `json:"schnorrSignature"`
}

func (m MuSigSignature) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"session_id":        m.SessionID.String(),
		"computed_at":       m.ComputedAt.In(time.UTC).Format(time.RFC3339),
		"commitment":        m.Commitment.String(),
		"schnorr_signature": hexutil.BigIntToHex(m.SchnorrSignature),
		"msg_type":          m.MsgType,
		"msg_body":          m.MsgBody.String(),
		"msg_meta":          m.MsgMeta.Meta,
		"signers":           m.Signers,
	})
}

func (m *MuSigSignature) toProtobuf() (*pb.MuSigSignatureMessage, error) {
	if m.MuSigMessage == nil {
		return nil, fmt.Errorf("empty message")
	}
	meta, err := m.MsgMeta.toProtobuf()
	if err != nil {
		return nil, err
	}
	msg := &pb.MuSigSignatureMessage{
		SessionID:           m.SessionID[:],
		ComputedAtTimestamp: m.ComputedAt.Unix(),
		MsgType:             m.MsgType,
		MsgBody:             m.MsgBody.Bytes(),
		MsgMeta:             meta,
		Commitment:          m.Commitment.Bytes(),
		Signers:             make([][]byte, len(m.Signers)),
		SchnorrSignature:    m.SchnorrSignature.Bytes(),
		AppInfo:             appInfoToProtobuf(m.AppInfo),
	}
	for i, signer := range m.Signers {
		msg.Signers[i] = signer.Bytes()
	}
	return msg, nil
}

func (m *MuSigSignature) fromProtobuf(msg *pb.MuSigSignatureMessage) error {
	m.MuSigMessage = &MuSigMessage{}
	if len(msg.MsgBody) != types.HashLength {
		return fmt.Errorf("invalid message body length")
	}
	if len(msg.SessionID) != types.HashLength {
		return fmt.Errorf("invalid session ID length")
	}
	com, err := types.AddressFromBytes(msg.Commitment)
	if err != nil {
		return err
	}
	m.SessionID = types.MustHashFromBytes(msg.SessionID, types.PadLeft)
	m.ComputedAt = time.Unix(msg.ComputedAtTimestamp, 0)
	m.MsgType = msg.MsgType
	m.MsgBody = types.MustHashFromBytes(msg.MsgBody, types.PadLeft)
	if err := m.MsgMeta.fromProtobuf(msg.MsgMeta); err != nil {
		return err
	}
	m.Commitment = com
	m.Signers = make([]types.Address, len(msg.Signers))
	for i, signer := range msg.Signers {
		m.Signers[i], err = types.AddressFromBytes(signer)
		if err != nil {
			return err
		}
	}
	m.SchnorrSignature = new(big.Int).SetBytes(msg.SchnorrSignature)
	m.AppInfo = appInfoFromProtobuf(msg.AppInfo)
	return nil
}

// MarshallBinary implements the transport.Message interface.
func (m MuSigSignature) MarshallBinary() ([]byte, error) {
	msg, err := m.toProtobuf()
	if err != nil {
		return nil, err
	}
	return proto.Marshal(msg)
}

// UnmarshallBinary implements the transport.Message interface.
func (m *MuSigSignature) UnmarshallBinary(bytes []byte) error {
	if len(bytes) == 0 {
		return fmt.Errorf("empty data")
	}
	msg := &pb.MuSigSignatureMessage{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return err
	}
	return m.fromProtobuf(msg)
}
