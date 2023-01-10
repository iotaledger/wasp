// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accessMgr

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chains/accessMgr/amDist"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

type AccessMgr interface {
	TrustedNodes(trusted []*cryptolib.PublicKey)
	ChainAccessNodes(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey)
	ChainDismissed(chainID isc.ChainID)
}

type accessMgrImpl struct {
	dist                    gpa.AckHandler
	dismissPeerBuf          []*cryptolib.PublicKey
	reqTrustedNodesPipe     pipe.Pipe
	reqChainAccessNodesPipe pipe.Pipe
	reqChainDismissedPipe   pipe.Pipe
	netRecvPipe             pipe.Pipe
	netPeeringID            peering.PeeringID
	netPeerPubs             map[gpa.NodeID]*cryptolib.PublicKey
	net                     peering.NetworkProvider
	log                     *logger.Logger
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

var _ AccessMgr = &accessMgrImpl{}

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
	log *logger.Logger,
) AccessMgr {
	// there is only one AccessMgr per Wasp node, so the identifier is a constant.
	netPeeringID := peering.HashPeeringIDFromBytes([]byte("AccessManager")) // AccessManager
	ami := &accessMgrImpl{
		dismissPeerBuf:          []*cryptolib.PublicKey{},
		reqTrustedNodesPipe:     pipe.NewDefaultInfinitePipe(),
		reqChainAccessNodesPipe: pipe.NewDefaultInfinitePipe(),
		reqChainDismissedPipe:   pipe.NewDefaultInfinitePipe(),
		netRecvPipe:             pipe.NewDefaultInfinitePipe(),
		netPeeringID:            netPeeringID,
		netPeerPubs:             map[gpa.NodeID]*cryptolib.PublicKey{},
		net:                     net,
		log:                     log,
	}
	me := ami.pubKeyAsNodeID(nodeIdentity.GetPublicKey())
	ami.dist = gpa.NewAckHandler(me, gpa.NewOwnHandler(
		me,
		amDist.NewAccessMgr(ami.pubKeyAsNodeID, serversUpdatedCB, ami.dismissPeerCB, log).AsGPA(),
	), resendPeriod)

	netRecvPipeInCh := ami.netRecvPipe.In()
	netAttachID := net.Attach(&netPeeringID, peering.PeerMessageReceiverAccessMgr, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeAccessMgr {
			ami.log.Warnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		netRecvPipeInCh <- recv
	})
	go ami.run(ctx, netAttachID)

	return ami
}

// Implements the AccessMgr interface.
func (ami *accessMgrImpl) TrustedNodes(trusted []*cryptolib.PublicKey) {
	ami.reqTrustedNodesPipe.In() <- &reqTrustedNodes{trusted: trusted}
}

// Implements the AccessMgr interface.
func (ami *accessMgrImpl) ChainAccessNodes(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey) {
	ami.reqChainAccessNodesPipe.In() <- &reqChainAccessNodes{chainID: chainID, accessNodes: accessNodes}
}

// Implements the AccessMgr interface.
func (ami *accessMgrImpl) ChainDismissed(chainID isc.ChainID) {
	ami.reqChainDismissedPipe.In() <- &reqChainDismissed{chainID: chainID}
}

// A callback for amDist.
func (ami *accessMgrImpl) dismissPeerCB(peerPubKey *cryptolib.PublicKey) {
	// Dismiss them after the messages are sent.
	ami.dismissPeerBuf = append(ami.dismissPeerBuf, peerPubKey)
}

func (ami *accessMgrImpl) run(ctx context.Context, netAttachID interface{}) {
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
			ami.handleReqTrustedNodes(recv.(*reqTrustedNodes))
		case recv, ok := <-reqChainAccessNodesPipeOutCh:
			if !ok {
				reqChainAccessNodesPipeOutCh = nil
				continue
			}
			ami.handleReqChainAccessNodes(recv.(*reqChainAccessNodes))
		case recv, ok := <-reqChainDismissedPipeOutCh:
			if !ok {
				reqChainDismissedPipeOutCh = nil
				continue
			}
			ami.handleReqChainDismissed(recv.(*reqChainDismissed))
		case recv, ok := <-netRecvPipeOutCh:
			if !ok {
				netRecvPipeOutCh = nil
				continue
			}
			ami.handleNetMessage(recv.(*peering.PeerMessageIn))
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
			ami.net.Detach(netAttachID)
			return
		}
	}
}

func (ami *accessMgrImpl) handleReqTrustedNodes(recv *reqTrustedNodes) {
	ami.sendMessages(ami.dist.Input(amDist.NewInputTrustedNodes(recv.trusted)))
}

func (ami *accessMgrImpl) handleReqChainAccessNodes(recv *reqChainAccessNodes) {
	ami.sendMessages(ami.dist.Input(amDist.NewInputAccessNodes(recv.chainID, recv.accessNodes)))
}

func (ami *accessMgrImpl) handleReqChainDismissed(recv *reqChainDismissed) {
	ami.sendMessages(ami.dist.Input(amDist.NewInputChainDisabled(recv.chainID)))
}

func (ami *accessMgrImpl) handleDistDebugTick() {
	ami.log.Debugf(
		"AccessMgr, dist=%v",
		ami.dist.StatusString(),
	)
}

func (ami *accessMgrImpl) handleDistTimeTick(timestamp time.Time) {
	ami.sendMessages(ami.dist.Input(ami.dist.MakeTickInput(timestamp)))
}

func (ami *accessMgrImpl) handleNetMessage(recv *peering.PeerMessageIn) {
	msg, err := ami.dist.UnmarshalMessage(recv.MsgData)
	if err != nil {
		ami.log.Warnf("cannot parse message: %v", err)
		return
	}
	msg.SetSender(ami.pubKeyAsNodeID(recv.SenderPubKey))
	outMsgs := ami.dist.Message(msg) // Output is handled via callbacks in this case.
	ami.sendMessages(outMsgs)
}

func (ami *accessMgrImpl) sendMessages(outMsgs gpa.OutMessages) {
	if len(ami.dismissPeerBuf) != 0 {
		for _, dismissPeerPub := range ami.dismissPeerBuf {
			ami.dist.DismissPeer(ami.pubKeyAsNodeID(dismissPeerPub))
		}
		ami.dismissPeerBuf = []*cryptolib.PublicKey{}
	}
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(m gpa.Message) {
		msgData, err := m.MarshalBinary()
		if err != nil {
			ami.log.Warnf("Failed to send a message: %v", err)
			return
		}
		pm := &peering.PeerMessageData{
			PeeringID:   ami.netPeeringID,
			MsgReceiver: peering.PeerMessageReceiverAccessMgr,
			MsgType:     msgTypeAccessMgr,
			MsgData:     msgData,
		}
		ami.net.SendMsgByPubKey(ami.netPeerPubs[m.Recipient()], pm)
	})
}

func (ami *accessMgrImpl) pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	nodeID := gpa.NodeIDFromPublicKey(pubKey)
	if _, ok := ami.netPeerPubs[nodeID]; !ok {
		ami.netPeerPubs[nodeID] = pubKey
	}
	return nodeID
}
