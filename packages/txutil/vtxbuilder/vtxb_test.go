package vtxbuilder

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	u := utxodb.New()

	ownerSigSheme := signaturescheme.RandBLS()
	ownerAddress := ownerSigSheme.Address()
	u.RequestFunds(ownerAddress)

	targetSigSheme := signaturescheme.RandBLS()
	targetAddress := targetSigSheme.Address()

	outs := u.GetAddressOutputs(ownerAddress)
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	err = txb.MoveTokensToAddress(targetAddress, balance.ColorIOTA, 1)
	assert.NoError(t, err)

	tx := txb.Build(false)
	tx.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	err = u.AddTransaction(tx)
	assert.NoError(t, err)
}

func TestColor(t *testing.T) {
	u := utxodb.New()

	ownerSigSheme := signaturescheme.RandBLS()
	ownerAddress := ownerSigSheme.Address()
	u.RequestFunds(ownerAddress)

	targetSigSheme := signaturescheme.RandBLS()
	targetAddress := targetSigSheme.Address()

	outs := u.GetAddressOutputs(ownerAddress)
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	err = txb.MintColoredTokens(targetAddress, balance.ColorIOTA, 10)
	assert.NoError(t, err)

	tx := txb.Build(false)
	tx.Sign(ownerSigSheme)
	assert.True(t, tx.SignaturesValid())

	err = u.AddTransaction(tx)
	assert.NoError(t, err)

	outs1 := u.GetAddressOutputs(targetAddress)
	txb1, err := NewFromOutputBalances(outs1)
	assert.NoError(t, err)

	color := (balance.Color)(tx.ID())
	assert.Equal(t, txb1.GetInputBalance(color), int64(10))

	err = txb1.EraseColor(targetAddress, color, 5)
	assert.NoError(t, err)

	tx1 := txb1.Build(true)
	tx1.Sign(targetSigSheme)

	err = u.AddTransaction(tx1)
	assert.NoError(t, err)

	outs2 := u.GetAddressOutputs(targetAddress)
	txb2, err := NewFromOutputBalances(outs2)
	assert.NoError(t, err)

	assert.Equal(t, txb2.GetInputBalance(color), int64(5))
}
