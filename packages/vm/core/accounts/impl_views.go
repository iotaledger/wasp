package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// viewBalance returns the balances of the account belonging to the AgentID
func viewBalance(ctx isc.SandboxView, optionalAgentID *isc.AgentID) isc.CoinBalances {
	ctx.Log().Debugf("accounts.viewBalance")
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getFungibleTokens(accountKey(aid, ctx.ChainID()))
}

// viewBalanceBaseToken returns the base tokens balance of the account belonging to the AgentID
func viewBalanceBaseToken(ctx isc.SandboxView, optionalAgentID *isc.AgentID) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).GetBaseTokensBalanceDiscardExtraDecimals(aid, ctx.ChainID())
}

// viewBalanceBaseTokenEVM returns the base tokens balance of the account belonging to the AgentID (in the EVM format with 18 decimals)
func viewBalanceBaseTokenEVM(ctx isc.SandboxView, optionalAgentID *isc.AgentID) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).GetBaseTokensBalanceFullDecimals(aid, ctx.ChainID())
}

func viewBalanceCoin(ctx isc.SandboxView, optionalAgentID *isc.AgentID, coinID isc.CoinType) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getCoinBalance(
		accountKey(aid, ctx.ChainID()),
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
	return NewStateReaderFromSandbox(ctx).AccountNonce(account, ctx.ChainID())
}

func viewGetCoinRegistry(ctx isc.SandboxView) []isc.CoinType {
	ntMap := NewStateReaderFromSandbox(ctx).coinRecordsMapR()
	ret := make([]isc.CoinType, 0, ntMap.Len())
	ntMap.IterateKeys(func(b []byte) bool {
		ntID := codec.CoinType.MustDecode(b)
		ret = append(ret, ntID)
		return true
	})
	return ret
}

func viewAccountTreasuries(ctx isc.SandboxView, optionalAgentID *isc.AgentID) []isc.CoinType {
	var ret []isc.CoinType
	account := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	NewStateReaderFromSandbox(ctx).accountTreasuryCapsMapR(account).IterateKeys(func(b []byte) bool {
		ret = append(ret, codec.CoinType.MustDecode(b))
		return true
	})
	return ret
}

var errTreasuryNotFound = coreerrors.Register("treasury not found").Create()

func viewTreasuryCapID(ctx isc.SandboxView, coinType isc.CoinType) sui.ObjectID {
	rec := NewStateReaderFromSandbox(ctx).GetTreasuryCap(coinType, ctx.ChainID())
	if rec == nil {
		panic(errTreasuryNotFound)
	}
	return rec.ID
}

// viewAccountObjects returns the ObjectIDs of Objects owned by an account
func viewAccountObjects(ctx isc.SandboxView, optionalAgentID *isc.AgentID) []sui.ObjectID {
	ctx.Log().Debugf("accounts.viewAccountObjects")
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getAccountObjects(aid)
}

func viewAccountObjectsInCollection(ctx isc.SandboxView, optionalAgentID *isc.AgentID, collectionID sui.ObjectID) []sui.ObjectID {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getAccountObjectsInCollection(aid, collectionID)
}

var errObjectNotFound = coreerrors.Register("object not found").Create()

// viewObjectBCS returns the Object data for a given ObjectID
func viewObjectBCS(ctx isc.SandboxView, objectID sui.ObjectID) []byte {
	data := NewStateReaderFromSandbox(ctx).GetObjectBCS(objectID)
	if data == nil {
		panic(errObjectNotFound)
	}
	return data
}
