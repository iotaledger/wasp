package transaction

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
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
func GetAnchorFromTransaction(tx *iotago.Transaction) (*iscp.StateAnchor, error) {
	anchorOutput, ok := tx.Essence.Outputs[0].(*iotago.AliasOutput)
	if !ok {
		return nil, ErrNoAliasOutputAtIndex0
	}
	txid, err := tx.ID()
	if err != nil {
		return nil, xerrors.Errorf("GetAnchorFromTransaction: %w", err)
	}
	aliasID := anchorOutput.AliasID
	isOrigin := false

	if aliasID == nilAliasID {
		isOrigin = true
		aliasID = iotago.AliasIDFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(*txid, 0))
	}
	sd, err := iscp.StateDataFromBytes(anchorOutput.StateMetadata)
	if err != nil {
		return nil, err
	}
	return &iscp.StateAnchor{
		IsOrigin:             isOrigin,
		OutputID:             iotago.OutputIDFromTransactionIDAndIndex(*txid, 0),
		ChainID:              iscp.ChainIDFromAliasID(aliasID),
		StateController:      anchorOutput.StateController,
		GovernanceController: anchorOutput.GovernanceController,
		StateIndex:           anchorOutput.StateIndex,
		StateData:            sd,
		Deposit:              anchorOutput.Amount,
	}, nil
}

// computeInputsAndRemainder computes inputs and reminder for given outputs balances.
// Takes into account minimum dust deposit requirements
// The inputs are consumed one by one in the order provided in the parameters.
// Consumes only what is needed to cover output balances
func computeInputsAndRemainder(
	senderAddress iotago.Address,
	iotasOut uint64,
	tokensOut map[iotago.NativeTokenID]*big.Int,
	allUnspentOutputs []iotago.Output,
	allInputs []*iotago.UTXOInput,
	deSeriParams *iotago.DeSerializationParameters,
) ([]*iotago.UTXOInput, *iotago.ExtendedOutput, error) {
	iotasIn := uint64(0)
	tokensIn := make(map[iotago.NativeTokenID]*big.Int)

	var remainder *iotago.ExtendedOutput
	var inputs []*iotago.UTXOInput

	for i, inp := range allUnspentOutputs {
		a := vmtxbuilder.AssetsFromOutput(inp)
		iotasIn += a.Iotas
		for _, nt := range a.Tokens {
			s, ok := tokensIn[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			tokensIn[nt.ID] = s
		}
		// calculate reminder. It will return != nil if input balances is enough, otherwise nil
		remainder = computeRemainderOutput(senderAddress, iotasIn, iotasOut, tokensIn, tokensOut, deSeriParams)
		if remainder != nil {
			inputs = allInputs[:i+1]
			break
		}
	}
	if remainder == nil {
		// not enough inputs to cover outputs
		return nil, nil, accounts.ErrNotEnoughFunds
	}
	return inputs, remainder, nil
}

// computeRemainderOutput calculates reminders for iotas and native tokens, returns skeleton reminder output
// which only contains assets filled in.
// - inIotas and inTokens is what is available in inputs
// - outIotas, outTokens is what is in outputs, except the reminder output itself with its dust deposit
// Returns nil if inputs are not enough (taking into account dust deposit requirements)
// If return not nil but Amount == 0, ite means reminder is a perfect match between inputs and outputs, remainder not needed
func computeRemainderOutput(senderAddress iotago.Address, inIotas, outIotas uint64, inTokens, outTokens map[iotago.NativeTokenID]*big.Int, deSeriParams *iotago.DeSerializationParameters) *iotago.ExtendedOutput {
	if inIotas < outIotas {
		return nil
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

	// calc reminders by outputs
	for id := range tokenIDs {
		bIn, okIn := inTokens[id]
		bOut, okOut := outTokens[id]
		if !okIn {
			return nil
		}
		switch {
		case okIn && okOut:
			// there are tokens in inputs and outputs. Check if it is enough
			if bIn.Cmp(bOut) < 0 {
				// not enough
				return nil
			}
			// bIn >= bOut
			s := new(big.Int).Sub(bIn, bOut)
			if !util.IsZeroBigInt(bIn) {
				remTokens[id] = s
			}
		case !okIn && okOut:
			// there's output but no input. Not enough
			return nil
		case okIn && !okOut:
			// native token is here by accident. All goes to reminder
			remTokens[id] = new(big.Int).Set(bIn)
			if util.IsZeroBigInt(bIn) {
				panic("bad input")
			}
		default:
			panic("inconsistency")
		}
	}
	ret := &iotago.ExtendedOutput{
		Address:      senderAddress,
		Amount:       remIotas,
		NativeTokens: iotago.NativeTokens{},
		Blocks:       nil,
	}
	if remIotas == 0 && len(remTokens) == 0 {
		// no need for remainder
		return ret
	}
	for id, b := range remTokens {
		ret.NativeTokens = append(ret.NativeTokens, &iotago.NativeToken{
			ID:     id,
			Amount: b,
		})
	}
	if len(remTokens) > 0 && remIotas < ret.VByteCost(deSeriParams.RentStructure, nil) {
		// iotas does not cover minimum dust requirements
		return nil
	}
	return ret
}

//func computeInputsAndRemainderOld(
//	amount uint64,
//	allUnspentOutputs []iotago.Output,
//	allInputs []*iotago.UTXOInput,
//	deSeriParams *iotago.DeSerializationParameters,
//) ([]*iotago.UTXOInput, uint64, error) {
//	remainderDustDeposit := (&iotago.ExtendedOutput{}).VByteCost(deSeriParams.RentStructure, nil)
//	var inputs []*iotago.UTXOInput
//	consumed := uint64(0)
//	for i, out := range allUnspentOutputs {
//		consumed += out.Deposit()
//		inputs = append(inputs, allInputs[i])
//		if consumed == amount {
//			return inputs, 0, nil
//		}
//		if consumed > amount {
//			remainder := amount - consumed
//			if remainder >= remainderDustDeposit {
//				return inputs, remainder, nil
//			}
//		}
//	}
//	return nil, 0, fmt.Errorf("insufficient funds")
//}
//
//// GetAliasOutput return output or nil if not found
//func GetAliasOutput(tx *ledgerstate.Transaction, aliasAddr ledgerstate.Address) *ledgerstate.AliasOutput {
//	return GetAliasOutputFromEssence(tx.Essence(), aliasAddr)
//}
//
