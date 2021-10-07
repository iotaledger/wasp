// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"

	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestAccountsBase(t *testing.T) {
	env := solo.New(t, true, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()
	chain.AssertTotalIotas(1)
	chain.AssertCommonAccountIotas(1)
}

func TestAccountsRepeatInit(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	req := solo.NewCallParams(accounts.Contract.Name, "init")
	_, err := chain.PostRequestSync(req.WithIotas(1), nil)
	require.Error(t, err)
	chain.CheckAccountLedger()
	chain.AssertTotalIotas(1)
	chain.AssertCommonAccountIotas(1)
}

func TestAccountsBase1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := iscp.NewAgentID(ownerAddr, 0)
	req := solo.NewCallParams(governance.Contract.Name, governance.FuncDelegateChainOwnership.Name, governance.ParamChainOwner, newOwnerAgentID)
	req.WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(newOwnerAgentID, 0)
	chain.AssertIotas(chain.ContractAgentID(root.Contract.Name), 0)
	chain.CheckAccountLedger()

	req = solo.NewCallParams(governance.Contract.Name, governance.FuncClaimChainOwnership.Name).WithIotas(1)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertIotas(newOwnerAgentID, 0)
	chain.AssertIotas(chain.ContractAgentID(root.Contract.Name), 0)
	chain.CheckAccountLedger()
	chain.AssertTotalIotas(3)
	chain.AssertCommonAccountIotas(3)
}

func TestAccountsDepositWithdrawToAddress(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	newOwner, newOwnerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := iscp.NewAgentID(newOwnerAddr, 0)
	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
	_, err := chain.PostRequestSync(req.WithIotas(42), newOwner)
	require.NoError(t, err)

	chain.AssertAccountBalance(newOwnerAgentID, colored.IOTA, 42)

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).WithIotas(1)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)
	chain.CheckAccountLedger()
	chain.AssertAccountBalance(newOwnerAgentID, colored.IOTA, 0)
	env.AssertAddressBalance(newOwnerAddr, colored.IOTA, solo.Saldo)
	chain.AssertTotalIotas(1)
	chain.AssertCommonAccountIotas(1)

	// withdraw owner's iotas
	_, ownerFromChain, _ := chain.GetInfo()
	require.True(t, chain.OriginatorAgentID.Equals(&ownerFromChain))
	t.Logf("origintor/owner: %s", chain.OriginatorAgentID.String())

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).WithIotas(1)
	_, err = chain.PostRequestSync(req, chain.OriginatorKeyPair)
	require.NoError(t, err)
	chain.AssertTotalIotas(2)
	chain.AssertIotas(&chain.OriginatorAgentID, 0)
	chain.AssertCommonAccountIotas(2)

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncHarvest.Name).WithIotas(1)
	_, err = chain.PostRequestSync(req, chain.OriginatorKeyPair)

	require.NoError(t, err)
	chain.AssertTotalIotas(3)
	chain.AssertIotas(&chain.OriginatorAgentID, 3)
	chain.AssertCommonAccountIotas(0)
}

func TestAccountsHarvest(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	// default EP will be called and tokens deposited
	req := solo.NewCallParams(blocklog.Contract.Name, "")
	_, err := chain.PostRequestSync(req.WithIotas(42), nil)
	require.NoError(t, err)

	chain.AssertTotalIotas(43)
	chain.AssertCommonAccountIotas(43)

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncHarvest.Name).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	chain.AssertTotalIotas(44)
	chain.AssertCommonAccountIotas(0)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-100-44)
}

func TestAccountsHarvestFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	// default EP will be called and tokens deposited
	req := solo.NewCallParams(blocklog.Contract.Name, "")
	_, err := chain.PostRequestSync(req.WithIotas(42), nil)
	require.NoError(t, err)

	chain.AssertTotalIotas(43)
	chain.AssertCommonAccountIotas(43)

	kp, _ := env.NewKeyPairWithFunds()

	req = solo.NewCallParams(accounts.Contract.Name, accounts.FuncHarvest.Name).WithIotas(1)
	_, err = chain.PostRequestSync(req, kp)
	require.Error(t, err)

	chain.AssertTotalIotas(43)
	chain.AssertCommonAccountIotas(43)
	env.AssertAddressIotas(chain.OriginatorAddress, solo.Saldo-100-43)
}

func TestAccountsDepositToCommon(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	chain.CheckAccountLedger()

	// default EP will be called and tokens deposited
	req := solo.NewCallParams("", "")
	_, err := chain.PostRequestSync(req.WithIotas(42), nil)
	require.NoError(t, err)

	chain.AssertTotalIotas(43)
	chain.AssertCommonAccountIotas(43)
}

func getAccountNonce(t *testing.T, chain *solo.Chain, address ledgerstate.Address) uint64 {
	ret, err := chain.CallView(accounts.Contract.Name, accounts.FuncGetAccountNonce.Name,
		accounts.ParamAgentID, iscp.NewAgentID(address, 0),
	)
	require.NoError(t, err)
	nonce, _, err := codec.DecodeUint64(ret.MustGet(accounts.ParamAccountNonce))
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
		solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).WithIotas(1000),
		userWallet,
	)
	require.NoError(t, err)
	require.Zero(t, getAccountNonce(t, chain, userAddress))

	nowNanoTs := uint64(time.Now().UnixNano())

	// offledger requests are constructed with nonce = current TS in nanoseconds
	chain.PostRequestOffLedger(
		solo.NewCallParams("", "").WithIotas(100),
		userWallet,
	)

	require.GreaterOrEqual(t, getAccountNonce(t, chain, userAddress), nowNanoTs)
}
