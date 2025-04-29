package webapi_validation

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
	"github.com/stretchr/testify/assert"
)

type EvmValidation struct {
	client base.EvmClientWrapper
}

func NewEvmValidation(s_clientUri string, r_clientUri string, validationContext base.ValidationContext) EvmValidation {
	evmStardustUri := fmt.Sprintf("%s/v1/chains/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5/evm", s_clientUri)
	evmRebasedUri := fmt.Sprintf("%s/v1/chain/evm", r_clientUri)

	evmClient, err := base.NewEvmClientWrapper(evmStardustUri, evmRebasedUri, validationContext)
	if err != nil {
		panic(err)
	}

	return EvmValidation{
		client: *evmClient,
	}
}

func (v *EvmValidation) ValidateEvm(blockNumber uint64) error {
	stardustTraces, rebasedTraces, err := v.client.GetTxTraces(blockNumber)
	if err != nil {
		return fmt.Errorf("failed to get traces for block %d: %w", blockNumber, err)
	}

	if !assert.Equal(base.T, len(stardustTraces), len(rebasedTraces), "number of traces for block %d is not equal", blockNumber) {
		return fmt.Errorf("number of traces for block %d is not equal", blockNumber)
	}

	for i, sTrace := range stardustTraces {

		if !assert.Equal(base.T, sTrace.TxHash, rebasedTraces[i].TxHash, "transaction hashes for block %d index %d are not equal", blockNumber, i) {
			return fmt.Errorf("transaction hashes for block %d index %d are not equal", blockNumber, i)
		}

		rTrace, ok := rebasedTraces[i].Trace.(map[string]any)
		if !ok {
			return fmt.Errorf("failed to cast rebased trace for block %d: %w", blockNumber, err)
		}

		sTrace, ok := sTrace.Trace.(map[string]any)
		if !ok {
			return fmt.Errorf("failed to cast stardust trace for block %d: %w", blockNumber, err)
		}

		from := sTrace["from"].(string)
		to := sTrace["to"].(string)

		// require.Equal(base.T, sFrom, rFrom, "transaction origins for block %d are not equal, index %d, stardust: %s, rebased: %s", blockNumber, i, sFrom, rFrom)
		// require.Equal(base.T, sTo, rTo, "transaction destinations for block %d are not equal, index %d, stardust: %s, rebased: %s", blockNumber, i, sTo, rTo)

		if !assert.EqualValues(base.T, sTrace, rTrace, "traces for block %d are not equal", blockNumber) {
			sTraceJson, err := json.MarshalIndent(sTrace, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal stardust trace for block %d: %w", blockNumber, err)
			}

			rTraceJson, err := json.MarshalIndent(rTrace, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal rebased trace for block %d: %w", blockNumber, err)
			}

			os.WriteFile(fmt.Sprintf("%s_stardust_trace_(block=%d)_(index=%d).json", rebasedTraces[i].TxHash, blockNumber, i), sTraceJson, 0644)
			os.WriteFile(fmt.Sprintf("%s_rebased_trace_(block=%d)_(index=%d).json", rebasedTraces[i].TxHash, blockNumber, i), rTraceJson, 0644)

			return fmt.Errorf("traces for block %d are not equal", blockNumber)
		}

		stardustBalance, rebasedBalance, err := v.client.GetBalances(common.HexToAddress(from), blockNumber)
		if err != nil {
			return fmt.Errorf("failed to get balances for address %s in block %d: %w", from, blockNumber, err)
		}

		if !assert.Equal(base.T, stardustBalance, rebasedBalance, "balances for address %s in block %d are not equal, index %d, stardust: %s, rebased: %s", from, blockNumber, i, stardustBalance, rebasedBalance) {
			return fmt.Errorf("balances for address %s in block %d are not equal, index %d, stardust: %s, rebased: %s", from, blockNumber, i, stardustBalance, rebasedBalance)
		}

		stardustBalance, rebasedBalance, err = v.client.GetBalances(common.HexToAddress(to), blockNumber)
		if err != nil {
			return fmt.Errorf("failed to get balances for address %s in block %d: %w", to, blockNumber, err)
		}

		if !assert.Equal(base.T, stardustBalance, rebasedBalance, "balances for address %s in block %d are not equal, index %d, stardust: %s, rebased: %s", to, blockNumber, i, stardustBalance, rebasedBalance) {
			return fmt.Errorf("balances for address %s in block %d are not equal, index %d, stardust: %s, rebased: %s", to, blockNumber, i, stardustBalance, rebasedBalance)
		}
	}

	return nil
}
