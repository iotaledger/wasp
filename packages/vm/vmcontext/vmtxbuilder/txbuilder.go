package vmtxbuilder

import (
	"bytes"
	"math/big"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// implements transaction builder used internally by the VM during batch run

// TODO dust protection not covered yet !!!

// tokenBalanceLoader externally supplied function which loads balance of the native token from the state
// it returns nil if balance exists. If balance exist, it also returns output ID which holds the balance for the token_id
type tokenBalanceLoader func(iotago.NativeTokenID) (*big.Int, iotago.UTXOInput)

// AnchorTransactionBuilder represents structure which handles all the data needed to eventually
// build an essence of the anchor transaction
type AnchorTransactionBuilder struct {
	// balance loader for native tokens
	loadNativeTokensOnChain tokenBalanceLoader
	// anchorOutput output of the chain
	anchorOutput *iotago.AliasOutput
	// anchorOutputID is the ID of the anchor output
	anchorOutputID iotago.UTXOInput
	// already consumed outputs, specified by entire RequestData. It is needed for checking validity
	consumed []iscp.RequestData
	// balance of iotas is kept in the anchor output. TODO dust considerations
	balanceIotas uint64
	// balances of native tokens touched during the batch run
	balanceNativeTokens map[iotago.NativeTokenID]*nativeTokenBalance
	// requests posted by smart contracts
	postedOutputs []iotago.Output
	// TODO
}

// nativeTokenBalance represents on-chain account of the specific native token
type nativeTokenBalance struct {
	input   iotago.UTXOInput // if initial != nil
	initial *big.Int         // if nil it means output does not exist, this is new account for the token_id
	balance *big.Int         // current balance of the token_id on the chain
}

// error codes used for handled panics
var (
	ErrInputLimitExceeded  = xerrors.Errorf("exceeded maximum number of inputs in transaction. iotago.MaxInputsCount = %d", iotago.MaxInputsCount)
	ErrOutputLimitExceeded = xerrors.Errorf("exceeded maximum number of outputs in transaction. iotago.MaxOutputsCount = %d", iotago.MaxOutputsCount)
)

// NewAnchorTransactionBuilder creates new AnchorTransactionBuilder object
func NewAnchorTransactionBuilder(anchorOutput *iotago.AliasOutput, anchorOutputID iotago.UTXOInput, balanceIotas uint64, tokenBalanceLoader tokenBalanceLoader) *AnchorTransactionBuilder {
	return &AnchorTransactionBuilder{
		loadNativeTokensOnChain: tokenBalanceLoader,
		anchorOutput:            anchorOutput,
		anchorOutputID:          anchorOutputID,
		consumed:                make([]iscp.RequestData, 0, iotago.MaxInputsCount-1),
		balanceIotas:            balanceIotas,
		balanceNativeTokens:     make(map[iotago.NativeTokenID]*nativeTokenBalance),
		postedOutputs:           make([]iotago.Output, 0, iotago.MaxOutputsCount-1),
	}
}

// Clone clones the AnchorTransactionBuilder object. Used to snapshot/recover
func (txb *AnchorTransactionBuilder) Clone() *AnchorTransactionBuilder {
	ret := NewAnchorTransactionBuilder(txb.anchorOutput, txb.anchorOutputID, txb.balanceIotas, txb.loadNativeTokensOnChain)
	ret.consumed = append(ret.consumed, txb.consumed...)
	ret.balanceIotas = txb.balanceIotas
	for k, v := range txb.balanceNativeTokens {
		ret.balanceNativeTokens[k] = &nativeTokenBalance{
			input:   v.input,
			initial: new(big.Int).Set(v.balance),
			balance: new(big.Int).Set(v.balance),
		}
	}
	ret.postedOutputs = append(ret.postedOutputs, txb.postedOutputs...)
	return ret
}

// Consume adds an input to the transaction. Return its index.
// It panics if transaction cannot hold that many inputs
// All explicitly consumed inputs will hold fixed index in the transaction
// It updates total assets held by the chain. So it may panic due to exceed output counts
func (txb *AnchorTransactionBuilder) Consume(inp iscp.RequestData) int {
	if inp.IsOffLedger() {
		panic(xerrors.New("Consume: must be UTXO"))
	}
	if txb.InputsAreFull() {
		panic(ErrInputLimitExceeded)
	}
	txb.consumed = append(txb.consumed, inp)
	txb.addDeltaIotas(inp.Assets().Iotas)
	for _, nt := range inp.Assets().Tokens {
		txb.addDeltaNativeToken(nt.ID, nt.Amount)
	}
	return len(txb.consumed) - 1
}

// inputs generate a deterministic list of inputs for the transaction essence
// The explicitly consumed inputs hold fixed indices.
// The consumed UTXO of internal accounts are sorted by tokenID for determinism
// Consumed only internal UTXOs with changed token balances. The rest is left untouched
func (txb *AnchorTransactionBuilder) inputs() iotago.Inputs {
	ret := make(iotago.Inputs, 0, len(txb.consumed)+len(txb.balanceNativeTokens))
	ret = append(ret, &txb.anchorOutputID)
	for i := range txb.consumed {
		ret = append(ret, &txb.consumed[i].Unwrap().UTXO().Metadata().UTXOInput)
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
			// entry didn't existed before. Not consumed
			continue
		}
		if nt.initial.Cmp(nt.balance) == 0 {
			// no need for input because nothing changed
			continue
		}
		ret = append(ret, &nt.input)
	}
	if len(ret) != txb.numInputs() {
		panic("AnchorTransactionBuilder.inputs: internal inconsistency")
	}
	return ret
}

