package dkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/peering"
	"go.dedis.ch/kyber/v3"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
)

//
// Stands for a DKG procedure instance on a particular node.
//
type proc struct {
	dkgID                 string            // DKG procedure ID we are participating in.
	chainID               coretypes.ChainID // Same as dkgID, just parsed.
	step                  string            // The current step.
	node                  *node             // DKG node we are running in.
	nodeIndex             int               // Index of this node.
	peerLocs              []string
	peerPubs              []kyber.Point
	coordPub              kyber.Point
	stepTimeout           time.Duration
	netGroup              peering.GroupProvider
	dkgImpl               *rabin_dkg.DistKeyGenerator
	cancelCh              chan bool        // To stop a timer on completion.
	coordStepCh           chan coordStepCh // To process all the step requests in a single thread.
	peerMsgCh             chan peerMsgCh   //
	done                  bool             // Indicates, if the process is completed.
	recvDealMsgs          map[int]*rabinDealMsg
	recvResponseMsgs      map[int]*rabinResponseMsg
	recvJustificationMsgs map[int]*rabinJustificationMsg
}
type coordStepCh struct { // Only for communicating with the main thread.
	msg  *StepReq
	resp chan error
}
type peerMsgCh struct { // Only for communicating with the main thread.
	from    peering.PeerSender
	fromIdx int
	msg     *peering.PeerMessage
}

func onCoordInit(dkgID string, msg *InitReq, node *node) (*proc, error) {
	var err error
	var chainID coretypes.ChainID
	if chainID, err = coretypes.NewChainIDFromBase58(dkgID); err != nil {
		return nil, err
	}
	var peerPubs []kyber.Point
	if peerPubs, err = pubsFromBytes(msg.PeerPubs, node.suite); err != nil {
		return nil, err
	}
	var coordPub kyber.Point
	if coordPub, err = pubFromBytes(msg.CoordPub, node.suite); err != nil {
		return nil, err
	}
	var netGroup peering.GroupProvider
	if netGroup, err = node.netProvider.Group(msg.PeerLocs); err != nil {
		return nil, err
	}
	var dkgImpl *rabin_dkg.DistKeyGenerator
	if dkgImpl, err = rabin_dkg.NewDistKeyGenerator(node.suite, node.secKey, peerPubs, len(peerPubs)); err != nil {
		return nil, err
	}
	var nodeIndex int
	if nodeIndex, err = netGroup.PeerIndex(node.netProvider.Self()); err != nil {
		return nil, err
	}
	timeout := time.Millisecond * time.Duration(msg.TimeoutMS)
	p := proc{
		dkgID:       dkgID,
		chainID:     chainID,
		node:        node,
		nodeIndex:   nodeIndex,
		peerLocs:    msg.PeerLocs,
		peerPubs:    peerPubs,
		coordPub:    coordPub,
		stepTimeout: timeout,
		netGroup:    netGroup,
		dkgImpl:     dkgImpl,
		cancelCh:    make(chan bool),
		coordStepCh: make(chan coordStepCh),
		peerMsgCh:   make(chan peerMsgCh),
		done:        false,
	}
	go p.processLoop(timeout)
	node.netProvider.Attach(chainID, p.onPeerMessage)
	return &p, nil
}

// Handles a `step` call from the coordinator and pass it to the main thread.
func (p *proc) onCoordStep(msg *StepReq) error {
	resp := make(chan error)
	p.coordStepCh <- coordStepCh{
		msg:  msg,
		resp: resp,
	}
	return <-resp
}

// Handles a `pubKey` call from the coordinator.
func (p *proc) onCoordPubKey() (*PubKeyResp, error) {
	return &PubKeyResp{}, nil // TODO
}

// Handles a message from a peer and pass it to the main thread.
func (p *proc) onPeerMessage(from peering.PeerSender, msg *peering.PeerMessage) {
	var err error
	var fromIdx int
	if fromIdx, err = p.netGroup.PeerIndex(from); err != nil {
		fmt.Printf("WARNING: Dropping message from unexpected peer %v: %v\n", from.Location(), msg)
		return
	}
	p.peerMsgCh <- peerMsgCh{
		from:    from,
		fromIdx: fromIdx,
		msg:     msg,
	}
}

