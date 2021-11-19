package vmtxbuilder

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
)

// implements transaction builder used internally by the VM during batch run

// TODO dust protection not covered yet !!!

type tokenBalanceLoader func(iotago.NativeTokenID) (*big.Int, iotago.UTXOInput)

type AnchorTransactionBuilder struct {
	loadNativeTokensOnChain tokenBalanceLoader
	// anchorOutput output of the chain
	anchorOutput *iotago.AliasOutput
	// already consumed Inputs
	consumedInputs []iotago.UTXOInput
	// balance of iotas is kept in the chain output
	balanceIotas uint64
	// balances of each native token is kept each in separate ExtendedOutput
	balanceNativeTokens map[iotago.NativeTokenID]*nativeTokenBalance
	// posted requests
	postedRequests []*iscp.PostRequestData
	// TODO
}

type nativeTokenBalance struct {
	input   iotago.UTXOInput
	initial *big.Int // if 0, it means output does not exists
	balance *big.Int
}

// error codes used for handled panics
var (
	ErrInputLimitExceeded  = xerrors.Errorf("exceeded maximum number of Inputs in transaction: %d", iotago.MaxInputsCount)
	ErrOutputLimitExceeded = xerrors.Errorf("exceeded maximum number of outputs in transaction: %d", iotago.MaxOutputsCount)
)

func (txb *AnchorTransactionBuilder) Clone() *AnchorTransactionBuilder {
	ret := NewAnchorTransactionBuilder(txb.anchorOutput, txb.consumedInputs[0], txb.balanceIotas, txb.loadNativeTokensOnChain)
	ret.consumedInputs = append(ret.consumedInputs, txb.consumedInputs...)
	ret.balanceIotas = txb.balanceIotas
	for k, v := range txb.balanceNativeTokens {
		ret.balanceNativeTokens[k] = &nativeTokenBalance{
			initial: new(big.Int).Set(v.balance),
			balance: new(big.Int).Set(v.balance),
		}
	}
	ret.postedRequests = append(ret.postedRequests, txb.postedRequests...)
	// TODO the rest
	return ret
}

func (txb *AnchorTransactionBuilder) addConsumedInput(inp iotago.UTXOInput) int {
	if txb.inputsAreFull() {
		panic(ErrInputLimitExceeded)
	}
	txb.consumedInputs = append(txb.consumedInputs, inp)
	return len(txb.consumedInputs) - 1
}

func (txb *AnchorTransactionBuilder) Inputs() iotago.Inputs {
	ret := make(iotago.Inputs, 0, len(txb.consumedInputs)+len(txb.balanceNativeTokens))
	for i := range txb.consumedInputs {
		ret = append(ret, &txb.consumedInputs[i])
	}
	// sort to avoid non-determinism of the map iteration
	tokenIDs := make([]iotago.NativeTokenID, 0, len(txb.balanceNativeTokens))
	for id := range txb.balanceNativeTokens {
		tokenIDs = append(tokenIDs, id)
	}
	sort.Slice(tokenIDs, func(i, j int) bool {
		return bytes.Compare(tokenIDs[i][:], tokenIDs[j][:]) < 0
	})
	for _, id := range tokenIDs {
		nt := txb.balanceNativeTokens[id]
		if nt.initial == nil {
			// entry didn't existed before
			continue
		}
		if nt.initial.Cmp(nt.balance) == 0 {
			// no need for input because nothing changed
		}
		ret = append(ret, &nt.input)
	}
	if len(ret) != txb.numInputsConsumed() {
		panic("AnchorTransactionBuilder.Inputs: internal inconsistency")
	}
	return ret
}

// sortedListOfInternalOutputs is needed to know exact position of each internal output for each token ID
// before it is stored into the state and state commitment is calculated
func (txb *AnchorTransactionBuilder) sortedListOfTokenIDsForOutputs() []iotago.NativeTokenID {
	ret := make([]iotago.NativeTokenID, 0, len(txb.balanceNativeTokens))
	for id, nt := range txb.balanceNativeTokens {
		if nt.initial != nil && nt.initial.Cmp(nt.balance) == 0 {
			continue
		}
		if util.IsZeroBigInt(nt.balance) {
			continue
		}
		ret = append(ret, id)
	}
	return ret
}

const dustAmountForInternalAccountUTXO = 100

