package runvm

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/packages/vm/vmnil"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "VM"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger

	vmDaemon        = daemon.New()
	processors      = make(map[string]processor.Processor)
	processorsMutex sync.RWMutex
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		// globally initialize VM
		go vmDaemon.Run()

		<-shutdownSignal

		vmDaemon.Shutdown()
		log.Infof("shutdown VM...  Done")
	})
	if err != nil {
		log.Errorf("failed to start NodeConn worker")
	}
}

// LoadProcessorAsync creates and registers processor for program hash
// asynchronously
// possibly, locates Wasm program code in IPFS and caches here
func LoadProcessorAsync(programHash string, onFinish func(err error)) {
	go func() {
		processorsMutex.Lock()
		defer processorsMutex.Unlock()

		switch programHash {
		case vmnil.ProgramHash:
			processors[programHash] = vmnil.New()
			onFinish(nil)

		case logsc.ProgramHash:
			processors[programHash] = logsc.New()
			onFinish(nil)

		default:
			onFinish(fmt.Errorf("can't create processor for program hash %s", programHash))
		}
	}()
}

func CheckProcessor(programHash string) bool {
	_, err := getProcessor(programHash)
	return err == nil
}

func getProcessor(programHash string) (processor.Processor, error) {
	processorsMutex.RLock()
	defer processorsMutex.RUnlock()

	ret, ok := processors[programHash]
	if !ok {
		return nil, fmt.Errorf("no such processor: %v", programHash)
	}
	return ret, nil
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

	reqids := sctransaction.TakeRequestIds(ctx.Requests)
	bh := vm.BatchHash(reqids, ctx.Timestamp)
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

	vmctx := &vm.VMContext{
		Address:       ctx.Address,
		OwnerAddress:  ctx.OwnerAddress,
		RewardAddress: ctx.RewardAddress,
		ProgramHash:   ctx.ProgramHash,
		MinimumReward: ctx.MinimumReward,
		TxBuilder:     txb, //mutates
		Timestamp:     ctx.Timestamp,
		VirtualState:  state.NewVirtualState(ctx.VirtualState), // clone
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
	}
	if len(stateUpdates) == 0 {
		// should not happen
		ctx.Log.Errorf("RunVM: no state updates were produced")
		return
	}

	var err error

	// create batch from state updates.
	ctx.ResultBatch, err = state.NewBatch(stateUpdates)
	if err != nil {
		ctx.Log.Errorf("RunVM: %v", err)
		return
	}
	ctx.ResultBatch.WithStateIndex(ctx.VirtualState.StateIndex() + 1)

	// calculate resulting state hash
	vsClone := state.NewVirtualState(ctx.VirtualState)
	if err = vsClone.ApplyBatch(ctx.ResultBatch); err != nil {
		ctx.Log.Errorf("RunVM: %v", err)
		return
	}
	stateHash := vsClone.Hash()

	// add state block
	err = vmctx.TxBuilder.AddStateBlock(sctransaction.NewStateBlock(sctransaction.NewStateBlockParams{
		Color:      ctx.Color,
		StateIndex: ctx.VirtualState.StateIndex() + 1,
		StateHash:  vsClone.Hash(),
		Timestamp:  vsClone.Timestamp(),
	}))
	if err != nil {
		ctx.Log.Errorf("RunVM.AddStateBlock: %v", err)
		return
	}
	// create result transaction
	ctx.ResultTransaction, err = vmctx.TxBuilder.Build(false)
	if err != nil {
		ctx.Log.Errorf("RunVM.txbuilder.Build: %v", err)
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
	ctx.OnFinish()
}
