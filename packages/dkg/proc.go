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
	dkgID                    string            // DKG procedure ID we are participating in.
	chainID                  coretypes.ChainID // Same as dkgID, just parsed.
	step                     string            // The current step.
	node                     *node             // DKG node we are running in.
	nodeIndex                int               // Index of this node.
	nodeLoc                  string
	peerLocs                 []string
	peerPubs                 []kyber.Point
	coordPub                 kyber.Point
	stepTimeout              time.Duration
	netGroup                 peering.GroupProvider
	dkgImpl                  *rabin_dkg.DistKeyGenerator
	cancelCh                 chan bool        // To stop a timer on completion.
	coordStepCh              chan coordStepCh // To process all the step requests in a single thread.
	peerMsgCh                chan peerMsgCh   //
	done                     bool             // Indicates, if the process is completed.
	recvDealMsgs             map[int]*rabinDealMsg
	recvResponseMsgs         map[int]*rabinResponseMsg
	recvJustificationMsgs    map[int]*rabinJustificationMsg
	recvSecretCommitsMsgs    map[int]*rabinSecretCommitsMsg
	recvComplaintCommitsMsgs map[int]*rabinComplaintCommitsMsg
	distKeyShare             *rabin_dkg.DistKeyShare
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
		nodeLoc:     node.netProvider.Self().Location(),
		peerLocs:    msg.PeerLocs,
		peerPubs:    peerPubs,
		coordPub:    coordPub,
		stepTimeout: timeout,
		netGroup:    netGroup,
		dkgImpl:     dkgImpl,
		cancelCh:    make(chan bool),
		coordStepCh: make(chan coordStepCh, 10),
		peerMsgCh:   make(chan peerMsgCh, len(peerPubs)),
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
	var err error
	if p.distKeyShare == nil {
		if err = p.onCoordStep(&StepReq{Step: "6-R6-SendReconstructCommits"}); err != nil {
			return nil, err
		}
	}
	var pubBytes []byte
	if pubBytes, err = p.distKeyShare.Public().MarshalBinary(); err != nil {
		return nil, err
	}
	return &PubKeyResp{PubKey: pubBytes}, nil
}

