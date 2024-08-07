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
		coreutil.FieldWithCodecOptional(codec.Uint64),
	)
	FuncTransferAllowanceTo = coreutil.NewEP1(Contract, "transferAllowanceTo",
		coreutil.FieldWithCodec(codec.AgentID),
	)
	FuncWithdraw = coreutil.NewEP0(Contract, "withdraw")

	// Views
	// TODO: implement pagination
	ViewAccountTreasuries = coreutil.NewViewEP11(Contract, "accountTreasuries",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		FieldArrayOf[isc.CoinType]{codec: codec.CoinType},
	)
	// TODO: implement pagination
	ViewAccountObjects = coreutil.NewViewEP11(Contract, "accountObjects",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		FieldArrayOf[sui.ObjectID]{codec: codec.ObjectID},
	)
	// TODO: implement pagination
	ViewAccountObjectsInCollection = coreutil.NewViewEP21(Contract, "accountObjectsInCollection",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.ObjectID),
		FieldArrayOf[sui.ObjectID]{codec: codec.ObjectID},
	)
	// TODO: implement pagination
	ViewBalance = coreutil.NewViewEP11(Contract, "balance",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		OutputCoinBalances{},
	)
	ViewBalanceBaseToken = coreutil.NewViewEP11(Contract, "balanceBaseToken",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.BigIntAbs),
	)
	ViewBalanceBaseTokenEVM = coreutil.NewViewEP11(Contract, "balanceBaseTokenEVM",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.BigIntAbs),
	)
	ViewBalanceCoin = coreutil.NewViewEP21(Contract, "balanceCoin",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.CoinType),
		coreutil.FieldWithCodec(codec.BigIntAbs),
	)
	ViewTreasuryCapID = coreutil.NewViewEP11(Contract, "treasuryCapID",
		coreutil.FieldWithCodec(codec.CoinType),
		coreutil.FieldWithCodec(codec.ObjectID),
	)

	ViewGetAccountNonce = coreutil.NewViewEP11(Contract, "getAccountNonce",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.Uint64),
	)
	// TODO: implement pagination
	ViewGetCoinRegistry = coreutil.NewViewEP01(Contract, "getCoinRegistry",
		FieldArrayOf[isc.CoinType]{codec: codec.CoinType},
	)
	ViewObjectBCS = coreutil.NewViewEP11(Contract, "objectBCS",
		coreutil.FieldWithCodec(codec.ObjectID),
		coreutil.FieldWithCodec(codec.Bytes),
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

func (a FieldArrayOf[T]) Encode(slice []T) []byte {
	if len(slice) > math.MaxUint16 {
		panic("too many values")
	}
	return codec.SliceToArray(a.codec, slice)
}

func (a FieldArrayOf[T]) Decode(r []byte) ([]T, error) {
	return codec.SliceFromArray(a.codec, r)
}

type OutputCoinBalances struct{}

func (OutputCoinBalances) Encode(fts isc.CoinBalances) []byte {
	panic("refactor me: serialization CoinBalances")
	return []byte{}
}

func (OutputCoinBalances) Decode(r []byte) (isc.CoinBalances, error) {
	panic("refactor me: serialization CoinBalances")
	return isc.CoinBalancesFromDict(dict.New())

}