// That's the main thread executing all the procedure steps.
// We use a single process to make all the actions sequential.
func (p *proc) processLoop(timeout time.Duration) {
	timeoutCh := time.After(timeout)
	done := false
	acceptPeerMsgType := byte(0)
	acceptPeerMsgCh := make(chan peerMsgCh)
	for !done {
		select {
		case coordStepCall := <-p.coordStepCh:
			switch coordStepCall.msg.Step {
			case "1-R2.1-SendDeals":
				acceptPeerMsgType = rabinDealMsgType
				go p.doCoordStepSendDeals(acceptPeerMsgCh, coordStepCall.resp)
			case "2-R2.2-SendResponses":
				acceptPeerMsgType = rabinResponseMsgType
				go p.doCoordStepSendResponses(acceptPeerMsgCh, coordStepCall.resp)
			case "3-R2.3-SendJustifications":
				acceptPeerMsgType = rabinJustificationMsgType
				go p.doCoordStepSendJustifications(acceptPeerMsgCh, coordStepCall.resp)
			case "4-R4-SendSecretCommits":
				acceptPeerMsgType = rabinSecretCommitsMsgType
				go p.doCoordStepSendSecretCommits(acceptPeerMsgCh, coordStepCall.resp)
			default:
				coordStepCall.resp <- errors.New("unknown_step")
			}
		case cast := <-p.peerMsgCh:
			if cast.msg.MsgType != acceptPeerMsgType {
				fmt.Printf("Dropping unexpected peer message.\n")
				continue
			}
			acceptPeerMsgCh <- cast
		case <-p.cancelCh:
			p.node.dropProcess(p)
			done = true
		case <-timeoutCh:
			if p.node.dropProcess(p) {
				fmt.Printf("Deleting a DkgProc on timeout.\n")
			}
			done = true
		}
	}
	p.done = true
}

func (p *proc) doCoordStepSendDeals(peerMsgCh chan peerMsgCh, errorCh chan error) {
	var err error
	var deals map[int]*rabin_dkg.Deal
	if deals, err = p.dkgImpl.Deals(); err != nil {
		fmt.Printf("[%v] ERROR: Deals -> %+v\n", p.nodeIndex, err)
		errorCh <- err
		return
	}
	//
	// Send own deals.
	for d := range deals {
		p.castPeerByIndex(d, &rabinDealMsg{
			deal: deals[d],
		})
	}
	//
	// Receive other's deals.
	var receivedDealMsgs map[int]*peering.PeerMessage
	if receivedDealMsgs, err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		errorCh <- err
		return
	}
	//
	// Decode the received deals.
	receivedDeals := make(map[int]*rabinDealMsg, len(receivedDealMsgs))
	for i := range receivedDealMsgs {
		peerDealMsg := rabinDealMsg{}
		if err = peerDealMsg.fromBytes(receivedDealMsgs[i].MsgData, p.node.suite); err != nil {
			errorCh <- err
			return
		}
		receivedDeals[i] = &peerDealMsg
	}
	p.recvDealMsgs = receivedDeals // Stored for the next step.
	fmt.Printf("[%v] All deals received.\n", p.nodeIndex)
	errorCh <- nil
}

func (p *proc) doCoordStepSendResponses(peerMsgCh chan peerMsgCh, errorCh chan error) {
	var err error
	//
	// Process the received deals and produce responses.
	ourResponses := []*rabin_dkg.Response{}
	for i := range p.recvDealMsgs {
		var r *rabin_dkg.Response
		if r, err = p.dkgImpl.ProcessDeal(p.recvDealMsgs[i].deal); err != nil {
			fmt.Printf("[%v] ERROR: ProcessDeal(%v) -> %+v\n", p.nodeIndex, i, err)
			errorCh <- err
			return
		}
		ourResponses = append(ourResponses, r)
	}
	//
	// Send our responses.
	for i := range p.recvDealMsgs { // To all other peers.
		p.castPeerByIndex(i, &rabinResponseMsg{
			responses: ourResponses,
		})
	}
	//
	// Receive other's responses.
	var receivedRespMsgs map[int]*peering.PeerMessage
	if receivedRespMsgs, err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		errorCh <- err
		return
	}
	//
	// Decode the received responces.
	receivedResponses := make(map[int]*rabinResponseMsg, len(receivedRespMsgs))
	for i := range receivedRespMsgs {
		peerResponseMsg := rabinResponseMsg{}
		if err = peerResponseMsg.fromBytes(receivedRespMsgs[i].MsgData); err != nil {
			errorCh <- err
			return
		}
		receivedResponses[i] = &peerResponseMsg
	}
	p.recvResponseMsgs = receivedResponses // Stored for the next step.
	fmt.Printf("[%v] All responses received.\n", p.nodeIndex)
	errorCh <- nil
}

