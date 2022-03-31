//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package cobra

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/cmd/keeman/eth"
	"github.com/chronicleprotocol/oracle-suite/cmd/keeman/rand"
	"github.com/chronicleprotocol/oracle-suite/cmd/keeman/ssb"
)

const (
	Ethereum                  = "m/44'/60'/0'/0"
	EthereumClassic           = "m/44'/61'/0'/0"
	EthereumTestnetRopsten    = "m/44'/1'/0'/0"
	EthereumLedger            = "m/44'/60'/0'"
	EthereumClassicLedger     = "m/44'/60'/160720'/0"
	EthereumLedgerLive        = "m/44'/60'"
	EthereumClassicLedgerLive = "m/44'/61'"
	RSKMainnet                = "m/44'/137'/0'/0"
	Expanse                   = "m/44'/40'/0'/0"
	Ubiq                      = "m/44'/108'/0'/0"
	Ellaism                   = "m/44'/163'/0'/0"
	EtherGem                  = "m/44'/1987'/0'/0"
	Callisto                  = "m/44'/820'/0'/0"
	EthereumSocial            = "m/44'/1128'/0'/0"
	Musicoin                  = "m/44'/184'/0'/0"
	EOSClassic                = "m/44'/2018'/0'/0"
	Akroma                    = "m/44'/200625'/0'/0"
	EtherSocialNetwork        = "m/44'/31102'/0'/0"
	PIRL                      = "m/44'/164'/0'/0"
	GoChain                   = "m/44'/6060'/0'/0"
	Ether                     = "m/44'/1313114'/0'/0"
	Atheios                   = "m/44'/1620'/0'/0"
	TomoChain                 = "m/44'/889'/0'/0"
	MixBlockchain             = "m/44'/76'/0'/0"
	Iolite                    = "m/44'/1171337'/0'/0"
	ThunderCore               = "m/44'/1001'/0'/0"
)

// Old paths:
// m/<env=[0,1]>'/<purpose>/<role>/<idx>
// key purpose
// eth.account path: m/0'/0/0/0
// p2p path: m/0'/1/0/0
// ssb path: m/0'/2/0/0
// caps.shs path: m/0'/3/0
// caps.sign path: m/0'/3/1
// nodeRoles = {
// "eth" = 0;
// "boot" = 1;
// "feed" = 2;
// "feed_lb" = 3;
// "bb" = 4;
// "relay" = 5;
// "spectre" = 6;
// "ghost" = 7;
// "monitor" = 8;
// };

func NewDerive(opts *Options) *cobra.Command {
	var prefix, password, format string
	var index int
	cmd := &cobra.Command{
		Use:     "derive [--prefix path] [--format eth|ssb|b32] [--password] path...",
		Aliases: []string{"der", "d"},
		Short:   "Derive a key pair from the provided mnemonic phrase",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{"0"}
			}
			mnemonic, err := lineFromFile(opts.InputFile, index)
			if err != nil {
				return err
			}
			wallet, err := hdwallet.NewFromMnemonic(mnemonic)
			if err != nil {
				return err
			}
			if prefix != "" && !strings.HasSuffix(prefix, "/") {
				prefix += "/"
			}
			for _, arg := range args {
				dp, err := accounts.ParseDerivationPath(prefix + arg)
				if err != nil {
					return err
				}
				log.Println(dp.String())
				acc, err := wallet.Derive(dp, false)
				if err != nil {
					return err
				}
				log.Println(acc.Address.String())
				privateKey, err := wallet.PrivateKey(acc)
				if err != nil {
					return err
				}
				b, err := formattedBytes(format, privateKey, password)
				if err != nil {
					return err
				}
				fmt.Println(string(b))
			}
			return nil
		},
	}
	cmd.Flags().IntVar(
		&index,
		"index",
		0,
		"data index (i.e. which seed line to take from input)",
	)
	cmd.Flags().StringVar(
		&prefix,
		"prefix",
		"",
		"derivation path prefix",
	)
	cmd.Flags().StringVar(
		&password,
		"password",
		"",
		"encryption password",
	)
	cmd.Flags().StringVar(
		&format,
		"format",
		FormatKeystore,
		"output format",
	)
	return cmd
}

const (
	FormatKeystore = "eth"
	FormatSSB      = "ssb"
	FormatSSBSHS   = "shs"
	FormatSSBCaps  = "caps"
	FormatBytes32  = "b32"
	FormatPrivHex  = "privhex"
	FormatLibP2P   = "libp2p"
)

func formattedBytes(format string, privateKey *ecdsa.PrivateKey, password string) ([]byte, error) {
	switch format {
	case FormatBytes32, FormatSSBSHS:
		randBytes, err := seededRandBytesFunc(privateKey, 32)
		if err != nil {
			return nil, err
		}
		return b64Encode(randBytes()), nil
	case FormatSSB:
		o, err := ssb.NewSecret(crypto.FromECDSA(privateKey))
		if err != nil {
			return nil, err
		}
		return json.Marshal(o)
	case FormatSSBCaps:
		o, err := ssb.NewCaps(crypto.FromECDSA(privateKey))
		if err != nil {
			return nil, err
		}
		return json.Marshal(o)
	case FormatKeystore:
		key, err := eth.NewKeyWithID(privateKey)
		if err != nil {
			return nil, err
		}
		return keystore.EncryptKey(
			key,
			password,
			keystore.StandardScryptN,
			keystore.StandardScryptP,
		)
	case FormatPrivHex:
		return hexEncodeBytes(crypto.FromECDSA(privateKey)), nil
	case FormatLibP2P:
		randBytes, err := seededRandBytesFunc(privateKey, 32)
		if err != nil {
			return nil, err
		}
		return hexEncodeBytes(randBytes()), nil
	}
	return nil, fmt.Errorf("unknown format: %s", format)
}
func hexEncodeBytes(b []byte) []byte {
	buff := make([]byte, len(b)*2)
	hex.Encode(buff, b)
	return buff
}
func b64Encode(b []byte) []byte {
	enc := base64.StdEncoding
	buff := make([]byte, enc.EncodedLen(len(b)))
	enc.Encode(buff, b)
	return buff
}
func seededRandBytesFunc(privateKey *ecdsa.PrivateKey, len int) (func() []byte, error) {
	return rand.SeededRandBytesGen(crypto.FromECDSA(privateKey), len)
}
