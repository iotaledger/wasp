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

	scid := NewContractID(chid, 5)
	scid58 := scid.Base58()
	t.Logf("scid58 = %s", scid58)
	scidstr := scid.String()
	t.Logf("scidstr = %s", scidstr)

	t.Logf("scid short = %s", scid.Short())

	scidBack, err := NewContractIDFromBytes(scid[:])
	assert.NoError(t, err)
	assert.EqualValues(t, scidBack, scid)

	scidBack, err = NewContractIDFromBase58(scid58)
	assert.NoError(t, err)
	assert.EqualValues(t, scidBack, scid)

	ep := NewEntryPointCodeFromFunctionName("dummyFunction")
	epbytes := ep.Bytes()
	epstr := ep.String()

	t.Logf("epstr = %s", epstr)

	epback, err := NewEntryPointCodeFromBytes(epbytes)
	assert.NoError(t, err)
	assert.EqualValues(t, ep, epback)
}

func TestRequestID(t *testing.T) {
	txid := (valuetransaction.ID)(*hashing.RandomHash(nil))
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

	contrId := NewContractID(chid, 22)
	aid1 := NewAgentIDFromContractID(contrId)
	require.True(t, !aid1.IsAddress())

	contrIdBack := aid1.MustContractID()
	require.EqualValues(t, contrId, contrIdBack)
}
