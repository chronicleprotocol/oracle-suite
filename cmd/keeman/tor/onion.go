package tor

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/base32"
	"encoding/json"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Onion struct {
	Address   string
	PublicKey ed25519.PublicKey
	SecretKey [64]byte
}

func (c Onion) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Prefix    string `json:"prefix"`
		Hostname  string `json:"hostname"`
		PublicKey []byte `json:"public_key"`
		SecretKey []byte `json:"secret_key"`
	}{
		Prefix:   "hs_ed25519",
		Hostname: c.Address + ".onion",
		PublicKey: append(
			[]byte("== ed25519v1-public: type0 ==\x00\x00\x00"),
			c.PublicKey...,
		),
		SecretKey: append(
			[]byte("== ed25519v1-secret: type0 ==\x00\x00\x00"),
			c.SecretKey[:]...,
		),
	})
}

func NewOnion(seedBytes []byte) (*Onion, error) {
	publicKey, secretKey, err := ed25519.GenerateKey(bytes.NewBuffer(seedBytes))
	if err != nil {
		return nil, err
	}
	return &Onion{
		Address:   encodePublicKey(publicKey),
		PublicKey: publicKey,
		SecretKey: expandSecretKey(secretKey),
	}, nil
}

func expandSecretKey(secretKey ed25519.PrivateKey) [64]byte {
	hash := sha512.Sum512(secretKey[:32])
	hash[0] &= 248
	hash[31] &= 127
	hash[31] |= 64
	return hash
}

func encodePublicKey(publicKey ed25519.PublicKey) string {
	// checksum = H(".onion checksum" || pubkey || version)
	var checksumBytes bytes.Buffer
	checksumBytes.Write([]byte(".onion checksum"))
	checksumBytes.Write(publicKey)
	checksumBytes.Write([]byte{0x03})
	checksum := sha3.Sum256(checksumBytes.Bytes())

	// onion_address = base32(pubkey || checksum || version)
	var onionAddressBytes bytes.Buffer
	onionAddressBytes.Write(publicKey)
	onionAddressBytes.Write(checksum[:2])
	onionAddressBytes.Write([]byte{0x03})
	onionAddress := base32.StdEncoding.EncodeToString(onionAddressBytes.Bytes())

	return strings.ToLower(onionAddress)
}
