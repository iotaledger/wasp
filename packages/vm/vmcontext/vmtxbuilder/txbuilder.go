package vmtxbuilder

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
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
	ErrCreateFoundryMaxSupplyMustBePositive = xerrors.New("max supply must be positive")
	ErrCreateFoundryMaxSupplyTooBig         = xerrors.New("max supply is too big")
	ErrFoundryDoesNotExist                  = xerrors.New("foundry does not exist")
	ErrCantModifySupplyOfTheToken           = xerrors.New("supply of the token is not controlled by the chain")
	ErrNativeTokenSupplyOutOffBounds        = xerrors.New("token supply is out of bounds")
)

// tokenOutputLoader externally supplied function which loads stored output from the state
// Should return nil if does not exist
type tokenOutputLoader func(*iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput)

// foundryLoader externally supplied function which returns foundry output and id by its serial number
// Should return nil if foundry does not exist
type foundryLoader func(uint32) (*iotago.FoundryOutput, *iotago.UTXOInput)

// AnchorTransactionBuilder represents structure which handles all the data needed to eventually
// build an essence of the anchor transaction
type AnchorTransactionBuilder struct {
	// anchorOutput output of the chain
	anchorOutput *iotago.AliasOutput
	// anchorOutputID is the ID of the anchor output
	anchorOutputID *iotago.UTXOInput
	// already consumed outputs, specified by entire RequestData. It is needed for checking validity
	consumed []iscp.RequestData
	// iotas which are on-chain. It does not include dust deposits on anchor and on internal outputs
	totalIotasOnChain uint64
	// cached number of iotas can't go below this
	dustDepositOnAnchor uint64
	// cached dust deposit constant for one internal output
	dustDepositOnInternalTokenOutput uint64
	// balance loader for native tokens
	loadTokenOutput tokenOutputLoader
	// foundry loader
	loadFoundry foundryLoader
	// balances of native tokens touched during the batch run
	balanceNativeTokens map[iotago.NativeTokenID]*nativeTokenBalance
	// invoked foundries. Foundry serial number is used as a key
	invokedFoundries map[uint32]*foundryInvoked
	// requests posted by smart contracts
	postedOutputs []iotago.Output
	// structure to calculate byte costs
	rentStructure *iotago.RentStructure
}

// nativeTokenBalance represents on-chain account of the specific native token
type nativeTokenBalance struct {
	tokenID            iotago.NativeTokenID
	input              iotago.UTXOInput // if in != nil
	dustDepositCharged bool
	in                 *iotago.ExtendedOutput // if nil it means output does not exist, this is new account for the token_id
	out                *iotago.ExtendedOutput // current balance of the token_id on the chain
}

type foundryInvoked struct {
	serialNumber uint32
	input        iotago.UTXOInput      // if in != nil
	in           *iotago.FoundryOutput // nil if created
	out          *iotago.FoundryOutput // nil if destroyed
}

// NewAnchorTransactionBuilder creates new AnchorTransactionBuilder object
func NewAnchorTransactionBuilder(
	anchorOutput *iotago.AliasOutput,
	anchorOutputID *iotago.UTXOInput,
	tokenBalanceLoader tokenOutputLoader,
	foundryLoader foundryLoader,
	rentStructure *iotago.RentStructure,
) *AnchorTransactionBuilder {
	anchorDustDeposit := anchorOutput.VByteCost(parameters.RentStructure(), nil)
	if anchorOutput.Amount < anchorDustDeposit {
		panic("internal inconsistency")
	}
	return &AnchorTransactionBuilder{
		anchorOutput:                     anchorOutput,
		anchorOutputID:                   anchorOutputID,
		totalIotasOnChain:                anchorOutput.Amount - anchorDustDeposit,
		dustDepositOnAnchor:              anchorDustDeposit,
		dustDepositOnInternalTokenOutput: calcVByteCostOfNativeTokenBalance(rentStructure),
		loadTokenOutput:                  tokenBalanceLoader,
		loadFoundry:                      foundryLoader,
		consumed:                         make([]iscp.RequestData, 0, iotago.MaxInputsCount-1),
		balanceNativeTokens:              make(map[iotago.NativeTokenID]*nativeTokenBalance),
		postedOutputs:                    make([]iotago.Output, 0, iotago.MaxOutputsCount-1),
		invokedFoundries:                 make(map[uint32]*foundryInvoked),
		rentStructure:                    rentStructure,
	}
}

