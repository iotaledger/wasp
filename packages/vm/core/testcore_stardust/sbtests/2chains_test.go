package sbtests

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

func Test2Chains(t *testing.T) { run2(t, test2Chains) }
func test2Chains(t *testing.T, w bool) {
	core.PrintWellKnownHnames()

	env := solo.New(t, &solo.InitOptions{
		AutoAdjustDustDeposit: true,
	}).
		WithNativeContract(sbtestsc.Processor)
	chain1 := env.NewChain(nil, "ch1")
	chain2 := env.NewChain(nil, "ch2")
	chain1.CheckAccountLedger()
	chain2.CheckAccountLedger()

	contractAgentID1 := setupTestSandboxSC(t, chain1, nil, w)
	contractAgentID2 := setupTestSandboxSC(t, chain2, nil, w)

	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := iscp.NewAgentID(userAddress, 0)
	env.AssertL1Iotas(userAddress, solo.Saldo)

	chain1CommonAccountIotas := chain1.L2Iotas(chain1.CommonAccount())
	chain2CommonAccountIotas := chain2.L2Iotas(chain2.CommonAccount())

	chain1.AssertL2Iotas(chain1.CommonAccount(), chain1CommonAccountIotas)
	chain1.AssertL2TotalIotas(chain1CommonAccountIotas + chain1.L2Iotas(chain1.OriginatorAgentID))

	chain2.AssertL2Iotas(chain2.CommonAccount(), chain2CommonAccountIotas)
	chain2.AssertL2TotalIotas(chain2CommonAccountIotas + chain2.L2Iotas(chain2.OriginatorAgentID))

	chain1TotalIotas := chain1.L2TotalIotas()
	chain2TotalIotas := chain2.L2TotalIotas()

	// send 42 iotas to contractAgentID2 on chain1
	iotasToSend := uint64(1000)
	iotasCreditedToSc2OnChain1 := uint64(321)
	req := solo.NewCallParams(
		accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name,
		accounts.ParamAgentID, contractAgentID2,
		accounts.ParamForceOpenAccount, true,
	).
		AddAssetsIotas(iotasToSend).
		AddIotaAllowance(iotasCreditedToSc2OnChain1).
		WithGasBudget(math.MaxUint64)

	_, err := chain1.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	receipt1 := chain1.LastReceipt()

	env.AssertL1Iotas(userAddress, solo.Saldo-iotasToSend)
	chain1.AssertL2Iotas(userAgentID, iotasToSend-iotasCreditedToSc2OnChain1-receipt1.GasFeeCharged)
	chain1.AssertL2Iotas(contractAgentID1, 0)
	chain1.AssertL2Iotas(contractAgentID2, iotasCreditedToSc2OnChain1)
	chain1.AssertL2Iotas(chain1.CommonAccount(), chain1CommonAccountIotas+receipt1.GasFeeCharged)
	chain1CommonAccountIotas += receipt1.GasFeeCharged
	chain1.AssertL2TotalIotas(chain1TotalIotas + iotasToSend)
	chain1TotalIotas += iotasToSend

	chain2.AssertL2Iotas(userAgentID, 0)
	chain2.AssertL2Iotas(contractAgentID1, 0)
	chain2.AssertL2Iotas(contractAgentID2, 0)
	chain2.AssertL2Iotas(chain2.CommonAccount(), chain2CommonAccountIotas)
	chain2.AssertL2TotalIotas(chain2TotalIotas)

	// estimate gas cost of a withdrawal on chain1
	// estimatedWdGas, _, _ := chain1.EstimateGasOnLedger(solo.NewCallParams(
	// 	accounts.Contract.Name,accounts.FuncWithdraw.Name
	// ))
	estimatedWdGas := uint64(0)

	// make chain2 send a call to chain1 to withdraw iotas
	reqAllowance := uint64(500)
	iotasToWithdrawalFromChain1 := uint64(10) // TODO increase... this is not enough for dust deposit, just to test something
	req = solo.NewCallParams(ScName, sbtestsc.FuncWithdrawFromChain.Name,
		sbtestsc.ParamChainID, chain1.ChainID,
		sbtestsc.ParamGasBudgetToSend, estimatedWdGas, // TODO sc not using this...
		sbtestsc.ParamIotasToWithdrawal, iotasToWithdrawalFromChain1).
		AddAssetsIotas(iotasToSend).
		WithAllowance(iscp.NewAssetsIotas(reqAllowance)).
		WithGasBudget(math.MaxUint64)

	_, err = chain2.PostRequestSync(req, userWallet)
	require.NoError(t, err)
	chain2SendWithdrawalReceipt := chain2.LastReceipt()

	extra := 0
	if w {
		extra = 1
	}
	// TODO should be seconds
	require.True(t, chain1.WaitForRequestsThrough(5+extra, 10*time.Minute))
	require.True(t, chain2.WaitForRequestsThrough(4+extra, 10*time.Minute)) // TODO needs to be 5

	chain1WithdrawalReceipt := chain1.LastReceipt() // TODO seems like this is not the correct receipt, wtf?
	require.Equal(t, chain1WithdrawalReceipt.Request.CallTarget().EntryPoint, accounts.Contract.Hname())
	require.Nil(t, chain1WithdrawalReceipt.Error())

	env.AssertL1Iotas(userAddress, solo.Saldo-2*iotasToSend)

	chain1.AssertL2Iotas(userAgentID, iotasToSend-iotasCreditedToSc2OnChain1-receipt1.GasFeeCharged)
	chain1.AssertL2Iotas(contractAgentID1, 0)
	chain1.AssertL2Iotas(contractAgentID2, 0)
	// chain1.AssertL2Iotas(chain1.CommonAccount(), chain1CommonAccountIotas+0) // TODO ???
	// chain1.AssertL2TotalIotas(chain1TotalIotas + 0)                          // TODO ???

	chain2.AssertL2Iotas(userAgentID, iotasToSend-chain2SendWithdrawalReceipt.GasFeeCharged)
	chain2.AssertL2Iotas(contractAgentID1, 0)
	chain2.AssertL2Iotas(contractAgentID2, 44)
	// chain2.AssertL2Iotas(chain2.CommonAccount(), chain2CommonAccountIotas+chain2SendWithdrawalReceipt.GasFeeCharged+0) // TODO ????
	chain2.AssertL2TotalIotas(chain2TotalIotas + chain2SendWithdrawalReceipt.GasFeeCharged + iotasCreditedToSc2OnChain1)
}
