package transaction

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

var (
	ErrNoAliasOutputAtIndex0 = xerrors.New("origin AliasOutput not found at index 0")
	nilAliasID               iotago.AliasID
)

type OutputFilter func(output iotago.Output) bool

// FilterOutputIndices returns a slice of indices of those outputs which satisfy all predicates
// Uses same underlying arrays slices
func FilterOutputIndices(outputs []iotago.Output, ids []*iotago.UTXOInput, filters ...OutputFilter) ([]iotago.Output, []*iotago.UTXOInput) {
	if len(outputs) != len(ids) {
		panic("FilterOutputIndices: number of outputs must be equal to the number of IDs")
	}
	ret := outputs[:0]
	retIDs := ids[:0]

	for i, out := range outputs {
		satisfyAll := true
		for _, f := range filters {
			if !f(out) {
				satisfyAll = false
				break
			}
		}
		if satisfyAll {
			ret = append(ret, out)
			retIDs = append(retIDs, ids[i])
		}
	}
	return ret, retIDs
}

func FilterType(t iotago.OutputType) OutputFilter {
	return func(out iotago.Output) bool {
		return out.Type() == t
	}
}

// GetAnchorFromTransaction analyzes the output at index 0 and extracts anchor information. Otherwise error
func GetAnchorFromTransaction(tx *iotago.Transaction) (*iscp.StateAnchor, *iotago.AliasOutput, error) {
	anchorOutput, ok := tx.Essence.Outputs[0].(*iotago.AliasOutput)
	if !ok {
		return nil, nil, ErrNoAliasOutputAtIndex0
	}
	txid, err := tx.ID()
	if err != nil {
		return nil, anchorOutput, xerrors.Errorf("GetAnchorFromTransaction: %w", err)
	}
	aliasID := anchorOutput.AliasID
	isOrigin := false

	if aliasID == nilAliasID {
		isOrigin = true
		aliasID = iotago.AliasIDFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(*txid, 0))
	}
	sd, err := iscp.StateDataFromBytes(anchorOutput.StateMetadata)
	if err != nil {
		return nil, anchorOutput, err
	}
	return &iscp.StateAnchor{
		IsOrigin:             isOrigin,
		OutputID:             iotago.OutputIDFromTransactionIDAndIndex(*txid, 0),
		ChainID:              iscp.ChainIDFromAliasID(aliasID),
		StateController:      anchorOutput.StateController(),
		GovernanceController: anchorOutput.GovernorAddress(),
		StateIndex:           anchorOutput.StateIndex,
		StateData:            sd,
		Deposit:              anchorOutput.Amount,
	}, anchorOutput, nil
}

// computeInputsAndRemainder computes inputs and remainder for given outputs balances.
// Takes into account minimum dust deposit requirements
// The inputs are consumed one by one in the order provided in the parameters.
// Consumes only what is needed to cover output balances
// Returned reminder is nil if not needed
func computeInputsAndRemainder(
	senderAddress iotago.Address,
	iotasOut uint64,
	tokensOut map[iotago.NativeTokenID]*big.Int,
	nftsOut map[*iotago.NFTID]bool,
	unspentOutputs iotago.OutputSet,
	unspentOutputIDs iotago.OutputIDs,
	rentStructure *iotago.RentStructure,
) (
	iotago.OutputIDs,
	*iotago.BasicOutput,
	error,
) {
	iotasIn := uint64(0)
	tokensIn := make(map[iotago.NativeTokenID]*big.Int)
	NFTsIn := make(map[*iotago.NFTID]bool)

	var remainder *iotago.BasicOutput

	var errLast error
	var inputCount int
	for _, id := range unspentOutputIDs {
		inp, ok := unspentOutputs[id]
		if !ok {
			return nil, nil, xerrors.New("computeInputsAndRemainder: outputID is not in the set ")
		}
		inputCount++
		if nftInp, ok := inp.(*iotago.NFTOutput); ok {
			nftID := util.NFTIDFromNFTOutput(nftInp, id)
			if nftsOut[&nftID] {
				NFTsIn[&nftID] = true
			}
		}
		a := AssetsFromOutput(inp)
		iotasIn += a.Iotas
		for _, nativeToken := range a.Tokens {
			nativeTokenAmountSum, ok := tokensIn[nativeToken.ID]
			if !ok {
				nativeTokenAmountSum = new(big.Int)
			}
			nativeTokenAmountSum.Add(nativeTokenAmountSum, nativeToken.Amount)
			tokensIn[nativeToken.ID] = nativeTokenAmountSum
		}
		// calculate remainder. It will return  err != nil if inputs not enough.
		remainder, errLast = computeRemainderOutput(senderAddress, iotasIn, iotasOut, tokensIn, tokensOut, rentStructure)
		if errLast == nil && len(NFTsIn) == len(nftsOut) {
			break
		}
	}
	if errLast != nil {
		return nil, nil, errLast
	}
	inputs := make(iotago.OutputIDs, inputCount)
	for j := range inputs {
		inputs[j] = unspentOutputIDs[j]
	}
	return inputs, remainder, nil
}

