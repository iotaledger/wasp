package offledger

import (
	"time"

	"github.com/iotaledger/hive.go/timeutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/plugins/chains"
	"golang.org/x/xerrors"
)

const cleanReqsTimeInterval = 10 * time.Minute
const gossipUpToNPeers = 10

type Gossip struct {
	forwardedRequests map[coretypes.RequestID]bool
}

func NewGossip() *Gossip {
	ret := &Gossip{
		forwardedRequests: make(map[coretypes.RequestID]bool),
	}

	// clear forwardedRequests list periodically
	timeutil.NewTicker(func() {
		ret.clearForwardedRequests()
	}, cleanReqsTimeInterval)

	return ret
}

func (g *Gossip) clearForwardedRequests() {
	g.forwardedRequests = make(map[coretypes.RequestID]bool)
}

func gossipMsg(ch chain.Chain, msgType byte, msgData []byte) {
	(*ch.Peers()).SendMsgToRandomPeersSimple(gossipUpToNPeers, msgType, msgData)
}

func (g *Gossip) ProcessOffLedgerRequest(chainID *coretypes.ChainID, req *request.RequestOffLedger) error {
	if g.forwardedRequests[req.ID()] {
		return xerrors.Errorf("received request twice. ID: %s", req.ID())
	}

	ch := chains.AllChains().Get(chainID)
	if ch == nil {
		return xerrors.Errorf("Unknown chain: %s", chainID.Base58())
	}

	g.forwardedRequests[req.ID()] = true
	msgData := NewOffledgerRequestMsg(chainID, req).Bytes()

	committee := ch.Committee()
	// TODO verify that above^ returns nil when no committee
	if committee != nil {
		// this node is a committee node, process the request (add to mempool) and forward it to the other committee nodes
		ch.ReceiveOffLedgerRequests(req)
		(*committee).SendMsgToPeers(chain.MsgOffLedgerRequest, msgData, time.Now().UnixNano())
		return nil
	}
	gossipMsg(ch, chain.MsgOffLedgerRequest, msgData)
	return nil
}
