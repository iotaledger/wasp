// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

//func TestAccountsBase(t *testing.T) {
//	env := solo.New(t, false, false)
//	chain := env.NewChain(nil, "chain1")
//	chain.CheckAccountLedger()
//	chain.AssertL2TotalIotas(1)
//	chain.AssertCommonAccountIotas(1)
//}

//func TestAccountsRepeatInit(t *testing.T) {
//	env := solo.New(t, false, false)
//	chain := env.NewChain(nil, "chain1")
//	req := solo.NewCallParams(accounts.Contract.Name, "init")
//	_, err := chain.PostRequestSync(req.AddAssetsIotas(1), nil)
//	require.Error(t, err)
//	chain.CheckAccountLedger()
//	chain.AssertL2TotalIotas(1)
//	chain.AssertCommonAccountIotas(1)
//}

func TestAccountsBase1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := iscp.NewAgentID(ownerAddr, 0)
	req := solo.NewCallParams(governance.Contract.Name, governance.FuncDelegateChainOwnership.Name, governance.ParamChainOwner, newOwnerAgentID)
	req.AddAssetsIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
	chain.AssertL2AccountIotas(newOwnerAgentID, 0)
	chain.AssertL2AccountIotas(chain.ContractAgentID(root.Contract.Name), 0)
	chain.CheckAccountLedger()

	req = solo.NewCallParams(governance.Contract.Name, governance.FuncClaimChainOwnership.Name).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
	chain.AssertL2AccountIotas(newOwnerAgentID, 0)
	chain.AssertL2AccountIotas(chain.ContractAgentID(root.Contract.Name), 0)
	chain.CheckAccountLedger()
	chain.AssertL2TotalIotas(3)
	chain.AssertCommonAccountIotas(3)
}

func TestAccountsDepositWithdrawToAddress(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, newOwnerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := iscp.NewAgentID(newOwnerAddr, 0)
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
	_, err := chain.PostRequestSync(req.AddAssetsIotas(42), newOwner)
	require.NoError(t, err)

	chain.AssertL2AccountNativeToken(newOwnerAgentID, colored.IOTA, 42)

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)
	chain.CheckAccountLedger()
	chain.AssertL2AccountNativeToken(newOwnerAgentID, colored.IOTA, 0)
	env.AssertAddressNativeTokenBalance(newOwnerAddr, colored.IOTA, solo.Saldo)
	chain.AssertL2TotalIotas(1)
	chain.AssertCommonAccountIotas(1)

	// withdraw owner's iotas
	_, ownerFromChain, _ := chain.GetInfo()
	require.True(t, chain.OriginatorAgentID.Equals(ownerFromChain))
	t.Logf("origintor/owner: %s", chain.OriginatorAgentID.String())

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, chain.OriginatorPrivateKey)
	require.NoError(t, err)
	chain.AssertL2TotalIotas(2)
	chain.AssertL2AccountIotas(chain.OriginatorAgentID, 0)
	chain.AssertCommonAccountIotas(2)

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncHarvest.Name).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, chain.OriginatorPrivateKey)

	require.NoError(t, err)
	chain.AssertL2TotalIotas(3)
	chain.AssertL2AccountIotas(chain.OriginatorAgentID, 3)
	chain.AssertCommonAccountIotas(0)
}

func TestAccountsHarvest(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	// default EP will be called and tokens deposited
	req := solo.NewCallParams(blocklog.Contract.Name, "")
	_, err := chain.PostRequestSync(req.AddAssetsIotas(42), nil)
	require.NoError(t, err)

	chain.AssertL2TotalIotas(43)
	chain.AssertCommonAccountIotas(43)

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncHarvest.Name).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertL2TotalIotas(44)
	chain.AssertCommonAccountIotas(0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-100-44)
}

func TestAccountsHarvestFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	// default EP will be called and tokens deposited
	req := solo.NewCallParams(blocklog.Contract.Name, "")
	_, err := chain.PostRequestSync(req.AddAssetsIotas(42), nil)
	require.NoError(t, err)

	chain.AssertL2TotalIotas(43)
	chain.AssertCommonAccountIotas(43)

	kp, _ := env.NewKeyPairWithFunds()

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncHarvest.Name).AddAssetsIotas(1)
	_, err = chain.PostRequestSync(req, kp)
	require.Error(t, err)

	chain.AssertL2TotalIotas(43)
	chain.AssertCommonAccountIotas(43)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-100-43)
}

func TestAccountsDepositToCommon(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	// default EP will be called and tokens deposited
	req := solo.NewCallParams("", "")
	_, err := chain.PostRequestSync(req.AddAssetsIotas(42), nil)
	require.NoError(t, err)

	chain.AssertL2TotalIotas(43)
	chain.AssertCommonAccountIotas(43)
}

func getAccountNonce(t *testing.T, chain *solo.Chain, address iotago.Address) uint64 {
	ret, err := chain.CallView(accounts.Contract.Name, accounts.FuncGetAccountNonce.Name,
		accounts.ParamAgentID, iscp.NewAgentID(address, 0),
	)
	require.NoError(t, err)
	nonce, err := codec.DecodeUint64(ret.MustGet(accounts.ParamAccountNonce), 0)
	require.NoError(t, err)
	return nonce
}

func TestGetAccountNonce(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	userWallet, userAddress := env.NewKeyPairWithFunds()

	// initial nonce should be 0
	require.Zero(t, getAccountNonce(t, chain, userAddress))

	// deposit funds to be able to issue offledger requests
	_, err := chain.PostRequestSync(
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).AddAssetsIotas(1000),
		userWallet,
	)
	require.NoError(t, err)
	require.Zero(t, getAccountNonce(t, chain, userAddress))

	nowNanoTs := uint64(time.Now().UnixNano())

	// offledger requests are constructed with nonce = current TS in nanoseconds
	chain.PostRequestOffLedger(
		solo.NewCallParams("", "").AddAssetsIotas(100),
		userWallet,
	)

	require.GreaterOrEqual(t, getAccountNonce(t, chain, userAddress), nowNanoTs)
}
