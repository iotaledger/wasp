package accounts

// Package accounts implements a core contract that manages on-chain accounts and their assets.
// It provides functionality for tracking, transferring, and managing various types of assets
// (tokens, native tokens, NFTs) that belong to on-chain accounts identified by AgentIDs.
//
// The accounts contract maintains balances and handles transfers between accounts,
// serves as the entry point for depositing assets to the chain, and provides
// views for querying account information and balances.
import (
	"math/big"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/parameters"
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
	FuncTransferAllowanceTo = coreutil.NewEP1(Contract, "transferAllowanceTo",
		coreutil.Field[isc.AgentID]("agentID"),
	)
	FuncWithdraw    = coreutil.NewEP0(Contract, "withdraw")
	SetCoinMetadata = coreutil.NewEP1(Contract, "setCoinMetadata",
		coreutil.Field[*parameters.IotaCoinInfo]("coinInfo"),
	)
	DeleteCoinMetadata = coreutil.NewEP1(Contract, "deleteCoinMetadata",
		coreutil.Field[coin.Type]("coinType"),
	)

	// Views
	// TODO: implement pagination
	ViewAccountObjects = coreutil.NewViewEP11(Contract, "accountObjects",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[[]isc.IotaObject]("accountObjects"),
	)
	// TODO: implement pagination
	ViewBalance = coreutil.NewViewEP11(Contract, "balance",
		coreutil.FieldOptional[isc.AgentID]("agentID"),
		coreutil.Field[*isc.Assets]("coinBalances"),
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
	// TODO: implement pagination
	ViewTotalAssets = coreutil.NewViewEP01(Contract, "totalAssets",
		coreutil.Field[*isc.Assets]("coinBalances"),
	)
)
