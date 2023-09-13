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
	"sort"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

type MessageMap map[string]transport.Message

// Keys returns a sorted list of keys.
func (mm MessageMap) Keys() []string {
	return maputil.SortKeys(mm, sort.Strings)
}

// SelectByTopic returns a new MessageMap with messages selected by topic.
// Empty topic list will yield an empty map.
func (mm MessageMap) SelectByTopic(topics ...string) (MessageMap, error) {
	return maputil.Select(mm, topics)
}

var AllMessagesMap = MessageMap{
	PriceV0MessageName:                 (*Price)(nil),
	PriceV1MessageName:                 (*Price)(nil),
	DataPointV1MessageName:             (*DataPoint)(nil),
	GreetV1MessageName:                 (*Greet)(nil),
	MuSigStartV1MessageName:            (*MuSigInitialize)(nil),
	MuSigTerminateV1MessageName:        (*MuSigTerminate)(nil),
	MuSigCommitmentV1MessageName:       (*MuSigCommitment)(nil),
	MuSigPartialSignatureV1MessageName: (*MuSigPartialSignature)(nil),
	MuSigSignatureV1MessageName:        (*MuSigSignature)(nil),
}

func appInfoToProtobuf(a transport.AppInfo) *pb.AppInfo {
	return &pb.AppInfo{
		Name:    a.Name,
		Version: a.Version,
	}
}

func appInfoFromProtobuf(a *pb.AppInfo) transport.AppInfo {
	return transport.AppInfo{
		Name:    a.Name,
		Version: a.Version,
	}
}

func decFloatPointToBytes(d *bn.DecFloatPointNumber) ([]byte, error) {
	return d.MarshalBinary()
}

func bytesToDecFloatPoint(b []byte) (*bn.DecFloatPointNumber, error) {
	d := new(bn.DecFloatPointNumber)
	if err := d.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return d, nil
}

func decFixedPointToBytes(d *bn.DecFixedPointNumber) ([]byte, error) {
	return d.MarshalBinary()
}

func bytesToDecFixedPoint(b []byte) (*bn.DecFixedPointNumber, error) {
	d := new(bn.DecFixedPointNumber)
	if err := d.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return d, nil
}
