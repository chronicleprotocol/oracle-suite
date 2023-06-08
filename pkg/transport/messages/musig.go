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
	MuSigNonceV1MessageName            = "musig_nonce/v1"
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
	MsgBody []byte

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
		MsgBody:            m.MsgBody,
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
	copy(m.SessionID[:], msg.SessionID)
	m.StartedAt = time.Unix(msg.StartedAtTimestamp, 0)
	m.MsgType = msg.MsgType
	m.MsgBody = msg.MsgBody
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

	// Commitment of the MuSig session.
	Commitment types.Address

	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

func (m *MuSigCommitment) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigCommitmentMessage{
		SessionID:  m.SessionID[:],
		Commitment: m.Commitment.Bytes(),
		PubKeyX:    m.PublicKeyX.Bytes(),
		PubKeyY:    m.PublicKeyY.Bytes(),
	})
}

func (m *MuSigCommitment) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigCommitmentMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	if len(msg.Commitment) != types.AddressLength {
		return fmt.Errorf("invalid commitment length: %d", len(msg.Commitment))
	}
	copy(m.SessionID[:], msg.SessionID)
	m.Commitment = types.MustAddressFromBytes(msg.Commitment)
	m.PublicKeyX = new(big.Int).SetBytes(msg.PubKeyX)
	m.PublicKeyY = new(big.Int).SetBytes(msg.PubKeyY)
	return nil
}

type MuSigNonce struct {
	// Unique SessionID of the MuSig session.
	SessionID [32]byte

	// Nonce of the MuSig session.
	Nonce *big.Int
}

func (m *MuSigNonce) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigNonceMessage{
		SessionID: m.SessionID[:],
		Nonce:     m.Nonce.Bytes(),
	})
}

func (m *MuSigNonce) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigNonceMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	copy(m.SessionID[:], msg.SessionID)
	m.Nonce = new(big.Int).SetBytes(msg.Nonce)
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
	Type string

	// Data that was signed.
	Data []byte

	// Signature of the MuSig session.
	Signature *big.Int
}

func (m *MuSigSignature) MarshallBinary() ([]byte, error) {
	return proto.Marshal(&pb.MuSigSignatureMessage{
		SessionID: m.SessionID[:],
		Type:      m.Type,
		Data:      m.Data,
		Signature: m.Signature.Bytes(),
	})
}

func (m *MuSigSignature) UnmarshallBinary(bytes []byte) error {
	msg := pb.MuSigSignatureMessage{}
	if err := proto.Unmarshal(bytes, &msg); err != nil {
		return err
	}
	copy(m.SessionID[:], msg.SessionID)
	m.Type = msg.Type
	m.Data = msg.Data
	m.Signature = new(big.Int).SetBytes(msg.Signature)
	return nil
}
