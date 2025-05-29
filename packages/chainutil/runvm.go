package chainutil

import (
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/vmimpl"
)

func runISCTask(
	anchor *isc.StateAnchor,
	l1Params *parameters.L1Params,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log log.Logger,
	blockTime time.Time,
	entropy hashing.HashValue,
	reqs []isc.Request,
	enforceGasBurned []vm.EnforceGasBurned,
	estimateGasMode bool,
	evmTracer *tracers.Tracer,
) ([]*vm.RequestResult, error) {
	migs, err := getMigrationsForBlock(store, anchor)
	if err != nil {
		return nil, err
	}
	task := &vm.VMTask{
		Processors:           processors,
		Anchor:               anchor,
		GasCoin:              nil,
		L1Params:             l1Params,
		Store:                store,
		Requests:             reqs,
		Timestamp:            blockTime,
		Entropy:              entropy,
		ValidatorFeeTarget:   accounts.CommonAccount(),
		EnableGasBurnLogging: estimateGasMode,
		EstimateGasMode:      estimateGasMode,
		EnforceGasBurned:     enforceGasBurned,
		EVMTracer:            evmTracer,
		Log:                  log,
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
	anchor *isc.StateAnchor,
	l1Params *parameters.L1Params,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log log.Logger,
	blockTime time.Time,
	entropy hashing.HashValue,
	req isc.Request,
	estimateGasMode bool,
) (*vm.RequestResult, error) {
	results, err := runISCTask(
		anchor,
		l1Params,
		store,
		processors,
		log,
		blockTime,
		entropy,
		[]isc.Request{req},
		nil,
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
