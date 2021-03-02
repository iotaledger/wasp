package ipfs

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/iotaledger/hive.go/logger"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/multiformats/go-multiaddr"
)

var (
	ipfsNode ipfslite.Peer // IPFS node to send and receive data

	ctx *context.Context // Context for IPFS node

	log *logger.Logger // Logger of IPFS functionality
)

// Start the ipfs node
func Start(inLogger *logger.Logger, inContext *context.Context) error {
	log = inLogger
	ctx = inContext

	// Bootstrappers are using 1024 keys. See:
	// https://github.com/ipfs/infra/issues/378
	crypto.MinRsaKeyBits = 1024

	var ds datastore.Batching
	var err error
	ds, err = ipfslite.BadgerDatastore("test")
	if err != nil {
		return err
	}

	var privKey crypto.PrivKey
	privKey, _, err = crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return err
	}

	var listen multiaddr.Multiaddr
	listen, err = multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005")

	var h host.Host
	var dht *dualdht.DHT
	h, dht, err = ipfslite.SetupLibp2p(
		*inContext,
		privKey,
		nil,
		[]multiaddr.Multiaddr{listen},
		ds,
		ipfslite.Libp2pOptionsExtra...,
	)
	if err != nil {
		return err
	}

	ipfsNode, err := ipfslite.New(*inContext, ds, h, dht, nil)
	if err != nil {
		return err
	}

	ipfsNode.Bootstrap(ipfslite.DefaultBootstrapPeers())

	return nil
}

// Download information from IPFS by its cid
func Download(inCid string) ([]byte, error) {
	var c cid.Cid
	var err error
	var iorc io.ReadCloser
	var result []byte

	c, _ = cid.Decode(inCid)
	iorc, err = ipfsNode.GetFile(*ctx, c)
	if err != nil {
		return nil, err
	}
	defer iorc.Close()
	result, err = ioutil.ReadAll(iorc)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Upload data to IPFS
func Upload(data []byte) (cid string, e error) {
	var ior io.Reader = bytes.NewReader(data)

	var err error
	var result ipld.Node
	result, err = ipfsNode.AddFile(*ctx, ior, nil)
	if err != nil {
		return "", err
	}

	return result.Cid().String(), nil
}
