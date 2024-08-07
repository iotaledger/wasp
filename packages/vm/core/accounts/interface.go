package accounts

import (
	"math"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/sui-go/sui"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts)

var (
	// Funcs
	FuncDeposit = coreutil.NewEP0(Contract, "deposit")
	// TODO: adapt to iota-rebased
	//   FuncFoundryCreateNew
	//   FuncCoinCreate
	//   FuncCoinModifySupply
	//   FuncCoinDestroy
	//   FuncMintObject
	FuncTransferAccountToChain = coreutil.NewEP1(Contract, "transferAccountToChain",
		coreutil.FieldWithCodecOptional("i1", codec.Uint64),
	)
	FuncTransferAllowanceTo = coreutil.NewEP1(Contract, "transferAllowanceTo",
		coreutil.FieldWithCodec("i1", codec.AgentID),
	)
	FuncWithdraw = coreutil.NewEP0(Contract, "withdraw")

	// Views
	// TODO: implement pagination
	ViewAccountTreasuries = coreutil.NewViewEP11(Contract, "accountTreasuries",
		coreutil.FieldWithCodecOptional("i1", codec.AgentID),
		FieldArrayOf[isc.CoinType]{codec: codec.CoinType},
	)
	// TODO: implement pagination
	ViewAccountObjects = coreutil.NewViewEP11(Contract, "accountObjects",
		coreutil.FieldWithCodecOptional("i1", codec.AgentID),
		FieldArrayOf[sui.ObjectID]{codec: codec.ObjectID},
	)
	// TODO: implement pagination
	ViewAccountObjectsInCollection = coreutil.NewViewEP21(Contract, "accountObjectsInCollection",
		coreutil.FieldWithCodecOptional("i1", codec.AgentID),
		coreutil.FieldWithCodec("i2", codec.ObjectID),
		FieldArrayOf[sui.ObjectID]{codec: codec.ObjectID},
	)
	// TODO: implement pagination
	ViewBalance = coreutil.NewViewEP11(Contract, "balance",
		coreutil.FieldWithCodecOptional("i1", codec.AgentID),
		OutputCoinBalances{},
	)
	ViewBalanceBaseToken = coreutil.NewViewEP11(Contract, "balanceBaseToken",
		coreutil.FieldWithCodecOptional("i1", codec.AgentID),
		coreutil.FieldWithCodec("o1", codec.BigIntAbs),
	)
	ViewBalanceBaseTokenEVM = coreutil.NewViewEP11(Contract, "balanceBaseTokenEVM",
		coreutil.FieldWithCodecOptional("i1", codec.AgentID),
		coreutil.FieldWithCodec("o1", codec.BigIntAbs),
	)
	ViewBalanceCoin = coreutil.NewViewEP21(Contract, "balanceCoin",
		coreutil.FieldWithCodecOptional("i1", codec.AgentID),
		coreutil.FieldWithCodec("i2", codec.CoinType),
		coreutil.FieldWithCodec("o1", codec.BigIntAbs),
	)
	ViewTreasuryCapID = coreutil.NewViewEP11(Contract, "treasuryCapID",
		coreutil.FieldWithCodec("i1", codec.CoinType),
		coreutil.FieldWithCodec("o1", codec.ObjectID),
	)

	ViewGetAccountNonce = coreutil.NewViewEP11(Contract, "getAccountNonce",
		coreutil.FieldWithCodecOptional("i1", codec.AgentID),
		coreutil.FieldWithCodec("o1", codec.Uint64),
	)
	// TODO: implement pagination
	ViewGetCoinRegistry = coreutil.NewViewEP01(Contract, "getCoinRegistry",
		FieldArrayOf[isc.CoinType]{codec: codec.CoinType},
	)
	ViewObjectBCS = coreutil.NewViewEP11(Contract, "objectBCS",
		coreutil.FieldWithCodec("i1", codec.ObjectID),
		coreutil.FieldWithCodec("o1", codec.Bytes),
	)
	// TODO: implement pagination
	ViewTotalAssets = coreutil.NewViewEP01(Contract, "totalAssets",
		OutputCoinBalances{},
	)
)

// TODO: move to coreutil
// TODO: add pagination
type FieldArrayOf[T any] struct {
	codec codec.Codec[T]
}

func (a FieldArrayOf[T]) Encode(slice []T) dict.Dict {
	if len(slice) > math.MaxUint16 {
		panic("too many values")
	}
	return codec.SliceToArray(a.codec, slice, "o1")
}

func (a FieldArrayOf[T]) Decode(r dict.Dict) ([]T, error) {
	return codec.SliceFromArray(a.codec, r, "o1")
}

type OutputCoinBalances struct{}

func (OutputCoinBalances) Encode(fts isc.CoinBalances) dict.Dict {
	return fts.ToDict()
}

func (OutputCoinBalances) Decode(r dict.Dict) (isc.CoinBalances, error) {
	return isc.CoinBalancesFromDict(r)
}
