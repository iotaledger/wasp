package origin

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
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
)

func L1Commitment(initParams dict.Dict, originDeposit uint64) *state.L1Commitment {
	store := InitChain(state.NewStore(mapdb.NewMapDB()), initParams, originDeposit)
	block, err := store.LatestBlock()
	if err != nil {
		panic(err)
	}
	return block.L1Commitment()
}

const (
	ParamEVMChainID   = "a"
	ParamEVMBlockKeep = "b"
	ParamChainOwner   = "c"
)

func InitChain(store state.Store, initParams dict.Dict, originDeposit uint64) state.Store {
	if initParams == nil {
		initParams = dict.New()
	}
	d := store.NewOriginStateDraft()
	d.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode(uint32(0)))
	d.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(time.Unix(0, 0)))

	contractState := func(contract *coreutil.ContractInfo) kv.KVStore {
		return subrealm.New(d, kv.Key(contract.Hname().Bytes()))
	}

	evmChainID := evmtypes.MustDecodeChainID(initParams.MustGet(ParamEVMChainID), evm.DefaultChainID)
	blockKeepAmount := codec.MustDecodeInt32(initParams.MustGet(ParamEVMBlockKeep), evm.BlockKeepAmountDefault)

	chainOwner := codec.MustDecodeAgentID(initParams.MustGet(ParamChainOwner), &isc.NilAgentID{})

	// init the state of each core contract
	rootimpl.SetInitialState(contractState(root.Contract))
	blob.SetInitialState(contractState(blob.Contract))
	accounts.SetInitialState(contractState(accounts.Contract), originDeposit)
	blocklog.SetInitialState(contractState(blocklog.Contract))
	errors.SetInitialState(contractState(errors.Contract))
	governanceimpl.SetInitialState(contractState(governance.Contract), chainOwner)
	evmimpl.SetInitialState(contractState(evm.Contract), evmChainID, blockKeepAmount)

	// set block context subscriptions
	root.SubscribeBlockContext(
		contractState(root.Contract),
		evm.Contract.Hname(),
		evm.FuncOpenBlockContext.Hname(),
		evm.FuncCloseBlockContext.Hname(),
	)

	block := store.Commit(d)
	if err := store.SetLatest(block.TrieRoot()); err != nil {
		panic(err)
	}
	return store
}

func InitChainByAliasOutput(chainStore state.Store, aliasOutput *isc.AliasOutputWithID) {
	var initParams dict.Dict
	if originMetadata := aliasOutput.GetAliasOutput().FeatureSet().MetadataFeature(); originMetadata != nil {
		var err error
		initParams, err = dict.FromBytes(originMetadata.Data)
		if err != nil {
			panic(fmt.Sprintf("invalid parameters on origin AO, %s", err.Error()))
		}
	}
	aoSD := transaction.NewStorageDepositEstimate().AnchorOutput
	InitChain(chainStore, initParams, aliasOutput.GetAliasOutput().Amount-aoSD)
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
) (*iotago.Transaction, *iotago.AliasOutput, isc.ChainID, error) {
	if len(unspentOutputs) != len(unspentOutputIDs) {
		panic("mismatched lengths of outputs and inputs slices")
	}

	walletAddr := keyPair.GetPublicKey().AsEd25519Address()

	if initParams == nil {
		initParams = dict.New()
	}
	if initParams.MustGet(ParamChainOwner) == nil {
		// default chain owner to the gov address
		initParams.Set(ParamChainOwner, isc.NewAgentID(governanceControllerAddress).Bytes())
	}

	minSD := transaction.NewStorageDepositEstimate().AnchorOutput
	minAmount := minSD + accounts.MinimumBaseTokensOnCommonAccount
	if deposit < minAmount {
		deposit = minAmount
	}

	aliasOutput := &iotago.AliasOutput{
		Amount:        deposit,
		StateMetadata: L1Commitment(initParams, deposit-minSD).Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateControllerAddress},
			&iotago.GovernorAddressUnlockCondition{Address: governanceControllerAddress},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: walletAddr,
			},
			&iotago.MetadataFeature{Data: initParams.Bytes()},
		},
	}

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
