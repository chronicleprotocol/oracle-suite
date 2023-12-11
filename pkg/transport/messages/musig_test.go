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
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestMuSigInitialize_MarshallBinary(t *testing.T) {
	tests := []struct {
		name       string
		initialize MuSigInitialize
		wantErr    bool
	}{
		{
			name: "valid Initialization",
			initialize: MuSigInitialize{
				SessionID: types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
				StartedAt: time.Unix(1630458972, 0),
				MuSigMessage: &MuSigMessage{
					MsgType: "testType",
					MsgBody: types.MustHashFromHex("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", types.PadNone),
					MsgMeta: MuSigMeta{
						Meta: MuSigMetaTickV1{
							Wat:       "TestAsset",
							Val:       bn.DecFixedPoint(100, 2),
							Age:       time.Unix(1630458972, 0),
							ECDSAData: types.MustSignatureFromHexPtr("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
							FeedTicks: []MuSigMetaFeedTick{
								{
									Val: bn.DecFixedPoint(100, 2),
									Age: time.Unix(1630458972, 0),
									VRS: types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
								},
							},
						},
					},
					Signers: []types.Address{},
				},
			},
			wantErr: false,
		},
		{
			name: "empty slices",
			initialize: MuSigInitialize{
				SessionID: types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
				StartedAt: time.Unix(1630458972, 0),
				MuSigMessage: &MuSigMessage{
					MsgType: "",
					MsgBody: types.MustHashFromHex("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", types.PadNone),
					MsgMeta: MuSigMeta{
						Meta: MuSigMetaTickV1{
							Wat:       "",
							Val:       nil,
							Age:       time.Unix(0, 0),
							ECDSAData: nil,
							FeedTicks: []MuSigMetaFeedTick{},
						},
					},
					Signers: []types.Address{},
				},
			},
			wantErr: false,
		},
		{
			name: "nil values",
			initialize: MuSigInitialize{
				SessionID: types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
				StartedAt: time.Unix(0, 0),
				MuSigMessage: &MuSigMessage{
					MsgType: "",
					MsgBody: types.MustHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000000", types.PadNone),
					MsgMeta: MuSigMeta{
						Meta: nil,
					},
					Signers: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "empty message",
			initialize: MuSigInitialize{
				SessionID:    types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
				StartedAt:    time.Unix(0, 0),
				MuSigMessage: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.initialize.MarshallBinary()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMuSigInitialize_UnmarshallBinary(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		wantErr bool
	}{
		{
			name: "valid bytes",
			bytes: func() []byte {
				init := MuSigInitialize{
					SessionID: types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
					StartedAt: time.Unix(1630458972, 0),
					MuSigMessage: &MuSigMessage{
						MsgType: "testType",
						MsgBody: types.MustHashFromHex("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", types.PadNone),
						MsgMeta: MuSigMeta{
							Meta: MuSigMetaTickV1{
								Wat:       "TestAsset",
								Val:       bn.DecFixedPoint(100, 2),
								Age:       time.Unix(1630458972, 0),
								ECDSAData: types.MustSignatureFromHexPtr("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
								FeedTicks: []MuSigMetaFeedTick{
									{
										Val: bn.DecFixedPoint(100, 2),
										Age: time.Unix(1630458972, 0),
										VRS: types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
									},
								},
							},
						},
						Signers: []types.Address{
							types.MustAddressFromHex("0x1234567890abcdef1234567890abcdef12345678"),
							types.MustAddressFromHex("0xabcdef1234567890abcdef1234567890abcdef12"),
						},
					},
				}
				bytes, _ := init.MarshallBinary()
				return bytes
			}(),
			wantErr: false,
		},
		{
			name:    "invalid bytes",
			bytes:   []byte("invalid bytes"),
			wantErr: true,
		},
		{
			name:    "empty bytes",
			bytes:   []byte{},
			wantErr: true,
		},
		{
			name: "valid bytes with empty fields",
			bytes: func() []byte {
				init := MuSigInitialize{
					SessionID: types.MustHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000000", types.PadNone),
					StartedAt: time.Unix(0, 0),
					MuSigMessage: &MuSigMessage{
						MsgType: "",
						MsgBody: types.MustHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000000", types.PadNone),
						MsgMeta: MuSigMeta{
							Meta: MuSigMetaTickV1{
								Wat:       "",
								Val:       nil,
								Age:       time.Unix(0, 0),
								ECDSAData: nil,
								FeedTicks: []MuSigMetaFeedTick{},
							},
						},
						Signers: []types.Address{},
					},
				}
				bytes, _ := init.MarshallBinary()
				return bytes
			}(),
			wantErr: false,
		},
		{
			name: "valid bytes with nil meta",
			bytes: func() []byte {
				init := MuSigInitialize{
					SessionID: types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
					StartedAt: time.Unix(1630458972, 0),
					MuSigMessage: &MuSigMessage{
						MsgType: "testType",
						MsgBody: types.MustHashFromHex("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", types.PadNone),
						MsgMeta: MuSigMeta{
							Meta: nil,
						},
						Signers: []types.Address{},
					},
				}
				bytes, _ := init.MarshallBinary()
				return bytes
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			init := &MuSigInitialize{}
			err := init.UnmarshallBinary(tt.bytes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func FuzzMuSigInitialize_UnmarshallBinary(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = (&MuSigInitialize{}).UnmarshallBinary(data)
	})
}

func TestMuSigCommitment_MarshallBinary(t *testing.T) {
	tests := []struct {
		name       string
		commitment MuSigCommitment
		wantErr    bool
	}{
		{
			name: "valid Commitment",
			commitment: MuSigCommitment{
				SessionID:      types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
				CommitmentKeyX: big.NewInt(1234567890),
				CommitmentKeyY: big.NewInt(1234567890),
				PublicKeyX:     big.NewInt(1234567890),
				PublicKeyY:     big.NewInt(1234567890),
			},
			wantErr: false,
		},
		{
			name: "nil values",
			commitment: MuSigCommitment{
				SessionID:      types.MustHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000000", types.PadNone),
				CommitmentKeyX: nil,
				CommitmentKeyY: nil,
				PublicKeyX:     nil,
				PublicKeyY:     nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.commitment.MarshallBinary()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMuSigCommitment_UnmarshallBinary(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		wantErr bool
	}{
		{
			name: "valid bytes",
			bytes: func() []byte {
				commitment := MuSigCommitment{
					SessionID:      types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
					CommitmentKeyX: big.NewInt(12345),
					CommitmentKeyY: big.NewInt(67890),
					PublicKeyX:     big.NewInt(112233),
					PublicKeyY:     big.NewInt(445566),
				}
				bytes, _ := commitment.MarshallBinary()
				return bytes
			}(),
			wantErr: false,
		},
		{
			name:    "invalid bytes",
			bytes:   []byte("invalid bytes"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var commitment MuSigCommitment
			err := commitment.UnmarshallBinary(tt.bytes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func FuzzMuSigCommitment_UnmarshallBinary(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = (&MuSigCommitment{}).UnmarshallBinary(data)
	})
}

func TestMuSigPartialSignature_MarshallBinary(t *testing.T) {
	tests := []struct {
		name             string
		partialSignature MuSigPartialSignature
		wantErr          bool
	}{
		{
			name: "valid PartialSignature",
			partialSignature: MuSigPartialSignature{
				SessionID:        types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
				PartialSignature: big.NewInt(1234567890),
			},
			wantErr: false,
		},
		{
			name: "nil PartialSignature",
			partialSignature: MuSigPartialSignature{
				SessionID:        types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
				PartialSignature: nil,
			},
			wantErr: false,
		},
		{
			name: "zero SessionID",
			partialSignature: MuSigPartialSignature{
				SessionID:        types.MustHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000000", types.PadNone),
				PartialSignature: big.NewInt(1234567890),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.partialSignature.MarshallBinary()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMuSigPartialSignature_UnmarshallBinary(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		wantErr bool
	}{
		{
			name: "valid bytes",
			bytes: func() []byte {
				partialSignature := MuSigPartialSignature{
					SessionID:        types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
					PartialSignature: big.NewInt(1234567890),
				}
				bytes, _ := partialSignature.MarshallBinary()
				return bytes
			}(),
			wantErr: false,
		},
		{
			name:    "invalid bytes",
			bytes:   []byte("invalid bytes"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var partialSignature MuSigPartialSignature
			err := partialSignature.UnmarshallBinary(tt.bytes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func FuzzMuSigPartialSignature_UnmarshallBinary(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = (&MuSigPartialSignature{}).UnmarshallBinary(data)
	})
}

func TestMuSigSignature_MarshallBinary(t *testing.T) {
	tests := []struct {
		name      string
		signature MuSigSignature
		wantErr   bool
	}{
		{
			name: "valid Signature",
			signature: MuSigSignature{
				SessionID:  types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
				ComputedAt: time.Unix(1630458972, 0),
				MuSigMessage: &MuSigMessage{
					MsgType: "testType",
					MsgBody: types.MustHashFromHex("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", types.PadNone),
					MsgMeta: MuSigMeta{
						Meta: MuSigMetaTickV1{
							Wat: "TestAsset",
							Val: bn.DecFixedPoint(100, 2),
							Age: time.Unix(1630458972, 0),
							FeedTicks: []MuSigMetaFeedTick{
								{
									Val: bn.DecFixedPoint(100, 2),
									Age: time.Unix(1630458972, 0),
									VRS: types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
								},
							},
						},
					},
					Signers: []types.Address{},
				},
				Commitment:       types.MustAddressFromHex("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				SchnorrSignature: big.NewInt(1234567890),
			},
			wantErr: false,
		},
		{
			name: "empty slices and zero values",
			signature: MuSigSignature{
				SessionID:  types.MustHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000000", types.PadNone),
				ComputedAt: time.Unix(0, 0),
				MuSigMessage: &MuSigMessage{
					MsgType: "",
					MsgBody: types.MustHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000000", types.PadNone),
					MsgMeta: MuSigMeta{Meta: nil},
					Signers: []types.Address{},
				},
				Commitment:       types.MustAddressFromHex("0x0000000000000000000000000000000000000000"),
				SchnorrSignature: big.NewInt(0),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.signature.MarshallBinary()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMuSigSignature_UnmarshallBinary(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		wantErr bool
	}{
		{
			name: "valid bytes",
			bytes: func() []byte {
				signature := MuSigSignature{
					SessionID:  types.MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", types.PadNone),
					ComputedAt: time.Unix(1630458972, 0),
					MuSigMessage: &MuSigMessage{
						MsgType: "testType",
						MsgBody: types.MustHashFromHex("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", types.PadNone),
						MsgMeta: MuSigMeta{
							Meta: MuSigMetaTickV1{
								Wat: "TestAsset",
								Val: bn.DecFixedPoint(100, 2),
								Age: time.Unix(1630458972, 0),
								FeedTicks: []MuSigMetaFeedTick{
									{
										Val: bn.DecFixedPoint(100, 2),
										Age: time.Unix(1630458972, 0),
										VRS: types.MustSignatureFromHex("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"),
									},
								},
							},
						},
						Signers: []types.Address{types.MustAddressFromHex("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")},
					},
					Commitment:       types.MustAddressFromHex("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
					SchnorrSignature: big.NewInt(1234567890),
				}
				bytes, _ := signature.MarshallBinary()
				return bytes
			}(),
			wantErr: false,
		},
		{
			name:    "invalid bytes",
			bytes:   []byte("invalid bytes"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var signature MuSigSignature
			err := signature.UnmarshallBinary(tt.bytes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func FuzzMuSigSignature_UnmarshallBinary(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = (&MuSigSignature{}).UnmarshallBinary(data)
	})
}
