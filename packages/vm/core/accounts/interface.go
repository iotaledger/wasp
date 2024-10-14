package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
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
		coreutil.FieldOptional[uint64](),
	)
	FuncTransferAllowanceTo = coreutil.NewEP1(Contract, "transferAllowanceTo",
		coreutil.Field[isc.AgentID](),
	)
	FuncWithdraw    = coreutil.NewEP0(Contract, "withdraw")
	SetCoinMetadata = coreutil.NewEP1(Contract, "setCoinMetadata",
		coreutil.Field[*isc.SuiCoinInfo](),
	)
	DeleteCoinMetadata = coreutil.NewEP1(Contract, "deleteCoinMetadata",
		coreutil.Field[coin.Type](),
	)

	// Views
	// TODO: implement pagination
	ViewAccountObjects = coreutil.NewViewEP11(Contract, "accountObjects",
		coreutil.FieldOptional[isc.AgentID](),
		coreutil.Field[[]sui.ObjectID](),
	)
	// TODO: implement pagination
	ViewAccountObjectsInCollection = coreutil.NewViewEP21(Contract, "accountObjectsInCollection",
		coreutil.FieldOptional[isc.AgentID](),
		coreutil.Field[sui.ObjectID](),
		coreutil.Field[[]sui.ObjectID](),
	)
	// TODO: implement pagination
	ViewBalance = coreutil.NewViewEP11(Contract, "balance",
		coreutil.FieldOptional[isc.AgentID](),
		coreutil.Field[isc.CoinBalances](),
	)
	ViewBalanceBaseToken = coreutil.NewViewEP11(Contract, "balanceBaseToken",
		coreutil.FieldOptional[isc.AgentID](),
		coreutil.Field[coin.Value](),
	)
	ViewBalanceBaseTokenEVM = coreutil.NewViewEP11(Contract, "balanceBaseTokenEVM",
		coreutil.FieldOptional[isc.AgentID](),
		coreutil.Field[*big.Int](),
	)
	ViewBalanceCoin = coreutil.NewViewEP21(Contract, "balanceCoin",
		coreutil.FieldOptional[isc.AgentID](),
		coreutil.Field[coin.Type](),
		coreutil.Field[coin.Value](),
	)

	ViewGetAccountNonce = coreutil.NewViewEP11(Contract, "getAccountNonce",
		coreutil.FieldOptional[isc.AgentID](),
		coreutil.Field[uint64](),
	)
	ViewObjectBCS = coreutil.NewViewEP11(Contract, "objectBCS",
		coreutil.Field[sui.ObjectID](),
		coreutil.Field[[]byte](),
	)
	// TODO: implement pagination
	ViewTotalAssets = coreutil.NewViewEP01(Contract, "totalAssets",
		coreutil.Field[isc.CoinBalances](),
	)
)
