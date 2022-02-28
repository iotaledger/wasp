package transaction

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

type NewRequestTransactionParams struct {
	SenderKeyPair                *cryptolib.KeyPair
	UnspentOutputs               iotago.OutputSet
	UnspentOutputIDs             iotago.OutputIDs
	Requests                     []*iscp.RequestParameters
	RentStructure                *iotago.RentStructure
	DisableAutoAdjustDustDeposit bool // if true, the minimal dust deposit won't be adjusted automatically
}

// NewRequestTransaction creates a transaction including one or more requests to a chain.
// Empty assets in the request data defaults to 1 iota, which later is adjusted to the dust minimum
// Assumes all UnspentOutputs and corresponding UnspentOutputIDs can be used as inputs, i.e. are
// unlockable for the sender address
func NewRequestTransaction(par NewRequestTransactionParams) (*iotago.Transaction, error) {
	outputs := iotago.Outputs{}
	sumIotasOut := uint64(0)
	sumTokensOut := make(map[iotago.NativeTokenID]*big.Int)
	sumNFTsOut := make([]*iotago.NFTID, 0)

	senderAddress := par.SenderKeyPair.Address()

	// create outputs, sum totals needed
	for _, req := range par.Requests {
		assets := req.Assets
		if assets == nil {
			// if assets not specified, the minimum dust deposit will be adjusted by vmtxbuilder.MakeBasicOutput
			assets = &iscp.Assets{}
		}
		// will adjust to minimum dust deposit
		out := MakeOutput(
			req.TargetAddress,
			senderAddress,
			assets,
			&iscp.RequestMetadata{
				SenderContract: 0,
				TargetContract: req.Metadata.TargetContract,
				EntryPoint:     req.Metadata.EntryPoint,
				Params:         req.Metadata.Params,
				Allowance:      req.Metadata.Allowance,
				GasBudget:      req.Metadata.GasBudget,
			},
			req.NFTID,
			req.Options,
			par.RentStructure,
			par.DisableAutoAdjustDustDeposit,
		)
		requiredDustDeposit := out.VByteCost(par.RentStructure, nil)
		if out.Deposit() < requiredDustDeposit {
			return nil, xerrors.Errorf("%v: available %d < required %d iotas",
				ErrNotEnoughIotasForDustDeposit, out.Deposit(), requiredDustDeposit)
		}
		outputs = append(outputs, out)
		sumIotasOut += out.Deposit()
		for _, nt := range out.NativeTokenSet() {
			s, ok := sumTokensOut[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			sumTokensOut[nt.ID] = s
		}

		if req.NFTID != nil {
			sumNFTsOut = append(sumNFTsOut, req.NFTID)
		}
	}
	// TODO needs refactoring - so that computeInputsAndRemainder includes the correct NFT inputs
	inputIDs, remainder, err := computeInputsAndRemainder(senderAddress, sumIotasOut, sumTokensOut, sumNFTsOut, par.UnspentOutputs, par.UnspentOutputIDs, par.RentStructure)
	if err != nil {
		return nil, err
	}
	if remainder != nil {
		outputs = append(outputs, remainder)
	}

	inputsCommitment := inputIDs.OrderedSet(par.UnspentOutputs).MustCommitment()
	return CreateAndSignTx(inputIDs, inputsCommitment, outputs, par.SenderKeyPair)
}
