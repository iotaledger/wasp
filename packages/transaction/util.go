package transaction

import (
	"fmt"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
)

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
		remainder = computeReminderOutput(senderAddress, iotasIn, iotasOut, tokensIn, tokensOut, deSeriParams)
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

// computeReminderOutput calculates reminders for iotas and native tokens, returns skeleton reminder output
// which only contains assets filled in.
// - inIotas and inTokens is what is available in inputs
// - outIotas, outTokens is what is in outputs, except the reminder output itself with its dust deposit
// Returns nil if inputs are not enough (taking into account dust deposit requirements)
// If return not nil but Amount == 0, ite means reminder is a perfect match between inputs and outputs, remainder not needed
func computeReminderOutput(senderAddress iotago.Address, inIotas, outIotas uint64, inTokens, outTokens map[iotago.NativeTokenID]*big.Int, deSeriParams *iotago.DeSerializationParameters) *iotago.ExtendedOutput {
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

func computeInputsAndRemainderOld(
	amount uint64,
	allUnspentOutputs []iotago.Output,
	allInputs []*iotago.UTXOInput,
	deSeriParams *iotago.DeSerializationParameters,
) ([]*iotago.UTXOInput, uint64, error) {
	remainderDustDeposit := (&iotago.ExtendedOutput{}).VByteCost(deSeriParams.RentStructure, nil)
	var inputs []*iotago.UTXOInput
	consumed := uint64(0)
	for i, out := range allUnspentOutputs {
		consumed += out.Deposit()
		inputs = append(inputs, allInputs[i])
		if consumed == amount {
			return inputs, 0, nil
		}
		if consumed > amount {
			remainder := amount - consumed
			if remainder >= remainderDustDeposit {
				return inputs, remainder, nil
			}
		}
	}
	return nil, 0, fmt.Errorf("insufficient funds")
}

//// GetAliasOutput return output or nil if not found
//func GetAliasOutput(tx *ledgerstate.Transaction, aliasAddr ledgerstate.Address) *ledgerstate.AliasOutput {
//	return GetAliasOutputFromEssence(tx.Essence(), aliasAddr)
//}
//
//func GetAliasOutputFromEssence(essence *ledgerstate.TransactionEssence, aliasAddr ledgerstate.Address) *ledgerstate.AliasOutput {
//	for _, o := range essence.Outputs() {
//		if out, ok := o.(*ledgerstate.AliasOutput); ok {
//			out1 := out.UpdateMintingColor().(*ledgerstate.AliasOutput)
//			if out1.GetAliasAddress().Equals(aliasAddr) {
//				return out1
//			}
//		}
//	}
//	return nil
//}
