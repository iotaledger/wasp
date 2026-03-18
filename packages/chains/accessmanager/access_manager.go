// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package accessmanager implements chain access control and permission management.
package accessmanager

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/chains/accessmanager/dist"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/util/pipe"
)

type AccessMgr struct {
	dist                    gpa.AckHandler
	dismissPeerBuf          []*cryptolib.PublicKey
	reqTrustedNodesPipe     pipe.Pipe[*reqTrustedNodes]
	reqChainAccessNodesPipe pipe.Pipe[*reqChainAccessNodes]
	reqChainDismissedPipe   pipe.Pipe[*reqChainDismissed]
	netRecvPipe             pipe.Pipe[*peering.PeerMessageIn]
	netPeeringID            peering.PeeringID
	netPeerPubs             map[gpa.NodeID]*cryptolib.PublicKey
	net                     peering.NetworkProvider
	log                     log.Logger
}

type reqTrustedNodes struct {
	trusted []*cryptolib.PublicKey
}

type reqChainAccessNodes struct {
	chainID     isc.ChainID
	accessNodes []*cryptolib.PublicKey
}

type reqChainDismissed struct {
	chainID isc.ChainID
}

const (
	msgTypeAccessMgr byte = iota
)

const (
	resendPeriod  = 3 * time.Second
	distDebugTick = 10 * time.Second
	distTimeTick  = 1 * time.Second
)

func New(
	ctx context.Context,
	serversUpdatedCB func(chainID isc.ChainID, servers []*cryptolib.PublicKey),
	nodeIdentity *cryptolib.KeyPair,
	net peering.NetworkProvider,
	log log.Logger,
) *AccessMgr {
	// there is only one AccessMgr per Wasp node, so the identifier is a constant.
	netPeeringID := peering.HashPeeringIDFromBytes([]byte("AccessManager")) // AccessManager
	ami := &AccessMgr{
		dismissPeerBuf:          []*cryptolib.PublicKey{},
		reqTrustedNodesPipe:     pipe.NewInfinitePipe[*reqTrustedNodes](),
		reqChainAccessNodesPipe: pipe.NewInfinitePipe[*reqChainAccessNodes](),
		reqChainDismissedPipe:   pipe.NewInfinitePipe[*reqChainDismissed](),
		netRecvPipe:             pipe.NewInfinitePipe[*peering.PeerMessageIn](),
		netPeeringID:            netPeeringID,
		netPeerPubs:             map[gpa.NodeID]*cryptolib.PublicKey{},
		net:                     net,
		log:                     log,
	}
	me := ami.pubKeyAsNodeID(nodeIdentity.GetPublicKey())
	ami.dist = gpa.NewAckHandler(me, gpa.NewOwnHandler(
		me,
		dist.NewAccessMgr(ami.pubKeyAsNodeID, serversUpdatedCB, ami.dismissPeerCB, log).AsGPA(),
	), resendPeriod)

	netRecvPipeInCh := ami.netRecvPipe.In()
	unhook := net.Attach(&netPeeringID, peering.ReceiverAccessMgr, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeAccessMgr {
			ami.log.LogWarnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		netRecvPipeInCh <- recv
	})
	go ami.run(ctx, unhook)

	return ami
}

// TrustedNodes implements the AccessMgr interface.
func (ami *AccessMgr) TrustedNodes(trusted []*cryptolib.PublicKey) {
	ami.reqTrustedNodesPipe.In() <- &reqTrustedNodes{trusted: trusted}
}

// ChainAccessNodes implements the AccessMgr interface.
func (ami *AccessMgr) ChainAccessNodes(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey) {
	ami.reqChainAccessNodesPipe.In() <- &reqChainAccessNodes{chainID: chainID, accessNodes: accessNodes}
}

// ChainDismissed implements the AccessMgr interface.
func (ami *AccessMgr) ChainDismissed(chainID isc.ChainID) {
	ami.reqChainDismissedPipe.In() <- &reqChainDismissed{chainID: chainID}
}

// A callback for amDist.
func (ami *AccessMgr) dismissPeerCB(peerPubKey *cryptolib.PublicKey) {
	// Dismiss them after the messages are sent.
	ami.dismissPeerBuf = append(ami.dismissPeerBuf, peerPubKey)
}