func (p *proc) doCoordStepSendJustifications(peerMsgCh chan peerMsgCh, errorCh chan error) {
	var err error
	//
	// Process the received responses and produce justifications.
	ourJustifications := []*rabin_dkg.Justification{}
	for i := range p.recvResponseMsgs {
		for _, r := range p.recvResponseMsgs[i].responses {
			var j *rabin_dkg.Justification
			if j, err = p.dkgImpl.ProcessResponse(r); err != nil {
				fmt.Printf("[%v] ERROR: ProcessResponse(%v) -> %+v\n", p.nodeIndex, i, err)
				errorCh <- err
				return
			}
			if j != nil {
				ourJustifications = append(ourJustifications, j)
			}
		}
	}
	//
	// Send our justifications.
	for i := range p.recvDealMsgs { // To all other peers.
		p.castPeerByIndex(i, &rabinJustificationMsg{
			justifications: ourJustifications,
		})
	}
	//
	// Receive other's justifications.
	var receivedMsgs map[int]*peering.PeerMessage
	if receivedMsgs, err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		errorCh <- err
		return
	}
	//
	// Decode the received justifications.
	receivedJustifications := make(map[int]*rabinJustificationMsg, len(receivedMsgs))
	for i := range receivedMsgs {
		peerJustificationMsg := rabinJustificationMsg{}
		if err = peerJustificationMsg.fromBytes(receivedMsgs[i].MsgData, p.node.suite); err != nil {
			errorCh <- err
			return
		}
		receivedJustifications[i] = &peerJustificationMsg
	}
	p.recvJustificationMsgs = receivedJustifications // Stored for the next step.
	fmt.Printf("[%v] All justifications received.\n", p.nodeIndex)
	errorCh <- nil
}

func (p *proc) doCoordStepSendSecretCommits(peerMsgCh chan peerMsgCh, errorCh chan error) {
	var err error
	//
	// Process the received justifications.
	for i := range p.recvJustificationMsgs {
		for _, j := range p.recvJustificationMsgs[i].justifications {
			if err = p.dkgImpl.ProcessJustification(j); err != nil {
				errorCh <- err
				return
			}
		}
	}
	fmt.Printf("[%v] All justifications processed.\n", p.nodeIndex)

	// TODO: ...
	p.dkgImpl.SetTimeout()
	qual := p.dkgImpl.QUAL()
	certified := p.dkgImpl.Certified()
	fmt.Printf("[%v] Certified=%v, qual=%v\n", p.nodeIndex, certified, qual)

	// TODO: ...
	errorCh <- nil
}

func (p *proc) castPeerByIndex(peerIdx int, msg msgByteCoder) {
	p.netGroup.SendMsgByIndex(peerIdx, &peering.PeerMessage{
		ChainID:     p.chainID,
		SenderIndex: 0, // TODO: Should be resolved on the receiving side.
		Timestamp:   0, // TODO: What to do with this?
		MsgType:     msg.MsgType(),
		MsgData:     util.MustBytes(msg),
	})
}

func (p *proc) waitForOthers(peerMsgCh chan peerMsgCh, timeout time.Duration) (map[int]*peering.PeerMessage, error) {
	flags := make([]bool, len(p.peerPubs))
	flags[p.nodeIndex] = true // Assume our message is already present.
	msgs := make(map[int]*peering.PeerMessage)
	timeoutCh := time.After(timeout)
	for !haveAll(flags) {
		select {
		case peerMsg := <-peerMsgCh:
			msgs[peerMsg.fromIdx] = peerMsg.msg
			flags[peerMsg.fromIdx] = true
		case <-timeoutCh:
			return nil, errors.New("step_timeout")
		}
	}
	return msgs, nil
}
