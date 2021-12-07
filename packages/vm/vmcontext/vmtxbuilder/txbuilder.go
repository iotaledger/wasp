package vmtxbuilder

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// error codes used for handled panics
var (
	ErrInputLimitExceeded                   = xerrors.Errorf("exceeded maximum number of inputs in transaction. iotago.MaxInputsCount = %d", iotago.MaxInputsCount)
	ErrOutputLimitExceeded                  = xerrors.Errorf("exceeded maximum number of outputs in transaction. iotago.MaxOutputsCount = %d", iotago.MaxOutputsCount)
	ErrNotEnoughFundsForInternalDustDeposit = xerrors.New("not enough funds for internal dust deposit")
	ErrOverflow                             = xerrors.New("overflow")
	ErrNotEnoughIotaBalance                 = xerrors.New("not enough iota balance")
	ErrNotEnoughNativeAssetBalance          = xerrors.New("not enough native assets balance")
)

// tokenBalanceLoader externally supplied function which loads balance of the native token from the state
// it returns nil if balance exists. If balance exist, it also returns output ID which holds the balance for the token_id
type tokenBalanceLoader func(iotago.NativeTokenID) (*big.Int, *iotago.UTXOInput)

// AnchorTransactionBuilder represents structure which handles all the data needed to eventually
// build an essence of the anchor transaction
type AnchorTransactionBuilder struct {
	// cache for dust deposit constant for one internal output
	dustAmountForInternalAccountUTXO uint64
	// on-chain balance loader for native tokens
	loadNativeTokensOnChain tokenBalanceLoader
	// anchorOutput output of the chain
	anchorOutput *iotago.AliasOutput
	// anchorOutputID is the ID of the anchor output
	anchorOutputID *iotago.UTXOInput
	// already consumed outputs, specified by entire RequestData. It is needed for checking validity
	consumed []iscp.RequestData
	// balance of iotas is kept in the anchor output. Does not include iotas which are on other outputs
	currentBalanceIotasOnAnchor uint64
	// balances of native tokens touched during the batch run
	balanceNativeTokens map[iotago.NativeTokenID]*nativeTokenBalance
	// requests posted by smart contracts
	postedOutputs []iotago.Output
	// TODO
}

// nativeTokenBalance represents on-chain account of the specific native token
type nativeTokenBalance struct {
	input              iotago.UTXOInput // if initial != nil
	initial            *big.Int         // if nil it means output does not exist, this is new account for the token_id
	balance            *big.Int         // current balance of the token_id on the chain
	dustDepositCharged bool
}

// producesOutput if value update produces UTXO of the corresponding total native token balance
func (n *nativeTokenBalance) producesOutput() bool {
	if n.initial != nil && n.balance.Cmp(n.initial) == 0 {
		// value didn't change
		return false
	}
	if util.IsZeroBigInt(n.balance) {
		// end value is 0
		return false
	}
	return true
}

// requiresInput returns if value change requires input in the transaction
func (n *nativeTokenBalance) requiresInput() bool {
	if n.initial == nil {
		// there's no input
		return false
	}
	if n.balance.Cmp(n.initial) == 0 {
		// value didn't change
		return false
	}
	return true
}

// NewAnchorTransactionBuilder creates new AnchorTransactionBuilder object
func NewAnchorTransactionBuilder(anchorOutput *iotago.AliasOutput, anchorOutputID *iotago.UTXOInput, balanceIotasOnAnchor uint64, tokenBalanceLoader tokenBalanceLoader) *AnchorTransactionBuilder {
	return &AnchorTransactionBuilder{
		loadNativeTokensOnChain:     tokenBalanceLoader,
		anchorOutput:                anchorOutput,
		anchorOutputID:              anchorOutputID,
		consumed:                    make([]iscp.RequestData, 0, iotago.MaxInputsCount-1),
		currentBalanceIotasOnAnchor: balanceIotasOnAnchor,
		balanceNativeTokens:         make(map[iotago.NativeTokenID]*nativeTokenBalance),
		postedOutputs:               make([]iotago.Output, 0, iotago.MaxOutputsCount-1),
	}
}

