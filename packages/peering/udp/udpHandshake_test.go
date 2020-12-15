// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package udp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestHandshakeCodec(t *testing.T) {
	var err error
	suite := pairing.NewSuiteBn256()
	pair := key.NewKeyPair(suite)
	a := handshakeMsg{
		netID:   "some",
		pubKey:  pair.Public,
		respond: true,
	}
	var buf []byte
	buf, err = a.bytes(pair.Private, suite)
	require.Nil(t, err)
	require.NotNil(t, buf)
	//
	// Correct message.
	var b *handshakeMsg
	b, err = handshakeMsgFromBytes(buf, suite)
	require.Nil(t, err)
	require.NotNil(t, b)
	require.Equal(t, a.netID, b.netID)
	require.True(t, a.pubKey.Equal(b.pubKey))
	require.Equal(t, a.respond, b.respond)
	//
	// Damaged message.
	buf[2] = buf[2] + 1
	var c *handshakeMsg
	c, err = handshakeMsgFromBytes(buf, suite)
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
