package transaction

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"golang.org/x/xerrors"
)

type NewRequestTransactionParams struct {
	SenderKeyPair                *cryptolib.KeyPair
	UnspentOutputs               iotago.OutputSet
	UnspentOutputIDs             iotago.OutputIDs
	Request                      *iscp.RequestParameters
	NFT                          *iscp.NFT
	L1                           *parameters.L1
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
	sumNFTsOut := make(map[iotago.NFTID]bool)

	senderAddress := par.SenderKeyPair.Address()

	req := par.Request

	// create outputs, sum totals needed
	assets := req.Assets
	if assets == nil {
		// if assets not specified, the minimum dust deposit will be adjusted by vmtxbuilder.MakeBasicOutput
		assets = &iscp.Assets{}
	}
	var out iotago.Output
	// will adjust to minimum dust deposit
	out = MakeBasicOutput(
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
		req.Options,
		par.L1.RentStructure(),
		par.DisableAutoAdjustDustDeposit,
	)
	if par.NFT != nil {
		out = NftOutputFromBasicOutput(out.(*iotago.BasicOutput), par.NFT)
	}

	requiredDustDeposit := out.VByteCost(par.L1.RentStructure(), nil)
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
	if par.NFT != nil {
		sumNFTsOut[par.NFT.ID] = true
	}

	inputIDs, remainder, err := computeInputsAndRemainder(senderAddress, sumIotasOut, sumTokensOut, sumNFTsOut, par.UnspentOutputs, par.UnspentOutputIDs, par.L1.RentStructure())
	if err != nil {
		return nil, err
	}
	if remainder != nil {
		outputs = append(outputs, remainder)
	}

	inputsCommitment := inputIDs.OrderedSet(par.UnspentOutputs).MustCommitment()
	return CreateAndSignTx(inputIDs, inputsCommitment, outputs, par.SenderKeyPair, par.L1.NetworkID)
}
