package chainutil

import (
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/vmimpl"
)

func runISCTask(
	ch chain.ChainCore,
	anchor *isc.StateAnchor,
	blockTime time.Time,
	reqs []isc.Request,
	estimateGasMode bool,
	evmTracer *isc.EVMTracer,
) ([]*vm.RequestResult, error) {
	store := ch.Store()
	migs, err := getMigrationsForBlock(store, anchor)
	if err != nil {
		return nil, err
	}
	task := &vm.VMTask{
		Processors:           ch.Processors(),
		Anchor:               anchor,
		Store:                store,
		Requests:             reqs,
		Timestamp:            blockTime,
		Entropy:              hashing.PseudoRandomHash(nil),
		ValidatorFeeTarget:   accounts.CommonAccount(),
		EnableGasBurnLogging: estimateGasMode,
		EstimateGasMode:      estimateGasMode,
		EVMTracer:            evmTracer,
		Log:                  ch.Log().Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
		Migrations:           migs,
	}
	res, err := vmimpl.Run(task)
	if err != nil {
		return nil, err
	}
	return res.RequestResults, nil
}

func getMigrationsForBlock(store indexedstore.IndexedStore, anchor *isc.StateAnchor) (*migrations.MigrationScheme, error) {
	prevL1Commitment, err := transaction.L1CommitmentFromAnchor(anchor)
	if err != nil {
		panic(err)
	}
	prevState, err := store.StateByTrieRoot(prevL1Commitment.TrieRoot())
	if err != nil {
		if errors.Is(err, state.ErrTrieRootNotFound) {
			return allmigrations.DefaultScheme, nil
		}
		panic(err)
	}
	if lo.Must(store.LatestBlockIndex()) == prevState.BlockIndex() {
		return allmigrations.DefaultScheme, nil
	}
	newState := lo.Must(store.StateByIndex(prevState.BlockIndex() + 1))
	targetSchemaVersion := newState.SchemaVersion()
	return allmigrations.DefaultScheme.WithTargetSchemaVersion(targetSchemaVersion)
}

func runISCRequest(
	ch chain.ChainCore,
	anchor *isc.StateAnchor,
	blockTime time.Time,
	req isc.Request,
	estimateGasMode bool,
) (*vm.RequestResult, error) {
	results, err := runISCTask(
		ch,
		anchor,
		blockTime,
		[]isc.Request{req},
		estimateGasMode,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("request was skipped")
	}
	return results[0], nil
}
