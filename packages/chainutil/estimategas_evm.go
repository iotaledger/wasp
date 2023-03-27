package chainutil

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/params"
	"go.uber.org/zap"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

func executeISCVM(ch chain.ChainCore, aliasOutput *isc.AliasOutputWithID, req isc.Request) (*vm.RequestResult, error) {
	vmRunner := runvm.NewVMRunner()

	task := &vm.VMTask{
		Processors:           ch.Processors(),
		AnchorOutput:         aliasOutput.GetAliasOutput(),
		AnchorOutputID:       aliasOutput.OutputID(),
		Store:                ch.Store(),
		Requests:             []isc.Request{req},
		TimeAssumption:       time.Now(),
		Entropy:              hashing.PseudoRandomHash(nil),
		ValidatorFeeTarget:   isc.NewContractAgentID(ch.ID(), 0),
		EnableGasBurnLogging: false,
		EstimateGasMode:      true,
		Log:                  ch.Log().Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
	err := vmRunner.Run(task)
	if err != nil {
		return nil, err
	}
	if len(task.Results) == 0 {
		return nil, errors.New("request was skipped")
	}
	return task.Results[0], nil
}

var evmErrorsRegex = regexp.MustCompile("out of gas|intrinsic gas too low|(execution reverted$)")

// EstimateGas executes the given request and discards the resulting chain state. It is useful
// for estimating gas.
func EstimateGas(ch chain.ChainCore, aliasOutput *isc.AliasOutputWithID, call ethereum.CallMsg) (uint64, error) { //nolint:gocyclo
	// Determine the lowest and highest possible gas limits to binary search in between
	var (
		lo     uint64 = params.TxGas - 1
		hi     uint64
		gasCap uint64
	)

	maximumPossibleGas, err := getMaxCallGasLimit(ch)
	if err != nil {
		return 0, err
	}

	if call.Gas >= params.TxGas {
		hi = call.Gas
	} else {
		hi = maximumPossibleGas
	}

	gasCap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (failed bool, used uint64, err error) {
		call.Gas = gas
		iscReq := isc.NewEVMOffLedgerCallRequest(ch.ID(), call)
		res, err := executeISCVM(ch, aliasOutput, iscReq)
		if err != nil {
			return true, 0, err
		}
		if res.Receipt.Error != nil {
			if res.Receipt.Error.ErrorCode == vm.ErrGasBudgetExceeded.Code() {
				// out of gas when charging ISC gas
				return true, 0, nil
			}
			vmerr, resolvingErr := ResolveError(ch, res.Receipt.Error)
			if resolvingErr != nil {
				panic(fmt.Errorf("error resolving vmerror %w", resolvingErr))
			}
			if evmErrorsRegex.Match([]byte(vmerr.Error())) {
				// increase gas
				return true, 0, nil
			}
			return true, 0, vmerr
		}
		return false, res.Receipt.GasBurned, nil
	}

	// Execute the binary search and hone in on an executable gas limit
	var lastUsed uint64

	const maxLastUsedAttempts = 2
	lastUsedAttempts := 0

	for lo+1 < hi {
		mid := (hi + lo) / 2
		if lastUsed > lo && lastUsed != mid && lastUsed < hi && lastUsedAttempts < maxLastUsedAttempts {
			// use the last used gas as a better estimation to home in faster
			mid = lastUsed
			// this may turn the binary search into a linear search for some
			// edge cases. We put a limit and after that we default to the
			// binary search.
			lastUsedAttempts++
		}

		var failed bool
		var err error
		failed, lastUsed, err = executable(mid)
		if err != nil {
			return 0, err
		}
		if failed {
			lo = mid
		} else {
			hi = mid
			if lastUsed == mid {
				// if used gas == gas limit, then use this as the estimation.
				// It may not be the most precise estimation (e.g. lowering the gas
				// limit may end up using less gas), but it's "good enough" and
				// saves a lot of iterations in the binary search.
				break
			}
		}
	}

	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == gasCap {
		failed, _, err := executable(hi)
		if err != nil {
			return 0, err
		}
		if failed {
			if hi == maximumPossibleGas {
				return 0, fmt.Errorf("request might require more gas than it is allowed by the VM (%d), or will never succeed", gasCap)
			}
			// the specified gas cap is too low
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", gasCap)
		}
	}
	return hi, nil
}

func getMaxCallGasLimit(ch chain.ChainCore) (uint64, error) {
	ret, err := CallView(
		mustLatestState(ch),
		ch,
		governance.Contract.Hname(),
		governance.ViewGetChainInfo.Hname(),
		nil,
	)
	if err != nil {
		return 0, err
	}
	fp, err := gas.FeePolicyFromBytes(ret.MustGet(governance.VarGasFeePolicyBytes))
	if err != nil {
		return 0, err
	}
	gl, err := gas.LimitsFromBytes(ret.MustGet(governance.VarGasLimitsBytes))
	if err != nil {
		return 0, err
	}
	return gas.EVMCallGasLimit(gl, &fp.EVMGasRatio), nil
}
