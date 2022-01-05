package testcore

import (
	"errors"
	"testing"

	"github.com/iotaledger/wasp/packages/utxodb"

	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"

	"github.com/iotaledger/wasp/packages/vm/gas"

	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/stretchr/testify/require"
)

// TestLedgerBaseConsistency deploys chain and check consistency of L1 and L2 ledgers
func TestLedgerBaseConsistency(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	genesisAddr := env.L1Ledger().GenesisAddress()
	assets := env.L1AddressBalances(genesisAddr)
	require.EqualValues(t, env.L1Ledger().Supply(), assets.Iotas)

	// create chain
	ch, _, initTx := env.NewChainExt(nil, "chain1")
	defer ch.Log.Sync()
	env.WaitPublisher()
	ch.AssertControlAddresses()
	t.Logf("originator address iotas: %d (spent %d)",
		env.L1IotaBalance(ch.OriginatorAddress), solo.Saldo-env.L1IotaBalance(ch.OriginatorAddress))

	// get all native tokens. Must be empty
	nativeTokenIDs := ch.GetOnChainTokenIDs()
	require.EqualValues(t, 0, len(nativeTokenIDs))

	// query dust parameters of the latest block
	totalIotasInfo := ch.GetTotalIotaInfo()
	totalIotasOnChain := ch.L2TotalIotasInAccounts()
	// all goes to dust and to total iotas on chain
	totalSpent := totalIotasInfo.TotalDustDeposit + totalIotasInfo.TotalIotasInL2Accounts
	t.Logf("total on chain: dust deposit: %d, total iotas on chain: %d, total spent: %d",
		totalIotasInfo.TotalDustDeposit, totalIotasOnChain, totalSpent)
	// what has left on L1 address
	env.AssertL1AddressIotas(ch.OriginatorAddress, solo.Saldo-totalSpent)

	// let's analise dust deposit on origin and init transactions
	vByteCostInit := transaction.GetVByteCosts(initTx, env.RentStructure())[0]
	dustCosts := vmtxbuilder.NewDepositEstimate(env.RentStructure())
	// what we spent is only for dust deposits for those 2 transactions
	require.EqualValues(t, int(totalSpent), int(dustCosts.AnchorOutput+vByteCostInit))

	// check if there's a single alias output on chain's address
	aliasOutputs, _ := env.L1Ledger().GetAliasOutputs(ch.ChainID.AsAddress())
	require.EqualValues(t, 1, len(aliasOutputs))

	// check total on chain assets
	totalAssets := ch.L2TotalAssetsInAccounts()
	// no native tokens expected
	require.EqualValues(t, 0, len(totalAssets.Tokens))
	// what spent all goes to the alias output
	require.EqualValues(t, int(totalSpent), int(aliasOutputs[0].Amount))
	// total iotas on L2 must be equal to alias output iotas - dust deposit
	ch.AssertL2TotalIotas(aliasOutputs[0].Amount - dustCosts.AnchorOutput)

	// all dust deposit of the init request goes to the sender account
	ch.AssertL2AccountIotas(ch.OriginatorAgentID, vByteCostInit)
	// common account is empty
	require.EqualValues(t, 0, ch.L2CommonAccountIotas())
}

