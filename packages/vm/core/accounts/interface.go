package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts)

var (
	// Funcs
	FuncDeposit = coreutil.NewEP0(Contract, "deposit")
	// TODO: adapt to iotago-rebased
	//   FuncCoinCreate
	//   FuncCoinModifySupply
	//   FuncCoinDestroy
	//   FuncMintObject
	FuncTransferAccountToChain = coreutil.NewEP1(Contract, "transferAccountToChain",
		coreutil.FieldOptional[uint64]("optionalGasReserve"),
	)
	FuncTransferAllowanceTo = coreutil.NewEP1(Contract, "transferAllowanceTo",
		coreutil.Field[isc.AgentID]("agentID"),
	)
	FuncWithdraw    = coreutil.NewEP0(Contract, "withdraw")
	SetCoinMetadata = coreutil.NewEP1(Contract, "setCoinMetadata",
		coreutil.Field[*isc.IotaCoinInfo]("coinInfo"),
	)
	DeleteCoinMetadata = coreutil.NewEP1(Contract, "deleteCoinMetadata",
		coreutil.Field[coin.Type]("coinType"),
	)

	// Views
	// TODO: implement pagination
	ViewAccountObjects = coreutil.NewViewEP11(Contract, "accountObjects",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[[]iotago.ObjectID]("accountObjects"),
	)
	// TODO: implement pagination
	ViewAccountObjectsInCollection = coreutil.NewViewEP21(Contract, "accountObjectsInCollection",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[iotago.ObjectID]("collectionID"),
		coreutil.Field[[]iotago.ObjectID]("accountObjects"),
	)
	// TODO: implement pagination
	ViewBalance = coreutil.NewViewEP11(Contract, "balance",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[isc.CoinBalances]("coinBalances"),
	)
	ViewBalanceBaseToken = coreutil.NewViewEP11(Contract, "balanceBaseToken",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[coin.Value]("baseTokenBalance"),
	)
	ViewBalanceBaseTokenEVM = coreutil.NewViewEP11(Contract, "balanceBaseTokenEVM",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[*big.Int]("evmBaseTokenBalance"),
	)
	ViewBalanceCoin = coreutil.NewViewEP21(Contract, "balanceCoin",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[coin.Type]("coinType"),
		coreutil.Field[coin.Value]("coinBalance"),
	)

	ViewGetAccountNonce = coreutil.NewViewEP11(Contract, "getAccountNonce",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[uint64]("nonce"),
	)
	ViewObjectBCS = coreutil.NewViewEP11(Contract, "objectBCS",
		coreutil.Field[iotago.ObjectID]("objectID"),
		coreutil.Field[[]byte]("bcsEncodedBytes"),
	)
	// TODO: implement pagination
	ViewTotalAssets = coreutil.NewViewEP01(Contract, "totalAssets",
		coreutil.Field[isc.CoinBalances]("coinBalances"),
	)
)
