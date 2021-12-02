package transaction

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

// NewChainOriginTransaction creates new origin transaction for the self-governed chain
// returns the transaction and newly minted chain ID
func NewChainOriginTransaction(
	key ed25519.PrivateKey,
	stateControllerAddress iotago.Address,
	governanceControllerAddress iotago.Address,
	deposit uint64,
	allUnspentOutputs []iotago.Output,
	allInputs []*iotago.UTXOInput,
	deSeriParams *iotago.DeSerializationParameters,
) (*iotago.Transaction, iotago.ChainID, error) {
	if len(allUnspentOutputs) != len(allInputs) {
		panic("mismatched lengths of outputs and inputs slices")
	}

	walletAddr := iotago.Ed25519AddressFromPubKey(key.Public().(ed25519.PublicKey))

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

	inputs, remainder, err := computeInputsAndRemainder(
		aliasOutput.Amount,
		allUnspentOutputs,
		allInputs,
		deSeriParams,
	)
	if err != nil {
		return nil, nil, err
	}
	for _, input := range inputs {
		txb.AddInput(&iotago.ToBeSignedUTXOInput{Address: &walletAddr, Input: input})
	}
	if remainder > 0 {
		txb.AddOutput(&iotago.ExtendedOutput{
			Address: &walletAddr,
			Amount:  remainder,
		})
	}

	signer := iotago.NewInMemoryAddressSigner(iotago.NewAddressKeysForEd25519Address(&walletAddr, key))
	tx, err := txb.Build(deSeriParams, signer)
	if err != nil {
		return nil, nil, err
	}
	return tx, aliasOutput.Chain(), nil
}

func computeInputsAndRemainder(
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

// NewRootInitRequestTransaction is a first request to be sent to the uninitialized
// chain. At this moment it only is able to process this specific request
// the request contains minimum data needed to bootstrap the chain
// TransactionEssence must be signed by the same address which created origin transaction
func NewRootInitRequestTransaction(
	keyPair *ed25519.KeyPair,
	chainID *iscp.ChainID,
	description string,
	timestamp time.Time,
	allInputs ...ledgerstate.Output,
) (*ledgerstate.Transaction, error) {
	txb := utxoutil.NewBuilder(allInputs...).WithTimestamp(timestamp)

	args := dict.New()

	args.Set(governance.ParamChainID, codec.EncodeChainID(chainID))
	args.Set(governance.ParamDescription, codec.EncodeString(description))

	metadata := request.NewMetadata().
		WithTarget(iscp.Hn("root")).
		WithEntryPoint(iscp.EntryPointInit).
		WithArgs(args).
		Bytes()

	err := txb.AddExtendedOutputConsume(chainID.AsAddress(), metadata, colored.Balances1IotaL1)
	if err != nil {
		return nil, err
	}
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	if err := txb.AddRemainderOutputIfNeeded(addr, nil, true); err != nil {
		return nil, err
	}
	tx, err := txb.BuildWithED25519(keyPair)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
