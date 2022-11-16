package mempool

import (
	"context"

	consGR "github.com/iotaledger/wasp/packages/chain/aaa2/cons/gr"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type Mempool interface {
	consGR.Mempool // TODO should this be unified with the Mempool interface
	// Invoked by the chain, when new alias output is considered as a tip/head
	// of the chain. Mempool can reorganize its state by removing/rejecting
	// or re-adding some requests, depending on how the head has changed.
	// It can mean simple advance of the chain, or a rollback or a reorg.
	// This function is guaranteed to be called in the order, which is
	// considered the chain block order by the ChainMgr.
	TrackNewChainHead(chainHeadAO isc.AliasOutputWithID)
	// Invoked by the chain when a new off-ledger request is received from a node user.
	// Inter-node off-ledger dissemination is NOT performed via this function.
	ReceiveOnLedgerRequest(request isc.OnLedgerRequest)
	// Invoked by the chain when a set of access nodes has changed.
	// These nodes should be used to disseminate the off-ledger requests.
	AccessNodesUpdated(committeePubKeys []*cryptolib.PublicKey, accessNodePubKeys []*cryptolib.PublicKey)
	//
	ReceiveRequests(reqs ...isc.Request) []bool
	RemoveRequests(reqs ...isc.RequestID)
	HasRequest(id isc.RequestID) bool
	GetRequest(id isc.RequestID) isc.Request
	Info() MempoolInfo
}

type MempoolInfo struct {
	TotalPool      int
	InPoolCounter  int
	OutPoolCounter int
}

// Interface the mempool needs form the StateMgr.
type MempoolStateMgr interface {
	// The StateMgr has to find a common ancestor for the prevAO and nextAO, then return
	// the state for Next ao and reject blocks in range (commonAO, prevAO]. The StateMgr
	// can determine relative positions of the corresponding blocks based on their state
	// indexes.
	MempoolStateRequest(ctx context.Context, prevAO, nextAO state.L1Commitment) (vs state.VirtualStateAccess, added, removed []state.Block)
}
