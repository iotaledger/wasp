// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO fix - broken because of crypto stuff
// func TestAgentID(t *testing.T) {
// 	aid := NewRandomAgentID()

// 	t.Logf("random AgentID = %s", aid.String())

// 	kp := cryptolib.NewKeyPair()
// 	addr := cryptolib.Ed25519AddressFromPubKey(kp.PublicKey)

// 	hname := Hn("dummy")
// 	aid = NewAgentID(addr, hname)

// 	t.Logf("agent ID string: %s", aid.String())
// 	t.Logf("agent ID base58: %s", aid.Base58())

// 	aidBack, err := AgentIDFromBytes(aid.Bytes())
// 	require.NoError(t, err)
// 	require.True(t, aid.Equals(aidBack))

// 	aid = NewAgentID(addr, hname)
// 	require.True(t, addr.Equals(aid.Address()))
// 	require.EqualValues(t, aid.Hname(), hname)

// 	aidBack, err = NewAgentIDFromBase58EncodedString(aid.Base58())
// 	require.NoError(t, err)
// 	require.True(t, aid.Equals(aidBack))

// 	aidBack, err = NewAgentIDFromString(aid.String())
// 	require.NoError(t, err)
// 	require.True(t, aid.Equals(aidBack))
// }

func TestHname(t *testing.T) {
	hn1 := Hn("first")

	hn1bytes := hn1.Bytes()
	hn1back, err := HnameFromBytes(hn1bytes)
	require.NoError(t, err)
	require.EqualValues(t, hn1, hn1back)

	s := hn1.String()
	hn1back, err = HnameFromString(s)
	require.NoError(t, err)
	require.EqualValues(t, hn1, hn1back)
}

func TestHnameCollision(t *testing.T) {
	hn1 := Hn("doNothing")
	hn2 := Hn("incCounter")

	require.NotEqualValues(t, hn1, hn2)
}
