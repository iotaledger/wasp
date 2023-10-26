package vmtxbuilder

import (
	"fmt"
	"math/big"

	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/vm"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

type AccountsContractRead struct {
	// nativeTokenOutputLoaderFunc loads stored output from the state
	// Should return nil if does not exist
	NativeTokenOutput func(iotago.NativeTokenID) (*iotago.BasicOutput, iotago.OutputID)

	// foundryLoaderFunc returns foundry output and id by its serial number
	// Should return nil if foundry does not exist
	FoundryOutput func(uint32) (*iotago.FoundryOutput, iotago.OutputID)

	// NFTOutput returns the stored NFT output from the state
	// Should return nil if NFT is not accounted for
	NFTOutput func(id iotago.NFTID) (*iotago.NFTOutput, iotago.OutputID)

	// TotalFungibleTokens returns the total base tokens and native tokens accounted by the chain
	TotalFungibleTokens func() *isc.FungibleTokens
}

// AnchorTransactionBuilder represents structure which handles all the data needed to eventually
// build an essence of the anchor transaction
type AnchorTransactionBuilder struct {
	// anchorOutput output of the chain
	anchorOutput               *iotago.AccountOutput
	anchorOutputStorageDeposit iotago.BaseToken

	// result new AO of the chain, filled by "BuildTransactionEssence"
	resultAnchorOutput *iotago.AccountOutput

	// anchorOutputID is the ID of the anchor output
	anchorOutputID iotago.OutputID
	// already consumed outputs, specified by entire Request. It is needed for checking validity
	consumed []isc.OnLedgerRequest

	// view the accounts contract state
	accountsView AccountsContractRead

	// balances of native tokens loaded during the batch run
	balanceNativeTokens map[iotago.NativeTokenID]*nativeTokenBalance
	// all nfts loaded during the batch run
	nftsIncluded map[iotago.NFTID]*nftIncluded
	// all nfts minted
	nftsMinted iotago.TxEssenceOutputs
	// invoked foundries. Foundry serial number is used as a key
	invokedFoundries map[uint32]*foundryInvoked
	// requests posted by smart contracts
	postedOutputs iotago.TxEssenceOutputs
}

// NewAnchorTransactionBuilder creates new AnchorTransactionBuilder object
func NewAnchorTransactionBuilder(
	anchorOutput *iotago.AccountOutput,
	anchorOutputID iotago.OutputID,
	anchorOutputStorageDeposit iotago.BaseToken, // because we don't know what L1 parameters were used to calculate the last AO, we need to infer it from the accounts state
	accounts AccountsContractRead,
) *AnchorTransactionBuilder {
	return &AnchorTransactionBuilder{
		anchorOutput:               anchorOutput,
		anchorOutputID:             anchorOutputID,
		anchorOutputStorageDeposit: anchorOutputStorageDeposit,
		accountsView:               accounts,
		consumed:                   make([]isc.OnLedgerRequest, 0, iotago.MaxInputsCount-1),
		balanceNativeTokens:        make(map[iotago.NativeTokenID]*nativeTokenBalance),
		postedOutputs:              make(iotago.TxEssenceOutputs, 0, iotago.MaxOutputsCount-1),
		invokedFoundries:           make(map[uint32]*foundryInvoked),
		nftsIncluded:               make(map[iotago.NFTID]*nftIncluded),
		nftsMinted:                 make(iotago.TxEssenceOutputs, 0),
	}
}

// Clone clones the AnchorTransactionBuilder object. Used to snapshot/recover
func (txb *AnchorTransactionBuilder) Clone() *AnchorTransactionBuilder {
	anchorOutputID := iotago.OutputID{}
	copy(anchorOutputID[:], txb.anchorOutputID[:])

	return &AnchorTransactionBuilder{
		anchorOutput:               txb.anchorOutput.Clone().(*iotago.AccountOutput),
		anchorOutputID:             anchorOutputID,
		anchorOutputStorageDeposit: txb.anchorOutputStorageDeposit,
		accountsView:               txb.accountsView,
		consumed:                   util.CloneSlice(txb.consumed),
		balanceNativeTokens:        util.CloneMap(txb.balanceNativeTokens),
		postedOutputs: lo.Map(txb.postedOutputs, func(o iotago.TxEssenceOutput, _ int) iotago.TxEssenceOutput {
			return o.Clone()
		}),
		invokedFoundries: util.CloneMap(txb.invokedFoundries),
		nftsIncluded:     util.CloneMap(txb.nftsIncluded),
		nftsMinted: lo.Map(txb.nftsMinted, func(o iotago.TxEssenceOutput, _ int) iotago.TxEssenceOutput {
			return o.Clone()
		}),
	}
}

// splitAssetsIntoInternalOutputs splits the native Tokens/NFT from a given (request) output.
// returns the resulting outputs and the list of new outputs
// (some of the native tokens might already have an accounting output owned by the chain, so we don't need new outputs for those)
func (txb *AnchorTransactionBuilder) splitAssetsIntoInternalOutputs(req isc.OnLedgerRequest) iotago.BaseToken {
	requiredSD := iotago.BaseToken(0)
	for _, nativeToken := range req.Assets().NativeTokens {
		// ensure this NT is in the txbuilder, update it
		nt := txb.ensureNativeTokenBalance(nativeToken.ID)
		sdBefore := nt.accountingOutput.Amount
		if util.IsZeroBigInt(nt.getOutValue()) {
			sdBefore = 0 // accounting output was zero'ed this block, meaning the existing SD was released
		}
		nt.add(nativeToken.Amount)
		nt.updateMinSD()
		sdAfter := nt.accountingOutput.Amount
		// user pays for the difference (in case SD has increased, will be the full SD cost if the output is new)
		requiredSD += sdAfter - sdBefore
	}

	if req.NFT() != nil {
		// create new output
		nftIncl := txb.internalNFTOutputFromRequest(req.Output().(*iotago.NFTOutput), req.OutputID())
		requiredSD += nftIncl.resultingOutput.Amount
	}

	txb.consumed = append(txb.consumed, req)
	return requiredSD
}

func (txb *AnchorTransactionBuilder) assertLimits() {
	if txb.InputsAreFull() {
		panic(vmexceptions.ErrInputLimitExceeded)
	}
	if txb.outputsAreFull() {
		panic(vmexceptions.ErrOutputLimitExceeded)
	}
}

// Consume adds an input to the transaction.
// It panics if transaction cannot hold that many inputs
// All explicitly consumed inputs will hold fixed index in the transaction
// It updates total assets held by the chain. So it may panic due to exceed output counts
// Returns  the amount of baseTokens needed to cover SD costs for the NTs/NFT contained by the request output
func (txb *AnchorTransactionBuilder) Consume(req isc.OnLedgerRequest) iotago.BaseToken {
	defer txb.assertLimits()
	// deduct the minSD for all the outputs that need to be created
	requiredSD := txb.splitAssetsIntoInternalOutputs(req)
	return requiredSD
}

// ConsumeUnprocessable adds an unprocessable request to the txbuilder,
// consumes the original request and cretes a new output keeping assets intact
// return the position of the resulting output in `txb.postedOutputs`
func (txb *AnchorTransactionBuilder) ConsumeUnprocessable(req isc.OnLedgerRequest) int {
	defer txb.assertLimits()
	txb.consumed = append(txb.consumed, req)
	txb.postedOutputs = append(txb.postedOutputs, retryOutputFromOnLedgerRequest(req, txb.anchorOutput.AccountID))
	return len(txb.postedOutputs) - 1
}

// AddOutput adds an information about posted request. It will produce output
// Return adjustment needed for the L2 ledger (adjustment on base tokens related to storage deposit)
func (txb *AnchorTransactionBuilder) AddOutput(o iotago.Output) int64 {
	defer txb.assertLimits()

	storageDeposit, err := parameters.Storage().MinDeposit(o)
	if err != nil {
		panic(err)
	}
	if o.BaseTokenAmount() < storageDeposit {
		panic(fmt.Errorf("%v: available %d < required %d base tokens",
			transaction.ErrNotEnoughBaseTokensForStorageDeposit, o.BaseTokenAmount(), storageDeposit))
	}
	assets := transaction.AssetsFromOutput(o)

	sdAdjustment := int64(0)
	for _, nativeToken := range assets.NativeTokens {
		sdAdjustment += txb.addNativeTokenBalanceDelta(nativeToken.ID, new(big.Int).Neg(nativeToken.Amount))
	}
	if nftout, ok := o.(*iotago.NFTOutput); ok {
		sdAdjustment += txb.sendNFT(nftout)
	}
	txb.postedOutputs = append(txb.postedOutputs, o)
	return sdAdjustment
}

// InputsAreFull returns if transaction cannot hold more inputs
func (txb *AnchorTransactionBuilder) InputsAreFull() bool {
	return txb.numInputs() >= iotago.MaxInputsCount
}

// BuildTransactionEssence builds transaction essence from tx builder data
func (txb *AnchorTransactionBuilder) BuildTransactionEssence(stateMetadata []byte, creationSlot iotago.SlotIndex) *iotago.Transaction {
	inputs, inputIDs := txb.inputs()
	return &iotago.Transaction{
		API: parameters.L1API(),
		TransactionEssence: &iotago.TransactionEssence{
			CreationSlot: creationSlot,
			NetworkID:    parameters.Protocol().NetworkID(),
			Inputs:       inputIDs.UTXOInputs(),
		},
		Outputs: txb.outputs(stateMetadata, creationSlot, inputs),
	}
}

// inputIDs generates a deterministic list of inputs for the transaction essence
// - index 0 is always alias output
// - then goes consumed external BasicOutput UTXOs, the requests, in the order of processing
// - then goes inputs of native token UTXOs, sorted by token id
// - then goes inputs of foundries sorted by serial number
func (txb *AnchorTransactionBuilder) inputs() (iotago.OutputSet, iotago.OutputIDs) {
	outputIDs := make(iotago.OutputIDs, 0, len(txb.consumed)+len(txb.balanceNativeTokens)+len(txb.invokedFoundries))
	inputs := make(iotago.OutputSet)

	// alias output
	outputIDs = append(outputIDs, txb.anchorOutputID)
	inputs[txb.anchorOutputID] = txb.anchorOutput

	// consumed on-ledger requests
	for i := range txb.consumed {
		req := txb.consumed[i]
		outputID := req.OutputID()
		output := req.Output()
		if retrReq, ok := req.(*isc.RetryOnLedgerRequest); ok {
			outputID = retrReq.RetryOutputID()
			output = retryOutputFromOnLedgerRequest(req, txb.anchorOutput.AccountID)
		}
		outputIDs = append(outputIDs, outputID)
		inputs[outputID] = output
	}

	// internal native token outputs
	for _, nativeTokenBalance := range txb.nativeTokenOutputsSorted() {
		if nativeTokenBalance.requiresExistingAccountingUTXOAsInput() {
			outputID := nativeTokenBalance.accountingInputID
			outputIDs = append(outputIDs, outputID)
			inputs[outputID] = nativeTokenBalance.accountingInput
		}
	}

	// foundries
	for _, foundry := range txb.foundriesSorted() {
		if foundry.requiresExistingAccountingUTXOAsInput() {
			outputID := foundry.accountingInputID
			outputIDs = append(outputIDs, outputID)
			inputs[outputID] = foundry.accountingInput
		}
	}

	// nfts
	for _, nft := range txb.nftsSorted() {
		if !isc.IsEmptyOutputID(nft.accountingInputID) {
			outputID := nft.accountingInputID
			outputIDs = append(outputIDs, outputID)
			inputs[outputID] = nft.accountingInput
		}
	}

	if len(outputIDs) != txb.numInputs() {
		panic(fmt.Sprintf("AnchorTransactionBuilder.inputs: internal inconsistency. expected: %d actual:%d", len(outputIDs), txb.numInputs()))
	}
	return inputs, outputIDs
}

func (txb *AnchorTransactionBuilder) CreateAnchorOutput(
	stateMetadata []byte,
	creationSlot iotago.SlotIndex,
	inputs iotago.OutputSet,
) *iotago.AccountOutput {
	aliasID := txb.anchorOutput.AccountID
	if aliasID.Empty() {
		aliasID = iotago.AccountIDFromOutputID(txb.anchorOutputID)
	}
	anchorOutput := &iotago.AccountOutput{
		Amount:         0,
		AccountID:      aliasID,
		StateIndex:     txb.anchorOutput.StateIndex + 1,
		StateMetadata:  stateMetadata,
		FoundryCounter: txb.nextFoundryCounter(),
		Conditions: iotago.AccountOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: txb.anchorOutput.StateController()},
			&iotago.GovernorAddressUnlockCondition{Address: txb.anchorOutput.GovernorAddress()},
		},
		Features: iotago.AccountOutputFeatures{
			&iotago.SenderFeature{
				Address: aliasID.ToAddress(),
			},
		},
	}
	if metadata := txb.anchorOutput.FeatureSet().Metadata(); metadata != nil {
		anchorOutput.Features = append(anchorOutput.Features,
			&iotago.MetadataFeature{
				Data: metadata.Data,
			},
		)
	}

	minSD, err := parameters.Storage().MinDeposit(anchorOutput)
	if err != nil {
		panic(err)
	}
	anchorOutput.Amount = txb.accountsView.TotalFungibleTokens().BaseTokens + minSD
	mana, err := vm.TotalManaIn(
		parameters.L1API().ManaDecayProvider(),
		parameters.Storage(),
		creationSlot,
		vm.InputSet(inputs),
	)
	if err != nil {
		panic(err)
	}
	anchorOutput.Mana = mana
	return anchorOutput
}

