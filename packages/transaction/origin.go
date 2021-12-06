package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// NewChainOriginTransaction creates new origin transaction for the self-governed chain
// returns the transaction and newly minted chain ID
func NewChainOriginTransaction(
	keyPair cryptolib.KeyPair,
	stateControllerAddress iotago.Address,
	governanceControllerAddress iotago.Address,
	deposit uint64,
	allUnspentOutputs []iotago.Output,
	allInputs []*iotago.UTXOInput,
	deSeriParams *iotago.DeSerializationParameters,
) (*iotago.Transaction, *iscp.ChainID, error) {
	if len(allUnspentOutputs) != len(allInputs) {
		panic("mismatched lengths of outputs and inputs slices")
	}

	walletAddr := cryptolib.Ed25519AddressFromPubKey(keyPair.PublicKey)

	txb := iotago.NewTransactionBuilder()

	aliasOutput := &iotago.AliasOutput{
		Amount:               deposit,
		StateController:      stateControllerAddress,
		GovernanceController: governanceControllerAddress,
		StateMetadata:        state.OriginStateHash().Bytes(),
	}
	{
		aliasDustDeposit := aliasOutput.VByteCost(deSeriParams.RentStructure, nil)
		if aliasOutput.Amount < aliasDustDeposit {
			aliasOutput.Amount = aliasDustDeposit
		}
	}
	txb.AddOutput(aliasOutput)

	inputs, remainderOutput, err := computeInputsAndRemainder(
		walletAddr,
		aliasOutput.Amount,
		nil,
		allUnspentOutputs,
		allInputs,
		deSeriParams,
	)
	if err != nil {
		return nil, nil, err
	}
	if remainderOutput != nil {
		txb.AddOutput(remainderOutput)
	}
	for _, input := range inputs {
		txb.AddInput(&iotago.ToBeSignedUTXOInput{Address: walletAddr, Input: input})
	}

	signer := iotago.NewInMemoryAddressSigner(iotago.NewAddressKeysForEd25519Address(walletAddr, keyPair.PrivateKey))
	tx, err := txb.Build(deSeriParams, signer)
	if err != nil {
		return nil, nil, err
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
	keyPair cryptolib.KeyPair,
	chainID *iscp.ChainID,
	description string,
	allUnspentOutputs []iotago.Output,
	allInputs []*iotago.UTXOInput,
	deSeriParams *iotago.DeSerializationParameters,
) (*iotago.Transaction, error) {
	//
	tx, err := NewRequestTransaction(NewRequestTransactionParams{
		SenderKeyPair:    keyPair,
		UnspentOutputs:   allUnspentOutputs,
		UnspentOutputIDs: allInputs,
		Requests: []*iscp.RequestParameters{{
			TargetAddress: chainID.AsAddress(),
			Metadata: &iscp.SendMetadata{
				TargetContract: root.Contract.Hname(),
				EntryPoint:     iscp.EntryPointInit,
				GasBudget:      0, // TODO. Probably we need minimum fixed budget for core contract calls. 0 for init call
				Params: dict.Dict{
					governance.ParamDescription: codec.EncodeString(description),
				},
			},
		}},
		DeSeriParams: deSeriParams,
	})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

//func NewRootInitRequestTransactionOld(
//	key ed25519.PrivateKey,
//	chainID *iscp.ChainID,
//	description string,
//	allUnspentOutputs []iotago.Output,
//	allInputs []*iotago.UTXOInput,
//	deSeriParams *iotago.DeSerializationParameters,
//) (*iotago.Transaction, error) {
//	walletAddr := iotago.Ed25519AddressFromPubKey(key.Public().(ed25519.PublicKey))
//
//	args := dict.Dict{
//		governance.ParamChainID:     codec.EncodeChainID(chainID),
//		governance.ParamDescription: codec.EncodeString(description),
//	}
//
//	metadata := &iscp.RequestMetadata{
//		TargetContract: iscp.Hn("root"),
//		EntryPoint:     iscp.EntryPointInit,
//		Params:         args,
//	}
//
//	txb := iotago.NewTransactionBuilder()
//
//	requestOutput := &iotago.ExtendedOutput{
//		Address: chainID.AsAddress(),
//		Amount:  0,
//		Blocks: []iotago.FeatureBlock{
//			&iotago.MetadataFeatureBlock{
//				Data: metadata.Bytes(),
//			},
//		},
//	}
//	requestOutput.Amount = requestOutput.VByteCost(deSeriParams.RentStructure, nil)
//	txb.AddOutput(requestOutput)
//
//	inputs, remainder, err := computeInputsAndRemainderOld(
//		requestOutput.Amount,
//		allUnspentOutputs,
//		allInputs,
//		deSeriParams,
//	)
//	if err != nil {
//		return nil, err
//	}
//	for _, input := range inputs {
//		txb.AddInput(&iotago.ToBeSignedUTXOInput{Address: &walletAddr, Input: input})
//	}
//	if remainder > 0 {
//		txb.AddOutput(&iotago.ExtendedOutput{
//			Address: &walletAddr,
//			Amount:  remainder,
//		})
//	}
//
//	signer := iotago.NewInMemoryAddressSigner(iotago.NewAddressKeysForEd25519Address(&walletAddr, key))
//	tx, err := txb.Build(deSeriParams, signer)
//	if err != nil {
//		return nil, err
//	}
//	return tx, nil
//}