// SortedListOfTokenIDsForOutputs is needed in order to know the exact index in the transaction of each
// internal output for each token ID before it is stored into the state and state commitment is calculated
// In the and `blocklog` and `accounts` contract each internal account is tracked by:
// - knowing anchor transactionID for each block index (in `blocklog`)
// - knowing block number of the anchor transaction where the last UTXO was produced for the tokenID (in `accounts`)
// - knowing the index of the output in the anchor transaction (in `accounts`). This is calculated from SortedListOfTokenIDsForOutputs()
func (txb *AnchorTransactionBuilder) SortedListOfTokenIDsForOutputs() []iotago.NativeTokenID {
	ret := make([]iotago.NativeTokenID, 0, len(txb.balanceNativeTokens))
	for id, nt := range txb.balanceNativeTokens {
		if nt.initial != nil && nt.initial.Cmp(nt.balance) == 0 {
			// if nothing changed, this is not included in outputs
			continue
		}
		if util.IsZeroBigInt(nt.balance) {
			// if final balance is 0, output is not produced, i.e. chain will not hold any tokens of this tokenID
			continue
		}
		ret = append(ret, id)
	}
	sort.Slice(ret, func(i, j int) bool {
		return bytes.Compare(ret[i][:], ret[j][:]) < 0
	})
	return ret
}

const dustAmountForInternalAccountUTXO = 100 // TODO dust

// outputs generates outputs for the transaction essence
func (txb *AnchorTransactionBuilder) outputs(stateData *iscp.StateData) iotago.Outputs {
	ret := make(iotago.Outputs, 0, 1+len(txb.balanceNativeTokens)+len(txb.postedOutputs))
	// creating the anchor output
	anchorOutput := &iotago.AliasOutput{
		Amount:               txb.balanceIotas,
		NativeTokens:         nil, // anchorOutput output does not contain native tokens
		AliasID:              txb.anchorOutput.AliasID,
		StateController:      txb.anchorOutput.StateController,
		GovernanceController: txb.anchorOutput.GovernanceController,
		StateIndex:           txb.anchorOutput.StateIndex + 1,
		StateMetadata:        stateData.Bytes(),
		FoundryCounter:       txb.anchorOutput.FoundryCounter, // TODO should come from minting logic
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: txb.anchorOutput.AliasID.ToAddress(), // TODO not needed? But why not?
			},
		},
	}
	ret = append(ret, anchorOutput)

	// creating outputs for updated internal accounts
	tokenIdsToBeUpdated := txb.SortedListOfTokenIDsForOutputs()

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
			},
		}
		ret = append(ret, o)
	}
	// creating outputs for posted on-ledger requests
	ret = append(ret, txb.postedOutputs...)
	return ret
}

// numInputs number of inputs in the future transaction
func (txb *AnchorTransactionBuilder) numInputs() int {
	ret := len(txb.consumed) + 1 // + 1 for anchor UTXO
	for _, v := range txb.balanceNativeTokens {
		if v.initial != nil && v.balance.Cmp(v.initial) != 0 {
			ret++
		}
	}
	return ret
}

