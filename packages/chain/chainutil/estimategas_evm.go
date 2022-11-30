package chainutil

import (
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/params"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

func executeIscVM(ch chain.Chain, req isc.Request) (*vm.RequestResult, error) {
	vmRunner := runvm.NewVMRunner()
	var ret *vm.RequestResult
	err := optimism.RetryOnStateInvalidated(func() (err error) {
		anchorOutput := ch.GetAnchorOutput()
		vs, ok, err := ch.GetVirtualState()
		if err != nil {
			return err
		}
		if !ok {
			return xerrors.Errorf("solid state does not exist")
		}
		task := &vm.VMTask{
			Processors:         ch.Processors(),
			AnchorOutput:       anchorOutput.GetAliasOutput(),
			AnchorOutputID:     anchorOutput.OutputID(),
			Requests:           []isc.Request{req},
			TimeAssumption:     time.Now(),
			VirtualStateAccess: vs,
			Entropy:            hashing.RandomHash(nil),
			ValidatorFeeTarget: isc.NewContractAgentID(ch.ID(), 0),
			Log:                ch.Log().Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
			// state baseline is always valid in Solo
			SolidStateBaseline:   ch.GlobalStateSync().GetSolidIndexBaseline(),
			EnableGasBurnLogging: true,
			EstimateGasMode:      true,
		}
		err = vmRunner.Run(task)
		if err != nil {
			return err
		}
		if len(task.Results) == 0 {
			return xerrors.Errorf("request was skipped")
		}
		ret = task.Results[0]
		return nil
	})
	return ret, err
}

var evmErrorsRegex = regexp.MustCompile("out of gas|intrinsic gas too low|(execution reverted$)")

// EstimateGas executes the given request and discards the resulting chain state. It is useful
// for estimating gas.
func EstimateGas(ch chain.Chain, call ethereum.CallMsg) (uint64, error) {
	// Determine the lowest and highest possible gas limits to binary search in between
	var (
		lo     uint64 = params.TxGas - 1
		hi     uint64
		gasCap uint64
	)

	ret, err := CallView(ch, evm.Contract.Hname(), evm.FuncGetCallGasLimit.Hname(), nil)
	if err != nil {
		return 0, err
	}
	maximumPossibleGas := codec.MustDecodeUint64(ret.MustGet(evm.FieldResult))

	if call.Gas >= params.TxGas {
		hi = call.Gas
	} else {
		hi = maximumPossibleGas
	}

	gasCap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (failed bool, err error) {
		call.Gas = gas
		iscReq := isc.NewEVMOffLedgerEstimateGasRequest(ch.ID(), call)
		res, err := executeIscVM(ch, iscReq)
		if err != nil {
			return true, err
		}
		if res.Receipt.Error != nil {
			if res.Receipt.Error.ErrorCode == vm.ErrGasBudgetExceeded.Code() {
				// out of gas when charging ISC gas
				return true, nil
			}
			vmerr, resolvingErr := ch.ResolveError(res.Receipt.Error)
			if resolvingErr != nil {
				panic(fmt.Errorf("error resolving vmerror %v", resolvingErr))
			}
			if evmErrorsRegex.Match([]byte(vmerr.Error())) {
				// increase gas
				return true, nil
			}
			return true, vmerr
		}
		return false, nil
	}
	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		failed, err := executable(mid)
		if err != nil {
			return 0, err
		}
		if failed {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == gasCap {
		failed, err := executable(hi)
		if err != nil {
			return 0, err
		}
		if failed {
			if hi == maximumPossibleGas {
				return 0, fmt.Errorf("request might require more gas than it is allowed by the VM (%d)", gasCap)
			}
			// the specified gas cap is too low
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", gasCap)
		}
	}
	return hi, nil
}
