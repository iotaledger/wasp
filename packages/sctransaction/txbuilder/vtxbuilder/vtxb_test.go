package vtxbuilder

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	outs := utxodb.GetAddressOutputs(utxodb.GetGenesisAddress())
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	err = txb.MoveToAddress(utxodb.GetAddress(1), balance.ColorIOTA, 1)
	assert.NoError(t, err)

	tx := txb.Build(false)
	tx.Sign(utxodb.GetGenesisSigScheme())
	assert.True(t, tx.SignaturesValid())

	err = utxodb.AddTransaction(tx)
	assert.NoError(t, err)
}

func TestColor(t *testing.T) {
	outs := utxodb.GetAddressOutputs(utxodb.GetGenesisAddress())
	txb, err := NewFromOutputBalances(outs)
	assert.NoError(t, err)

	err = txb.MintColor(utxodb.GetAddress(1), balance.ColorIOTA, 10)
	assert.NoError(t, err)

	tx := txb.Build(false)
	tx.Sign(utxodb.GetGenesisSigScheme())
	assert.True(t, tx.SignaturesValid())

	err = utxodb.AddTransaction(tx)
	assert.NoError(t, err)

	outs1 := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb1, err := NewFromOutputBalances(outs1)
	assert.NoError(t, err)

	color := (balance.Color)(tx.ID())
	assert.Equal(t, txb1.GetInputBalance(color), int64(10))

	err = txb1.EraseColor(utxodb.GetAddress(1), color, 5)
	assert.NoError(t, err)

	tx1 := txb1.Build(true)
	tx1.Sign(utxodb.GetSigScheme(utxodb.GetAddress(1)))

	err = utxodb.AddTransaction(tx1)
	assert.NoError(t, err)

	outs2 := utxodb.GetAddressOutputs(utxodb.GetAddress(1))
	txb2, err := NewFromOutputBalances(outs2)
	assert.NoError(t, err)

	assert.Equal(t, txb2.GetInputBalance(color), int64(5))
}
