package chainutil

import (
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

func runISCTask(
	ch chain.ChainCore,
	aliasOutput *isc.AliasOutputWithID,
	blockTime time.Time,
	reqs []isc.Request,
	estimateGasMode bool,
	evmTracer *isc.EVMTracer,
) ([]*vm.RequestResult, error) {
	vmRunner := runvm.NewVMRunner()
	task := &vm.VMTask{
		Processors:           ch.Processors(),
		AnchorOutput:         aliasOutput.GetAliasOutput(),
		AnchorOutputID:       aliasOutput.OutputID(),
		Store:                ch.Store(),
		Requests:             reqs,
		TimeAssumption:       blockTime,
		Entropy:              hashing.PseudoRandomHash(nil),
		ValidatorFeeTarget:   isc.NewContractAgentID(ch.ID(), 0),
		EnableGasBurnLogging: false,
		EstimateGasMode:      estimateGasMode,
		EVMTracer:            evmTracer,
		Log:                  ch.Log().Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
	err := vmRunner.Run(task)
	if err != nil {
		return nil, err
	}
	return task.Results, nil
}

func runISCRequest(
	ch chain.ChainCore,
	aliasOutput *isc.AliasOutputWithID,
	blockTime time.Time,
	req isc.Request,
) (*vm.RequestResult, error) {
	results, err := runISCTask(
		ch,
		aliasOutput,
		blockTime,
		[]isc.Request{req},
		true,
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
