package tor

import (
	"sort"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/tor/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

type messageSigner struct {
	signer ethereum.Signer
}

func newMessageSigner(signer ethereum.Signer) *messageSigner {
	return &messageSigner{
		signer: signer,
	}
}

func (ms *messageSigner) SignMessage(msg *pb.MessagePack) error {
	signature, err := ms.signer.Signature(ms.signingData(msg))
	if err != nil {
		return err
	}
	msg.Signature = signature.Bytes()
	return nil
}

func (ms *messageSigner) VerifyMessage(msg *pb.MessagePack) (*ethereum.Address, error) {
	return ms.signer.Recover(ethereum.SignatureFromBytes(msg.Signature), ms.signingData(msg))
}

func (ms *messageSigner) signingData(msg *pb.MessagePack) []byte {
	var signingData []byte
	topics := maputil.SortKeys(msg.Messages, sort.Strings)
	for _, topic := range topics {
		for _, data := range msg.Messages[topic].Data {
			signingData = append(signingData, data...)
		}
	}
	return signingData
}
