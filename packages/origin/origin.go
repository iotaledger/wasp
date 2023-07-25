package origin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/legacymigration"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/root/rootimpl"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// L1Commitment calculates the L1 commitment for the origin state
// originDeposit must exclude the minSD for the AliasOutput
func L1Commitment(initParams dict.Dict, originDeposit uint64) *state.L1Commitment {
	block := InitChain(state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()), initParams, originDeposit)
	return block.L1Commitment()
}

const (
	ParamEVMChainID      = "a"
	ParamBlockKeepAmount = "b"
	ParamChainOwner      = "c"
	ParamWaspVersion     = "d"
)

func InitChain(store state.Store, initParams dict.Dict, originDeposit uint64) state.Block {
	if initParams == nil {
		initParams = dict.New()
	}
	d := store.NewOriginStateDraft()
	d.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode(uint32(0)))
	d.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(time.Unix(0, 0)))

	contractState := func(contract *coreutil.ContractInfo) kv.KVStore {
		return subrealm.New(d, kv.Key(contract.Hname().Bytes()))
	}

	evmChainID := codec.MustDecodeUint16(initParams.Get(ParamEVMChainID), evm.DefaultChainID)
	blockKeepAmount := codec.MustDecodeInt32(initParams.Get(ParamBlockKeepAmount), governance.DefaultBlockKeepAmount)
	chainOwner := codec.MustDecodeAgentID(initParams.Get(ParamChainOwner), &isc.NilAgentID{})

	// init the state of each core contract
	rootimpl.SetInitialState(contractState(root.Contract))
	blob.SetInitialState(contractState(blob.Contract))
	accounts.SetInitialState(contractState(accounts.Contract), originDeposit)
	blocklog.SetInitialState(contractState(blocklog.Contract))
	errors.SetInitialState(contractState(errors.Contract))
	governanceimpl.SetInitialState(contractState(governance.Contract), chainOwner, blockKeepAmount)
	evmimpl.SetInitialState(contractState(evm.Contract), evmChainID)
	legacymigration.SetInitialState(contractState(legacymigration.Contract))

	block := store.Commit(d)
	if err := store.SetLatest(block.TrieRoot()); err != nil {
		panic(err)
	}
	return block
}

func InitChainByAliasOutput(chainStore state.Store, aliasOutput *isc.AliasOutputWithID) (state.Block, error) {
	var initParams dict.Dict
	if originMetadata := aliasOutput.GetAliasOutput().FeatureSet().MetadataFeature(); originMetadata != nil {
		var err error
		initParams, err = dict.FromBytes(originMetadata.Data)
		if err != nil {
			return nil, fmt.Errorf("invalid parameters on origin AO, %w", err)
		}
	}
	l1params := parameters.L1()
	aoMinSD := l1params.Protocol.RentStructure.MinRent(aliasOutput.GetAliasOutput())
	commonAccountAmount := aliasOutput.GetAliasOutput().Amount - aoMinSD
	originBlock := InitChain(chainStore, initParams, commonAccountAmount)

	originAOStateMetadata, err := transaction.StateMetadataFromBytes(aliasOutput.GetStateMetadata())
	if err != nil {
		return nil, fmt.Errorf("invalid state metadata on origin AO: %w", err)
	}
	if originAOStateMetadata.Version != transaction.StateMetadataSupportedVersion {
		return nil, fmt.Errorf("unsupported StateMetadata Version: %v, expect %v", originAOStateMetadata.Version, transaction.StateMetadataSupportedVersion)
	}
	if !originBlock.L1Commitment().Equals(originAOStateMetadata.L1Commitment) {
		l1paramsJSON, err := json.Marshal(l1params)
		if err != nil {
			l1paramsJSON = []byte(fmt.Sprintf("unable to marshalJson l1params: %s", err.Error()))
		}
		return nil, fmt.Errorf(
			"l1Commitment mismatch between originAO / originBlock: %s / %s, AOminSD: %d, L1params: %s",
			originAOStateMetadata.L1Commitment,
			originBlock.L1Commitment(),
			aoMinSD,
			string(l1paramsJSON),
		)
	}
	return originBlock, nil
}

