package testcore

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMintOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chain.CheckChain()

	wallet := env.NewKeyPairWithFunds()
	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)

	color1, err := env.MintTokens(wallet, 1000)
	require.NoError(t, err)
	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000)
	env.AssertAddressBalance(wallet.Address(), color1, 1000)

	color2, err := env.MintTokens(wallet, 100)
	require.NoError(t, err)
	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000-100)
	env.AssertAddressBalance(wallet.Address(), color1, 1000)
	env.AssertAddressBalance(wallet.Address(), color2, 100)
}

func TestMintFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chain.CheckChain()

	wallet := env.NewKeyPairWithFunds()
	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)

	color1, err := env.MintTokens(wallet, 1000)
	require.NoError(t, err)
	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000)
	env.AssertAddressBalance(wallet.Address(), color1, 1000)

	_, err = env.MintTokens(wallet, 500)
	require.Error(t, err)
	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000)
	env.AssertAddressBalance(wallet.Address(), color1, 1000)
}

//func TestDestroyColoredOk1(t *testing.T) {
//	env := solo.New(t, false, false)
//	chain := env.NewChain(nil, "chain1")
//
//	chain.CheckChain()
//
//	wallet := env.NewKeyPairWithFunds()
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
//
//	color1, err := env.MintTokens(wallet, 1000)
//	require.NoError(t, err)
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000)
//	env.AssertAddressBalance(wallet.Address(), color1, 1000)
//
//	err = env.DestroyColoredTokens(wallet, color1, 100)
//	require.NoError(t, err)
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000+100)
//	env.AssertAddressBalance(wallet.Address(), color1, 1000-100)
//}
//
//func TestDestroyColoredOk2(t *testing.T) {
//	env := solo.New(t, false, false)
//	chain := env.NewChain(nil, "chain1")
//
//	chain.CheckChain()
//
//	wallet := env.NewKeyPairWithFunds()
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
//
//	color1, err := env.MintTokens(wallet, 1000)
//	require.NoError(t, err)
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000)
//	env.AssertAddressBalance(wallet.Address(), color1, 1000)
//
//	err = env.DestroyColoredTokens(wallet, color1, 1000)
//	require.NoError(t, err)
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
//	env.AssertAddressBalance(wallet.Address(), color1, 0)
//}
//
//func TestDestroyColoredFail(t *testing.T) {
//	env := solo.New(t, false, false)
//	chain := env.NewChain(nil, "chain1")
//
//	chain.CheckChain()
//
//	wallet := env.NewKeyPairWithFunds()
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
//
//	color1, err := env.MintTokens(wallet, 1000)
//	require.NoError(t, err)
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000)
//	env.AssertAddressBalance(wallet.Address(), color1, 1000)
//
//	err = env.DestroyColoredTokens(wallet, color1, 1100)
//	require.Error(t, err)
//	env.AssertAddressBalance(wallet.Address(), balance.ColorIOTA, testutil.RequestFundsAmount-1000)
//	env.AssertAddressBalance(wallet.Address(), color1, 1000)
//}
