package utxodb

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	u := New()
	genTx := u.GenesisTransaction()
	id, err := genTx.ID()
	require.NoError(t, err)
	require.EqualValues(t, *id, u.GenesisTransactionID())
}

func TestGenesis(t *testing.T) {
	u := New()
	require.EqualValues(t, u.Supply(), u.GetAddressBalance(u.GenesisAddress()))
	u.checkLedgerBalance()
}

func TestRequestFunds(t *testing.T) {
	u := New()
	addr := tpkg.RandEd25519Address()
	_, err := u.RequestFunds(addr)
	require.NoError(t, err)
	require.EqualValues(t, u.Supply()-RequestFundsAmount, u.GetAddressBalance(u.GenesisAddress()))
	require.EqualValues(t, RequestFundsAmount, u.GetAddressBalance(addr))
	u.checkLedgerBalance()
}

func TestAddTransactionFail(t *testing.T) {
	u := New()

	addr := tpkg.RandEd25519Address()
	tx, err := u.RequestFunds(addr)
	require.NoError(t, err)

	err = u.AddTransaction(tx)
	require.Error(t, err)
	u.checkLedgerBalance()
}

func TestGetOutput(t *testing.T) {
	u := New()
	addr := tpkg.RandEd25519Address()
	tx, err := u.RequestFunds(addr)
	require.NoError(t, err)

	txID, err := tx.ID()
	require.NoError(t, err)

	outid0 := iotago.OutputIDFromTransactionIDAndIndex(*txID, 0)
	out0 := u.GetOutput(outid0)
	require.EqualValues(t, RequestFundsAmount, out0.Deposit())

	outid1 := iotago.OutputIDFromTransactionIDAndIndex(*txID, 1)
	out1 := u.GetOutput(outid1)
	require.EqualValues(t, u.Supply()-RequestFundsAmount, out1.Deposit())

	outidFail := iotago.OutputIDFromTransactionIDAndIndex(*txID, 5)
	require.Nil(t, u.GetOutput(outidFail))
}
