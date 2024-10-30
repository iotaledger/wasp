package chainutil

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

// EVMCall executes an EVM contract call and returns its output, discarding any state changes
func EVMCall(
	anchor *isc.StateAnchor,
	store indexedstore.IndexedStore,
	processors *processors.Config,
	log *logger.Logger,
	call ethereum.CallMsg,
) ([]byte, error) {
	chainID := anchor.ChainID()

	latestState, err := store.LatestState()
	if err != nil {
		return nil, err
	}
	info := getChainInfo(chainID, latestState)

	// 0 means view call
	gasLimit := gas.EVMCallGasLimit(info.GasLimits, &info.GasFeePolicy.EVMGasRatio)
	if call.Gas != 0 && call.Gas > gasLimit {
		call.Gas = gasLimit
	}

	if call.GasPrice == nil {
		call.GasPrice = info.GasFeePolicy.DefaultGasPriceFullDecimals(parameters.Decimals)
	}

	iscReq := isc.NewEVMOffLedgerCallRequest(chainID, call)
	// TODO: setting EstimateGasMode = true feels wrong here
	res, err := runISCRequest(
		anchor,
		store,
		processors,
		log,
		time.Now(),
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
	return res.Return.Bytes(), nil
}