// outputs generates outputs for the transaction essence
// IMPORTANT: the order that assets are added here must not change, otherwise vmctx.saveInternalUTXOs will be broken.
// 0. Anchor Output
// 1. NativeTokens
// 2. Foundries
// 3. received NFTs
// 4. minted NFTs
// 5. other outputs (posted from requests)
func (txb *AnchorTransactionBuilder) outputs(
	stateMetadata []byte,
	creationSlot iotago.SlotIndex,
	inputs iotago.OutputSet,
) iotago.TxEssenceOutputs {
	ret := make(iotago.TxEssenceOutputs, 0, 1+len(txb.balanceNativeTokens)+len(txb.postedOutputs))

	txb.resultAnchorOutput = txb.CreateAnchorOutput(stateMetadata, creationSlot, inputs)
	ret = append(ret, txb.resultAnchorOutput)

	// creating outputs for updated internal accounts
	nativeTokensToBeUpdated, _ := txb.NativeTokenRecordsToBeUpdated()
	for _, id := range nativeTokensToBeUpdated {
		// create one output for each token ID of internal account
		ret = append(ret, txb.balanceNativeTokens[id].accountingOutput)
	}
	// creating outputs for updated foundries
	foundriesToBeUpdated, _ := txb.FoundriesToBeUpdated()
	for _, sn := range foundriesToBeUpdated {
		ret = append(ret, txb.invokedFoundries[sn].accountingOutput)
	}
	// creating outputs for received NFTs
	nftOuts := txb.NFTOutputs()
	for _, nftOut := range nftOuts {
		ret = append(ret, nftOut)
	}
	// creating outputs for minted NFTs
	ret = append(ret, txb.nftsMinted...)

	// creating outputs for posted on-ledger requests
	ret = append(ret, txb.postedOutputs...)
	return ret
}

