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

package transport

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/net/proxy"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/chain"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/recoverer"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/webapi"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const LoggerTag = "CONFIG_LIBP2P"

type Dependencies struct {
	Keys     ethereum.KeyRegistry
	Clients  ethereum.ClientRegistry
	Messages map[string]transport.Message
	Logger   log.Logger

	// Application info:
	AppName    string
	AppVersion string
}

type BootstrapDependencies struct {
	Logger log.Logger

	// Application info:
	AppName    string
	AppVersion string
}

type Config struct {
	LibP2P *libP2PConfig `hcl:"libp2p,block,optional"`
	WebAPI *webAPIConfig `hcl:"webapi,block,optional"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured transport:
	transport transport.Service
}

type libP2PConfig struct {
	// Feeds is a list of Ethereum addresses that are allowed to send messages
	// to the node.
	Feeds []types.Address `hcl:"feeds"`

	DisableFeedFilter bool `hcl:"feeds_filter_disable,optional"`

	// ListenAddrs is the list of listening addresses for libp2p node encoded
	// using the multiaddress format.
	ListenAddrs []string `hcl:"listen_addrs"`

	// ExternalAddr is the external IP address of the node. It will be added to the local address list
	ExternalAddr string `hcl:"external_addr,optional"`

	// PrivKeySeed is the random hex-encoded 32 bytes. It is used to generate
	// a unique identity on the libp2p network. The value may be empty to
	// generate a random seed.
	PrivKeySeed string `hcl:"priv_key_seed,optional"`

	// BootstrapAddrs is the list of bootstrap addresses for libp2p node
	// encoded using the multiaddress format.
	BootstrapAddrs []string `hcl:"bootstrap_addrs,optional"`

	// DirectPeersAddrs is the list of direct peer addresses to which messages
	// will be sent directly. Addresses are encoded using the format the
	// multiaddress format. This option must be configured symmetrically on
	// both ends.
	DirectPeersAddrs []string `hcl:"direct_peers_addrs,optional"`

	// BlockedAddrs is the list of blocked addresses encoded using the
	// multiaddress format.
	BlockedAddrs []string `hcl:"blocked_addrs,optional"`

	// DisableDiscovery disables node discovery. If enabled, the IP address of
	// a node will not be broadcast to other peers. This option must be used
	// together with `directPeersAddrs`.
	DisableDiscovery bool `hcl:"disable_discovery,optional"`

	// EthereumKey is the name of the Ethereum key to use for signing messages.
	// Required if the transport is used for sending messages.
	EthereumKey string `hcl:"ethereum_key,optional"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type webAPIConfig struct {
	// Feeds is a list of Ethereum addresses that are allowed to send messages
	// to the node.
	Feeds []types.Address `hcl:"feeds"`

	// ListenAddr is the address on which the WebAPI server will listen for
	// incoming connections. The address must be in the format `host:port`.
	// When used with a TOR hidden service, the server should listen on
	// localhost.
	ListenAddr string `hcl:"listen_addr"`

	// Socks5ProxyAddr is the address of the SOCKS5 proxy server. The address
	// must be in the format `host:port`.
	Socks5ProxyAddr string `hcl:"socks5_proxy_addr,optional"`

	// EthereumKey is the name of the Ethereum key to use for signing messages.
	// Required if the transport is used for sending messages.
	EthereumKey string `hcl:"ethereum_key"`

	// AddressBook configuration. Address book provides a list of addresses
	// to which messages will be sent.

	// EthereumAddressBook is the configuration for the Ethereum address book.
	EthereumAddressBook *webAPIEthereumAddressBook `hcl:"ethereum_address_book,block,optional"`

	// StaticAddressBook is the configuration for the static address book.
	StaticAddressBook *webAPIStaticAddressBook `hcl:"static_address_book,block,optional"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type webAPIEthereumAddressBook struct {
	// ContractAddr is the Ethereum address of the address book contract.
	ContractAddr types.Address `hcl:"contract_addr"`

	// EthereumClient is the name of the Ethereum client to use for reading
	// the address book.
	EthereumClient string `hcl:"ethereum_client"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type webAPIStaticAddressBook struct {
	// Addresses is the list of static addresses to which messages will be
	// sent.
	Addresses []string `hcl:"addresses"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

func (c *Config) Transport(d Dependencies) (transport.Service, error) {
	if c.transport != nil {
		return c.transport, nil
	}
	var transports []transport.Service
	if c.LibP2P != nil {
		t, err := c.configureLibP2P(d)
		if err != nil {
			return nil, err
		}
		transports = append(transports, t)
	}
	if c.WebAPI != nil {
		t, err := c.configureWebAPI(d)
		if err != nil {
			return nil, err
		}
		transports = append(transports, t)
	}
	switch {
	case len(transports) == 0:
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "At least one transport must be configured.",
			Subject:  &c.Range,
		}
	case len(transports) == 1:
		c.transport = transports[0]
	default:
		c.transport = chain.New(transports...)
	}
	return logger.New(c.transport, d.Logger), nil
}

func (c *Config) LibP2PBootstrap(d BootstrapDependencies) (transport.Service, error) {
	if c.LibP2P == nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "LibP2P transport must be configured.",
			Subject:  &c.Range,
		}
	}
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}
	var extAddr multiaddr.Multiaddr
	if c.LibP2P.ExternalAddr != "" {
		extAddr, err = parseAddr(c.LibP2P.ExternalAddr, c.LibP2P.ListenAddrs)
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Could not determine external address from %s", c.LibP2P.ExternalAddr),
				Subject:  c.LibP2P.Content.Attributes["external_addr"].Range.Ptr(),
			}
		}
	}
	cfg := libp2p.Config{
		Mode:             libp2p.BootstrapMode,
		PeerPrivKey:      peerPrivKey,
		ListenAddrs:      c.LibP2P.ListenAddrs,
		ExternalAddr:     extAddr,
		BootstrapAddrs:   c.LibP2P.BootstrapAddrs,
		DirectPeersAddrs: c.LibP2P.DirectPeersAddrs,
		BlockedAddrs:     c.LibP2P.BlockedAddrs,
		Logger:           d.Logger,
		AppName:          d.AppName,
		AppVersion:       d.AppVersion,
	}
	p, err := libp2p.New(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Cannot create LibP2P bootstrap node: %v", err),
			Subject:  &c.LibP2P.Range,
		}
	}
	return p, nil
}

func (c *Config) configureWebAPI(d Dependencies) (transport.Service, error) {
	l := d.Logger.WithField("tag", "CONFIG_"+webapi.LoggerTag)

	// Configure HTTP client:
	httpClient := &http.Client{}
	if len(c.WebAPI.Socks5ProxyAddr) != 0 {
		dialer, err := proxy.SOCKS5("tcp", c.WebAPI.Socks5ProxyAddr, nil, proxy.Direct)
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Cannot create SOCKS5 proxy: %v", err),
				Subject:  &c.WebAPI.Content.Attributes["socks5_proxy_addr"].Range,
			}
		}
		httpClient.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.Dial(network, address)
			},
		}
		l.WithField("address", c.WebAPI.Socks5ProxyAddr).
			Info("SOCKS5 proxy")
	}

	// Configure address book:
	var addressBooks []webapi.AddressBook
	if c.WebAPI.EthereumAddressBook != nil {
		l.WithField("address", c.WebAPI.EthereumAddressBook.ContractAddr).
			Info("Ethereum address book")

		rpcClient := d.Clients[c.WebAPI.EthereumAddressBook.EthereumClient]
		if rpcClient == nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Ethereum client %q is not configured", c.WebAPI.EthereumAddressBook.EthereumClient),
				Subject:  c.WebAPI.EthereumAddressBook.Content.Attributes["ethereum_client"].Range.Ptr(),
			}
		}
		addressBooks = append(addressBooks, webapi.NewEthereumAddressBook(
			rpcClient,
			c.WebAPI.EthereumAddressBook.ContractAddr,
			time.Hour,
		))
	}
	if c.WebAPI.StaticAddressBook != nil {
		addressBooks = append(
			addressBooks,
			webapi.NewStaticAddressBook(c.WebAPI.StaticAddressBook.Addresses),
		)
	}

	var addressBook webapi.AddressBook
	switch {
	case len(addressBooks) == 0:
		addressBook = webapi.NullAddressBook{}
	case len(addressBooks) == 1:
		addressBook = addressBooks[0]
	default:
		addressBook = webapi.NewMultiAddressBook(addressBooks...)
	}

	// Log consumers:
	consumers, err := addressBook.Consumers(context.Background())
	if err != nil {
		l.WithError(err).Error("Failed to get consumers")
	}
	for _, c := range consumers {
		l.WithField("address", c).
			Info("Consumer")
	}

	// Configure signer:
	key := d.Keys[c.WebAPI.EthereumKey]
	if c.WebAPI.EthereumKey != "" && key == nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Ethereum key %q is not configured", c.WebAPI.EthereumKey),
			Subject:  c.WebAPI.Content.Attributes["ethereum_key"].Range.Ptr(),
		}
	}

	// Configure transport:
	webapiTransport, err := webapi.New(webapi.Config{
		ListenAddr:      c.WebAPI.ListenAddr,
		AddressBook:     addressBook,
		Topics:          d.Messages,
		AuthorAllowlist: c.WebAPI.Feeds,
		FlushTicker:     timeutil.NewTicker(time.Minute),
		Signer:          key,
		Client:          httpClient,
		Logger:          d.Logger,
		AppName:         d.AppName,
		AppVersion:      d.AppVersion,
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the WebAPI transport: %v", err),
			Subject:  &c.WebAPI.Range,
		}
	}
	return recoverer.New(webapiTransport, d.Logger), nil
}

func (c *Config) configureLibP2P(d Dependencies) (transport.Service, error) {
	// Configure signer:
	key := d.Keys[c.LibP2P.EthereumKey]
	if c.LibP2P.EthereumKey != "" && key == nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Ethereum key %q is not configured", c.LibP2P.EthereumKey),
			Subject:  c.LibP2P.Content.Attributes["ethereum_key"].Range.Ptr(),
		}
	}

	// Configure LibP2P private keys:
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}
	var messagePrivKey crypto.PrivKey
	if key != nil {
		messagePrivKey = ethkey.NewPrivKey(key)
		if !sliceutil.Contains(c.LibP2P.Feeds, key.Address()) {
			c.LibP2P.Feeds = append(c.LibP2P.Feeds, key.Address())
		}
	}

	if !c.LibP2P.DisableFeedFilter {
		if len(c.LibP2P.Feeds) == 0 {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   "At least one feed must be configured",
				Subject:  c.LibP2P.Content.Attributes["feeds"].Range.Ptr(),
			}
		}
	} else if len(c.LibP2P.Feeds) != 0 {
		d.Logger.
			WithField("feeds", c.LibP2P.Feeds).
			Warn("Feeds filter is disabled, the list of feeds will be ignored")
		c.LibP2P.Feeds = nil
	}

	logger := d.Logger.WithField("tag", LoggerTag)
	for _, addr := range c.LibP2P.Feeds {
		logger.
			WithField("address", addr.String()).
			Info("Feed")
	}
	for _, addr := range c.LibP2P.BootstrapAddrs {
		logger.
			WithField("address", addr).
			Info("Bootstrap")
	}

	var extAddr multiaddr.Multiaddr
	if c.LibP2P.ExternalAddr != "" {
		extAddr, err = parseAddr(c.LibP2P.ExternalAddr, c.LibP2P.ListenAddrs)
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Could not determine external address from %s", c.LibP2P.ExternalAddr),
				Subject:  c.LibP2P.Content.Attributes["external_addr"].Range.Ptr(),
			}
		}
		logger.
			WithField("address", extAddr).
			Info("External Ingress")
	}

	// Configure LibP2P transport:
	cfg := libp2p.Config{
		Mode:             libp2p.ClientMode,
		Topics:           d.Messages,
		PeerPrivKey:      peerPrivKey,
		MessagePrivKey:   messagePrivKey,
		ListenAddrs:      c.LibP2P.ListenAddrs,
		ExternalAddr:     extAddr,
		BootstrapAddrs:   c.LibP2P.BootstrapAddrs,
		DirectPeersAddrs: c.LibP2P.DirectPeersAddrs,
		BlockedAddrs:     c.LibP2P.BlockedAddrs,
		AuthorAllowlist:  c.LibP2P.Feeds,
		Discovery:        !c.LibP2P.DisableDiscovery,
		Signer:           key,
		Logger:           d.Logger,
		AppName:          d.AppName,
		AppVersion:       d.AppVersion,
	}
	libP2PTransport, err := libp2p.New(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the LibP2P transport: %v", err),
			Subject:  &c.LibP2P.Range,
		}
	}
	return recoverer.New(libP2PTransport, d.Logger), nil
}

// parseAddr creates an external listening address
// TODO: this would be a good place to ensure that the external does not have a tcp component
// TODO: or use it instead the original one
func parseAddr(addr string, addrs []string) (multiaddr.Multiaddr, error) {
	if len(addrs) == 0 {
		return nil, errors.New("no listen addresses configured")
	}

	listen, err := multiaddr.NewMultiaddr(addrs[0])
	if err != nil {
		return nil, err
	}

	external, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		var ip net.IP
		if err := ip.UnmarshalText([]byte(addr)); err != nil {
			return nil, fmt.Errorf("could not create IP from %s: %w", addr, err)
		}
		external, err = multiaddr.NewComponent("ip4", ip.String())
		if err != nil {
			return nil, fmt.Errorf("could not create addr from %s: %w", ip, err)
		}
	}

	_, port := multiaddr.SplitLast(listen)
	return external.Encapsulate(port), nil
}

func (c *Config) generatePrivKey() (crypto.PrivKey, error) {
	seedReader := rand.Reader
	if len(c.LibP2P.PrivKeySeed) != 0 {
		seed, err := hex.DecodeString(c.LibP2P.PrivKeySeed)
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid privKeySeed value: %v", err),
				Subject:  c.LibP2P.Content.Attributes["priv_key_seed"].Range.Ptr(),
			}
		}
		if len(seed) != ed25519.SeedSize {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   "Invalid privKeySeed value, 32 bytes expected",
				Subject:  c.LibP2P.Content.Attributes["priv_key_seed"].Range.Ptr(),
			}
		}
		seedReader = bytes.NewReader(seed)
	}
	privKey, _, err := crypto.GenerateEd25519Key(seedReader)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to generate LibP2P private key: %v", err),
			Subject:  c.LibP2P.Content.Attributes["priv_key_seed"].Range.Ptr(),
		}
	}
	return privKey, nil
}