func (ami *AccessMgr) run(ctx context.Context, cleanupFunc context.CancelFunc) {
	reqTrustedNodesOutCh := ami.reqTrustedNodesPipe.Out()
	reqChainAccessNodesPipeOutCh := ami.reqChainAccessNodesPipe.Out()
	reqChainDismissedPipeOutCh := ami.reqChainDismissedPipe.Out()
	netRecvPipeOutCh := ami.netRecvPipe.Out()
	debugTicker := time.NewTicker(distDebugTick)
	timeTicker := time.NewTicker(distTimeTick)
	for {
		select {
		case recv, ok := <-reqTrustedNodesOutCh:
			if !ok {
				reqTrustedNodesOutCh = nil
				continue
			}
			ami.handleReqTrustedNodes(recv)
		case recv, ok := <-reqChainAccessNodesPipeOutCh:
			if !ok {
				reqChainAccessNodesPipeOutCh = nil
				continue
			}
			ami.handleReqChainAccessNodes(recv)
		case recv, ok := <-reqChainDismissedPipeOutCh:
			if !ok {
				reqChainDismissedPipeOutCh = nil
				continue
			}
			ami.handleReqChainDismissed(recv)
		case recv, ok := <-netRecvPipeOutCh:
			if !ok {
				netRecvPipeOutCh = nil
				continue
			}
			ami.handleNetMessage(recv)
		case <-debugTicker.C:
			ami.handleDistDebugTick()
		case timestamp := <-timeTicker.C:
			ami.handleDistTimeTick(timestamp)
		case <-ctx.Done():
			// close(reqTrustedNodesOutCh) // TODO: Causes panic: send on closed channel
			// close(reqChainAccessNodesOutCh)
			// close(reqChainDismissedOutCh)
			debugTicker.Stop()
			timeTicker.Stop()
			util.ExecuteIfNotNil(cleanupFunc)
			return
		}
	}
}

func (ami *AccessMgr) handleReqTrustedNodes(recv *reqTrustedNodes) {
	ami.log.LogDebugf("handleReqTrustedNodes: trusted=%v", recv.trusted)
	ami.sendMessages(ami.dist.Input(dist.NewInputTrustedNodes(recv.trusted)))
}

func (ami *AccessMgr) handleReqChainAccessNodes(recv *reqChainAccessNodes) {
	ami.log.LogDebugf("handleReqChainAccessNodes: chainID=%v, access=%v", recv.chainID, recv.accessNodes)
	ami.sendMessages(ami.dist.Input(dist.NewInputAccessNodes(recv.chainID, recv.accessNodes)))
}

func (ami *AccessMgr) handleReqChainDismissed(recv *reqChainDismissed) {
	ami.log.LogDebugf("handleReqChainDismissed: chainID=%v", recv.chainID)
	ami.sendMessages(ami.dist.Input(dist.NewInputChainDisabled(recv.chainID)))
}

func (ami *AccessMgr) handleDistDebugTick() {
	ami.log.LogDebugf(
		"AccessMgr, dist=%v",
		ami.dist.StatusString(),
	)
}

func (ami *AccessMgr) handleDistTimeTick(timestamp time.Time) {
	ami.sendMessages(ami.dist.Input(ami.dist.MakeTickInput(timestamp)))
}

func (ami *AccessMgr) handleNetMessage(recv *peering.PeerMessageIn) {
	msg, err := ami.dist.UnmarshalPayload(recv.MsgData)
	if err != nil {
		ami.log.LogWarnf("cannot parse message: %v", err)
		return
	}
	// Output is handled via callbacks in this case.
	outMsgs := ami.dist.Message(gpa.NewMessageIn(ami.pubKeyAsNodeID(recv.SenderPubKey), msg))
	ami.sendMessages(outMsgs)
}

func (ami *AccessMgr) sendMessages(outMsgs []gpa.MessageOut) {
	if len(ami.dismissPeerBuf) != 0 {
		for _, dismissPeerPub := range ami.dismissPeerBuf {
			ami.dist.DismissPeer(ami.pubKeyAsNodeID(dismissPeerPub))
		}
		ami.dismissPeerBuf = []*cryptolib.PublicKey{}
	}
	for _, msg := range outMsgs {
		msgBytes := lo.Must(gpa.MarshalPayload(msg.Payload))
		pm := peering.NewPeerMessageData(ami.netPeeringID, peering.ReceiverAccessMgr, msgTypeAccessMgr, msgBytes)
		ami.net.SendMsgByPubKey(ami.netPeerPubs[msg.Recipient], pm)
	}
}

func (ami *AccessMgr) pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	nodeID := gpa.NodeIDFromPublicKey(pubKey)
	if _, ok := ami.netPeerPubs[nodeID]; !ok {
		ami.netPeerPubs[nodeID] = pubKey
	}
	return nodeID
}