// InputsAreFull returns if transaction cannot bear more inputs
func (txb *AnchorTransactionBuilder) InputsAreFull() bool {
	return txb.numInputs() >= iotago.MaxInputsCount
}

// numOutputs in the transaction
func (txb *AnchorTransactionBuilder) numOutputs() int {
	ret := 1 // for chain output
	for _, v := range txb.balanceNativeTokens {
		if v.balance.Cmp(v.initial) != 0 && !util.IsZeroBigInt(v.balance) {
			ret++
		}
	}
	ret += len(txb.postedOutputs)
	return ret
}

// outputsAreFull return if transaction cannot bear more outputs
func (txb *AnchorTransactionBuilder) outputsAreFull() bool {
	return txb.numOutputs() >= iotago.MaxOutputsCount
}

// sumInputs sums up all assets in inputs
func (txb *AnchorTransactionBuilder) sumInputs() (uint64, map[iotago.NativeTokenID]*big.Int) {
	sumIotas := txb.anchorOutput.Amount
	sumTokens := make(map[iotago.NativeTokenID]*big.Int)
	// sum up all initial values of internal accounts
	for id, ntb := range txb.balanceNativeTokens {
		if ntb.initial != nil {
			s, ok := sumTokens[id]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, ntb.initial)
			sumTokens[id] = s
		}
	}
	// sum up all explicitly consumed outputs, except anchor output
	for _, out := range txb.consumed {
		a := out.Assets()
		sumIotas += a.Iotas
		for _, nt := range a.Tokens {
			s, ok := sumTokens[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			sumTokens[nt.ID] = s
		}
	}
	return sumIotas, sumTokens
}

// sumOutputs sums all balances in outputs
func (txb *AnchorTransactionBuilder) sumOutputs() (uint64, map[iotago.NativeTokenID]*big.Int) {
	sumIotas := txb.balanceIotas
	sumTokens := make(map[iotago.NativeTokenID]*big.Int)
	// sum up all initial values of internal accounts
	for id, ntb := range txb.balanceNativeTokens {
		if ntb.balance != nil && !util.IsZeroBigInt(ntb.balance) {
			s, ok := sumTokens[id]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, ntb.balance)
			sumTokens[id] = s
		}
	}
	for _, o := range txb.postedOutputs {
		assets := assetsFromOutput(o)
		sumIotas += assets.Iotas
		for _, nt := range assets.Tokens {
			s, ok := sumTokens[nt.ID]
			if !ok {
				s = new(big.Int)
			}
			s.Add(s, nt.Amount)
			sumTokens[nt.ID] = s
		}
	}
	return sumIotas, sumTokens
}

// TotalAssets check consistency. If input total equals with output totals, returns (iota total, native token totals, true)
// Otherwise returns (0, nil, false)
func (txb *AnchorTransactionBuilder) TotalAssets() (uint64, map[iotago.NativeTokenID]*big.Int, bool) {
	sumIotasIN, sumTokensIN := txb.sumInputs()
	sumIotasOUT, sumTokensOUT := txb.sumOutputs()

	if sumIotasIN != sumIotasOUT {
		return 0, nil, false
	}
	if len(sumTokensIN) != len(sumTokensOUT) {
		return 0, nil, false
	}
	for id, bIN := range sumTokensIN {
		bOUT, ok := sumTokensOUT[id]
		if !ok {
			return 0, nil, false
		}
		if bIN.Cmp(bOUT) != 0 {
			return 0, nil, false
		}
	}
	return sumIotasIN, sumTokensIN, true
}

// addDeltaIotas increases number of on-chain iotas by delta
func (txb *AnchorTransactionBuilder) addDeltaIotas(delta uint64) {
	if delta == 0 {
		return
	}
	// safe arithmetics
	n := txb.balanceIotas + delta
	if n < txb.balanceIotas {
		panic(xerrors.New("addDeltaIotas: overflow"))
	}
	txb.balanceIotas = n
}

// subDeltaIotas decreases number of on-chain iotas
func (txb *AnchorTransactionBuilder) subDeltaIotas(delta uint64) {
	if delta == 0 {
		return
	}
	// safe arithmetics
	if delta > txb.balanceIotas {
		panic(xerrors.New("subDeltaIotas: overflow"))
	}
	txb.balanceIotas -= delta
}

