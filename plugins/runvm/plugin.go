package runvm

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

// PluginName is the name of the RunVM plugin.
const PluginName = "RunVM"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger

	vmDaemon = daemon.New()
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
	vmtypes.SetDefaultVMType(parameters.GetString(parameters.VMDefaultVmType))
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		// globally initialize VM
		go vmDaemon.Run()

		<-shutdownSignal

		vmDaemon.Shutdown()
		log.Infof("shutdown RunVM...  Done")
	})
	if err != nil {
		log.Errorf("failed to start RunVM worker")
	}
}

// RunComputationsAsync runs computations for the batch of requests in the background
func RunComputationsAsync(ctx *vm.VMTask) error {
	if len(ctx.Requests) == 0 {
		return fmt.Errorf("must be at least 1 request")
	}

	txb, err := txbuilder.NewFromAddressBalances(&ctx.Address, ctx.Balances)
	if err != nil {
		ctx.Log.Debugf("NewTxBuilder: %v\n%s", err, util.BalancesToString(ctx.Balances))
		return err
	}

	//ctx.Log.Infof("$$$$ dump orig txbuilder\n%s\n", txb.Dump())

	reqids := sctransaction.TakeRequestIds(ctx.Requests)
	bh := vm.BatchHash(reqids, ctx.Timestamp, ctx.LeaderPeerIndex)
	taskName := ctx.Address.String() + "." + bh.String()

	err = vmDaemon.BackgroundWorker(taskName, func(shutdownSignal <-chan struct{}) {
		runTask(ctx, txb, shutdownSignal)
	})
	return err
}

// runs batch
func runTask(ctx *vm.VMTask, txb *txbuilder.Builder, shutdownSignal <-chan struct{}) {
	ctx.Log.Debugw("runTask IN",
		"addr", ctx.Address.String(),
		"finalTimestamp", ctx.Timestamp,
		"state index", ctx.VirtualState.StateIndex(),
		"num req", len(ctx.Requests),
		"leader", ctx.LeaderPeerIndex,
	)

	// create state block and move smart contract state token
	if err := txb.CreateStateBlock(ctx.Color); err != nil {
		ctx.Log.Debugf("handleRequestTokens: %v\nDump txbuilder accounts:\n%s\n", err, txb.Dump())
		ctx.OnFinish(err)
		return
	}

	// erase all request tokens
	for _, reqRef := range ctx.Requests {
		if err := txb.EraseColor(ctx.Address, (balance.Color)(reqRef.Tx.ID()), 1); err != nil {
			ctx.Log.Debugf("handleRequestTokens: %v\nDump txbuilder accounts:\n%s\n", err, txb.Dump())
			ctx.OnFinish(err)
			return
		}
	}

	vmctx := &vm.VMContext{
		Address:       ctx.Address,
		OwnerAddress:  ctx.OwnerAddress,
		RewardAddress: ctx.RewardAddress,
		ProgramHash:   ctx.ProgramHash,
		MinimumReward: ctx.MinimumReward,
		Entropy:       *hashing.HashData(ctx.Entropy[:]), // mutates deterministically
		TxBuilder:     txb,                               // mutates
		Timestamp:     ctx.Timestamp,                     // mutate by incrementing 1 nanosec
		VirtualState:  ctx.VirtualState.Clone(),
		Log:           ctx.Log,
	}
	stateUpdates := make([]state.StateUpdate, 0, len(ctx.Requests))
	for _, reqRef := range ctx.Requests {

		vmctx.RequestRef = reqRef
		vmctx.StateUpdate = state.NewStateUpdate(reqRef.RequestId()).WithTimestamp(vmctx.Timestamp)

		runTheRequest(vmctx)

		stateUpdates = append(stateUpdates, vmctx.StateUpdate)
		// update state
		vmctx.VirtualState.ApplyStateUpdate(vmctx.StateUpdate)
		if vmctx.Timestamp != 0 {
			// increasing (nonempty) timestamp for 1 nanosecond for each request in the batch
			// the reason is to provide a different timestamp for each VM call and remain deterministic
			vmctx.Timestamp += 1
		}
		vmctx.Entropy = *hashing.HashData(vmctx.Entropy[:])
	}
	if len(stateUpdates) == 0 {
		// should not happen
		ctx.OnFinish(fmt.Errorf("RunVM: no state updates were produced"))
		return
	}

	var err error

	// create batch from state updates.
	ctx.ResultBatch, err = state.NewBatch(stateUpdates)
	if err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM: %v", err))
		return
	}
	ctx.ResultBatch.WithStateIndex(ctx.VirtualState.StateIndex() + 1)

	// calculate resulting state hash
	vsClone := ctx.VirtualState.Clone()
	if err = vsClone.ApplyBatch(ctx.ResultBatch); err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM: %v", err))
		return
	}
	stateHash := vsClone.Hash()

	// add state block
	err = vmctx.TxBuilder.SetStateParams(ctx.VirtualState.StateIndex()+1, &stateHash, vsClone.Timestamp())
	if err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM.txbuilder.SetStateParams: %v", err))
		return
	}
	// create result transaction
	ctx.ResultTransaction, err = vmctx.TxBuilder.Build(false)
	if err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM.txbuilder.Build: %v", err))
		return
	}

	// deprecate
	// check of all provided inputs were properly consumed
	//if err := ctx.ResultTransaction.ValidateConsumptionOfInputs(&ctx.Address, ctx.Balances); err != nil {
	//	ctx.Log.Errorf("RunVM.ValidateConsumptionOfInputs: wrong result transaction: %v", err)
	//	return
	//}

	ctx.Log.Debugw("runTask OUT",
		"result batch size", ctx.ResultBatch.Size(),
		"result batch state index", ctx.ResultBatch.StateIndex(),
		"result variable state hash", stateHash.String(),
		"result essence hash", hashing.HashData(ctx.ResultTransaction.EssenceBytes()).String(),
		"result tx finalTimestamp", time.Unix(0, ctx.ResultTransaction.MustState().Timestamp()),
	)
	// call back
	ctx.OnFinish(nil)
}
