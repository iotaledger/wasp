package transaction

import (
	"bytes"
	"testing"

	"github.com/iotaledger/wasp/packages/state"

	"github.com/iotaledger/wasp/packages/testutil/testdeserparams"

	"github.com/iotaledger/wasp/packages/utxodb"

	"github.com/stretchr/testify/require"
)

func TestCreateOrigin(t *testing.T) {
	u := utxodb.New()
	user, addr := u.NewKeyPairByIndex(1)
	_, err := u.RequestFunds(addr)
	require.NoError(t, err)

	_, stateAddr := u.NewKeyPairByIndex(2)

	require.EqualValues(t, utxodb.RequestFundsAmount, u.GetAddressBalanceIotas(addr))
	require.EqualValues(t, 0, u.GetAddressBalanceIotas(stateAddr))

	allOutputs, ids := u.GetUnspentOutputs(addr)
	tx, chainID, err := NewChainOriginTransaction(
		user,
		stateAddr,
		stateAddr,
		100,
		allOutputs,
		ids,
		testdeserparams.DeSerializationParameters(),
	)
	require.NoError(t, err)

	err = u.AddTransaction(tx)
	require.NoError(t, err)

	txid, err := tx.ID()
	require.NoError(t, err)

	txBack, ok := u.GetTransaction(*txid)
	require.True(t, ok)
	txidBack, err := txBack.ID()
	require.NoError(t, err)
	require.EqualValues(t, *txid, *txidBack)

	t.Logf("New chain ID: %s", chainID.String())

	anchor, err := GetAnchorFromTransaction(tx)
	require.NoError(t, err)
	require.True(t, anchor.IsOrigin)
	require.EqualValues(t, *chainID, anchor.ChainID)
	require.EqualValues(t, 0, anchor.StateIndex)
	require.True(t, stateAddr.Equal(anchor.StateController))
	require.True(t, stateAddr.Equal(anchor.GovernanceController))
	require.True(t, bytes.Equal(state.OriginStateHash().Bytes(), anchor.StateData.Bytes()))

	// only one output is expected in the ledger under the address of chainID
	outs, ids := u.GetUnspentOutputs(chainID.AsAddress())
	require.EqualValues(t, 1, len(outs))
	require.EqualValues(t, 1, len(ids))

	out := u.GetOutput(anchor.OutputID)
	require.NotNil(t, out)
}

func TestInitChainTx(t *testing.T) {
	u := utxodb.New()
	user, addr := u.NewKeyPairByIndex(1)
	_, err := u.RequestFunds(addr)
	require.NoError(t, err)

	_, stateAddr := u.NewKeyPairByIndex(2)

	require.EqualValues(t, utxodb.RequestFundsAmount, u.GetAddressBalanceIotas(addr))
	require.EqualValues(t, 0, u.GetAddressBalanceIotas(stateAddr))

	allOutputs, ids := u.GetUnspentOutputs(addr)
	originTx, chainID, err := NewChainOriginTransaction(
		user,
		stateAddr,
		stateAddr,
		100,
		allOutputs,
		ids,
		testdeserparams.DeSerializationParameters(),
	)
	require.NoError(t, err)

	err = u.AddTransaction(originTx)
	require.NoError(t, err)

	chainIotas := originTx.Essence.Outputs[0].Deposit()

	allOutputs, ids = u.GetUnspentOutputs(addr)
	txInit, err := NewRootInitRequestTransaction(
		user,
		chainID,
		"test chain",
		allOutputs,
		ids,
		testdeserparams.DeSerializationParameters(),
	)
	require.NoError(t, err)

	err = u.AddTransaction(txInit)
	require.NoError(t, err)

	initIotas := txInit.Essence.Outputs[0].Deposit()

	t.Logf("chainIotas: %d initIotas: %d", chainIotas, initIotas)

	require.EqualValues(t, utxodb.RequestFundsAmount-chainIotas-initIotas, int(u.GetAddressBalanceIotas(addr)))
	require.EqualValues(t, 0, u.GetAddressBalanceIotas(stateAddr))
	allOutputs, ids = u.GetUnspentOutputs(chainID.AsAddress())
	require.EqualValues(t, 2, len(allOutputs))
	require.EqualValues(t, 2, len(ids))
}
