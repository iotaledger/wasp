package sctransaction

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/magiconair/properties/assert"
	"math/rand"
	"testing"
)

const (
	testAddress = "mtNnGt72bZd25v291TjEzw5uTonExip24cAjtB38x4tq"
)

//
//func TestGenData(t *testing.T) {
//	addr1 := address.RandomOfType(address.VersionED25519)
//	t.Logf("addrEC = %s", addr1.String())
//	addr2 := address.RandomOfType(address.VersionBLS)
//	t.Logf("addrBLS = %s", addr2.String())
//	color := RandomColor()
//	t.Logf("color = %s", color.String())
//}

func randomColor() (ret balance.Color) {
	if _, err := rand.Read(ret[:]); err != nil {
		panic(err)
	}
	return
}

func TestTransactionStateBlock1(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)

	txb := NewTransactionBuilder()
	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)

	o1 := valuetransaction.NewOutputID(addr, valuetransaction.RandomID())
	txb.AddInputs(o1)
	bal := balance.New(balance.ColorIOTA, 1)
	txb.AddBalanceToOutput(addr, bal)

	color := randomColor()
	txb.AddStateBlock(NewStateBlockParams{
		Color:      color,
		StateIndex: 42,
	})

	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}

func TestTransactionStateBlock2(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)

	txb := NewTransactionBuilder()
	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)

	o1 := valuetransaction.NewOutputID(addr, valuetransaction.RandomID())
	txb.AddInputs(o1)
	bal := balance.New(balance.ColorIOTA, 1)
	txb.AddBalanceToOutput(addr, bal)

	color := randomColor()

	txb.AddStateBlock(NewStateBlockParams{
		Color:      color,
		StateIndex: 42,
	})

	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}

func TestTransactionRequestBlock(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)

	txb := NewTransactionBuilder()
	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)

	o1 := valuetransaction.NewOutputID(addr, valuetransaction.RandomID())
	txb.AddInputs(o1)
	bal := balance.New(balance.ColorIOTA, 1)
	txb.AddBalanceToOutput(addr, bal)

	reqBlk := NewRequestBlock(addr)
	txb.AddRequestBlock(reqBlk)

	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}

func TestTransactionMultiBlocks(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)

	txb := NewTransactionBuilder()
	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)

	o1 := valuetransaction.NewOutputID(addr, valuetransaction.RandomID())
	txb.AddInputs(o1)
	bal := balance.New(balance.ColorIOTA, 1)
	txb.AddBalanceToOutput(addr, bal)

	color := randomColor()

	txb.AddStateBlock(NewStateBlockParams{
		Color:      color,
		StateIndex: 42,
	})

	reqBlk := NewRequestBlock(addr)
	txb.AddRequestBlock(reqBlk)

	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}
