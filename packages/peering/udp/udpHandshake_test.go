// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package udp

import (
	"net"
	"testing"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/stretchr/testify/require"
)

func TestHandshakeCodec(t *testing.T) {
	var err error
	pair := ed25519.GenerateKeyPair()
	a := handshakeMsg{
		netID:   "some",
		pubKey:  pair.PublicKey,
		respond: true,
	}
	var buf []byte
	buf, err = a.bytes(pair.PrivateKey)
	require.Nil(t, err)
	require.NotNil(t, buf)
	//
	// Correct message.
	var b *handshakeMsg
	b, err = handshakeMsgFromBytes(buf)
	require.Nil(t, err)
	require.NotNil(t, b)
	require.Equal(t, a.netID, b.netID)
	require.Equal(t, a.pubKey, b.pubKey)
	require.Equal(t, a.respond, b.respond)
	//
	// Damaged message.
	buf[2]++
	var c *handshakeMsg
	c, err = handshakeMsgFromBytes(buf)
	require.NotNil(t, err)
	require.Nil(t, c)
}

func TestUDPAddrString(t *testing.T) {
	var err error
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr("udp", "localhost:1248")
	require.Nil(t, err)
	require.Equal(t, "127.0.0.1:1248", addr.String())
}