// numInputs number of inputs in the future transaction
func (txb *AnchorTransactionBuilder) numInputs() int {
	ret := len(txb.consumed) + 1 // + 1 for anchor UTXO
	for _, v := range txb.balanceNativeTokens {
		if v.requiresExistingAccountingUTXOAsInput() {
			ret++
		}
	}
	for _, f := range txb.invokedFoundries {
		if f.requiresExistingAccountingUTXOAsInput() {
			ret++
		}
	}
	for _, nft := range txb.nftsIncluded {
		if !isc.IsEmptyOutputID(nft.accountingInputID) {
			ret++
		}
	}
	return ret
}

// numOutputs in the transaction
func (txb *AnchorTransactionBuilder) numOutputs() int {
	ret := 1 // for chain output
	for _, v := range txb.balanceNativeTokens {
		if v.producesAccountingOutput() {
			ret++
		}
	}
	ret += len(txb.postedOutputs)
	for _, f := range txb.invokedFoundries {
		if f.producesAccountingOutput() {
			ret++
		}
	}
	ret += len(txb.nftsMinted)
	return ret
}

// outputsAreFull return if transaction cannot bear more outputs
func (txb *AnchorTransactionBuilder) outputsAreFull() bool {
	return txb.numOutputs() >= iotago.MaxOutputsCount
}

