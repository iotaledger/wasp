package transaction

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

type NewRequestTransactionParams struct {
	SenderKeyPair                   *cryptolib.KeyPair
	SenderAddress                   iotago.Address // might be different from the senderKP address (when sending as NFT or alias)
	UnspentOutputs                  iotago.OutputSet
	UnspentOutputIDs                iotago.OutputIDs
	Request                         *isc.RequestParameters
	NFT                             *isc.NFT
	DisableAutoAdjustStorageDeposit bool // if true, the minimal storage deposit won't be adjusted automatically
}

type NewTransferTransactionParams struct {
	DisableAutoAdjustStorageDeposit bool // if true, the minimal storage deposit won't be adjusted automatically
	FungibleTokens                  *isc.Assets
	SendOptions                     isc.SendOptions
	SenderAddress                   iotago.Address
	SenderKeyPair                   *cryptolib.KeyPair
	TargetAddress                   iotago.Address
	UnspentOutputs                  iotago.OutputSet
	UnspentOutputIDs                iotago.OutputIDs
}

// NewTransferTransaction creates a basic output transaction that sends L1 Token to another L1 address
func NewTransferTransaction(params NewTransferTransactionParams) (*iotago.Transaction, error) {
	output := MakeBasicOutput(
		params.TargetAddress,
		params.SenderAddress,
		params.FungibleTokens,
		nil,
		params.SendOptions,
	)
	if !params.DisableAutoAdjustStorageDeposit {
		output = AdjustToMinimumStorageDeposit(output)
	}

	storageDeposit := parameters.L1().Protocol.RentStructure.MinRent(output)
	if output.Deposit() < storageDeposit {
		return nil, fmt.Errorf("%v: available %d < required %d base tokens",
			ErrNotEnoughBaseTokensForStorageDeposit, output.Deposit(), storageDeposit)
	}

	sumBaseTokensOut := output.Deposit()
	sumTokensOut := make(map[iotago.NativeTokenID]*big.Int)
	sumTokensOut = addNativeTokens(sumTokensOut, output)

	tokenMap := map[iotago.NativeTokenID]*big.Int{}
	for _, nativeToken := range params.FungibleTokens.NativeTokens {
		tokenMap[nativeToken.ID] = nativeToken.Amount
	}

	inputIDs, remainder, err := computeInputsAndRemainder(params.SenderAddress,
		sumBaseTokensOut,
		sumTokensOut,
		map[iotago.NFTID]bool{},
		params.UnspentOutputs,
		params.UnspentOutputIDs,
	)
	if err != nil {
		return nil, err
	}

	outputs := []iotago.Output{output}

	if remainder != nil {
		outputs = append(outputs, remainder)
	}

	inputsCommitment := inputIDs.OrderedSet(params.UnspentOutputs).MustCommitment()

	return CreateAndSignTx(inputIDs, inputsCommitment, outputs, params.SenderKeyPair, parameters.L1().Protocol.NetworkID())
}

