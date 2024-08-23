package origin

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
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
func L1Commitment(v isc.SchemaVersion, initParams isc.CallArguments, originDeposit coin.Value) *state.L1Commitment {
	block := InitChain(v, state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB()), initParams, originDeposit)
	return block.L1Commitment()
}

func InitChain(v isc.SchemaVersion, store state.Store, initParams isc.CallArguments, originDeposit coin.Value) state.Block {
	d := store.NewOriginStateDraft()
	d.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode(uint32(0)))
	d.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.Time.Encode(time.Unix(0, 0)))

	chainOwner := codec.AgentID.MustDecode(initParams.MustAt(0))
	evmChainID := codec.Uint16.MustDecode(initParams.OrNil(1), evm.DefaultChainID)
	blockKeepAmount := codec.Int32.MustDecode(initParams.OrNil(2), governance.DefaultBlockKeepAmount)

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
	evmimpl.SetInitialState(evm.Contract.StateSubrealm(d), evmChainID)

	block := store.Commit(d)
	if err := store.SetLatest(block.TrieRoot()); err != nil {
		panic(err)
	}
	return block
}

func InitChainByAnchor(
	chainStore state.Store,
	anchor *iscmove.RefWithObject[iscmove.Anchor],
	anchorAssets isc.Assets,
) (state.Block, error) {
	stateMetadata, err := transaction.StateMetadataFromBytes(anchor.Object.StateMetadata)
	if err != nil {
		return nil, err
	}
	originBlock := InitChain(
		stateMetadata.SchemaVersion,
		chainStore,
		stateMetadata.InitParams,
		anchorAssets.BaseTokens(),
	)
	if !originBlock.L1Commitment().Equals(stateMetadata.L1Commitment) {
		return nil, fmt.Errorf(
			"l1Commitment mismatch between originAO / originBlock: %s / %s",
			stateMetadata.L1Commitment,
			originBlock.L1Commitment(),
		)
	}
	return originBlock, nil
}
