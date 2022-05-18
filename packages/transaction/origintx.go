package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// NewChainOriginTransaction creates new origin transaction for the self-governed chain
// returns the transaction and newly minted chain ID
func NewChainOriginTransaction(
	keyPair *cryptolib.KeyPair,
	stateControllerAddress iotago.Address,
	governanceControllerAddress iotago.Address,
	deposit uint64,
	unspentOutputs iotago.OutputSet,
	unspentOutputIDs iotago.OutputIDs,
) (*iotago.Transaction, *iscp.ChainID, error) {
	if len(unspentOutputs) != len(unspentOutputIDs) {
		panic("mismatched lengths of outputs and inputs slices")
	}

	walletAddr := keyPair.GetPublicKey().AsEd25519Address()

	aliasOutput := &iotago.AliasOutput{
		Amount:        deposit,
		StateMetadata: state.OriginL1Commitment().Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateControllerAddress},
			&iotago.GovernorAddressUnlockCondition{Address: governanceControllerAddress},
		},
		Blocks: iotago.FeatureBlocks{
			&iotago.SenderFeatureBlock{
				Address: walletAddr,
			},
		},
	}
	{
		aliasDustDeposit := NewDepositEstimate().AnchorOutput
		if aliasOutput.Amount < aliasDustDeposit {
			aliasOutput.Amount = aliasDustDeposit
		}
	}
	txInputs, remainderOutput, err := computeInputsAndRemainder(
		walletAddr,
		aliasOutput.Amount,
		nil,
		nil,
		unspentOutputs,
		unspentOutputIDs,
	)
	if err != nil {
		return nil, nil, err
	}
	outputs := iotago.Outputs{aliasOutput}
	if remainderOutput != nil {
		outputs = append(outputs, remainderOutput)
	}
	essence := &iotago.TransactionEssence{
		NetworkID: parameters.L1.Protocol.NetworkID(),
		Inputs:    txInputs.UTXOInputs(),
		Outputs:   outputs,
	}
	sigs, err := essence.Sign(
		txInputs.OrderedSet(unspentOutputs).MustCommitment(),
		keyPair.GetPrivateKey().AddressKeysForEd25519Address(walletAddr),
	)
	if err != nil {
		return nil, nil, err
	}
	tx := &iotago.Transaction{
		Essence:      essence,
		UnlockBlocks: MakeSignatureAndReferenceUnlockBlocks(len(txInputs), sigs[0]),
	}
	txid, err := tx.ID()
	if err != nil {
		return nil, nil, err
	}
	chainID := iscp.ChainIDFromAliasID(iotago.AliasIDFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(*txid, 0)))
	return tx, &chainID, nil
}

// NewRootInitRequestTransaction is a transaction with one request output.
// It is the first request to be sent to the uninitialized
// chain. At this moment it is only able to process this specific request.
// The request contains the minimum data needed to bootstrap the chain.
// The signer must be the same that created the origin transaction.
func NewRootInitRequestTransaction(
	keyPair *cryptolib.KeyPair,
	chainID *iscp.ChainID,
	description string,
	unspentOutputs iotago.OutputSet,
	unspentOutputIDs iotago.OutputIDs,
	initParams ...dict.Dict,
) (*iotago.Transaction, error) {
	params := dict.Dict{
		root.ParamDustDepositAssumptionsBin: NewDepositEstimate().Bytes(),
		governance.ParamDescription:         codec.EncodeString(description),
	}
	for _, p := range initParams {
		params.Extend(p)
	}
	tx, err := NewRequestTransaction(NewRequestTransactionParams{
		SenderKeyPair:    keyPair,
		SenderAddress:    keyPair.Address(),
		UnspentOutputs:   unspentOutputs,
		UnspentOutputIDs: unspentOutputIDs,
		Request: &iscp.RequestParameters{
			TargetAddress: chainID.AsAddress(),
			Metadata: &iscp.SendMetadata{
				TargetContract: root.Contract.Hname(),
				EntryPoint:     iscp.EntryPointInit,
				GasBudget:      0, // TODO. Probably we need minimum fixed budget for core contract calls. 0 for init call
				Params:         params,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return tx, nil
}
