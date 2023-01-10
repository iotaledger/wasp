package transaction

import (
	"fmt"
	"math/big"

	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

var ErrNoAliasOutputAtIndex0 = xerrors.New("origin AliasOutput not found at index 0")

// GetAnchorFromTransaction analyzes the output at index 0 and extracts anchor information. Otherwise error
func GetAnchorFromTransaction(tx *iotago.Transaction) (*isc.StateAnchor, *iotago.AliasOutput, error) {
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

	if aliasID.Empty() {
		isOrigin = true
		aliasID = iotago.AliasIDFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(txid, 0))
	}
	return &isc.StateAnchor{
		IsOrigin:             isOrigin,
		OutputID:             iotago.OutputIDFromTransactionIDAndIndex(txid, 0),
		ChainID:              isc.ChainIDFromAliasID(aliasID),
		StateController:      anchorOutput.StateController(),
		GovernanceController: anchorOutput.GovernorAddress(),
		StateIndex:           anchorOutput.StateIndex,
		StateData:            anchorOutput.StateMetadata,
		Deposit:              anchorOutput.Amount,
	}, anchorOutput, nil
}

// computeInputsAndRemainder computes inputs and remainder for given outputs balances.
// Takes into account minimum storage deposit requirements
// The inputs are consumed one by one in the order provided in the parameters.
// Consumes only what is needed to cover output balances
// Returned reminder is nil if not needed
func computeInputsAndRemainder(
	senderAddress iotago.Address,
	baseTokenOut uint64,
	tokensOut map[iotago.NativeTokenID]*big.Int,
	nftsOut map[iotago.NFTID]bool,
	unspentOutputs iotago.OutputSet,
	unspentOutputIDs iotago.OutputIDs,
) (
	iotago.OutputIDs,
	*iotago.BasicOutput,
	error,
) {
	baseTokensIn := uint64(0)
	tokensIn := make(map[iotago.NativeTokenID]*big.Int)
	NFTsIn := make(map[iotago.NFTID]bool)

	var remainder *iotago.BasicOutput

	var errLast error

	var inputIDs iotago.OutputIDs

	for _, outputID := range unspentOutputIDs {
		output, ok := unspentOutputs[outputID]
		if !ok {
			return nil, nil, xerrors.New("computeInputsAndRemainder: outputID is not in the set ")
		}
		if nftOutput, ok := output.(*iotago.NFTOutput); ok {
			nftID := util.NFTIDFromNFTOutput(nftOutput, outputID)
			if nftsOut[nftID] {
				NFTsIn[nftID] = true
			} else {
				// this is an UTXO that holds an NFT that is not relevant for this tx, should be skipped
				continue
			}
		}
		if _, ok := output.(*iotago.AliasOutput); ok {
			// this is an UTXO that holds an alias that is not relevant for this tx, should be skipped
			continue
		}
		if _, ok := output.(*iotago.FoundryOutput); ok {
			// this is an UTXO that holds an foundry that is not relevant for this tx, should be skipped
			continue
		}
		if output.UnlockConditionSet().StorageDepositReturn() != nil {
			// don't consume anything with SDRUC
			continue
		}
		inputIDs = append(inputIDs, outputID)
		a := AssetsFromOutput(output)
		baseTokensIn += a.BaseTokens
		for _, nativeToken := range a.NativeTokens {
			nativeTokenAmountSum, ok := tokensIn[nativeToken.ID]
			if !ok {
				nativeTokenAmountSum = new(big.Int)
			}
			nativeTokenAmountSum.Add(nativeTokenAmountSum, nativeToken.Amount)
			tokensIn[nativeToken.ID] = nativeTokenAmountSum
		}
		// calculate remainder. It will return  err != nil if inputs not enough.
		remainder, errLast = computeRemainderOutput(senderAddress, baseTokensIn, baseTokenOut, tokensIn, tokensOut)
		if errLast == nil && len(NFTsIn) == len(nftsOut) {
			break
		}
	}
	if errLast != nil {
		return nil, nil, errLast
	}
	return inputIDs, remainder, nil
}

// computeRemainderOutput calculates remainders for base tokens and native tokens, returns skeleton remainder output
// which only contains assets filled in.
// - inBaseTokens and inTokens is what is available in inputs
// - outBaseTokens, outTokens is what is in outputs, except the remainder output itself with its storage deposit
// Returns (nil, error) if inputs are not enough (taking into account storage deposit requirements)
// If return (nil, nil) it means remainder is a perfect match between inputs and outputs, remainder not needed
//