// Clone clones the AnchorTransactionBuilder object. Used to snapshot/recover
func (txb *AnchorTransactionBuilder) Clone() *AnchorTransactionBuilder {
	ret := NewAnchorTransactionBuilder(txb.anchorOutput, txb.anchorOutputID, txb.currentBalanceIotasOnAnchor, txb.loadNativeTokensOnChain)
	ret.dustAmountForInternalAccountUTXO = txb.dustAmountForInternalAccountUTXO
	ret.consumed = append(ret.consumed, txb.consumed...)
	ret.currentBalanceIotasOnAnchor = txb.currentBalanceIotasOnAnchor
	for k, v := range txb.balanceNativeTokens {
		initial := v.initial
		if initial != nil {
			initial = new(big.Int).Set(initial)
		}
		ret.balanceNativeTokens[k] = &nativeTokenBalance{
			input:              v.input,
			initial:            initial,
			balance:            new(big.Int).Set(v.balance),
			dustDepositCharged: v.dustDepositCharged,
		}
	}
	ret.postedOutputs = append(ret.postedOutputs, txb.postedOutputs...)
	return ret
}

// Consume adds an input to the transaction.
// It panics if transaction cannot hold that many inputs
// All explicitly consumed inputs will hold fixed index in the transaction
// It updates total assets held by the chain. So it may panic due to exceed output counts
// Returns delta of iotas needed to adjust the common account due to dust deposit requirement for internal UTXOs
// NOTE: if call panics with ErrNotEnoughFundsForInternalDustDeposit, the state of the builder becomes inconsistent
// It means, in the caller context it should be rolled back altogether
func (txb *AnchorTransactionBuilder) Consume(inp iscp.RequestData) int64 {
	if inp.IsOffLedger() {
		panic(xerrors.New("Consume: must be UTXO"))
	}
	if txb.InputsAreFull() {
		panic(ErrInputLimitExceeded)
	}
	txb.consumed = append(txb.consumed, inp)
	deltaIotasDustDepositAdjustment := int64(0)
	for _, nt := range inp.Assets().Tokens {
		deltaIotasDustDepositAdjustment += txb.addNativeTokenBalanceDelta(nt.ID, nt.Amount)
	}
	// first we add all iotas arrived with the output to anchor balance
	txb.addDeltaIotasToTotal(inp.Assets().Iotas)
	return deltaIotasDustDepositAdjustment
}

// AddOutput adds an information about posted request. It will produce output
func (txb *AnchorTransactionBuilder) AddOutput(o iotago.Output) {
	if txb.outputsAreFull() {
		panic(ErrOutputLimitExceeded)
	}
	assets := AssetsFromOutput(o)
	txb.subDeltaIotasFromTotal(assets.Iotas)
	bi := new(big.Int)
	for _, nt := range assets.Tokens {
		bi.Neg(nt.Amount)
		txb.addNativeTokenBalanceDelta(nt.ID, bi)
	}
	txb.postedOutputs = append(txb.postedOutputs, o)
}

// inputs generate a deterministic list of inputs for the transaction essence
// The explicitly consumed inputs hold fixed indices.
// The consumed UTXO of internal accounts are sorted by tokenID for determinism
// Consumed only internal UTXOs with changed token balances. The rest is left untouched
func (txb *AnchorTransactionBuilder) inputs() iotago.Inputs {
	ret := make(iotago.Inputs, 0, len(txb.consumed)+len(txb.balanceNativeTokens))
	ret = append(ret, txb.anchorOutputID)
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
		if !nt.requiresInput() {
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
		if !nt.producesOutput() {
			continue
		}
		ret = append(ret, id)
	}
	sort.Slice(ret, func(i, j int) bool {
		return bytes.Compare(ret[i][:], ret[j][:]) < 0
	})
	return ret
}

// outputs generates outputs for the transaction essence
func (txb *AnchorTransactionBuilder) outputs(stateData *iscp.StateData) iotago.Outputs {
	ret := make(iotago.Outputs, 0, 1+len(txb.balanceNativeTokens)+len(txb.postedOutputs))
	// creating the anchor output
	anchorOutput := &iotago.AliasOutput{
		Amount:               txb.currentBalanceIotasOnAnchor,
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
			Amount:  txb.vByteCostOfNativeTokenBalance(),
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
		if !v.requiresInput() {
			continue
		}
		ret++
	}
	return ret
}

// InputsAreFull returns if transaction cannot hold more inputs
func (txb *AnchorTransactionBuilder) InputsAreFull() bool {
	return txb.numInputs() >= iotago.MaxInputsCount
}

// numOutputs in the transaction
func (txb *AnchorTransactionBuilder) numOutputs() int {
	ret := 1 // for chain output
	for _, v := range txb.balanceNativeTokens {
		if !v.producesOutput() {
			continue
		}
		ret++
	}
	ret += len(txb.postedOutputs)
	return ret
}