// ensureNativeTokenBalance makes sure that cached balance information is in the builder
// if not, it asks for the initial balance by calling the loader function
// Panics if the call results to exceeded limits
func (txb *AnchorTransactionBuilder) ensureNativeTokenBalance(id iotago.NativeTokenID) *nativeTokenBalance {
	if b, ok := txb.balanceNativeTokens[id]; ok {
		return b
	}
	balance, input := txb.loadNativeTokensOnChain(id) // balance will be nil if no such token id accounted yet
	if balance != nil && txb.InputsAreFull() {
		panic(ErrInputLimitExceeded)
	}
	if balance != nil && txb.outputsAreFull() {
		panic(ErrOutputLimitExceeded)
	}
	b := &nativeTokenBalance{
		input:   input,
		initial: balance,
		balance: new(big.Int),
	}
	b.balance.Set(balance)
	txb.balanceNativeTokens[id] = b
	return b
}

// addDeltaNativeToken adds delta to the token balance.
// Use negative to subtract
func (txb *AnchorTransactionBuilder) addDeltaNativeToken(id iotago.NativeTokenID, delta *big.Int) {
	if util.IsZeroBigInt(delta) {
		return
	}
	b := txb.ensureNativeTokenBalance(id)
	// TODO safe arithmetic
	b.balance.Add(b.balance, delta)
}

// AddOutput adds an information about posted request. It will produce output
func (txb *AnchorTransactionBuilder) AddOutput(o iotago.Output) {
	if txb.outputsAreFull() {
		panic(ErrOutputLimitExceeded)
	}
	assets := assetsFromOutput(o)
	txb.subDeltaIotas(assets.Iotas)
	bi := new(big.Int)
	for _, nt := range assets.Tokens {
		bi.Neg(nt.Amount)
		txb.addDeltaNativeToken(nt.ID, bi)
	}

	txb.postedOutputs = append(txb.postedOutputs, o)
}

// ExtendedOutputFromPostData creates extended output object from the PostRequestData.
// It automatically adjusts amount of iotas required for the dust deposit
func ExtendedOutputFromPostData(
	targetAddress, senderAddress iotago.Address,
	senderContract iscp.Hname,
	assets *iscp.Assets,
	sendMetadata *iscp.SendMetadata,
	options ...*iscp.SendOptions) *iotago.ExtendedOutput {
	// --

	reqMetadata := &iscp.RequestMetadata{
		SenderContract: senderContract,
		TargetContract: sendMetadata.TargetContract,
		EntryPoint:     sendMetadata.EntryPoint,
		Params:         sendMetadata.Params,
		Transfer:       sendMetadata.Transfer,
		GasBudget:      sendMetadata.GasBudget,
	}

	ret := &iotago.ExtendedOutput{
		Address:      targetAddress,
		Amount:       assets.Iotas,
		NativeTokens: assets.Tokens,
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: senderAddress,
			},
			&iotago.MetadataFeatureBlock{
				Data: reqMetadata.Bytes(),
			},
			// TODO feature blocks as per SendOptions
		},
	}
	// calculate and adjust required dust deposit amount TODO it is fake rental structure
	fakeRentStructure := &iotago.RentStructure{
		VByteCost:    1,
		VBFactorData: 1,
		VBFactorKey:  1,
	}
	dustDeposit := ret.VByteCost(fakeRentStructure, nil)
	if dustDeposit < ret.Amount {
		ret.Amount = dustDeposit
	}
	return ret
}

func assetsFromOutput(o iotago.Output) *iscp.Assets {
	switch o := o.(type) {
	case *iotago.ExtendedOutput:
		return &iscp.Assets{
			Iotas:  o.Amount,
			Tokens: o.NativeTokens,
		}
	default:
		panic(xerrors.Errorf("assetsFromOutput: not supported output type: %T", o))
	}
}

func (txb *AnchorTransactionBuilder) BuildTransactionEssence(stateData *iscp.StateData) *iotago.TransactionEssence {
	return &iotago.TransactionEssence{
		Inputs:  txb.inputs(),
		Outputs: txb.outputs(stateData),
		Payload: nil,
	}
}
