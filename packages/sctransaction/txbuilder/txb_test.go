package txbuilder

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/stretchr/testify/assert"
	"testing"
)

const scAddressStr = "pHoaPehxf811Kg2nCHmkcXc7vjDMnBnBXnksTYXyhzXa"

var (
	scSigSheme = signaturescheme.RandBLS()
	scAddress  = scSigSheme.Address()
)

func initUtxodb() {
	utxodb.Init()
	scSigSheme = signaturescheme.RandBLS()
	scAddress = scSigSheme.Address()
}

func TestBasic(t *testing.T) {
	initUtxodb()

	outs := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateBlock(sh, &scAddress)
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
	initUtxodb()

	outs := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateBlock(sh, &scAddress)
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

func TestNextState(t *testing.T) {
	initUtxodb()

	outs := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateBlock(sh, &scAddress)
	assert.NoError(t, err)

	err = txb.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeInit))
	assert.NoError(t, err)

	err = txb.MoveToAddress(scAddress, balance.ColorIOTA, 5)
	assert.NoError(t, err)

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))
	assert.True(t, tx.SignaturesValid())

	err = utxodb.AddTransaction(tx.Transaction)
	assert.NoError(t, err)

	scColor := (balance.Color)(tx.ID())
	reqColor := (balance.Color)(tx.ID())

	outs = utxodb.GetAddressOutputs(scAddress)
	sumScCol := int64(0)
	sumIota := int64(0)
	for _, bals := range outs {
		sumScCol += util.BalanceOfColor(bals, scColor)
		sumIota += util.BalanceOfColor(bals, balance.ColorIOTA)
	}
	assert.Equal(t, int64(2), sumScCol)
	assert.Equal(t, int64(5), sumIota)

	txb, err = NewFromOutputBalances(outs)
	assert.NoError(t, err)

	err = txb.EraseColor(scAddress, reqColor, 1)
	assert.NoError(t, err)

	vtx := txb.BuildValueTransactionOnly(false)
	vtx.Sign(scSigSheme)
	assert.True(t, vtx.SignaturesValid())

	err = utxodb.AddTransaction(vtx)
	assert.NoError(t, err)

	outs = utxodb.GetAddressOutputs(scAddress)
	sumScCol = int64(0)
	sumIota = int64(0)
	for _, bals := range outs {
		sumScCol += util.BalanceOfColor(bals, scColor)
		sumIota += util.BalanceOfColor(bals, balance.ColorIOTA)
	}
	assert.Equal(t, int64(1), sumScCol)
	assert.Equal(t, int64(6), sumIota)

	txb, err = NewFromOutputBalances(outs)
	assert.NoError(t, err)

	err = txb.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeNOP))
	assert.NoError(t, err)

	tx, err = txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(scSigSheme)
	assert.True(t, tx.SignaturesValid())

	reqColor = (balance.Color)(tx.ID())

	err = utxodb.AddTransaction(tx.Transaction)
	assert.NoError(t, err)

	outs = utxodb.GetAddressOutputs(scAddress)
	sumScCol = int64(0)
	sumIota = int64(0)
	sumReq := int64(0)
	for _, bals := range outs {
		sumScCol += util.BalanceOfColor(bals, scColor)
		sumIota += util.BalanceOfColor(bals, balance.ColorIOTA)
		sumReq += util.BalanceOfColor(bals, reqColor)
	}
	assert.Equal(t, int64(1), sumScCol)
	assert.Equal(t, int64(5), sumIota)
	assert.Equal(t, int64(1), sumReq)
}

func TestClone(t *testing.T) {
	initUtxodb()

	outs := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateBlock(sh, &scAddress)
	assert.NoError(t, err)

	err = txb.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeInit))
	assert.NoError(t, err)

	txbClone := txb.Clone()

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))
	assert.True(t, tx.SignaturesValid())

	txClone, err := txbClone.Build(false)
	assert.NoError(t, err)

	txClone.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))
	assert.True(t, tx.SignaturesValid())

	assert.EqualValues(t, tx.ID(), txClone.ID())
}

func TestDeterminism(t *testing.T) {
	initUtxodb()

	outs := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateBlock(sh, &scAddress)
	assert.NoError(t, err)

	err = txb.MoveToAddress(scAddress, balance.ColorIOTA, 50)

	err = txb.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeInit))
	assert.NoError(t, err)

	txbClone := txb.Clone()

	err = txb.MoveToAddress(scAddress, balance.ColorIOTA, 50)
	assert.NoError(t, err)

	err = txb.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeNOP))
	assert.NoError(t, err)

	err = txbClone.AddRequestBlock(sctransaction.NewRequestBlock(scAddress, vmconst.RequestCodeNOP))
	assert.NoError(t, err)

	err = txbClone.MoveToAddress(scAddress, balance.ColorIOTA, 50)
	assert.NoError(t, err)

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))
	assert.True(t, tx.SignaturesValid())

	txClone, err := txbClone.Build(false)
	assert.NoError(t, err)

	txClone.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))
	assert.True(t, tx.SignaturesValid())

	assert.EqualValues(t, tx.ID(), txClone.ID())
}
