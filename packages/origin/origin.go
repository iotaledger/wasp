package origin

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// L1Commitment calculates the L1 commitment for the origin state
// originDeposit must exclude the minSD for the AliasOutput
func L1Commitment(v isc.SchemaVersion, initParams dict.Dict, originDeposit uint64) *state.L1Commitment {
	block := InitChain(v, state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()), initParams, originDeposit)
	return block.L1Commitment()
}

const (
	ParamEVMChainID               = "a"
	ParamBlockKeepAmount          = "b"
	ParamChainOwner               = "c"
	ParamWaspVersion              = "d"
	ParamDeployBaseTokenMagicWrap = "m"
)

func InitChain(v isc.SchemaVersion, store state.Store, initParams dict.Dict, originDeposit uint64) state.Block {
	if initParams == nil {
		initParams = dict.New()
	}
	d := store.NewOriginStateDraft()
	d.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode(uint32(0)))
	d.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.Time.Encode(time.Unix(0, 0)))

	evmChainID := codec.Uint16.MustDecode(initParams.Get(ParamEVMChainID), evm.DefaultChainID)
	blockKeepAmount := codec.Int32.MustDecode(initParams.Get(ParamBlockKeepAmount), governance.DefaultBlockKeepAmount)
	chainOwner := codec.AgentID.MustDecode(initParams.Get(ParamChainOwner), &isc.NilAgentID{})
	deployMagicWrap := codec.Bool.MustDecode(initParams.Get(ParamDeployBaseTokenMagicWrap), false)

	// init the state of each core contract
	root.NewStateWriter(root.Contract.StateSubrealm(d)).SetInitialState(v, []*coreutil.ContractInfo{
		root.Contract,
		blob.Contract,
		accounts.Contract,
		blocklog.Contract,
		errors.Contract,
		governance.Contract,
		evm.Contract,
	})
	blob.NewStateWriter(blob.Contract.StateSubrealm(d)).SetInitialState()
	accounts.NewStateWriter(v, accounts.Contract.StateSubrealm(d)).SetInitialState(originDeposit)
	blocklog.NewStateWriter(blocklog.Contract.StateSubrealm(d)).SetInitialState()
	errors.NewStateWriter(errors.Contract.StateSubrealm(d)).SetInitialState()
	governance.NewStateWriter(governance.Contract.StateSubrealm(d)).SetInitialState(chainOwner, blockKeepAmount)
	evmimpl.SetInitialState(evm.Contract.StateSubrealm(d), evmChainID, deployMagicWrap)

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

	panic("refactor me: parameters.L1() / RentStructure")
	_ = initParams
	// l1params := parameters.L1()
	//aoMinSD := l1params.Protocol.RentStructure.MinRent(aliasOutput.GetAliasOutput())

	/* commonAccountAmount := aliasOutput.GetAliasOutput().Amount - aoMinSD
	originAOStateMetadata, err := transaction.StateMetadataFromBytes(aliasOutput.GetStateMetadata())
	originBlock := InitChain(originAOStateMetadata.SchemaVersion, chainStore, initParams, commonAccountAmount)

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
	return originBlock, nil */

	return nil, nil
}