// outputsAreFull return if transaction cannot bear more outputs
func (txb *AnchorTransactionBuilder) outputsAreFull() bool {
	return txb.numOutputs() >= iotago.MaxOutputsCount
}

// addDeltaIotasToTotal increases number of on-chain main account iotas by delta
func (txb *AnchorTransactionBuilder) addDeltaIotasToTotal(delta uint64) {
	if delta == 0 {
		return
	}
	// safe arithmetics
	n := txb.currentBalanceIotasOnAnchor + delta
	if n < txb.currentBalanceIotasOnAnchor {
		panic(xerrors.Errorf("addDeltaIotasToTotal: %w", ErrOverflow))
	}
	txb.currentBalanceIotasOnAnchor = n
}

// subDeltaIotasFromTotal decreases number of on-chain main account iotas
func (txb *AnchorTransactionBuilder) subDeltaIotasFromTotal(delta uint64) {
	if delta == 0 {
		return
	}
	// safe arithmetics
	if delta > txb.currentBalanceIotasOnAnchor {
		panic(ErrNotEnoughIotaBalance)
	}
	txb.currentBalanceIotasOnAnchor -= delta
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
		input:              *input,
		initial:            balance,
		balance:            new(big.Int),
		dustDepositCharged: balance != nil,
	}
	if balance != nil {
		b.balance.Set(balance)
	}
	txb.balanceNativeTokens[id] = b
	return b
}

// addNativeTokenBalanceDelta adds delta to the token balance. Use negative delta to subtract.
// The call may result in adding new token ID to the ledger or disappearing one
// This impacts dust amount locked in the internal UTXOs which keep respective balances
// Returns delta of required dust deposit
func (txb *AnchorTransactionBuilder) addNativeTokenBalanceDelta(id iotago.NativeTokenID, delta *big.Int) int64 {
	if util.IsZeroBigInt(delta) {
		return 0
	}
	b := txb.ensureNativeTokenBalance(id)
	tmp := new(big.Int).Set(b.balance)
	tmp.Add(b.balance, delta)
	if tmp.Sign() < 0 {
		panic(xerrors.Errorf("addNativeTokenBalanceDelta: %w", ErrNotEnoughNativeAssetBalance))
	}
	b.balance = tmp
	switch {
	case b.dustDepositCharged && b.requiresInput() && !b.producesOutput():
		// this is an old token in the on-chain ledger. Now it disappears and dust deposit
		// is released and delta of anchor is positive
		b.dustDepositCharged = false
		dd := txb.vByteCostOfNativeTokenBalance()
		txb.addDeltaIotasToTotal(dd)
		return int64(dd)
	case !b.dustDepositCharged && !b.requiresInput() && b.producesOutput():
		// this is a new token in the on-chain ledger
		// There's a need for additional dust deposit on the respective UTXO, so delta for the anchor is negative
		b.dustDepositCharged = true
		dd := txb.vByteCostOfNativeTokenBalance()
		if dd >= txb.currentBalanceIotasOnAnchor {
			panic(ErrNotEnoughFundsForInternalDustDeposit)
		}
		txb.subDeltaIotasFromTotal(dd)
		return -int64(dd)
	}
	return 0
}

func stringUTXOInput(inp *iotago.UTXOInput) string {
	return fmt.Sprintf("[%d]%s", inp.TransactionOutputIndex, base58.Encode(inp.TransactionID[:]))
}

func stringNativeTokenID(id *iotago.NativeTokenID) string {
	return base58.Encode(id[:])
}

func (txb *AnchorTransactionBuilder) String() string {
	ret := ""
	ret += fmt.Sprintf("%s\n", stringUTXOInput(txb.anchorOutputID))
	ret += fmt.Sprintf("initial IOTA balance: %d\n", txb.anchorOutput.Amount)
	ret += fmt.Sprintf("current IOTA balance: %d\n", txb.currentBalanceIotasOnAnchor)
	ret += fmt.Sprintf("Native tokens (%d):\n", len(txb.balanceNativeTokens))
	for id, ntb := range txb.balanceNativeTokens {
		initial := "0"
		if ntb.initial != nil {
			initial = ntb.initial.String()
		}
		current := ntb.balance.String()
		ret += fmt.Sprintf("      %s: %s --> %s, dust deposit charged: %v\n",
			stringNativeTokenID(&id), initial, current, ntb.dustDepositCharged)
	}
	ret += fmt.Sprintf("consumed inputs (%d):\n", len(txb.consumed))
	//for _, inp := range txb.consumed {
	//	ret += fmt.Sprintf("      %s\n", inp.ID().String())
	//}
	ret += fmt.Sprintf("added outputs (%d):\n", len(txb.postedOutputs))
	ret += ">>>>>> TODO. Not finished....."
	return ret
}