func (txb *AnchorTransactionBuilder) Outputs(stateMetadata []byte) iotago.Outputs {
	ret := make(iotago.Outputs, 0, 1+len(txb.balanceNativeTokens)+len(txb.postedRequests))
	// creating chain output
	anchorOutput := &iotago.AliasOutput{
		Amount:               txb.balanceIotas,
		NativeTokens:         nil, // anchorOutput output does not contain native tokens
		AliasID:              txb.anchorOutput.AliasID,
		StateController:      txb.anchorOutput.StateController,
		GovernanceController: txb.anchorOutput.GovernanceController,
		StateIndex:           txb.anchorOutput.StateIndex + 1,
		StateMetadata:        stateMetadata,
		FoundryCounter:       txb.anchorOutput.FoundryCounter, // TODO
		Blocks:               nil,                             // TODO ??
	}
	ret = append(ret, anchorOutput)

	// creating outputs for updated internal accounts
	tokenIdsToBeUpdated := txb.sortedListOfTokenIDsForOutputs()

	for _, id := range tokenIdsToBeUpdated {
		o := &iotago.ExtendedOutput{
			Address: txb.anchorOutput.AliasID.ToAddress(),
			Amount:  dustAmountForInternalAccountUTXO,
			NativeTokens: iotago.NativeTokens{&iotago.NativeToken{
				ID:     id,
				Amount: txb.balanceNativeTokens[id].balance,
			}},
			Blocks: iotago.FeatureBlocks{
				&iotago.SenderFeatureBlock{
					Address: txb.anchorOutput.AliasID.ToAddress(),
				},
			}, // no blocks
		}
		ret = append(ret, o)
	}
	// creating outputs for posted on-ledger requests
	for _, pr := range txb.postedRequests {
		reqMetadata := iscp.NewRequestMetadata().
			WithSender(pr.SenderContract).
			WithTarget(pr.Metadata.TargetContract).
			WithEntryPoint(pr.Metadata.EntryPoint).
			WithArgs(pr.Metadata.Args).
			WithTransfer(pr.Assets)

		o := &iotago.ExtendedOutput{
			Address:      pr.TargetAddress,
			Amount:       pr.Assets.Amount, // TODO dust protection !!! Serialize, count bytes then put dust deposit
			NativeTokens: pr.Assets.Tokens,
			Blocks: iotago.FeatureBlocks{
				&iotago.SenderFeatureBlock{
					Address: txb.anchorOutput.AliasID.ToAddress(),
				},
				&iotago.MetadataFeatureBlock{
					Data: reqMetadata.Bytes(),
				},
				// TODO feature blocks as per SendOptions
			},
		}
		ret = append(ret, o)
	}
	return ret
}

func (txb *AnchorTransactionBuilder) numInputsConsumed() int {
	ret := len(txb.consumedInputs)
	for _, v := range txb.balanceNativeTokens {
		if v.initial != nil && v.balance.Cmp(v.initial) != 0 {
			ret++
		}
	}
	return ret
}

func (txb *AnchorTransactionBuilder) inputsAreFull() bool {
	return txb.numInputsConsumed() >= iotago.MaxInputsCount
}

func (txb *AnchorTransactionBuilder) numOutputsConsumed() int {
	ret := 1 // for chain output
	for _, v := range txb.balanceNativeTokens {
		if v.balance.Cmp(v.initial) != 0 && !util.IsZeroBigInt(v.balance) {
			ret++
		}
	}
	ret += len(txb.postedRequests)
	return ret
}

func (txb *AnchorTransactionBuilder) outputsAreFull() bool {
	return txb.numOutputsConsumed() >= iotago.MaxOutputsCount
}

func (txb *AnchorTransactionBuilder) addDeltaIotas(delta uint64) {
	// safe arithmetics
	n := txb.balanceIotas + delta
	if n < txb.balanceIotas {
		panic("addDeltaIotas: overflow")
	}
	txb.balanceIotas += n
}

func (txb *AnchorTransactionBuilder) subDeltaIotas(delta uint64) {
	// safe arithmetics
	if delta > txb.balanceIotas {
		panic("subDeltaIotas: overflow")
	}
	txb.balanceIotas -= delta
}

func (txb *AnchorTransactionBuilder) ensureNativeTokenBalance(id iotago.NativeTokenID) *nativeTokenBalance {
	if b, ok := txb.balanceNativeTokens[id]; ok {
		return b
	}
	balance, input := txb.loadNativeTokensOnChain(id)
	b := &nativeTokenBalance{
		input:   input,
		initial: balance,
		balance: new(big.Int),
	}
	b.balance.Set(balance)
	txb.balanceNativeTokens[id] = b
	if txb.outputsAreFull() {
		panic(ErrOutputLimitExceeded)
	}
	return b
}

// use negative to subtract
func (txb *AnchorTransactionBuilder) addDeltaNativeToken(id iotago.NativeTokenID, delta *big.Int) {
	b := txb.ensureNativeTokenBalance(id)
	// TODO safe arithmetic
	b.balance.Add(b.balance, delta)
}

func (txb *AnchorTransactionBuilder) addPostedRequest(par iscp.PostRequestData) {
	p := par
	txb.postedRequests = append(txb.postedRequests, &p)
}

/////////// vmcontext methods

func NewAnchorTransactionBuilder(anchorOutput *iotago.AliasOutput, anchorOutputID iotago.UTXOInput, balanceIotas uint64, tokenBalanceLoader tokenBalanceLoader) *AnchorTransactionBuilder {
	return &AnchorTransactionBuilder{
		loadNativeTokensOnChain: tokenBalanceLoader,
		anchorOutput:            anchorOutput,
		consumedInputs:          []iotago.UTXOInput{anchorOutputID},
		balanceIotas:            balanceIotas,
		balanceNativeTokens:     make(map[iotago.NativeTokenID]*nativeTokenBalance),
		postedRequests:          make([]*iscp.PostRequestData, 0),
	}
}
