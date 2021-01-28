package txbuilder

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	u := utxodb.New()
	ownerSigSheme := signaturescheme.RandBLS()
	ownerAddress := ownerSigSheme.Address()
	scSigSheme := signaturescheme.RandBLS()
	scAddress := scSigSheme.Address()
	_, err := u.RequestFunds(ownerAddress)
	assert.NoError(t, err)

	outs := u.GetAddressOutputs(ownerAddress)
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateSection(sh, &scAddress)
	assert.NoError(t, err)

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	err = u.AddTransaction(tx.Transaction)
	assert.NoError(t, err)

	outs = u.GetAddressOutputs(scAddress)
	sum := int64(0)
	for _, bals := range outs {
		sum += txutil.BalanceOfColor(bals, (balance.Color)(tx.ID()))
	}
	assert.Equal(t, int64(1), sum)
}

func TestWithRequest(t *testing.T) {
	u := utxodb.New()
	ownerSigSheme := signaturescheme.RandBLS()
	ownerAddress := ownerSigSheme.Address()
	scSigSheme := signaturescheme.RandBLS()
	scAddress := scSigSheme.Address()
	_, err := u.RequestFunds(ownerAddress)
	assert.NoError(t, err)

	outs := u.GetAddressOutputs(ownerAddress)
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateSection(sh, &scAddress)
	assert.NoError(t, err)

	err = txb.AddRequestSection(sctransaction.NewRequestSection(0, coretypes.NewContractID(coretypes.ChainID(scAddress), 0), 1))
	assert.NoError(t, err)

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	err = u.AddTransaction(tx.Transaction)
	assert.NoError(t, err)

	outs = u.GetAddressOutputs(scAddress)
	sum := int64(0)
	for _, bals := range outs {
		sum += txutil.BalanceOfColor(bals, (balance.Color)(tx.ID()))
	}
	assert.Equal(t, int64(2), sum)
}

func TestNextState(t *testing.T) {
	u := utxodb.New()
	ownerSigSheme := signaturescheme.RandBLS()
	ownerAddress := ownerSigSheme.Address()
	scSigSheme := signaturescheme.RandBLS()
	scAddress := scSigSheme.Address()
	_, err := u.RequestFunds(ownerAddress)
	assert.NoError(t, err)

	outs := u.GetAddressOutputs(ownerAddress)
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateSection(sh, &scAddress)
	assert.NoError(t, err)

	err = txb.AddRequestSection(sctransaction.NewRequestSection(0, coretypes.NewContractID(coretypes.ChainID(scAddress), 0), 1))
	assert.NoError(t, err)

	err = txb.MoveTokensToAddress(scAddress, balance.ColorIOTA, 5)
	assert.NoError(t, err)

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	err = u.AddTransaction(tx.Transaction)
	assert.NoError(t, err)

	scColor := (balance.Color)(tx.ID())
	reqColor := (balance.Color)(tx.ID())

	outs = u.GetAddressOutputs(scAddress)
	sumScCol := int64(0)
	sumIota := int64(0)
	for _, bals := range outs {
		sumScCol += txutil.BalanceOfColor(bals, scColor)
		sumIota += txutil.BalanceOfColor(bals, balance.ColorIOTA)
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

	err = u.AddTransaction(vtx)
	assert.NoError(t, err)

	outs = u.GetAddressOutputs(scAddress)
	sumScCol = int64(0)
	sumIota = int64(0)
	for _, bals := range outs {
		sumScCol += txutil.BalanceOfColor(bals, scColor)
		sumIota += txutil.BalanceOfColor(bals, balance.ColorIOTA)
	}
	assert.Equal(t, int64(1), sumScCol)
	assert.Equal(t, int64(6), sumIota)

	txb, err = NewFromOutputBalances(outs)
	assert.NoError(t, err)

	err = txb.AddRequestSection(sctransaction.NewRequestSection(0, coretypes.NewContractID(coretypes.ChainID(scAddress), 0), 1))
	assert.NoError(t, err)

	tx, err = txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(scSigSheme)
	assert.True(t, tx.SignaturesValid())

	reqColor = (balance.Color)(tx.ID())

	err = u.AddTransaction(tx.Transaction)
	assert.NoError(t, err)

	outs = u.GetAddressOutputs(scAddress)
	sumScCol = int64(0)
	sumIota = int64(0)
	sumReq := int64(0)
	for _, bals := range outs {
		sumScCol += txutil.BalanceOfColor(bals, scColor)
		sumIota += txutil.BalanceOfColor(bals, balance.ColorIOTA)
		sumReq += txutil.BalanceOfColor(bals, reqColor)
	}
	assert.Equal(t, int64(1), sumScCol)
	assert.Equal(t, int64(5), sumIota)
	assert.Equal(t, int64(1), sumReq)
}

func TestClone(t *testing.T) {
	u := utxodb.New()
	ownerSigSheme := signaturescheme.RandBLS()
	ownerAddress := ownerSigSheme.Address()
	scSigSheme := signaturescheme.RandBLS()
	scAddress := scSigSheme.Address()
	_, err := u.RequestFunds(ownerAddress)
	assert.NoError(t, err)

	outs := u.GetAddressOutputs(ownerAddress)
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateSection(sh, &scAddress)
	assert.NoError(t, err)

	err = txb.AddRequestSection(sctransaction.NewRequestSection(0, coretypes.NewContractID(coretypes.ChainID(scAddress), 0), 1))
	assert.NoError(t, err)

	txbClone := txb.Clone()

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	txClone, err := txbClone.Build(false)
	assert.NoError(t, err)

	txClone.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	assert.EqualValues(t, tx.ID(), txClone.ID())
}

func TestDeterminism(t *testing.T) {
	u := utxodb.New()
	ownerSigSheme := signaturescheme.RandBLS()
	ownerAddress := ownerSigSheme.Address()
	scSigSheme := signaturescheme.RandBLS()
	scAddress := scSigSheme.Address()
	_, err := u.RequestFunds(ownerAddress)
	assert.NoError(t, err)

	outs := u.GetAddressOutputs(ownerAddress)
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	sh := hashing.RandomHash(nil)
	err = txb.CreateOriginStateSection(sh, &scAddress)
	assert.NoError(t, err)

	err = txb.MoveTokensToAddress(scAddress, balance.ColorIOTA, 50)
	assert.NoError(t, err)

	err = txb.AddRequestSection(sctransaction.NewRequestSection(0, coretypes.NewContractID(coretypes.ChainID(scAddress), 0), 1))
	assert.NoError(t, err)

	txbClone := txb.Clone()

	err = txb.MoveTokensToAddress(scAddress, balance.ColorIOTA, 50)
	assert.NoError(t, err)

	err = txb.AddRequestSection(sctransaction.NewRequestSection(0, coretypes.NewContractID(coretypes.ChainID(scAddress), 0), 1))
	assert.NoError(t, err)

	err = txbClone.AddRequestSection(sctransaction.NewRequestSection(0, coretypes.NewContractID(coretypes.ChainID(scAddress), 0), 1))
	assert.NoError(t, err)

	err = txbClone.MoveTokensToAddress(scAddress, balance.ColorIOTA, 50)
	assert.NoError(t, err)

	tx, err := txb.Build(false)
	assert.NoError(t, err)

	tx.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	txClone, err := txbClone.Build(false)
	assert.NoError(t, err)

	txClone.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	assert.EqualValues(t, tx.ID(), txClone.ID())
}