// Handles a message from a peer and pass it to the main thread.
func (p *proc) onPeerMessage(from peering.PeerSender, msg *peering.PeerMessage) {
	var err error
	var fromIdx int
	if fromIdx, err = p.netGroup.PeerIndex(from); err != nil {
		fmt.Printf(
			"[%v] WARNING: Dropping message from unexpected peer %v: %v\n",
			p.nodeLoc, from.Location(), msg,
		)
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
	acceptPeerMsgCh := make(chan peerMsgCh, len(p.peerPubs))
	for !done {
		select {
		case coordStepCall := <-p.coordStepCh:
			switch coordStepCall.msg.Step {
			case "1-R2.1-SendDeals":
				go func() {
					acceptPeerMsgType = rabinDealMsgType
					res := p.doCoordStepSendDeals(acceptPeerMsgCh)
					acceptPeerMsgType = rabinResponseMsgType
					coordStepCall.resp <- res
				}()
			case "2-R2.2-SendResponses":
				go func() {
					res := p.doCoordStepSendResponses(acceptPeerMsgCh)
					acceptPeerMsgType = rabinJustificationMsgType
					coordStepCall.resp <- res
				}()
			case "3-R2.3-SendJustifications":
				go func() {
					res := p.doCoordStepSendJustifications(acceptPeerMsgCh)
					acceptPeerMsgType = rabinSecretCommitsMsgType
					coordStepCall.resp <- res
				}()
			case "4-R4-SendSecretCommits":
				go func() {
					res := p.doCoordStepSendSecretCommits(acceptPeerMsgCh)
					acceptPeerMsgType = rabinComplaintCommitsMsgType
					coordStepCall.resp <- res
				}()
			case "5-R5-SendComplaintCommits":
				go func() {
					res := p.doCoordStepSendComplaintCommits(acceptPeerMsgCh)
					acceptPeerMsgType = rabinReconstructCommitsMsgType
					coordStepCall.resp <- res
				}()
			case "6-R6-SendReconstructCommits": // Invoked from onCoordPubKey
				go func() {
					res := p.doCoordStepSendReconstructCommits(acceptPeerMsgCh)
					acceptPeerMsgType = 0 // None accepted.
					coordStepCall.resp <- res
				}()
			case "7-CommitAndTerminate":
				//
				// TODO: Save the keys.
				//
				p.node.dropProcess(p)
				done = true
				coordStepCall.resp <- nil
			default:
				fmt.Printf(
					"[%v] Dropping unexpected step message: %v\n",
					p.nodeLoc, coordStepCall.msg.Step,
				)
				coordStepCall.resp <- errors.New("unknown_step")
			}
		case cast := <-p.peerMsgCh:
			if cast.msg.MsgType != acceptPeerMsgType {
				fmt.Printf(
					"[%v] Dropping unexpected peer message: type=%v, expected=%v\n",
					p.nodeLoc, cast.msg.MsgType, acceptPeerMsgType,
				)
				continue
			}
			acceptPeerMsgCh <- cast
		case <-p.cancelCh:
			p.node.dropProcess(p)
			done = true
		case <-timeoutCh:
			if p.node.dropProcess(p) {
				fmt.Printf("[%v] Deleting a DkgProc on timeout.\n", p.nodeLoc)
			}
			done = true
		}
	}
	p.done = true
}

func (p *proc) doCoordStepSendDeals(peerMsgCh chan peerMsgCh) error {
	var err error
	var deals map[int]*rabin_dkg.Deal
	if deals, err = p.dkgImpl.Deals(); err != nil {
		fmt.Printf("[%v] ERROR: Deals -> %+v\n", p.nodeLoc, err)
		return err
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
		return err
	}
	//
	// Decode the received deals.
	receivedDeals := make(map[int]*rabinDealMsg, len(receivedDealMsgs))
	for i := range receivedDealMsgs {
		peerDealMsg := rabinDealMsg{}
		if err = peerDealMsg.fromBytes(receivedDealMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		receivedDeals[i] = &peerDealMsg
	}
	p.recvDealMsgs = receivedDeals // Stored for the next step.
	fmt.Printf("[%v] All deals received.\n", p.nodeLoc)
	return nil
}

func (p *proc) doCoordStepSendResponses(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Process the received deals and produce responses.
	ourResponses := []*rabin_dkg.Response{}
	for i := range p.recvDealMsgs {
		var r *rabin_dkg.Response
		if r, err = p.dkgImpl.ProcessDeal(p.recvDealMsgs[i].deal); err != nil {
			fmt.Printf("[%v] ERROR: ProcessDeal(%v) -> %+v\n", p.nodeLoc, i, err)
			return err
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
		return err
	}
	//
	// Decode the received responces.
	receivedResponses := make(map[int]*rabinResponseMsg, len(receivedRespMsgs))
	for i := range receivedRespMsgs {
		peerResponseMsg := rabinResponseMsg{}
		if err = peerResponseMsg.fromBytes(receivedRespMsgs[i].MsgData); err != nil {
			return err
		}
		receivedResponses[i] = &peerResponseMsg
	}
	p.recvResponseMsgs = receivedResponses // Stored for the next step.
	fmt.Printf("[%v] All responses received.\n", p.nodeLoc)
	return nil
}

func (p *proc) doCoordStepSendJustifications(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Process the received responses and produce justifications.
	ourJustifications := []*rabin_dkg.Justification{}
	for i := range p.recvResponseMsgs {
		for _, r := range p.recvResponseMsgs[i].responses {
			var j *rabin_dkg.Justification
			if j, err = p.dkgImpl.ProcessResponse(r); err != nil {
				fmt.Printf("[%v] ERROR: ProcessResponse(%v) -> %+v\n", p.nodeLoc, i, err)
				return err
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
		return err
	}
	//
	// Decode the received justifications.
	receivedJustifications := make(map[int]*rabinJustificationMsg, len(receivedMsgs))
	for i := range receivedMsgs {
		peerJustificationMsg := rabinJustificationMsg{}
		if err = peerJustificationMsg.fromBytes(receivedMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		receivedJustifications[i] = &peerJustificationMsg
	}
	p.recvJustificationMsgs = receivedJustifications // Stored for the next step.
	fmt.Printf("[%v] All justifications received.\n", p.nodeLoc)
	return nil
}

func (p *proc) doCoordStepSendSecretCommits(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Process the received justifications.
	for i := range p.recvJustificationMsgs {
		for _, j := range p.recvJustificationMsgs[i].justifications {
			if err = p.dkgImpl.ProcessJustification(j); err != nil {
				return err
			}
		}
	}
	fmt.Printf("[%v] All justifications processed.\n", p.nodeLoc)
	//
	// Take the QUAL set.
	p.dkgImpl.SetTimeout()
	if !p.dkgImpl.Certified() {
		return fmt.Errorf("node %v not certified", p.nodeIndex)
	}
	thisInQual := p.nodeInQUAL(p.nodeIndex)
	var ourSecretCommits *rabin_dkg.SecretCommits // Will be nil, if we are not in QUAL.
	if thisInQual {
		if ourSecretCommits, err = p.dkgImpl.SecretCommits(); err != nil {
			return err
		}
	}
	//
	// Send our secret commits to all the peers in the QUAL set.
	// Send the nil to all the other nodes, or if we are not in QUAL.
	// We send the messages to all the peers to make it easier to detect the end of step.
	for i := range p.recvDealMsgs { // To all other peers.
		if thisInQual && p.nodeInQUAL(i) {
			p.castPeerByIndex(i, &rabinSecretCommitsMsg{
				secretCommits: ourSecretCommits,
			})
		} else {
			p.castPeerByIndex(i, &rabinSecretCommitsMsg{
				secretCommits: nil,
			})
		}
	}
	//
	// Receive other's secret commits.
	var receivedMsgs map[int]*peering.PeerMessage
	if receivedMsgs, err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	//
	// Decode and process the received secret commits.
	receivedSecretCommits := make(map[int]*rabinSecretCommitsMsg, len(receivedMsgs))
	for i := range receivedMsgs {
		peerSecretCommitsMsg := rabinSecretCommitsMsg{}
		if err = peerSecretCommitsMsg.fromBytes(receivedMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		receivedSecretCommits[i] = &peerSecretCommitsMsg
	}
	p.recvSecretCommitsMsgs = receivedSecretCommits
	fmt.Printf("[%v] All secret commits received, QUAL=%v.\n", p.nodeLoc, p.dkgImpl.QUAL())
	return nil
}

func (p *proc) doCoordStepSendComplaintCommits(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Process the received secret commits.
	ourComplaintCommits := []*rabin_dkg.ComplaintCommits{}
	if p.nodeInQUAL(p.nodeIndex) {
		for i := range p.recvSecretCommitsMsgs {
			sc := p.recvSecretCommitsMsgs[i].secretCommits
			if sc != nil {
				var cc *rabin_dkg.ComplaintCommits
				if cc, err = p.dkgImpl.ProcessSecretCommits(sc); err != nil {
					return err
				}
				if cc != nil {
					ourComplaintCommits = append(ourComplaintCommits, cc)
				}
			}
		}
	}
	//
	// Send our complaint commits.
	for i := range p.recvDealMsgs { // To all other peers.
		if p.nodeInQUAL(i) {
			p.castPeerByIndex(i, &rabinComplaintCommitsMsg{
				complaintCommits: ourComplaintCommits,
			})
		} else {
			p.castPeerByIndex(i, &rabinComplaintCommitsMsg{
				complaintCommits: []*rabin_dkg.ComplaintCommits{},
			})
		}
	}
	//
	// Receive other's complaint commits.
	var receivedMsgs map[int]*peering.PeerMessage
	if receivedMsgs, err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	//
	// Decode and process the received secret commits.
	receivedComplaintCommits := make(map[int]*rabinComplaintCommitsMsg, len(receivedMsgs))
	for i := range receivedMsgs {
		peerComplaintCommitsMsg := rabinComplaintCommitsMsg{}
		if err = peerComplaintCommitsMsg.fromBytes(receivedMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		receivedComplaintCommits[i] = &peerComplaintCommitsMsg
	}
	p.recvComplaintCommitsMsgs = receivedComplaintCommits
	fmt.Printf("[%v] All complaint commits received.\n", p.nodeLoc)
	return nil
}

func (p *proc) doCoordStepSendReconstructCommits(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Process the received complaint commits.
	fmt.Printf("[%v] doCoordStepSendReconstructCommits - proc Start.\n", p.nodeLoc)
	ourReconstructCommits := []*rabin_dkg.ReconstructCommits{}
	if p.nodeInQUAL(p.nodeIndex) {
		for i := range p.recvComplaintCommitsMsgs {
			for _, cc := range p.recvComplaintCommitsMsgs[i].complaintCommits {
				var rc *rabin_dkg.ReconstructCommits
				if rc, err = p.dkgImpl.ProcessComplaintCommits(cc); err != nil {
					return err
				}
				if rc != nil {
					ourReconstructCommits = append(ourReconstructCommits, rc)
				}
			}
		}
	}
	fmt.Printf("[%v] doCoordStepSendReconstructCommits - proc Done.\n", p.nodeLoc)
	//
	// Send our reconstruct commits.
	for i := range p.recvDealMsgs { // To all other peers.
		if p.nodeInQUAL(i) {
			p.castPeerByIndex(i, &rabinReconstructCommitsMsg{
				reconstructCommits: ourReconstructCommits,
			})
		} else {
			p.castPeerByIndex(i, &rabinReconstructCommitsMsg{
				reconstructCommits: []*rabin_dkg.ReconstructCommits{},
			})
		}
	}
	fmt.Printf("[%v] doCoordStepSendReconstructCommits - send Done.\n", p.nodeLoc)
	//
	// Receive other's reconstruct commits.
	var receivedMsgs map[int]*peering.PeerMessage
	if receivedMsgs, err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	fmt.Printf("[%v] doCoordStepSendReconstructCommits - wait Done.\n", p.nodeLoc)
	//
	// Decode and process the received reconstruct commits.
	for i := range receivedMsgs {
		peerReconstructCommitsMsg := rabinReconstructCommitsMsg{}
		if err = peerReconstructCommitsMsg.fromBytes(receivedMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		// receivedReconstructCommits[i] = &peerReconstructCommitsMsg
		for _, rc := range peerReconstructCommitsMsg.reconstructCommits {
			if err = p.dkgImpl.ProcessReconstructCommits(rc); err != nil {
				return err
			}
		}
	}
	if !p.dkgImpl.Finished() {
		return fmt.Errorf("peer %v has not finished the procedure", p.nodeLoc)
	}
	if p.distKeyShare, err = p.dkgImpl.DistKeyShare(); err != nil {
		return err
	}
	fmt.Printf(
		"[%v] All reconstruct commits received, shared public: %v.\n",
		p.nodeLoc, p.distKeyShare.Public(),
	)
	return nil
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

func (p *proc) nodeInQUAL(nodeIdx int) bool {
	for _, q := range p.dkgImpl.QUAL() {
		if q == nodeIdx {
			return true
		}
	}
	return false
}
