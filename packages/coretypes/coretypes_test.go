// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChainID(t *testing.T) {
	chid := NewRandomChainID()

	chid58 := chid.Base58()
	t.Logf("chid58 = %s", chid58)

	chidString := chid.String()
	t.Logf("chidString = %s", chidString)

	chidback, err := NewChainIDFromBytes(chid.Bytes())
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)

	chidback, err = NewChainIDFromBase58(chid58)
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)
}

func TestContractIDHname(t *testing.T) {
	chid := NewRandomChainID()
	name := "test contract name"
	hname := Hn(name)
	scid := NewContractID(chid, hname)
	scid58 := scid.Base58()
	t.Logf("scname = '%s' schname = %s scid58 = %s", name, hname, scid58)
	scidstr := scid.String()
	t.Logf("scidstr = %s", scidstr)
	t.Logf("scid short = %s", scid.Short())

	scidBack, err := NewContractIDFromBytes(scid.Bytes())
	assert.NoError(t, err)
	assert.EqualValues(t, scidBack, scid)

	scidBack, err = NewContractIDFromBase58(scid58)
	assert.NoError(t, err)
	assert.EqualValues(t, scidBack, scid)

	ep := Hn("dummyFunction")
	epbytes := ep.Bytes()
	epstr := ep.String()

	t.Logf("epstr = %s", epstr)

	epback, err := NewHnameFromBytes(epbytes)
	assert.NoError(t, err)
	assert.EqualValues(t, ep, epback)
}

func TestAgentID(t *testing.T) {
	chid := NewRandomChainID()

	chid58 := chid.String()
	t.Logf("chid58 = %s", chid58)

	kp := ed25519.GenerateKeyPair()
	addr := ledgerstate.NewED25519Address(kp.PublicKey)

	aid := NewAgentIDFromAddress(chid.AsAddress())
	require.False(t, aid.IsContract())
	t.Logf("agent ID 1: %s", aid.String())
	aidBack, err := NewAgentIDFromBytes(aid.Bytes())
	require.NoError(t, err)
	require.EqualValues(t, aid.Bytes(), aidBack.Bytes())

	aid = NewAgentIDFromAddress(addr)
	require.False(t, aid.IsContract())
	t.Logf("agent ID 2: %s", aid.String())
	aidBack, err = NewAgentIDFromBytes(aid.Bytes())
	require.NoError(t, err)
	require.EqualValues(t, aid.Bytes(), aidBack.Bytes())

	cid := NewContractID(chid, Hn("dummy"))
	aid = NewAgentIDFromContractID(cid)
	require.True(t, aid.IsContract())
	t.Logf("agent ID 3: %s", aid.String())
	aidBack, err = NewAgentIDFromBytes(aid.Bytes())
	require.NoError(t, err)
	require.EqualValues(t, aid.Bytes(), aidBack.Bytes())

	cid = NewContractID(chid, 0)
	aid = NewAgentIDFromContractID(cid)
	require.False(t, aid.IsContract())
	t.Logf("agent ID 4: %s", aid.String())
	aidBack, err = NewAgentIDFromBytes(aid.Bytes())
	require.NoError(t, err)
	require.EqualValues(t, aid.Bytes(), aidBack.Bytes())
}

func TestHname(t *testing.T) {
	hn1 := Hn("first")

	hn1bytes := hn1.Bytes()
	hn1back, err := NewHnameFromBytes(hn1bytes)
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
