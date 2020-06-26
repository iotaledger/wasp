package txbuilder

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/stretchr/testify/assert"
	"testing"
)

const scAddressStr = "pHoaPehxf811Kg2nCHmkcXc7vjDMnBnBXnksTYXyhzXa"

var scAddress address.Address

func init() {
	var err error
	scAddress, err = address.FromBase58(scAddressStr)
	if err != nil {
		panic(err)
	}
}

func TestBasic(t *testing.T) {
	outs := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.AddOriginStateBlock(sh, &scAddress)
	assert.NoError(t, err)

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))
	assert.True(t, tx.SignaturesValid())

	err = utxodb.AddTransaction(tx.Transaction)
	assert.NoError(t, err)

	outs = utxodb.GetAddressOutputs(scAddress)
	sum := int64(0)
	for _, bals := range outs {
		sum += util.BalanceOfColor(bals, (balance.Color)(tx.ID()))
	}
	assert.Equal(t, int64(1), sum)
}

func TestWithRequest(t *testing.T) {
	outs := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.AddOriginStateBlock(sh, &scAddress)
	assert.NoError(t, err)

	err = txb.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeInit))
	assert.NoError(t, err)

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))
	assert.True(t, tx.SignaturesValid())

	err = utxodb.AddTransaction(tx.Transaction)
	assert.NoError(t, err)

	outs = utxodb.GetAddressOutputs(scAddress)
	sum := int64(0)
	for _, bals := range outs {
		sum += util.BalanceOfColor(bals, (balance.Color)(tx.ID()))
	}
	assert.Equal(t, int64(2), sum)
}

//
//func TestNextState(t *testing.T) {
//	outs := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
//	txb, err := NewFromOutputBalances(outs)
//	assert.NoError(t, err)
//
//	sh := hashing.RandomHash(nil)
//	err = txb.AddOriginStateBlock(sh, &scAddress)
//	assert.NoError(t, err)
//
//	err = txb.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeInit))
//	assert.NoError(t, err)
//
//	tx, err := txb.Build(false)
//	assert.NoError(t, err)
//
//	tx.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))
//	assert.True(t, tx.SignaturesValid())
//
//	err = utxodb.AddTransaction(tx.Transaction)
//	assert.NoError(t, err)
//
//	scColor := (balance.Color)(tx.ID())
//	outs = utxodb.GetAddressOutputs(scAddress)
//	sum := int64(0)
//	for _, bals := range outs{
//		sum += util.BalanceOfColor(bals, scColor)
//	}
//	assert.Equal(t, int64(2), sum)
//
//	outs = utxodb.GetAddressOutputs(scAddress)
//	txb, err = NewFromOutputBalances(outs)
//	assert.NoError(t, err)
//
//	err = txb.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeNOP))
//	assert.NoError(t, err)
//
//	assert.Equal(t, int64(1), txb.GetInputBalance(scColor))
//
//	tx, err = txb.Build(false)
//	assert.NoError(t, err)
//}
