package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp/colored"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestMintOk(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chain.CheckChain()

	wallet, walletAddr := env.NewKeyPairWithFunds()
	env.AssertAddressIotas(walletAddr, solo.Saldo)

	color1, err := env.MintTokens(wallet, 1000)
	require.NoError(t, err)
	env.AssertAddressIotas(walletAddr, solo.Saldo-1000)
	env.AssertAddressNativeTokenBalance(walletAddr, color1, 1000)

	color2, err := env.MintTokens(wallet, 100)
	require.NoError(t, err)
	env.AssertAddressNativeTokenBalance(walletAddr, colored.IOTA, solo.Saldo-1000-100)
	env.AssertAddressNativeTokenBalance(walletAddr, color1, 1000)
	env.AssertAddressNativeTokenBalance(walletAddr, color2, 100)
}

func TestMintFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chain.CheckChain()

	wallet, walletAddr := env.NewKeyPairWithFunds()
	env.AssertAddressNativeTokenBalance(walletAddr, colored.IOTA, solo.Saldo)

	color1, err := env.MintTokens(wallet, 1000)
	require.NoError(t, err)
	env.AssertAddressNativeTokenBalance(walletAddr, colored.IOTA, solo.Saldo-1000)
	env.AssertAddressNativeTokenBalance(walletAddr, color1, 1000)

	_, err = env.MintTokens(wallet, solo.Saldo-500)
	require.Error(t, err)
	env.AssertAddressNativeTokenBalance(walletAddr, colored.IOTA, solo.Saldo-1000)
	env.AssertAddressNativeTokenBalance(walletAddr, color1, 1000)
}

//func TestDestroyColoredOk1(t *testing.T) {
//	env := solo.New(t, false, false)
//	chain := env.NewChain(nil, "chain1")
//
//	chain.CheckChain()
//
//	wallet := env.NewKeyPairWithFunds()
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo)
//
//	color1, err := env.MintTokens(wallet, 1000)
//	require.NoError(t, err)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo-1000)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), color1, 1000)
//
//	err = env.DestroyColoredTokens(wallet, color1, 100)
//	require.NoError(t, err)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo-1000+100)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), color1, 1000-100)
//}
//
//func TestDestroyColoredOk2(t *testing.T) {
//	env := solo.New(t, false, false)
//	chain := env.NewChain(nil, "chain1")
//
//	chain.CheckChain()
//
//	wallet := env.NewKeyPairWithFunds()
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo)
//
//	color1, err := env.MintTokens(wallet, 1000)
//	require.NoError(t, err)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo-1000)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), color1, 1000)
//
//	err = env.DestroyColoredTokens(wallet, color1, 1000)
//	require.NoError(t, err)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), color1, 0)
//}
//
//func TestDestroyColoredFail(t *testing.T) {
//	env := solo.New(t, false, false)
//	chain := env.NewChain(nil, "chain1")
//
//	chain.CheckChain()
//
//	wallet := env.NewKeyPairWithFunds()
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo)
//
//	color1, err := env.MintTokens(wallet, 1000)
//	require.NoError(t, err)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo-1000)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), color1, 1000)
//
//	err = env.DestroyColoredTokens(wallet, color1, 1100)
//	require.ErrorStr(t, err)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), ledgerstate.ColorIOTA, solo.Saldo-1000)
//	env.AssertAddressNativeTokenBalance(wallet.Address(), color1, 1000)
//}