//nolint:gocyclo
func computeRemainderOutput(senderAddress iotago.Address, inBaseTokens, outBaseTokens uint64, inTokens, outTokens map[iotago.NativeTokenID]*big.Int) (*iotago.BasicOutput, error) {
	if inBaseTokens < outBaseTokens {
		return nil, ErrNotEnoughBaseTokens
	}
	// collect all token ids
	nativeTokenIDs := make(map[iotago.NativeTokenID]bool)
	for id := range inTokens {
		nativeTokenIDs[id] = true
	}
	for id := range outTokens {
		nativeTokenIDs[id] = true
	}
	remBaseTokens := inBaseTokens - outBaseTokens
	remTokens := make(map[iotago.NativeTokenID]*big.Int)

	// calc remainders by outputs
	for nativeTokenID := range nativeTokenIDs {
		bIn, okIn := inTokens[nativeTokenID]
		bOut, okOut := outTokens[nativeTokenID]
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
			if !util.IsZeroBigInt(s) {
				remTokens[nativeTokenID] = s
			}
		case !okIn && okOut:
			// there's output but no input. Not enough
			return nil, ErrNotEnoughNativeTokens
		case okIn && !okOut:
			// native token is here by accident. All goes to remainder
			remTokens[nativeTokenID] = new(big.Int).Set(bIn)
			if util.IsZeroBigInt(bIn) {
				panic("bad input")
			}
		default:
			panic("inconsistency")
		}
	}
	if remBaseTokens == 0 && len(remTokens) == 0 {
		// no need for remainder
		return nil, nil
	}
	ret := &iotago.BasicOutput{
		Amount:       remBaseTokens,
		NativeTokens: iotago.NativeTokens{},
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: senderAddress},
		},
	}
	for nativeTokenID, b := range remTokens {
		ret.NativeTokens = append(ret.NativeTokens, &iotago.NativeToken{
			ID:     nativeTokenID,
			Amount: b,
		})
	}
	storageDeposit := parameters.L1().Protocol.RentStructure.MinRent(ret)
	if ret.Amount < storageDeposit {
		return nil, xerrors.Errorf("%v: needed at least %d", ErrNotEnoughBaseTokensForStorageDeposit, storageDeposit)
	}
	return ret, nil
}

func MakeSignatureAndReferenceUnlocks(totalInputs int, sig iotago.Signature) iotago.Unlocks {
	ret := make(iotago.Unlocks, totalInputs)
	for i := range ret {
		if i == 0 {
			ret[0] = &iotago.SignatureUnlock{Signature: sig}
			continue
		}
		ret[i] = &iotago.ReferenceUnlock{Reference: 0}
	}
	return ret
}

func MakeSignatureAndAliasUnlockFeatures(totalInputs int, sig iotago.Signature) iotago.Unlocks {
	ret := make(iotago.Unlocks, totalInputs)
	for i := range ret {
		if i == 0 {
			ret[0] = &iotago.SignatureUnlock{Signature: sig}
			continue
		}
		ret[i] = &iotago.AliasUnlock{Reference: 0}
	}
	return ret
}

func MakeAnchorTransaction(essence *iotago.TransactionEssence, sig iotago.Signature) *iotago.Transaction {
	return &iotago.Transaction{
		Essence: essence,
		Unlocks: MakeSignatureAndAliasUnlockFeatures(len(essence.Inputs), sig),
	}
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
		Essence: essence,
		Unlocks: MakeSignatureAndReferenceUnlocks(len(inputs), sigs[0]),
	}, nil
}

func GetAliasOutput(tx *iotago.Transaction, aliasAddr iotago.Address) (*isc.AliasOutputWithID, error) {
	txID, err := tx.ID()
	if err != nil {
		return nil, err
	}

	for index, output := range tx.Essence.Outputs {
		if aliasOutput, ok := output.(*iotago.AliasOutput); ok {
			outputID := iotago.OutputIDFromTransactionIDAndIndex(txID, uint16(index))

			aliasID := aliasOutput.AliasID
			if aliasID.Empty() {
				aliasID = iotago.AliasIDFromOutputID(outputID)
			}

			if aliasID.ToAddress().Equal(aliasAddr) {
				// output found
				return isc.NewAliasOutputWithID(aliasOutput, outputID), nil
			}
		}
	}

	return nil, fmt.Errorf("cannot find alias output for address %v in transaction", aliasAddr.String())
}
