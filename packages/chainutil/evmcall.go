package chainutil

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/wasp/packages/chain/chaintypes"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// EVMCall executes an EVM contract call and returns its output, discarding any state changes
func EVMCall(
	ch chaintypes.ChainCore,
	accountOutput *isc.AccountOutputWithID,
	call ethereum.CallMsg,
) ([]byte, error) {
	info := getChainInfo(ch)

	// 0 means view call
	gasLimit := gas.EVMCallGasLimit(info.GasLimits, &info.GasFeePolicy.EVMGasRatio)
	if call.Gas != 0 && call.Gas > gasLimit {
		call.Gas = gasLimit
	}

	if call.GasPrice == nil {
		call.GasPrice = info.GasFeePolicy.GasPriceWei(parameters.BaseToken().Decimals)
	}

	iscReq := isc.NewEVMOffLedgerCallRequest(ch.ID(), call)
	// TODO: setting EstimateGasMode = true feels wrong here
	blockTime := vm.Time{
		SlotIndex: accountOutput.OutputID().CreationSlot() + 1,
		Timestamp: time.Now(),
	}
	res, err := runISCRequest(ch, accountOutput, blockTime, iscReq, true)
	if err != nil {
		return nil, err
	}
	if res.Receipt.Error != nil {
		vmerr, resolvingErr := ResolveError(ch, res.Receipt.Error)
		if resolvingErr != nil {
			panic(fmt.Errorf("error resolving vmerror %w", resolvingErr))
		}
		return nil, vmerr
	}
	return res.Return[evm.FieldResult], nil
}
