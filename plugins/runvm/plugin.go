package runvm

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmnil"
	"sync"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "VM"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger

	vmDaemon        = daemon.New()
	processors      = make(map[string]vm.Processor)
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

// RegisterProcessor creates and registers processor for program hash
// asynchronously
// possibly, locates Wasm program code in IPFS and caches here
func RegisterProcessor(programHash string, onFinish func(err error)) {
	go func() {
		processorsMutex.Lock()
		defer processorsMutex.Unlock()

		switch programHash {
		case "KSoWFbHwZuHG8B8HVcYVKR4WYVQ7MpoqeaXgKWfkBMF": // sc1
			processors[programHash] = vmnil.New()
			onFinish(nil)

		case "7xmPcECfZsSQq5eq7GCucuxmL2QpsgYwTjusuQcoK9GE": // sc2
			onFinish(fmt.Errorf("VM not implemented"))

		case "2tx7z36m9EhX3xBRGmEUD4FwTyP6R66zGPYY53EWc87k": // sc3
			onFinish(fmt.Errorf("VM not implemented"))

		default:
			onFinish(fmt.Errorf("can't create processor for progam hash %s", programHash))
		}
	}()
}

func getProcessor(programHash string) (vm.Processor, error) {
	processorsMutex.RLock()
	defer processorsMutex.RUnlock()

	ret, ok := processors[programHash]
	if !ok {
		return nil, errors.New("no such processor")
	}
	return ret, nil
}

// RunComputationsAsync runs computations for the batch of requests in the background
func RunComputationsAsync(ctx *vm.VMTask) error {
	if len(ctx.Requests) == 0 {
		return fmt.Errorf("must be at least 1 request")
	}
	processor, err := getProcessor(ctx.ProgramHash.String())
	if err != nil {
		return err
	}
	txbuilder, err := vm.NewTxBuilder(vm.TransactionBuilderParams{
		Balances:   ctx.Balances,
		OwnColor:   ctx.Color,
		OwnAddress: ctx.Address,
	})
	if err != nil {
		return err
	}

	reqids := sctransaction.TakeRequestIds(ctx.Requests)
	bh := vm.BatchHash(reqids, ctx.Timestamp)
	taskName := ctx.Address.String() + "." + bh.String()

	err = vmDaemon.BackgroundWorker(taskName, func(shutdownSignal <-chan struct{}) {
		runVM(ctx, txbuilder, processor, shutdownSignal)
	})
	return err
}

// runs batch
func runVM(ctx *vm.VMTask, txbuilder *vm.TransactionBuilder, processor vm.Processor, shutdownSignal <-chan struct{}) {
	ctx.Log.Debugf("input balances:\n%s", util.BalancesToString(ctx.Balances))

	vmctx := &vm.VMContext{
		Address:       ctx.Address,
		Color:         ctx.Color,
		TxBuilder:     txbuilder,
		Timestamp:     ctx.Timestamp,
		VariableState: state.NewVariableState(ctx.VariableState), // clone
		Log:           ctx.Log,
	}
	stateUpdates := make([]state.StateUpdate, 0, len(ctx.Requests))
	for _, reqRef := range ctx.Requests {
		// destroy token corresponding to request
		err := txbuilder.EraseColor(ctx.Address, (balance.Color)(reqRef.Tx.ID()), 1)
		if err != nil {
			// not enough balance for requests tokens
			// stop here
			ctx.Log.Errorf("something wrong with request token for reqid = %s. Not all requests were processed: %v",
				reqRef.RequestId().String(), err)
			break
		}
		// run processor
		vmctx.Request = reqRef
		vmctx.StateUpdate = state.NewStateUpdate(reqRef.RequestId())
		processor.Run(vmctx)

		stateUpdates = append(stateUpdates, vmctx.StateUpdate)
		// update state
		vmctx.VariableState.ApplyStateUpdate(vmctx.StateUpdate)
	}
	if len(stateUpdates) == 0 {
		ctx.Log.Errorf("no state updates were produced")
		return
	}

	var err error
	// create batch out of state updates. Note that batch can contain less state updates than number of requests
	ctx.ResultBatch, err = state.NewBatch(stateUpdates)
	if err != nil {
		ctx.Log.Error(err)
	}
	ctx.ResultBatch.WithStateIndex(ctx.VariableState.StateIndex() + 1)
	err = ctx.VariableState.ApplyBatch(ctx.ResultBatch)
	if err != nil {
		ctx.Log.Error(err)
	}
	// create final transaction
	ctx.ResultTransaction = vmctx.TxBuilder.Finalize(
		ctx.VariableState.StateIndex(),
		ctx.VariableState.Hash(),
		ctx.Timestamp.UnixNano(),
	)
	// call back
	ctx.OnFinish()
}