// TestNoTargetPostOnLedger test what happens when sending requests to non-existent contract or entry point
func TestNoTargetPostOnLedger(t *testing.T) {
	t.Run("no contract,originator==sender", func(t *testing.T) {
		env := solo.New(t)
		env.EnablePublisher(true)
		ch := env.NewChain(nil, "chain1")
		defer ch.Log.Sync()

		totalIotasBefore := ch.L2TotalIotasInAccounts()
		originatorsL2IotasBefore := ch.L2AccountIotas(ch.OriginatorAgentID)
		originatorsL1IotasBefore := env.L1IotaBalance(ch.OriginatorAddress)
		require.EqualValues(t, 0, ch.L2CommonAccountIotas())

		req := solo.NewCallParams("dummyContract", "dummyEP").
			WithGasBudget(1000)
		reqTx, _, err := ch.PostRequestSyncTx(req, nil)
		// expecting specific error
		require.True(t, errors.Is(err, vmcontext.ErrTargetContractNotFound))

		totalIotasAfter := ch.L2TotalIotasInAccounts()
		commonAccountIotasAfter := ch.L2CommonAccountIotas()

		reqDustDeposit := transaction.GetVByteCosts(reqTx, env.RentStructure())[0]

		// total iotas on chain increase by the dust deposit from the request tx
		require.EqualValues(t, int(totalIotasBefore+reqDustDeposit), int(totalIotasAfter))
		// sender on L1 is charged with dust deposit
		env.AssertL1AddressIotas(ch.OriginatorAddress, originatorsL1IotasBefore-reqDustDeposit)
		// originator (sender) is charged with gas fee on L2
		ch.AssertL2AccountIotas(ch.OriginatorAgentID, originatorsL2IotasBefore+reqDustDeposit-gas.NotFoundTarget)
		// all gas fee goes to the common account
		require.EqualValues(t, int(gas.NotFoundTarget), commonAccountIotasAfter)
		env.WaitPublisher()
	})
	t.Run("no contract,originator!=sender", func(t *testing.T) {
		env := solo.New(t)
		env.EnablePublisher(true)
		ch := env.NewChain(nil, "chain1")
		defer ch.Log.Sync()

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID := iscp.NewAgentID(senderAddr, 0)

		totalIotasBefore := ch.L2TotalIotasInAccounts()
		originatorsL2IotasBefore := ch.L2AccountIotas(ch.OriginatorAgentID)
		originatorsL1IotasBefore := env.L1IotaBalance(ch.OriginatorAddress)
		env.AssertL1AddressIotas(senderAddr, solo.Saldo)
		require.EqualValues(t, 0, ch.L2CommonAccountIotas())

		req := solo.NewCallParams("dummyContract", "dummyEP").
			WithGasBudget(1000)
		reqTx, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		// expecting specific error
		require.Contains(t, err.Error(), vmcontext.ErrTargetContractNotFound.Error())

		totalIotasAfter := ch.L2TotalIotasInAccounts()
		commonAccountIotasAfter := ch.L2CommonAccountIotas()

		reqDustDeposit := transaction.GetVByteCosts(reqTx, env.RentStructure())[0]

		// total iotas on chain increase by the dust deposit from the request tx
		require.EqualValues(t, int(totalIotasBefore+reqDustDeposit), int(totalIotasAfter))
		// originator on L1 does not change
		env.AssertL1AddressIotas(ch.OriginatorAddress, originatorsL1IotasBefore)
		// sender on L1 is charged with dust deposit
		env.AssertL1AddressIotas(senderAddr, solo.Saldo-reqDustDeposit)
		// originator account does not change
		ch.AssertL2AccountIotas(ch.OriginatorAgentID, originatorsL2IotasBefore)
		// sender is charged with gas fee on L2
		ch.AssertL2AccountIotas(senderAgentID, reqDustDeposit-gas.NotFoundTarget)
		// all gas fee goes to the common account
		require.EqualValues(t, int(gas.NotFoundTarget), commonAccountIotasAfter)
		env.WaitPublisher()
	})
	t.Run("no EP,originator==sender", func(t *testing.T) {
		env := solo.New(t)
		env.EnablePublisher(true)
		ch := env.NewChain(nil, "chain1")
		defer ch.Log.Sync()

		totalIotasBefore := ch.L2TotalIotasInAccounts()
		originatorsL2IotasBefore := ch.L2AccountIotas(ch.OriginatorAgentID)
		originatorsL1IotasBefore := env.L1IotaBalance(ch.OriginatorAddress)
		require.EqualValues(t, 0, ch.L2CommonAccountIotas())

		req := solo.NewCallParams(root.Contract.Name, "dummyEP").
			WithGasBudget(1000)
		reqTx, _, err := ch.PostRequestSyncTx(req, nil)
		// expecting specific error
		require.Contains(t, err.Error(), vmcontext.ErrTargetEntryPointNotFound.Error())

		totalIotasAfter := ch.L2TotalIotasInAccounts()
		commonAccountIotasAfter := ch.L2CommonAccountIotas()

		reqDustDeposit := transaction.GetVByteCosts(reqTx, env.RentStructure())[0]

		// total iotas on chain increase by the dust deposit from the request tx
		require.EqualValues(t, int(totalIotasBefore+reqDustDeposit), int(totalIotasAfter))
		// sender on L1 is charged with dust deposit
		env.AssertL1AddressIotas(ch.OriginatorAddress, originatorsL1IotasBefore-reqDustDeposit)
		// originator (sender) is charged with gas fee on L2
		ch.AssertL2AccountIotas(ch.OriginatorAgentID, originatorsL2IotasBefore+reqDustDeposit-gas.NotFoundTarget)
		// all gas fee goes to the common account
		require.EqualValues(t, int(gas.NotFoundTarget), commonAccountIotasAfter)
		env.WaitPublisher()
	})
	t.Run("no EP,originator!=sender", func(t *testing.T) {
		env := solo.New(t)
		env.EnablePublisher(true)
		ch := env.NewChain(nil, "chain1")
		defer ch.Log.Sync()

		senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
		senderAgentID := iscp.NewAgentID(senderAddr, 0)

		totalIotasBefore := ch.L2TotalIotasInAccounts()
		originatorsL2IotasBefore := ch.L2AccountIotas(ch.OriginatorAgentID)
		originatorsL1IotasBefore := env.L1IotaBalance(ch.OriginatorAddress)
		env.AssertL1AddressIotas(senderAddr, solo.Saldo)
		require.EqualValues(t, 0, ch.L2CommonAccountIotas())

		req := solo.NewCallParams(root.Contract.Name, "dummyEP").
			WithGasBudget(1000)
		reqTx, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
		// expecting specific error
		require.Contains(t, err.Error(), vmcontext.ErrTargetEntryPointNotFound.Error())

		totalIotasAfter := ch.L2TotalIotasInAccounts()
		commonAccountIotasAfter := ch.L2CommonAccountIotas()

		reqDustDeposit := transaction.GetVByteCosts(reqTx, env.RentStructure())[0]

		// total iotas on chain increase by the dust deposit from the request tx
		require.EqualValues(t, int(totalIotasBefore+reqDustDeposit), int(totalIotasAfter))
		// originator on L1 does not change
		env.AssertL1AddressIotas(ch.OriginatorAddress, originatorsL1IotasBefore)
		// sender on L1 is charged with dust deposit
		env.AssertL1AddressIotas(senderAddr, solo.Saldo-reqDustDeposit)
		// originator account does not change
		ch.AssertL2AccountIotas(ch.OriginatorAgentID, originatorsL2IotasBefore)
		// sender is charged with gas fee on L2
		ch.AssertL2AccountIotas(senderAgentID, reqDustDeposit-gas.NotFoundTarget)
		// all gas fee goes to the common account
		require.EqualValues(t, int(gas.NotFoundTarget), commonAccountIotasAfter)
		env.WaitPublisher()
	})
}

