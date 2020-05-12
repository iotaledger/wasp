package sctransaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/magiconair/properties/assert"
	"testing"
	"time"
)

func TestBasicScId(t *testing.T) {
	addr := address.RandomOfType(address.VersionBLS)
	color := RandomColor()
	scid := NewScId(color, addr)

	scidstr := scid.String()
	scid1, err := ScIdFromString(scidstr)
	assert.Equal(t, err, nil)
	assert.Equal(t, scidstr, scid1.String())

	assert.Equal(t, bytes.Equal(scid.Address().Bytes(), addr[:]), true)
	assert.Equal(t, bytes.Equal(scid.Color().Bytes(), color[:]), true)
}

const (
	testAddress = "mtNnGt72bZd25v291TjEzw5uTonExip24cAjtB38x4tq"
	testColor   = "3MrmupSNH8gPH2ZcEiLPno5dTNycgAE1qDs4cgbzgMLm"
	testScid    = "46Smwtm1jH2hQ4gEYb7skX1EBMQSi1oLvcDpUwEudB6qJV96GpWjucD398R3s3UJ6kgrsZWhHw6FiSCMiGZ47QsA9"
)

//
//func TestGenData(t *testing.T){
//	addr := address.RandomOfType(address.VERSION_BLS)
//	t.Logf("addr = %s", addr.String())
//	color := RandomColor()
//	t.Logf("color = %s", color.String())
//	scid := NewScId(color, addr)
//	t.Logf("scid = %s", scid.String())
//}

func TestScid(t *testing.T) {
	scid, err := ScIdFromString(testScid)
	assert.Equal(t, err, nil)
	addr := scid.Address()
	assert.Equal(t, addr.Version(), address.VersionBLS)
	color, err := ColorFromString(testColor)
	assert.Equal(t, err, nil)

	assert.Equal(t, color, scid.Color())
	assert.Equal(t, addr, scid.Address())

	scidBack := NewScId(color, addr).String()
	assert.Equal(t, scidBack, testScid)
}

func TestRandScid(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)
	assert.Equal(t, addr.Version(), address.VersionBLS)

	scid := RandomScId(addr)
	a := scid.Address().Bytes()
	assert.Equal(t, bytes.Equal(a, addr[:]), true)

	scid1, err := ScIdFromString(scid.String())
	assert.Equal(t, err, nil)
	assert.Equal(t, scid.Equal(scid1), true)
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

	color := RandomColor()
	txb.AddStateBlock(&color, 42)

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

	color := RandomColor()

	txb.AddStateBlock(&color, 42)
	txb.SetTimestamp(time.Now().UnixNano())
	txb.SetVariableStateHash(hashing.RandomHash(nil))

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

	reqBlk := NewRequestBlock(&addr)
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

	color := RandomColor()

	txb.AddStateBlock(&color, 42)
	txb.SetTimestamp(time.Now().UnixNano())
	txb.SetVariableStateHash(hashing.RandomHash(nil))

	reqBlk := NewRequestBlock(&addr)
	txb.AddRequestBlock(reqBlk)

	_, err = txb.Finalize()
	assert.Equal(t, err, nil)

	_, err = txb.Finalize()
	assert.Equal(t, err != nil, true)
}
