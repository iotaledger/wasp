package vm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxBuilderEmpty(t *testing.T) {
	_, err := NewTxBuilder(TransactionBuilderParams{
		Balances:   nil,
		OwnColor:   balance.Color{},
		OwnAddress: address.Address{},
	})
	assert.Equal(t, false, err == nil)
}

func TestTxBuilder(t *testing.T) {
	color := util.RandomColor()
	colorWrong := util.RandomColor()
	addr := address.RandomOfType(address.VersionBLS)
	txid := util.RandomTransactionID()

	balances := map[valuetransaction.ID][]*balance.Balance{
		txid: {balance.New(color, 1)},
	}
	_, err := NewTxBuilder(TransactionBuilderParams{
		Balances:   balances,
		OwnColor:   color,
		OwnAddress: addr,
	})
	assert.NoError(t, err)

	_, err = NewTxBuilder(TransactionBuilderParams{
		Balances:   balances,
		OwnColor:   colorWrong,
		OwnAddress: addr,
	})
	assert.Equal(t, err == nil, false)
}

func TestTxBuilder2(t *testing.T) {
	color := util.RandomColor()
	addr := address.RandomOfType(address.VersionBLS)
	txid1 := util.RandomTransactionID()
	txid2 := util.RandomTransactionID()

	balances := map[valuetransaction.ID][]*balance.Balance{
		txid1: {balance.New(color, 1)},
		txid2: {balance.New(balance.ColorIOTA, 100)},
	}
	txb, err := NewTxBuilder(TransactionBuilderParams{
		Balances:   balances,
		OwnColor:   color,
		OwnAddress: addr,
	})
	assert.NoError(t, err)

	err = txb.MoveTokens(addr, color, 1)
	assert.Equal(t, err == nil, false)

	b, ok := txb.GetInputBalance(color)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(0))

	b, ok = txb.GetInputBalance(balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(100))

	err = txb.MoveTokens(addr, balance.ColorIOTA, 99)
	assert.NoError(t, err)

	b, ok = txb.GetInputBalance(balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(1))

	b, ok = txb.GetOutputBalance(addr, balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(99))

	err = txb.MoveTokens(addr, balance.ColorIOTA, 5)
	assert.Equal(t, err == nil, false)

	b, ok = txb.GetInputBalance(balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(1))

	b, ok = txb.GetOutputBalance(addr, balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(99))

	err = txb.MoveTokens(addr, balance.ColorIOTA, 1)
	assert.NoError(t, err)

	err = txb.MoveTokens(addr, balance.ColorIOTA, 1)
	assert.Equal(t, err == nil, false)
}

func TestTxBuilder3(t *testing.T) {
	color := util.RandomColor()
	addr := address.RandomOfType(address.VersionBLS)
	txid1 := util.RandomTransactionID()
	txid2 := util.RandomTransactionID()

	balances := map[valuetransaction.ID][]*balance.Balance{
		txid1: {balance.New(color, 1)},
		txid2: {balance.New(balance.ColorIOTA, 100)},
	}
	txb, err := NewTxBuilder(TransactionBuilderParams{
		Balances:   balances,
		OwnColor:   color,
		OwnAddress: addr,
	})
	assert.NoError(t, err)

	err = txb.MoveTokens(addr, color, 1)
	assert.Equal(t, err == nil, false)

	b, ok := txb.GetInputBalance(color)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(0))

	b, ok = txb.GetInputBalance(balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(100))

	err = txb.MoveTokens(addr, balance.ColorIOTA, 95)
	assert.NoError(t, err)

	b, ok = txb.GetInputBalance(balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(5))

	b, ok = txb.GetOutputBalance(addr, balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(95))

	err = txb.MoveTokens(addr, balance.ColorIOTA, 6)
	assert.Equal(t, err == nil, false)

	b, ok = txb.GetInputBalance(balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(5))

	b, ok = txb.GetOutputBalance(addr, balance.ColorIOTA)
	assert.Equal(t, ok, true)
	assert.Equal(t, b, int64(95))

	tx := txb.Finalize(42, *hashing.NilHash, 5000)
	sidx := tx.MustState().StateIndex()
	assert.Equal(t, sidx, uint32(42))

	scol := tx.MustState().Color()
	assert.EqualValues(t, scol, color)

	ts := tx.MustState().Timestamp()
	assert.EqualValues(t, ts, int64(5000))
}