// computeRemainderOutput calculates remainders for iotas and native tokens, returns skeleton remainder output
// which only contains assets filled in.
// - inIotas and inTokens is what is available in inputs
// - outIotas, outTokens is what is in outputs, except the remainder output itself with its dust deposit
// Returns (nil, error) if inputs are not enough (taking into account dust deposit requirements)
// If return (nil, nil) it means remainder is a perfect match between inputs and outputs, remainder not needed
func computeRemainderOutput(senderAddress iotago.Address, inIotas, outIotas uint64, inTokens, outTokens map[iotago.NativeTokenID]*big.Int, rentStructure *iotago.RentStructure) (*iotago.BasicOutput, error) {
	if inIotas < outIotas {
		return nil, ErrNotEnoughIotas
	}
	// collect all token ids
	tokenIDs := make(map[iotago.NativeTokenID]bool)
	for id := range inTokens {
		tokenIDs[id] = true
	}
	for id := range outTokens {
		tokenIDs[id] = true
	}
	remIotas := inIotas - outIotas
	remTokens := make(map[iotago.NativeTokenID]*big.Int)

	// calc remainders by outputs
	for id := range tokenIDs {
		bIn, okIn := inTokens[id]
		bOut, okOut := outTokens[id]
		if !okIn {
			return nil, ErrNotEnoughNativeTokens
		}
		switch {
		case okIn && okOut:
			// there are tokens in inputs and outputs. Check if it is enough
			if bIn.Cmp(bOut) < 0 {
				// not enough
				return nil, ErrNotEnoughNativeTokens
			}
			// bIn >= bOut
			s := new(big.Int).Sub(bIn, bOut)
			if !util.IsZeroBigInt(bIn) {
				remTokens[id] = s
			}
		case !okIn && okOut:
			// there's output but no input. Not enough
			return nil, ErrNotEnoughNativeTokens
		case okIn && !okOut:
			// native token is here by accident. All goes to remainder
			remTokens[id] = new(big.Int).Set(bIn)
			if util.IsZeroBigInt(bIn) {
				panic("bad input")
			}
		default:
			panic("inconsistency")
		}
	}
	if remIotas == 0 && len(remTokens) == 0 {
		// no need for remainder
		return nil, nil
	}
	ret := &iotago.BasicOutput{
		Amount:       remIotas,
		NativeTokens: iotago.NativeTokens{},
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: senderAddress},
		},
	}
	for id, b := range remTokens {
		ret.NativeTokens = append(ret.NativeTokens, &iotago.NativeToken{
			ID:     id,
			Amount: b,
		})
	}
	bc := ret.VByteCost(rentStructure, nil)
	if ret.Amount < bc {
		return nil, xerrors.Errorf("%v: needed at least %d", ErrNotEnoughIotasForDustDeposit, bc)
	}
	return ret, nil
}

func MakeSignatureAndReferenceUnlockBlocks(totalInputs int, sig iotago.Signature) iotago.UnlockBlocks {
	ret := make(iotago.UnlockBlocks, totalInputs)
	for i := range ret {
		if i == 0 {
			ret[0] = &iotago.SignatureUnlockBlock{Signature: sig}
			continue
		}
		ret[i] = &iotago.ReferenceUnlockBlock{Reference: 0}
	}
	return ret
}

func MakeSignatureAndAliasUnlockBlocks(totalInputs int, sig iotago.Signature) iotago.UnlockBlocks {
	ret := make(iotago.UnlockBlocks, totalInputs)
	for i := range ret {
		if i == 0 {
			ret[0] = &iotago.SignatureUnlockBlock{Signature: sig}
			continue
		}
		ret[i] = &iotago.AliasUnlockBlock{Reference: 0}
	}
	return ret
}

func MakeAnchorTransaction(essence *iotago.TransactionEssence, sig iotago.Signature) *iotago.Transaction {
	return &iotago.Transaction{
		Essence:      essence,
		UnlockBlocks: MakeSignatureAndAliasUnlockBlocks(len(essence.Inputs), sig),
	}
}

func GetVByteCosts(tx *iotago.Transaction, rentStructure *iotago.RentStructure) []uint64 {
	ret := make([]uint64, len(tx.Essence.Outputs))
	for i, out := range tx.Essence.Outputs {
		ret[i] = out.VByteCost(rentStructure, nil)
	}
	return ret
}

func CreateAndSignTx(inputs iotago.OutputIDs, inputsCommitment []byte, outputs iotago.Outputs, wallet *cryptolib.KeyPair, networkID uint64) (*iotago.Transaction, error) {
	essence := &iotago.TransactionEssence{
		NetworkID: networkID,
		Inputs:    inputs.UTXOInputs(),
		Outputs:   outputs,
	}

	sigs, err := essence.Sign(
		inputsCommitment,
		wallet.GetPrivateKey().AddressKeysForEd25519Address(wallet.Address()),
	)
	if err != nil {
		return nil, err
	}

	return &iotago.Transaction{
		Essence:      essence,
		UnlockBlocks: MakeSignatureAndReferenceUnlockBlocks(len(inputs), sigs[0]),
	}, nil
}
