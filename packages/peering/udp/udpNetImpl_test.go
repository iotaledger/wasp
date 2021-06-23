// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package udp_test

import (
	"testing"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/sign/eddsa"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestUDPPeeringImpl(t *testing.T) {
	var err error
	log := testlogger.NewLogger(t)
	defer log.Sync()

	doneCh := make(chan bool)
	netIDs := []string{"localhost:9017", "localhost:9018", "localhost:9019"}
	nodes := make([]peering.NetworkProvider, len(netIDs))

	keys := make([]ed25519.KeyPair, len(netIDs))
	tnms := make([]peering.TrustedNetworkManager, len(netIDs))
	for i := range keys {
		keys[i] = ed25519.GenerateKeyPair()
		tnms[i] = testutil.NewTrustedNetworkManager()
	}
	for _, tnm := range tnms {
		for i := range netIDs {
			_, err = tnm.TrustPeer(keys[i].PublicKey, netIDs[1])
			require.NoError(t, err)
		}
	}

	nodes[0], err = udp.NewNetworkProvider(netIDs[0], 9017, ed25519.GenerateKeyPair(), tnms[0], log.Named("node0"))
	require.NoError(t, err)
	nodes[1], err = udp.NewNetworkProvider(netIDs[1], 9018, ed25519.GenerateKeyPair(), tnms[1], log.Named("node1"))
	require.NoError(t, err)
	nodes[2], err = udp.NewNetworkProvider(netIDs[2], 9019, ed25519.GenerateKeyPair(), tnms[1], log.Named("node2"))
	require.NoError(t, err)
	for i := range nodes {
		go nodes[i].Run(make(<-chan struct{}))
	}

	n0p2, err := nodes[0].PeerByNetID(netIDs[2])
	require.NoError(t, err)
	n1p1, err := nodes[1].PeerByNetID(netIDs[1])
	require.NoError(t, err)
	n2p0, err := nodes[2].PeerByNetID(netIDs[0])
	require.NoError(t, err)

	nodes[0].Attach(nil, func(recv *peering.RecvEvent) {
		doneCh <- true
	})

	chain1 := peering.RandomPeeringID()
	chain2 := peering.RandomPeeringID()
	n0p2.SendMsg(&peering.PeerMessage{PeeringID: chain1, MsgType: 125})
	n1p1.SendMsg(&peering.PeerMessage{PeeringID: chain1, MsgType: 125})
	n2p0.SendMsg(&peering.PeerMessage{PeeringID: chain2, MsgType: 125})

	<-doneCh
}

// TestHiveKyberInterop checks, if hive keys can be used in kyber.
//	```
//  public, private, _ := hive.GenerateKey()
// 	var kyber eddsa.EdDSA
// 	kyber.UnmarshalBinary(private.Bytes())
// 	kyberSig, _ := kyber.Sign(msg)
// 	hiveSig, _, _ := hive.SignatureFromBytes(kyberSig)
// 	public.VerifySignature(msg, hiveSig)
//	```
func TestHiveKyberInterop(t *testing.T) {
	var err error
	hiveKeyPair := ed25519.GenerateKeyPair()
	kyberSuite := edwards25519.NewBlakeSHA256Ed25519()
	kyberEdDSSA := eddsa.EdDSA{}
	require.NoError(t, kyberEdDSSA.UnmarshalBinary(hiveKeyPair.PrivateKey.Bytes()))
	kyberKeyPair := &key.Pair{
		Public:  kyberEdDSSA.Public,
		Private: kyberEdDSSA.Secret,
	}
	// Check, if pub key can be unmarshalled directly.
	kyberPubUnmarshaled := kyberSuite.Point()
	require.NoError(t, kyberPubUnmarshaled.UnmarshalBinary(hiveKeyPair.PublicKey.Bytes()))
	//
	// Check signatures.
	message := []byte{0, 1, 2}
	//
	// Hive-to-Hive
	hiveSig := hiveKeyPair.PrivateKey.Sign(message)
	require.True(t, hiveKeyPair.PublicKey.VerifySignature(message, hiveSig))
	//
	// Kyber-to-Kyber
	kyberSig, err := schnorr.Sign(kyberSuite, kyberKeyPair.Private, message)
	require.NoError(t, err)
	require.NoError(t, schnorr.Verify(kyberSuite, kyberKeyPair.Public, message, kyberSig))
	require.NoError(t, schnorr.Verify(kyberSuite, kyberPubUnmarshaled, message, kyberSig))
}