// NewRequestTransaction creates a transaction including one or more requests to a chain.
// Empty assets in the request data defaults to 1 base token, which later is adjusted to the minimum storage deposit
// Assumes all UnspentOutputs and corresponding UnspentOutputIDs can be used as inputs, i.e. are
// unlockable for the sender address
func NewRequestTransaction(par NewRequestTransactionParams) (*iotago.Transaction, error) {
	outputs := iotago.Outputs{}
	sumBaseTokensOut := uint64(0)
	sumTokensOut := make(map[iotago.NativeTokenID]*big.Int)
	sumNFTsOut := make(map[iotago.NFTID]bool)

	req := par.Request

	// create outputs, sum totals needed
	assets := req.Assets
	if assets == nil {
		// if assets not specified, the minimum storage deposit will be adjusted by vmtxbuilder.MakeBasicOutput
		assets = &isc.Assets{}
	}
	var out iotago.Output
	// will adjust to minimum storage deposit
	out = MakeBasicOutput(
		req.TargetAddress,
		par.SenderAddress,
		assets,
		&isc.RequestMetadata{
			SenderContract: 0,
			TargetContract: req.Metadata.TargetContract,
			EntryPoint:     req.Metadata.EntryPoint,
			Params:         req.Metadata.Params,
			Allowance:      req.Metadata.Allowance,
			GasBudget:      req.Metadata.GasBudget,
		},
		req.Options,
	)
	if par.NFT != nil {
		out = NftOutputFromBasicOutput(out.(*iotago.BasicOutput), par.NFT)
	}
	if !par.DisableAutoAdjustStorageDeposit {
		out = AdjustToMinimumStorageDeposit(out)
	}

	storageDeposit := parameters.L1().Protocol.RentStructure.MinRent(out)
	if out.Deposit() < storageDeposit {
		return nil, fmt.Errorf("%v: available %d < required %d base tokens",
			ErrNotEnoughBaseTokensForStorageDeposit, out.Deposit(), storageDeposit)
	}
	outputs = append(outputs, out)
	sumBaseTokensOut += out.Deposit()
	sumTokensOut = addNativeTokens(sumTokensOut, out)
	if par.NFT != nil {
		sumNFTsOut[par.NFT.ID] = true
	}

	outputs, sumBaseTokensOut, sumTokensOut, sumNFTsOut = updateOutputsWhenSendingOnBehalfOf(par, outputs, sumBaseTokensOut, sumTokensOut, sumNFTsOut)

	inputIDs, remainder, err := computeInputsAndRemainder(par.SenderKeyPair.Address(), sumBaseTokensOut, sumTokensOut, sumNFTsOut, par.UnspentOutputs, par.UnspentOutputIDs)
	if err != nil {
		return nil, err
	}

	if remainder != nil {
		outputs = append(outputs, remainder)
	}

	inputsCommitment := inputIDs.OrderedSet(par.UnspentOutputs).MustCommitment()
	return CreateAndSignTx(inputIDs, inputsCommitment, outputs, par.SenderKeyPair, parameters.L1().Protocol.NetworkID())
}

func outputMatchesSendAsAddress(output iotago.Output, outputID iotago.OutputID, address iotago.Address) bool {
	switch o := output.(type) {
	case *iotago.NFTOutput:
		if address.Equal(util.NFTIDFromNFTOutput(o, outputID).ToAddress()) {
			return true
		}
	case *iotago.AliasOutput:
		if address.Equal(o.AliasID.ToAddress()) {
			return true
		}
	}
	return false
}

func addNativeTokens(sumTokensOut map[iotago.NativeTokenID]*big.Int, out iotago.Output) map[iotago.NativeTokenID]*big.Int {
	for _, nt := range out.NativeTokenList() {
		s, ok := sumTokensOut[nt.ID]
		if !ok {
			s = new(big.Int)
		}
		s.Add(s, nt.Amount)
		sumTokensOut[nt.ID] = s
	}
	return sumTokensOut
}

func updateOutputsWhenSendingOnBehalfOf(
	par NewRequestTransactionParams,
	outputs iotago.Outputs,
	sumBaseTokensOut uint64,
	sumTokensOut map[iotago.NativeTokenID]*big.Int,
	sumNFTsOut map[iotago.NFTID]bool,
) (
	iotago.Outputs,
	uint64,
	map[iotago.NativeTokenID]*big.Int,
	map[iotago.NFTID]bool,
) {
	if par.SenderAddress.Equal(par.SenderKeyPair.Address()) {
		return outputs, sumBaseTokensOut, sumTokensOut, sumNFTsOut
	}
	// sending request "on behalf of" (need NFT or alias output as input/output)

	for _, output := range outputs {
		if outputMatchesSendAsAddress(output, iotago.OutputID{}, par.SenderAddress) {
			// if already present in the outputs, no need to do anything
			return outputs, sumBaseTokensOut, sumTokensOut, sumNFTsOut
		}
	}
	for outID, out := range par.UnspentOutputs {
		// find the output that matches the "send as" address
		if !outputMatchesSendAsAddress(out, outID, par.SenderAddress) {
			continue
		}
		if nftOut, ok := out.(*iotago.NFTOutput); ok {
			if nftOut.NFTID.Empty() {
				// if this is the first time the NFT output transitions, we need to fill the correct NFTID
				nftOut.NFTID = iotago.NFTIDFromOutputID(outID)
			}
			sumNFTsOut[nftOut.NFTID] = true
		}
		// found the needed output
		outputs = append(outputs, out)
		sumBaseTokensOut += out.Deposit()
		sumTokensOut = addNativeTokens(sumTokensOut, out)
		return outputs, sumBaseTokensOut, sumTokensOut, sumNFTsOut
	}
	panic("unable to build tx, 'sendAs' output not found")
}
