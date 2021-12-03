package transaction

import (
	"testing"

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

	t.Logf("New chain alias: %s", chainID.Base58())

	err = u.AddTransaction(tx)
	require.NoError(t, err)
}

//func TestInitChain(t *testing.T) {
//	u := utxodb.New()
//	user, addr := u.NewKeyPairByIndex(1)
//	_, err := u.RequestFunds(addr)
//	require.NoError(t, err)
//
//	_, stateAddr := u.NewKeyPairByIndex(2)
//
//	require.EqualValues(t, utxodb.RequestFundsAmount, u.BalanceIOTA(addr))
//	require.EqualValues(t, 0, u.BalanceIOTA(stateAddr))
//
//	bal100 := colored.NewBalancesForIotas(100)
//	tx, chainID, err := NewChainOriginTransaction(user, stateAddr, bal100, time.Now(), u.GetAddressOutputs(addr)...)
//	require.NoError(t, err)
//
//	t.Logf("New chain alias: %s", chainID.Base58())
//
//	err = u.AddTransaction(tx)
//	require.NoError(t, err)
//
//	tx, err = NewRootInitRequestTransaction(user, chainID, "test chain", time.Now(), u.GetAddressOutputs(addr)...)
//	require.NoError(t, err)
//
//	err = u.AddTransaction(tx)
//	require.NoError(t, err)
//
//	require.EqualValues(t, utxodb.RequestFundsAmount-100-1, u.BalanceIOTA(addr))
//	require.EqualValues(t, 0, u.BalanceIOTA(stateAddr))
//	require.EqualValues(t, 100+1, u.BalanceIOTA(chainID.AsAddress()))
//	require.EqualValues(t, 2, len(u.GetAddressOutputs(chainID.AsAddress())))
//}
