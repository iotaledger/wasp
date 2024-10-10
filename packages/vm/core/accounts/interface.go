package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts)

var (
	// Funcs
	FuncDeposit = coreutil.NewEP0(Contract, "deposit")
	// TODO: adapt to iotago-rebased
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
	FuncWithdraw    = coreutil.NewEP0(Contract, "withdraw")
	SetCoinMetadata = coreutil.NewEP1(Contract, "setCoinMetadata",
		coreutil.FieldWithCodec(codec.NewCodecFromBCS[*isc.SuiCoinInfo]()),
	)
	DeleteCoinMetadata = coreutil.NewEP1(Contract, "deleteCoinMetadata",
		coreutil.FieldWithCodec(codec.CoinType),
	)

	// Views
	// TODO: implement pagination
	ViewAccountObjects = coreutil.NewViewEP11(Contract, "accountObjects",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldArrayWithCodec(codec.ObjectID),
	)
	// TODO: implement pagination
	ViewAccountObjectsInCollection = coreutil.NewViewEP21(Contract, "accountObjectsInCollection",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.ObjectID),
		coreutil.FieldArrayWithCodec(codec.ObjectID),
	)
	// TODO: implement pagination
	ViewBalance = coreutil.NewViewEP11(Contract, "balance",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.CoinBalances),
	)
	ViewBalanceBaseToken = coreutil.NewViewEP11(Contract, "balanceBaseToken",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.CoinValue),
	)
	ViewBalanceBaseTokenEVM = coreutil.NewViewEP11(Contract, "balanceBaseTokenEVM",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.BigIntAbs),
	)
	ViewBalanceCoin = coreutil.NewViewEP21(Contract, "balanceCoin",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.CoinType),
		coreutil.FieldWithCodec(codec.CoinValue),
	)

	ViewGetAccountNonce = coreutil.NewViewEP11(Contract, "getAccountNonce",
		coreutil.FieldWithCodecOptional(codec.AgentID),
		coreutil.FieldWithCodec(codec.Uint64),
	)
	ViewObjectBCS = coreutil.NewViewEP11(Contract, "objectBCS",
		coreutil.FieldWithCodec(codec.ObjectID),
		coreutil.FieldWithCodec(codec.Bytes),
	)
	// TODO: implement pagination
	ViewTotalAssets = coreutil.NewViewEP01(Contract, "totalAssets",
		coreutil.FieldWithCodec(codec.CoinBalances),
	)
)
