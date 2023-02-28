package origin

import (
	"time"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
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

	chainOwner := codec.MustDecodeAgentID(initParams.MustGet(governance.ParamChainOwner), &isc.NilAgentID{})

	blockKeepAmount := codec.MustDecodeInt32(initParams.MustGet(evm.FieldBlockKeepAmount), evm.BlockKeepAmountDefault)
	evmChainID := evmtypes.MustDecodeChainID(initParams.MustGet(evm.FieldChainID), evm.DefaultChainID)

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
