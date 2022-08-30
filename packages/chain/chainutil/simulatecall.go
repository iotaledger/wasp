package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// SimulateCall executes the given request and discards the resulting chain state. It is useful
// for estimating gas.
func SimulateCall(ch chain.Chain, req isc.Request) (*vm.RequestResult, error) {
	vmRunner := runvm.NewVMRunner()
	var ret *vm.RequestResult
	err := optimism.RetryOnStateInvalidated(func() (err error) {
		anchorOutput := ch.GetAnchorOutput()
		virtualStateAccess, ok, err := state.LoadSolidState(ch.GetDB(), ch.ID())
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
			TimeAssumption:     ch.GetTimeData(),
			VirtualStateAccess: virtualStateAccess,
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
