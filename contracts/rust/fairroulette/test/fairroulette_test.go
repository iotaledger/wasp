// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/client/wallet/packages/seed"
	"github.com/iotaledger/wasp/contracts/common"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	chain := common.StartChainAndDeployWasmContractByName(t, ScName)
	_, err := chain.FindContract(ScName)
	require.NoError(t, err)
}

func getLastWinningNumber(chain *solo.Chain) (dict.Dict, error) {
	result, err := chain.CallView("fairroulette", HViewLastWinningNumber.String())

	return result, err
}

func TestTutorial2(t *testing.T) {
	chain := common.StartChainAndDeployWasmContractByName(t, ScName)
	chain.GetChainOutput().SetDelegationTimelock(time.Now())

	// agent := iscp.NewAgentID(chain.ChainID.AsAddress(), iscp.Hn(ScName))

	pSeed := seed.NewSeed()

	env := solo.New(t, false, false, pSeed.Seed)

	_, userAddress := env.NewKeyPairWithFunds(env.NewSeedFromIndex(1))
	t.Logf("address of the userWallet is: %s", userAddress.Base58())
	numIotas := env.GetAddressBalance(userAddress, colored.IOTA) // how many iotas the address contains
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	env.AssertAddressBalance(userAddress, colored.IOTA, solo.Saldo)

	req := solo.NewCallParams("fairroulette", FuncPlaceBet, "number", 1).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	req = solo.NewCallParams("fairroulette", FuncPayWinners).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	req = solo.NewCallParams("fairroulette", FuncPlaceBet, "number", 1).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
}