func TestNoTargetView(t *testing.T) {
	t.Run("no contract view", func(t *testing.T) {
		env := solo.New(t)
		env.EnablePublisher(true)
		chain := env.NewChain(nil, "chain1")
		chain.AssertControlAddresses()

		_, err := chain.CallView("dummyContract", "dummyEP")
		require.Error(t, err)
		env.WaitPublisher()
	})
	t.Run("no EP view", func(t *testing.T) {
		env := solo.New(t)
		env.EnablePublisher(true)
		chain := env.NewChain(nil, "chain1")
		chain.AssertControlAddresses()

		_, err := chain.CallView(root.Contract.Name, "dummyEP")
		require.Error(t, err)
		env.WaitPublisher()
	})
}

func TestOkCall(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	ch := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name).
		WithGasBudget(1000)
	_, err := ch.PostRequestSync(req, nil)
	require.NoError(t, err)
	env.WaitPublisher()
}

func TestRepeatInit(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")
		req := solo.NewCallParams(root.Contract.Name, "init")
		_, err := ch.PostRequestSync(req, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already initialized")
		ch.CheckAccountLedger()
	})
	t.Run("accounts", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")
		req := solo.NewCallParams(accounts.Contract.Name, "init")
		_, err := ch.PostRequestSync(req, nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, vmcontext.ErrRepeatingInitCall))
		ch.CheckAccountLedger()
	})
	t.Run("blocklog", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")
		req := solo.NewCallParams(blocklog.Contract.Name, "init")
		_, err := ch.PostRequestSync(req, nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, vmcontext.ErrRepeatingInitCall))
		ch.CheckAccountLedger()
	})
	t.Run("blob", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")
		req := solo.NewCallParams(blob.Contract.Name, "init")
		_, err := ch.PostRequestSync(req, nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, vmcontext.ErrRepeatingInitCall))
		ch.CheckAccountLedger()
	})
	t.Run("governance", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "chain1")
		req := solo.NewCallParams(governance.Contract.Name, "init")
		_, err := ch.PostRequestSync(req, nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, vmcontext.ErrRepeatingInitCall))
		ch.CheckAccountLedger()
	})
}

func TestDeployNativeContract(t *testing.T) {
	env := solo.New(t).WithNativeContract(sbtestsc.Processor)

	env.EnablePublisher(true)
	ch := env.NewChain(nil, "chain1")

	senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
	//senderAgentID := iscp.NewAgentID(senderAddr, 0)

	// get more iotas for originator
	originatorBalance := env.L1AddressBalances(ch.OriginatorAddress).Iotas
	_, err := env.L1Ledger().GetFundsFromFaucet(ch.OriginatorAddress)
	require.NoError(t, err)
	env.AssertL1AddressIotas(ch.OriginatorAddress, originatorBalance+utxodb.FundsFromFaucetAmount)

	req := solo.NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name,
		root.ParamDeployer, iscp.NewAgentID(senderAddr, 0)).
		AddIotas(1000).
		WithGasBudget(1000)
	_, err = ch.PostRequestSync(req, nil)
	require.NoError(t, err)

	err = ch.DeployContract(senderKeyPair, "sctest", sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)
	//
	//req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name)
	//_, err := ch.PostRequestSync(req, nil)
	env.WaitPublisher()
}

func TestFeeBasic(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")
	feePolicy := chain.GetGasFeePolicy()
	require.Nil(t, feePolicy.GasFeeTokenID)
	require.EqualValues(t, 0, feePolicy.ValidatorFeeShare)
}
