package utxodb

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	u := New()
	genTx := u.GenesisTransaction()
	id, err := genTx.ID()
	require.NoError(t, err)
	require.EqualValues(t, *id, u.GenesisTransactionID())
	require.EqualValues(t, u.Supply(), u.GetAddressBalance(u.GenesisAddress()))

	require.Same(t, genTx, u.MustGetTransaction(u.GenesisTransactionID()))

	m, ok := u.GetTransactionMilestoneInfo(u.GenesisTransactionID())
	require.True(t, ok)
	require.Equal(t, MilestoneInfo{Index: 0, Timestamp: 0}, m)
}

func TestRequestFunds(t *testing.T) {
	u := New()
	addr := tpkg.RandEd25519Address()
	tx, err := u.RequestFunds(addr)
	require.NoError(t, err)
	require.EqualValues(t, u.Supply()-RequestFundsAmount, u.GetAddressBalance(u.GenesisAddress()))
	require.EqualValues(t, RequestFundsAmount, u.GetAddressBalance(addr))

	txID, err := tx.ID()
	require.NoError(t, err)
	require.Same(t, tx, u.MustGetTransaction(*txID))

	m, ok := u.GetTransactionMilestoneInfo(*txID)
	require.True(t, ok)
	require.Equal(t, MilestoneInfo{Index: 1, Timestamp: 1}, m)
}

func TestAddTransactionFail(t *testing.T) {
	u := New()

	addr := tpkg.RandEd25519Address()
	tx, err := u.RequestFunds(addr)
	require.NoError(t, err)

	err = u.AddTransaction(tx)
	require.Error(t, err)
}

func TestDoubleSpend(t *testing.T) {
	key1 := tpkg.RandEd25519PrivateKey()
	addr1 := iotago.Ed25519AddressFromPubKey(key1.Public().(ed25519.PublicKey))
	key1Signer := iotago.NewInMemoryAddressSigner(iotago.NewAddressKeysForEd25519Address(&addr1, key1))

	addr2 := tpkg.RandEd25519Address()
	addr3 := tpkg.RandEd25519Address()

	u := New()

	tx1, err := u.RequestFunds(&addr1)
	require.NoError(t, err)
	tx1ID, err := tx1.ID()
	require.NoError(t, err)

	spend2, err := iotago.NewTransactionBuilder().
		AddInput(&iotago.ToBeSignedUTXOInput{Address: &addr1, Input: &iotago.UTXOInput{
			TransactionID:          *tx1ID,
			TransactionOutputIndex: 0,
		}}).
		AddOutput(&iotago.ExtendedOutput{Address: addr2, Amount: RequestFundsAmount}).
		Build(u.deSeriParas(), key1Signer)
	require.NoError(t, err)
	err = u.AddTransaction(spend2)
	require.NoError(t, err)

	spend3, err := iotago.NewTransactionBuilder().
		AddInput(&iotago.ToBeSignedUTXOInput{Address: &addr1, Input: &iotago.UTXOInput{
			TransactionID:          *tx1ID,
			TransactionOutputIndex: 0,
		}}).
		AddOutput(&iotago.ExtendedOutput{Address: addr3, Amount: RequestFundsAmount}).
		Build(u.deSeriParas(), key1Signer)
	require.NoError(t, err)
	err = u.AddTransaction(spend3)
	require.Error(t, err)
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
