package tor

import (
	"encoding/hex"
	"errors"
	"io"
	netURL "net/url"
	"strconv"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type urlSigner struct {
	signer ethereum.Signer
	rand   io.Reader
}

func newURLSigner(signer ethereum.Signer, rand io.Reader) *urlSigner {
	return &urlSigner{
		signer: signer,
		rand:   rand,
	}
}

func (us *urlSigner) SignURL(url string) (string, error) {
	r := make([]byte, 16)
	_, err := us.rand.Read(r)
	if err != nil {
		return "", err
	}
	t := strconv.FormatInt(time.Now().Unix(), 10)
	s, err := us.signer.Signature([]byte(t + hex.EncodeToString(r)))
	if err != nil {
		return "", err
	}
	return url + "?t=" + t + "&r=" + hex.EncodeToString(r) + "&s=" + hex.EncodeToString(s.Bytes()), nil
}

func (us *urlSigner) VerifyURL(url string) (*ethereum.Address, error) {
	p, err := netURL.Parse(url)
	if err != nil {
		panic(err)
	}
	q := p.Query()
	t, err := strconv.ParseInt(q.Get("t"), 10, 64)
	if err != nil {
		return nil, err
	}
	if time.Now().Unix()-t > 60 {
		return nil, errors.New("url expired")
	}
	s, err := hex.DecodeString(q.Get("s"))
	if err != nil {
		return nil, err
	}
	return us.signer.Recover(ethereum.SignatureFromBytes(s), []byte(q.Get("t")+q.Get("r")))
}
