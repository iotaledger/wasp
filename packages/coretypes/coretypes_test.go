// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBase(t *testing.T) {
	chid := (ChainID)(address.Random())

	chid58 := chid.String()
	t.Logf("chid58 = %s", chid58)

	chidback, err := NewChainIDFromBytes(chid[:])
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)

	chidback, err = NewChainIDFromBase58(chid58)
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)

	name := "test contract name"
	hname := Hn(name)
	scid := NewContractID(chid, hname)
	scid58 := scid.Base58()
	t.Logf("scname = '%s' schname = %s scid58 = %s", name, hname, scid58)
	scidstr := scid.String()
	t.Logf("scidstr = %s", scidstr)

	t.Logf("scid short = %s", scid.Short())

	scidBack, err := NewContractIDFromBytes(scid[:])
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

func TestRequestID(t *testing.T) {
	txid := (valuetransaction.ID)(hashing.RandomHash(nil))
	reqid := NewRequestID(txid, 3)

	t.Logf("txid = %s", txid.String())
	t.Logf("reqid = %s", reqid.String())
	t.Logf("reqidShort = %s", reqid.Short())

	reqidback, err := NewRequestIDFromBytes(reqid.Bytes())
	assert.NoError(t, err)
	assert.EqualValues(t, reqid, reqidback)

	reqid58 := reqid.Base58()
	t.Logf("reqid58 = %s", reqid58)
	reqidback, err = NewRequestIDFromBase58(reqid58)
	assert.NoError(t, err)
	assert.EqualValues(t, reqid, reqidback)
}

func TestAgentID(t *testing.T) {
	chid := (ChainID)(address.Random())

	chid58 := chid.String()
	t.Logf("chid58 = %s", chid58)

	addr := address.Random()
	t.Logf("addr = %s", addr.String())

	aid := NewAgentIDFromAddress(addr)
	require.True(t, aid.IsAddress())

	contrId := NewContractID(chid, Hn("22"))
	aid1 := NewAgentIDFromContractID(contrId)
	require.True(t, !aid1.IsAddress())

	contrIdBack := aid1.MustContractID()
	require.EqualValues(t, contrId, contrIdBack)

	t.Logf("addr agent ID = %s", aid.String())
	t.Logf("contract agent ID = %s", aid1.String())

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