func (txb *AnchorTransactionBuilder) BuildTransactionEssence(stateData *iscp.StateData) *iotago.TransactionEssence {
	return &iotago.TransactionEssence{
		Inputs:  txb.inputs(),
		Outputs: txb.outputs(stateData),
		Payload: nil,
	}
}

// cache for the dust amount constant

// vByteCostOfNativeTokenBalance return byte cost for the internal UTXO used to keep on chain native tokens.
// We assume that size of the UTXO will always be a constant
// TODO review and optimize together with the whole dust handling stuff
func (txb *AnchorTransactionBuilder) vByteCostOfNativeTokenBalance() uint64 {
	if txb.dustAmountForInternalAccountUTXO > 0 {
		return txb.dustAmountForInternalAccountUTXO
	}
	// a fake output with one native token in the balance
	// the NewExtendedOutput will adjust the dust deposit
	o, _ := NewExtendedOutput(
		txb.anchorOutput.AliasID.ToAddress(),
		&iscp.Assets{
			Iotas: 1,
			Tokens: iotago.NativeTokens{&iotago.NativeToken{
				ID:     iotago.NativeTokenID{},
				Amount: new(big.Int),
			}},
		},
		txb.anchorOutput.AliasID.ToAddress(),
		nil,
		nil,
	)
	txb.dustAmountForInternalAccountUTXO = o.Amount
	return txb.dustAmountForInternalAccountUTXO
}

// ExtendedOutputFromPostData creates extended output object from parameters.
// It automatically adjusts amount of iotas required for the dust deposit
func ExtendedOutputFromPostData(senderAddress iotago.Address, senderContract iscp.Hname, par iscp.RequestParameters) *iotago.ExtendedOutput {
	ret, _ := NewExtendedOutput(
		par.TargetAddress,
		par.Assets,
		senderAddress,
		&iscp.RequestMetadata{
			SenderContract: senderContract,
			TargetContract: par.Metadata.TargetContract,
			EntryPoint:     par.Metadata.EntryPoint,
			Params:         par.Metadata.Params,
			Transfer:       par.Metadata.Transfer,
			GasBudget:      par.Metadata.GasBudget,
		},
		par.Options,
	)
	return ret
}

// NewExtendedOutput creates new ExtendedOutput from input parameters.
// Adjusts dust deposit if needed and returns flag if adjusted
func NewExtendedOutput(
	targetAddress iotago.Address,
	assets *iscp.Assets,
	senderAddress iotago.Address,
	metadata *iscp.RequestMetadata,
	options *iscp.SendOptions) (*iotago.ExtendedOutput, bool) {
	ret := &iotago.ExtendedOutput{
		Address:      targetAddress,
		Amount:       assets.Iotas,
		NativeTokens: assets.Tokens,
		Blocks:       iotago.FeatureBlocks{},
	}
	if senderAddress != nil {
		ret.Blocks = append(ret.Blocks, &iotago.SenderFeatureBlock{
			Address: senderAddress,
		})
	}
	if metadata != nil {
		ret.Blocks = append(ret.Blocks, &iotago.MetadataFeatureBlock{
			Data: metadata.Bytes(),
		})
	}

	if options != nil {
		panic(" send options FeatureBlocks not implemented yet")
	}

	// Adjust to minimum dust deposit
	dustDepositAdjusted := false
	neededDustDeposit := ret.VByteCost(parameters.RentStructure(), nil)
	if ret.Amount < neededDustDeposit {
		ret.Amount = neededDustDeposit
		dustDepositAdjusted = true
	}
	return ret, dustDepositAdjusted
}

func AssetsFromOutput(o iotago.Output) *iscp.Assets {
	switch o := o.(type) {
	case *iotago.ExtendedOutput:
		return &iscp.Assets{
			Iotas:  o.Amount,
			Tokens: o.NativeTokens,
		}
	default:
		panic(xerrors.Errorf("AssetsFromOutput: not supported output type: %T", o))
	}
}