func calcStateMetadata(initParams dict.Dict, commonAccountAmount uint64, schemaVersion uint32) []byte {
	s := transaction.NewStateMetadata(
		L1Commitment(initParams, commonAccountAmount),
		gas.DefaultFeePolicy(),
		schemaVersion,
		"",
	)
	return s.Bytes()
}

// NewChainOriginTransaction creates new origin transaction for the self-governed chain
// returns the transaction and newly minted chain ID
func NewChainOriginTransaction(
	keyPair *cryptolib.KeyPair,
	stateControllerAddress iotago.Address,
	governanceControllerAddress iotago.Address,
	deposit uint64,
	initParams dict.Dict,
	unspentOutputs iotago.OutputSet,
	unspentOutputIDs iotago.OutputIDs,
	schemaVersion uint32,
) (*iotago.Transaction, *iotago.AliasOutput, isc.ChainID, error) {
	if len(unspentOutputs) != len(unspentOutputIDs) {
		panic("mismatched lengths of outputs and inputs slices")
	}

	walletAddr := keyPair.GetPublicKey().AsEd25519Address()

	if initParams == nil {
		initParams = dict.New()
	}
	if initParams.Get(ParamChainOwner) == nil {
		// default chain owner to the gov address
		initParams.Set(ParamChainOwner, isc.NewAgentID(governanceControllerAddress).Bytes())
	}

	aliasOutput := &iotago.AliasOutput{
		Amount:        deposit,
		StateMetadata: calcStateMetadata(initParams, deposit, schemaVersion), // NOTE: Updated below.
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateControllerAddress},
			&iotago.GovernorAddressUnlockCondition{Address: governanceControllerAddress},
		},
		Features: iotago.Features{
			&iotago.MetadataFeature{Data: initParams.Bytes()},
		},
	}

	minSD := parameters.L1().Protocol.RentStructure.MinRent(aliasOutput)
	minAmount := minSD + governance.DefaultMinBaseTokensOnCommonAccount
	if aliasOutput.Amount < minAmount {
		aliasOutput.Amount = minAmount
	}
	// update the L1 commitment to not include the minimumSD
	aliasOutput.StateMetadata = calcStateMetadata(initParams, aliasOutput.Amount-minSD, schemaVersion)

	txInputs, remainderOutput, err := transaction.ComputeInputsAndRemainder(
		walletAddr,
		aliasOutput.Amount,
		nil,
		nil,
		unspentOutputs,
		unspentOutputIDs,
	)
	if err != nil {
		return nil, aliasOutput, isc.ChainID{}, err
	}
	outputs := iotago.Outputs{aliasOutput}
	if remainderOutput != nil {
		outputs = append(outputs, remainderOutput)
	}
	essence := &iotago.TransactionEssence{
		NetworkID: parameters.L1().Protocol.NetworkID(),
		Inputs:    txInputs.UTXOInputs(),
		Outputs:   outputs,
	}
	sigs, err := essence.Sign(
		txInputs.OrderedSet(unspentOutputs).MustCommitment(),
		keyPair.GetPrivateKey().AddressKeysForEd25519Address(walletAddr),
	)
	if err != nil {
		return nil, aliasOutput, isc.ChainID{}, err
	}
	tx := &iotago.Transaction{
		Essence: essence,
		Unlocks: transaction.MakeSignatureAndReferenceUnlocks(len(txInputs), sigs[0]),
	}
	txid, err := tx.ID()
	if err != nil {
		return nil, aliasOutput, isc.ChainID{}, err
	}
	chainID := isc.ChainIDFromAliasID(iotago.AliasIDFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(txid, 0)))
	return tx, aliasOutput, chainID, nil
}
