package coretypes

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/stretchr/testify/assert"
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
	ep58 := ep.String()

	t.Logf("ep58 = %s", ep58)

	epback, err := NewEntryPointCodeFromBytes(ep[:])
	assert.NoError(t, err)
	assert.EqualValues(t, ep, epback)

	epuint := ep.Uint32()
	epback = NewEntryPointCodeFromUint32(epuint)
	assert.EqualValues(t, ep, epback)

	epback, err = NewEntryPointCodeFromBase58(ep58)
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
