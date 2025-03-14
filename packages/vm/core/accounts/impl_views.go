package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

// viewBalance returns the balances of the account belonging to the AgentID
func viewBalance(ctx isc.SandboxView, optionalAgentID *isc.AgentID) isc.CoinBalances {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	ctx.Log().Debugf("accounts.viewBalance %s", aid)
	return NewStateReaderFromSandbox(ctx).getFungibleTokens(accountKey(aid))
}

// viewBalanceBaseToken returns the base tokens balance of the account belonging to the AgentID
func viewBalanceBaseToken(ctx isc.SandboxView, optionalAgentID *isc.AgentID) coin.Value {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).GetBaseTokensBalanceDiscardExtraDecimals(aid)
}

// viewBalanceBaseTokenEVM returns the base tokens balance of the account belonging to the AgentID (in the EVM format with 18 decimals)
func viewBalanceBaseTokenEVM(ctx isc.SandboxView, optionalAgentID *isc.AgentID) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).GetBaseTokensBalanceFullDecimals(aid)
}

func viewBalanceCoin(ctx isc.SandboxView, optionalAgentID *isc.AgentID, coinID coin.Type) coin.Value {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getCoinBalance(
		accountKey(aid),
		coinID,
	)
}

// viewTotalAssets returns total balances controlled by the chain
func viewTotalAssets(ctx isc.SandboxView) isc.CoinBalances {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return NewStateReaderFromSandbox(ctx).getFungibleTokens(L2TotalsAccount)
}

// nonces are only sent with off-ledger requests
func viewGetAccountNonce(ctx isc.SandboxView, optionalAgentID *isc.AgentID) uint64 {
	account := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).AccountNonce(account)
}

// viewAccountObjects returns the ObjectIDs of Objects owned by an account
func viewAccountObjects(ctx isc.SandboxView, optionalAgentID *isc.AgentID) []iotago.ObjectID {
	ctx.Log().Debugf("accounts.viewAccountObjects")
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getAccountObjects(aid)
}

func viewAccountObjectsInCollection(ctx isc.SandboxView, optionalAgentID *isc.AgentID, collectionID iotago.ObjectID) []iotago.ObjectID {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getAccountObjectsInCollection(aid, collectionID)
}

var errObjectNotFound = coreerrors.Register("object not found").Create()

// viewObjectBCS returns the Object data for a given ObjectID
func viewObjectBCS(ctx isc.SandboxView, objectID iotago.ObjectID) []byte {
	data := NewStateReaderFromSandbox(ctx).GetObjectBCS(objectID)
	if data == nil {
		panic(errObjectNotFound)
	}
	return data
}