// Clone clones the AnchorTransactionBuilder object. Used to snapshot/recover
func (txb *AnchorTransactionBuilder) Clone() *AnchorTransactionBuilder {
	ret := &AnchorTransactionBuilder{
		anchorOutput:                     txb.anchorOutput,
		anchorOutputID:                   txb.anchorOutputID,
		totalIotasOnChain:                txb.totalIotasOnChain,
		dustDepositOnAnchor:              txb.dustDepositOnAnchor,
		dustDepositOnInternalTokenOutput: txb.dustDepositOnInternalTokenOutput,
		loadTokenOutput:                  txb.loadTokenOutput,
		loadFoundry:                      txb.loadFoundry,
		consumed:                         make([]iscp.RequestData, 0, cap(txb.consumed)),
		balanceNativeTokens:              make(map[iotago.NativeTokenID]*nativeTokenBalance),
		postedOutputs:                    make([]iotago.Output, 0, cap(txb.postedOutputs)),
		invokedFoundries:                 make(map[uint32]*foundryInvoked),
		rentStructure:                    txb.rentStructure,
	}

	ret.consumed = append(ret.consumed, txb.consumed...)
	for k, v := range txb.balanceNativeTokens {
		ret.balanceNativeTokens[k] = v.clone()
	}
	for k, v := range txb.invokedFoundries {
		ret.invokedFoundries[k] = v.clone()
	}
	ret.postedOutputs = append(ret.postedOutputs, txb.postedOutputs...)
	return ret
}

// TotalAvailableIotas returns number of on-chain iotas.
// It does not include minimum dust deposit needed for anchor output and other internal UTXOs
func (txb *AnchorTransactionBuilder) TotalAvailableIotas() uint64 {
	return txb.totalIotasOnChain
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

	// first we add all iotas arrived with the output to anchor balance
	txb.addDeltaIotasToTotal(inp.Unwrap().UTXO().Output().Deposit())
	// then we add all arriving native tokens to corresponding internal outputs
	deltaIotasDustDepositAdjustment := int64(0)
	for _, nt := range inp.Assets().Tokens {
		deltaIotasDustDepositAdjustment += txb.addNativeTokenBalanceDelta(&nt.ID, nt.Amount)
	}
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
		txb.addNativeTokenBalanceDelta(&nt.ID, bi)
	}
	txb.postedOutputs = append(txb.postedOutputs, o)
}

// inputs generate a deterministic list of inputs for the transaction essence
// - index 0 is always alias output
// - then goes consumed external ExtendedOutput UTXOs, the requests, in the order of processing
// - then goes inputs of native token UTXOs, sorted by token id
// - then goes inputs of foundries sorted by serial number
func (txb *AnchorTransactionBuilder) inputs() iotago.Inputs {
	ret := make(iotago.Inputs, 0, len(txb.consumed)+len(txb.balanceNativeTokens)+len(txb.invokedFoundries))
	// alias output
	ret = append(ret, txb.anchorOutputID)
	// consumed on-ledger requests
	for i := range txb.consumed {
		ret = append(ret, txb.consumed[i].ID().OutputID())
	}
	// internal native token outputs
	for _, nt := range txb.nativeTokenOutputsSorted() {
		if nt.requiresInput() {
			ret = append(ret, &nt.input)
		}
	}
	// foundries
	for _, f := range txb.foundriesSorted() {
		if f.requiresInput() {
			ret = append(ret, &f.input)
		}
	}
	if len(ret) != txb.numInputs() {
		panic("AnchorTransactionBuilder.inputs: internal inconsistency")
	}
	return ret
}