func (txb *AnchorTransactionBuilder) AnchorOutputStorageDeposit() iotago.BaseToken {
	return txb.anchorOutputStorageDeposit
}

func retryOutputFromOnLedgerRequest(req isc.OnLedgerRequest, chainAccountID iotago.AccountID) iotago.Output {
	out := req.Output().Clone()

	feature := &iotago.SenderFeature{
		Address: chainAccountID.ToAddress(), // must have the chain as the sender, so its recognized as an internalUTXO
	}

	unlock := &iotago.AddressUnlockCondition{
		Address: chainAccountID.ToAddress(),
	}

	// cleanup features and unlock conditions except metadata
	switch o := out.(type) {
	case *iotago.BasicOutput:
		o.Features = iotago.BasicOutputFeatures{feature}
		o.Conditions = iotago.BasicOutputUnlockConditions{unlock}
	case *iotago.NFTOutput:
		o.Features = iotago.NFTOutputFeatures{feature}
		o.Conditions = iotago.NFTOutputUnlockConditions{unlock}
	case *iotago.AccountOutput:
		o.Features = iotago.AccountOutputFeatures{feature}
		o.Conditions = iotago.AccountOutputUnlockConditions{unlock}
	default:
		panic("unexpected output type")
	}
	return out
}

func (txb *AnchorTransactionBuilder) chainAddress() iotago.Address {
	return txb.anchorOutput.AccountID.ToAddress()
}
