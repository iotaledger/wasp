package chainutil

import (
	"fmt"

	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

// Call executes an EVM contract call and returns its output, discarding any state changes
func Call(ch chain.ChainCore, aliasOutput *isc.AliasOutputWithID, call ethereum.CallMsg) ([]byte, error) {
	gasLimit, err := getMaxCallGasLimit(ch)
	if err != nil {
		return nil, err
	}
	if call.Gas == 0 || call.Gas > gasLimit {
		call.Gas = gasLimit
	}

	iscReq := isc.NewEVMOffLedgerCallRequest(ch.ID(), call)
	res, err := executeISCVM(ch, aliasOutput, iscReq)
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
