package chainutil

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

// EVMCall executes an EVM contract call and returns its output, discarding any state changes
func EVMCall(ch chain.ChainCore, aliasOutput *isc.AliasOutputWithID, call ethereum.CallMsg) ([]byte, error) {
	gasLimit := getMaxCallGasLimit(ch)

	// 0 means view call
	if call.Gas != 0 && call.Gas > gasLimit {
		call.Gas = gasLimit
	}

	iscReq := isc.NewEVMOffLedgerCallRequest(ch.ID(), call)
	// TODO: setting EstimateGasMode = true feels wrong here
	res, err := runISCRequest(ch, aliasOutput, time.Now(), iscReq, true)
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
