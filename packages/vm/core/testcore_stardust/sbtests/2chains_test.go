package sbtests

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

// TODO deposit fee needs to be constant, this test is using a placeholder value that will need to be changed

// test case:
// 2 chains
// SC deployed on chain 2
// funds are deposited by some user on chain 1, on behalf of SC
// SC tries to withdraw those funds from chain 1 to chain 2
func Test2Chains(t *testing.T) { run2(t, test2Chains) }

func test2Chains(t *testing.T, w bool) {
	if w {
		t.SkipNow()
	}
	corecontracts.PrintWellKnownHnames()

	env := solo.New(t, &solo.InitOptions{
		AutoAdjustDustDeposit: true,
	}).
		WithNativeContract(sbtestsc.Processor)
	chain1 := env.NewChain(nil, "ch1")
	chain2 := env.NewChain(nil, "ch2")
	chain1.CheckAccountLedger()
	chain2.CheckAccountLedger()

	setupTestSandboxSC(t, chain1, nil, w)
	contractAgentID := setupTestSandboxSC(t, chain2, nil, w)

	userWallet, userAddress := env.NewKeyPairWithFunds()
	userAgentID := iscp.NewAgentID(userAddress, 0)
	env.AssertL1Iotas(userAddress, utxodb.FundsFromFaucetAmount)

	chain1CommonAccountIotas := chain1.L2Iotas(chain1.CommonAccount())
	chain2CommonAccountIotas := chain2.L2Iotas(chain2.CommonAccount())

	chain1.AssertL2Iotas(chain1.CommonAccount(), chain1CommonAccountIotas)
	chain1.AssertL2TotalIotas(chain1CommonAccountIotas + chain1.L2Iotas(chain1.OriginatorAgentID))

	chain2.AssertL2Iotas(chain2.CommonAccount(), chain2CommonAccountIotas)
	chain2.AssertL2TotalIotas(chain2CommonAccountIotas + chain2.L2Iotas(chain2.OriginatorAgentID))

	chain1TotalIotas := chain1.L2TotalIotas()
	chain2TotalIotas := chain2.L2TotalIotas()

	// send iotas to contractAgentID (that is an entity of chain2) on chain1
	iotasToSend := uint64(2 * iscp.Mi)
	iotasCreditedToSc2OnChain1 := uint64(1 * iscp.Mi)
	req := solo.NewCallParams(
		accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name,
		accounts.ParamAgentID, contractAgentID,
		accounts.ParamForceOpenAccount, true,
	).
		AddIotas(iotasToSend).
		AddAllowanceIotas(iotasCreditedToSc2OnChain1).
		WithGasBudget(math.MaxUint64)

	_, err := chain1.PostRequestSync(req, userWallet)
	require.NoError(t, err)

	receipt1 := chain1.LastReceipt()

	env.AssertL1Iotas(userAddress, utxodb.FundsFromFaucetAmount-iotasToSend)
	chain1.AssertL2Iotas(userAgentID, iotasToSend-iotasCreditedToSc2OnChain1-receipt1.GasFeeCharged)
	chain1.AssertL2Iotas(contractAgentID, iotasCreditedToSc2OnChain1)
	chain1.AssertL2Iotas(chain1.CommonAccount(), chain1CommonAccountIotas+receipt1.GasFeeCharged)
	chain1CommonAccountIotas += receipt1.GasFeeCharged
	chain1.AssertL2TotalIotas(chain1TotalIotas + iotasToSend)
	chain1TotalIotas += iotasToSend

	chain2.AssertL2Iotas(userAgentID, 0)
	chain2.AssertL2Iotas(contractAgentID, 0)
	chain2.AssertL2Iotas(chain2.CommonAccount(), chain2CommonAccountIotas)
	chain2.AssertL2TotalIotas(chain2TotalIotas)

	println("----chain1------------------------------------------")
	println(chain1.DumpAccounts())
	println("-----chain2-----------------------------------------")
	println(chain2.DumpAccounts())
	println("----------------------------------------------")

	// make chain2 send a call to chain1 to withdraw iotas
	iotasToWithdrawalFromChain1 := iotasCreditedToSc2OnChain1 // try to withdraw all iotas deposited to chain1 on behalf of chain2's contract
	reqAllowance := accounts.ConstDepositFeeTmp + 0.5*iscp.Mi // reqAllowance is the allowance provided to the "withdraw from chain" contract that will be used as gas in chain1 (also needs to be enough for storage deposit)

	req = solo.NewCallParams(ScName, sbtestsc.FuncWithdrawFromChain.Name,
		sbtestsc.ParamChainID, chain1.ChainID,
		sbtestsc.ParamIotasToWithdrawal, iotasToWithdrawalFromChain1).
		AddIotas(iotasToSend).
		WithAllowance(iscp.NewAllowanceIotas(reqAllowance)).
		WithGasBudget(math.MaxUint64)

	_, err = chain2.PostRequestSync(req, userWallet)
	require.NoError(t, err)
	chain2SendWithdrawalReceipt := chain2.LastReceipt()

	extra := 0
	if w {
		extra = 1
	}
	require.True(t, chain1.WaitForRequestsThrough(5+extra, 10*time.Second))
	require.True(t, chain2.WaitForRequestsThrough(5+extra, 10*time.Second))

	println("----chain1------------------------------------------")
	println(chain1.DumpAccounts())
	println("-----chain2-----------------------------------------")
	println(chain2.DumpAccounts())
	println("----------------------------------------------")

	chain2DepositReceipt := chain2.LastReceipt()

	chain1WithdrawalReceipt := chain1.LastReceipt()
	require.Equal(t, chain1WithdrawalReceipt.Request.CallTarget().Contract, accounts.Contract.Hname())
	require.Equal(t, chain1WithdrawalReceipt.Request.CallTarget().EntryPoint, accounts.FuncWithdraw.Hname())
	require.Nil(t, chain1WithdrawalReceipt.Error)

	env.AssertL1Iotas(userAddress, utxodb.FundsFromFaucetAmount-2*iotasToSend)

	chain1.AssertL2Iotas(userAgentID, iotasToSend-iotasCreditedToSc2OnChain1-receipt1.GasFeeCharged)
	chain1.AssertL2Iotas(contractAgentID, reqAllowance-chain1WithdrawalReceipt.GasFeeCharged) // amount of iotas sent from chain2 to chain1 in order to call the "withdrawal" request
	chain1.AssertL2Iotas(chain1.CommonAccount(), chain1CommonAccountIotas+chain1WithdrawalReceipt.GasFeeCharged)
	chain1.AssertL2TotalIotas(chain1TotalIotas + reqAllowance - iotasToWithdrawalFromChain1)

	chain2.AssertL2Iotas(userAgentID, iotasToSend-reqAllowance-chain2SendWithdrawalReceipt.GasFeeCharged)
	chain2.AssertL2Iotas(contractAgentID, iotasToWithdrawalFromChain1-accounts.ConstDepositFeeTmp)
	chain2.AssertL2Iotas(chain2.CommonAccount(), chain2CommonAccountIotas+chain2SendWithdrawalReceipt.GasFeeCharged+chain2DepositReceipt.GasFeeCharged)
	println(chain2.DumpAccounts())
	chain2.AssertL2TotalIotas(chain2TotalIotas + iotasToSend - accounts.ConstDepositFeeTmp + iotasCreditedToSc2OnChain1)
}
