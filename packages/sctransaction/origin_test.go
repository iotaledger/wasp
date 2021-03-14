package sctransaction

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateOrigin(t *testing.T) {
	u := utxodb.New()
	user, addr := utxodb.NewKeyPairByIndex(1)
	_, err := u.RequestFunds(addr)
	require.NoError(t, err)

	_, stateAddr := utxodb.NewKeyPairByIndex(2)

	require.EqualValues(t, utxodb.RequestFundsAmount, u.BalanceIOTA(addr))
	require.EqualValues(t, 0, u.BalanceIOTA(stateAddr))

	bal100 := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100}
	tx, chainID, err := NewChainOriginTransaction(user, stateAddr, bal100, u.GetAddressOutputs(addr)...)
	require.NoError(t, err)

	t.Logf("New chain alias: %s", chainID.Base58())

	err = u.AddTransaction(tx)
	require.NoError(t, err)
}

func TestInitChain(t *testing.T) {
	u := utxodb.New()
	user, addr := utxodb.NewKeyPairByIndex(1)
	_, err := u.RequestFunds(addr)
	require.NoError(t, err)

	_, stateAddr := utxodb.NewKeyPairByIndex(2)

	require.EqualValues(t, utxodb.RequestFundsAmount, u.BalanceIOTA(addr))
	require.EqualValues(t, 0, u.BalanceIOTA(stateAddr))

	bal100 := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100}
	tx, chainID, err := NewChainOriginTransaction(user, stateAddr, bal100, u.GetAddressOutputs(addr)...)
	require.NoError(t, err)

	t.Logf("New chain alias: %s", chainID.Base58())

	err = u.AddTransaction(tx)
	require.NoError(t, err)

	tx, err = NewRootInitRequestTransaction(user, chainID, "test chain", u.GetAddressOutputs(addr)...)
	require.NoError(t, err)

	err = u.AddTransaction(tx)
	require.NoError(t, err)

	require.EqualValues(t, utxodb.RequestFundsAmount-100-1, u.BalanceIOTA(addr))
	require.EqualValues(t, 0, u.BalanceIOTA(stateAddr))
	require.EqualValues(t, 100+1, u.BalanceIOTA(chainID.AsAddress()))
	require.EqualValues(t, 2, len(u.GetAddressOutputs(chainID.AsAddress())))
}
