package chainutil

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// EVMCall executes an EVM contract call and returns its output, discarding any state changes
func EVMCall(
	anchor *isc.StateAnchor,
	l1Params *parameters.L1Params,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log log.Logger,
	call ethereum.CallMsg,
) ([]byte, error) {
	latestState, err := store.LatestState()
	if err != nil {
		return nil, err
	}
	info := getChainInfo(anchor.ChainID(), latestState)

	// 0 means view call
	gasLimit := gas.EVMCallGasLimit(info.GasLimits, &info.GasFeePolicy.EVMGasRatio)
	if call.Gas != 0 && call.Gas > gasLimit {
		call.Gas = gasLimit
	}

	if call.GasPrice == nil {
		call.GasPrice = info.GasFeePolicy.DefaultGasPriceFullDecimals(parameters.BaseTokenDecimals)
	}

	iscReq := isc.NewEVMOffLedgerCallRequest(info.ChainID, call)
	res, err := runISCRequest(
		anchor,
		l1Params,
		store,
		processors,
		log,
		time.Now(),
		hashing.PseudoRandomHash(nil),
		iscReq,
		true,
	)
	if err != nil {
		return nil, err
	}
	if res.Receipt.Error != nil {
		vmerr, resolvingErr := ResolveError(latestState, res.Receipt.Error)
		if resolvingErr != nil {
			panic(fmt.Errorf("error resolving vmerror %w", resolvingErr))
		}
		return nil, vmerr
	}
	return bcs.Unmarshal[[]byte](res.Return[0])
}