// outputs generates outputs for the transaction essence
func (txb *AnchorTransactionBuilder) outputs(stateData *iscp.StateData) iotago.Outputs {
	ret := make(iotago.Outputs, 0, 1+len(txb.balanceNativeTokens)+len(txb.postedOutputs))
	// creating the anchor output
	aliasID := txb.anchorOutput.AliasID
	if aliasID.Empty() {
		aliasID = iotago.AliasIDFromOutputID(txb.anchorOutputID.ID())
	}
	anchorOutput := &iotago.AliasOutput{
		Amount:               txb.totalIotasOnChain + txb.dustDepositOnAnchor,
		NativeTokens:         nil, // anchor output does not contain native tokens
		AliasID:              aliasID,
		StateController:      txb.anchorOutput.StateController,
		GovernanceController: txb.anchorOutput.GovernanceController,
		StateIndex:           txb.anchorOutput.StateIndex + 1,
		StateMetadata:        stateData.Bytes(),
		FoundryCounter:       txb.nextFoundryCounter(),
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: aliasID.ToAddress(),
			},
		},
	}
	ret = append(ret, anchorOutput)

	// creating outputs for updated internal accounts
	nativeTokensToBeUpdated, _ := txb.NativeTokenRecordsToBeUpdated()
	for _, id := range nativeTokensToBeUpdated {
		// create one output for each token ID of internal account
		ret = append(ret, txb.balanceNativeTokens[id].out)
	}
	// creating outputs for updated foundries
	foundriesToBeUpdated, _ := txb.FoundriesToBeUpdated()
	for _, serNum := range foundriesToBeUpdated {
		ret = append(ret, txb.invokedFoundries[serNum].out)
	}
	// creating outputs for posted on-ledger requests
	ret = append(ret, txb.postedOutputs...)
	return ret
}

// numInputs number of inputs in the future transaction
func (txb *AnchorTransactionBuilder) numInputs() int {
	ret := len(txb.consumed) + 1 // + 1 for anchor UTXO
	for _, v := range txb.balanceNativeTokens {
		if v.requiresInput() {
			ret++
		}
	}
	for _, f := range txb.invokedFoundries {
		if f.requiresInput() {
			ret++
		}
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
		if v.producesOutput() {
			ret++
		}
	}
	ret += len(txb.postedOutputs)
	for _, f := range txb.invokedFoundries {
		if f.producesOutput() {
			ret++
		}
	}
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
	n := txb.totalIotasOnChain + delta
	if n+txb.dustDepositOnAnchor < txb.totalIotasOnChain {
		panic(xerrors.Errorf("addDeltaIotasToTotal: %w", ErrOverflow))
	}
	txb.totalIotasOnChain = n
}

// subDeltaIotasFromTotal decreases number of on-chain main account iotas
func (txb *AnchorTransactionBuilder) subDeltaIotasFromTotal(delta uint64) {
	if delta == 0 {
		return
	}
	// safe arithmetics
	if delta > txb.totalIotasOnChain {
		panic(ErrNotEnoughIotaBalance)
	}
	txb.totalIotasOnChain -= delta
}

// ensureNativeTokenBalance makes sure that cached output is in the builder
// if not, it asks for the in balance by calling the loader function
// Panics if the call results to exceeded limits
func (txb *AnchorTransactionBuilder) ensureNativeTokenBalance(id *iotago.NativeTokenID) *nativeTokenBalance {
	if b, ok := txb.balanceNativeTokens[*id]; ok {
		return b
	}
	in, input := txb.loadTokenOutput(id) // output will be nil if no such token id accounted yet
	if in != nil && txb.InputsAreFull() {
		panic(ErrInputLimitExceeded)
	}
	if in != nil && txb.outputsAreFull() {
		panic(ErrOutputLimitExceeded)
	}
	var out *iotago.ExtendedOutput
	if in == nil {
		out = txb.newInternalTokenOutput(txb.anchorOutput.AliasID, *id)
	} else {
		out = cloneInternalExtendedOutput(in)
	}
	b := &nativeTokenBalance{
		tokenID:            out.NativeTokens[0].ID,
		in:                 in,
		out:                out,
		dustDepositCharged: in != nil,
	}
	if input != nil {
		b.input = *input
	}
	txb.balanceNativeTokens[*id] = b
	return b
}

// addNativeTokenBalanceDelta adds delta to the token balance. Use negative delta to subtract.
// The call may result in adding new token ID to the ledger or disappearing one
// This impacts dust amount locked in the internal UTXOs which keep respective balances
// Returns delta of required dust deposit
func (txb *AnchorTransactionBuilder) addNativeTokenBalanceDelta(id *iotago.NativeTokenID, delta *big.Int) int64 {
	if util.IsZeroBigInt(delta) {
		return 0
	}
	nt := txb.ensureNativeTokenBalance(id)
	tmp := new(big.Int).Add(nt.getOutValue(), delta)
	if tmp.Sign() < 0 {
		panic(xerrors.Errorf("addNativeTokenBalanceDelta: %w", ErrNotEnoughNativeAssetBalance))
	}
	if tmp.Cmp(abi.MaxUint256) > 0 {
		panic(xerrors.Errorf("addNativeTokenBalanceDelta: %w", ErrOverflow))
	}
	nt.setOutValue(tmp)
	switch {
	case nt.dustDepositCharged && !nt.producesOutput():
		// this is an old token in the on-chain ledger. Now it disappears and dust deposit
		// is released and delta of anchor is positive
		nt.dustDepositCharged = false
		txb.addDeltaIotasToTotal(txb.dustDepositOnInternalTokenOutput)
		return int64(txb.dustDepositOnInternalTokenOutput)
	case !nt.dustDepositCharged && nt.producesOutput():
		// this is a new token in the on-chain ledger
		// There's a need for additional dust deposit on the respective UTXO, so delta for the anchor is negative
		nt.dustDepositCharged = true
		if txb.dustDepositOnInternalTokenOutput > txb.totalIotasOnChain {
			panic(ErrNotEnoughFundsForInternalDustDeposit)
		}
		txb.subDeltaIotasFromTotal(txb.dustDepositOnInternalTokenOutput)
		return -int64(txb.dustDepositOnInternalTokenOutput)
	}
	return 0
}

