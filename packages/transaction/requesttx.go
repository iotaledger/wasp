package transaction

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
)

type NewRequestTransactionParams struct {
	SenderKeyPair    cryptolib.KeyPair
	UnspentOutputs   []iotago.Output
	UnspentOutputIDs []*iotago.UTXOInput
	Requests         []*iscp.RequestParameters
	RentStructure    *iotago.RentStructure
}

// NewRequestTransaction creates a transaction including one or more requests to a chain.
// Empty assets in the request data defaults to 1 iota, which later is adjusted to the dust minimum
// Assumes all UnspentOutputs and corresponding UnspentOutputIDs can be used as inputs, i.e. are
// unlockable for the sender address
func NewRequestTransaction(par NewRequestTransactionParams) (*iotago.Transaction, error) {
	outputs := iotago.Outputs{}
	sumIotasOut := uint64(0)
	sumTokensOut := make(map[iotago.NativeTokenID]*big.Int)

	senderAddress := cryptolib.Ed25519AddressFromPubKey(par.SenderKeyPair.PublicKey)

	// create outputs, sum totals needed
	for _, req := range par.Requests {
		assets := req.Assets
		if assets == nil {
			// if assets not specified, the minimum dust deposit will be adjusted by vmtxbuilder.MakeExtendedOutput
			assets = &iscp.Assets{}
		}
		// will adjust to minimum dust deposit
		out, _ := vmtxbuilder.MakeExtendedOutput(
			req.TargetAddress,
			senderAddress,
			assets,
			&iscp.RequestMetadata{
				SenderContract: 0,
				TargetContract: req.Metadata.TargetContract,
				EntryPoint:     req.Metadata.EntryPoint,
				Params:         req.Metadata.Params,
				Transfer:       req.Metadata.Transfer,
				GasBudget:      req.Metadata.GasBudget,
			},
			req.Options,
			par.RentStructure,
		)
		outputs = append(outputs, out)
		sumIotasOut += out.Amount
		for _, nt := range out.NativeTokens {
			s, ok := sumTokensOut[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			sumTokensOut[nt.ID] = s
		}
	}
	inputs, remainder, err := computeInputsAndRemainder(senderAddress, sumIotasOut, sumTokensOut, par.UnspentOutputs, par.UnspentOutputIDs, par.RentStructure)
	if err != nil {
		return nil, err
	}
	if remainder.Amount > 0 {
		outputs = append(outputs, remainder)
	}
	essence := &iotago.TransactionEssence{
		Inputs:  inputs,
		Outputs: outputs,
	}
	sigs, err := essence.Sign(iotago.NewAddressKeysForEd25519Address(senderAddress, par.SenderKeyPair.PrivateKey))
	if err != nil {
		return nil, err
	}

	return &iotago.Transaction{
		Essence:      essence,
		UnlockBlocks: MakeSignatureAndReferenceUnlockBlocks(len(inputs), sigs[0]),
	}, nil
}