func stringUTXOInput(inp *iotago.UTXOInput) string {
	return fmt.Sprintf("[%d]%s", inp.TransactionOutputIndex, hex.EncodeToString(inp.TransactionID[:]))
}

func stringNativeTokenID(id *iotago.NativeTokenID) string {
	return hex.EncodeToString(id[:])
}

func (txb *AnchorTransactionBuilder) String() string {
	ret := ""
	ret += fmt.Sprintf("%s\n", stringUTXOInput(txb.anchorOutputID))
	ret += fmt.Sprintf("in IOTA balance: %d\n", txb.anchorOutput.Amount)
	ret += fmt.Sprintf("current IOTA balance: %d\n", txb.totalIotasOnChain)
	ret += fmt.Sprintf("Native tokens (%d):\n", len(txb.balanceNativeTokens))
	for id, ntb := range txb.balanceNativeTokens {
		initial := "0"
		if ntb.in != nil {
			initial = ntb.getOutValue().String()
		}
		current := ntb.getOutValue().String()
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

// calcVByteCostOfNativeTokenBalance return byte cost for the internal UTXO used to keep on chain native tokens.
// We assume that size of the UTXO will always be a constant
func calcVByteCostOfNativeTokenBalance(rentStructure *iotago.RentStructure) uint64 {
	// a fake output with one native token in the balance
	// the MakeExtendedOutput will adjust the dust deposit
	addr := iotago.AliasAddressFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(iotago.TransactionID{}, 0))
	o, _ := MakeExtendedOutput(
		&addr,
		&addr,
		&iscp.Assets{
			Iotas: 1,
			Tokens: iotago.NativeTokens{&iotago.NativeToken{
				ID:     iotago.NativeTokenID{},
				Amount: new(big.Int),
			}},
		},
		nil,
		nil,
		rentStructure,
	)
	return o.Amount
}

func (txb *AnchorTransactionBuilder) DustDeposits() (uint64, uint64) {
	return txb.dustDepositOnAnchor, txb.dustDepositOnInternalTokenOutput
}

// ExtendedOutputFromPostData creates extended output object from parameters.
// It automatically adjusts amount of iotas required for the dust deposit
func ExtendedOutputFromPostData(
	senderAddress iotago.Address,
	senderContract iscp.Hname,
	par iscp.RequestParameters,
	rentStructure *iotago.RentStructure,
) *iotago.ExtendedOutput {
	ret, _ := MakeExtendedOutput(
		par.TargetAddress,
		senderAddress,
		par.Assets,
		&iscp.RequestMetadata{
			SenderContract: senderContract,
			TargetContract: par.Metadata.TargetContract,
			EntryPoint:     par.Metadata.EntryPoint,
			Params:         par.Metadata.Params,
			Transfer:       par.Metadata.Transfer,
			GasBudget:      par.Metadata.GasBudget,
		},
		par.Options,
		rentStructure,
	)
	return ret
}

// MakeExtendedOutput creates new ExtendedOutput from input parameters.
// Adjusts dust deposit if needed and returns flag if adjusted
func MakeExtendedOutput(
	targetAddress iotago.Address,
	senderAddress iotago.Address,
	assets *iscp.Assets,
	metadata *iscp.RequestMetadata,
	options *iscp.SendOptions,
	rentStructure *iotago.RentStructure,
) (*iotago.ExtendedOutput, bool) {
	if assets == nil {
		assets = &iscp.Assets{}
	}
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
	neededDustDeposit := ret.VByteCost(rentStructure, nil)
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
