package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/peering"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/schnorr"
)

const (
	rabinStep1R21SendDeals             = 1
	rabinStep2R22SendResponses         = 2
	rabinStep3R23SendJustifications    = 3
	rabinStep4R4SendSecretCommits      = 4
	rabinStep5R5SendComplaintCommits   = 5
	rabinStep6R6SendReconstructCommits = 6
	rabinStep7CommitAndTerminate       = 7
)

//
// Stands for a DKG procedure instance on a particular node.
//
type proc struct {
	dkgID                    string            // DKG procedure ID we are participating in.
	chainID                  coretypes.ChainID // Same as dkgID, just parsed.
	dkShare                  *DKShare          // This will be generated as a result of this procedure.
	step                     string            // The current step.
	node                     *node             // DKG node we are running in.
	nodeIndex                int               // Index of this node.
	nodeLoc                  string
	peerLocs                 []string
	peerPubs                 []kyber.Point
	initiatorPub             kyber.Point
	threshold                uint32
	version                  address.Version
	stepTimeout              time.Duration
	netGroup                 peering.GroupProvider
	dkgImpl                  *rabin_dkg.DistKeyGenerator
	cancelCh                 chan bool            // To stop a timer on completion.
	initiatorStepCh          chan initiatorStepCh // To process all the step requests in a single thread.
	peerMsgCh                chan peerMsgCh       //
	done                     bool                 // Indicates, if the process is completed.
	log                      *logger.Logger
	recvDealMsgs             map[int]*rabinDealMsg
	recvResponseMsgs         map[int]*rabinResponseMsg
	recvJustificationMsgs    map[int]*rabinJustificationMsg
	recvSecretCommitsMsgs    map[int]*rabinSecretCommitsMsg
	recvComplaintCommitsMsgs map[int]*rabinComplaintCommitsMsg
	distKeyShare             *rabin_dkg.DistKeyShare
}
type initiatorStepCh struct { // Only for communicating with the main thread.
	msg  *StepReq
	resp chan error
}
type peerMsgCh struct { // Only for communicating with the main thread.
	from    peering.PeerSender
	fromIdx int
	msg     *peering.PeerMessage
}

func onInitiatorInit(dkgID string, msg *InitReq, node *node) (*proc, error) {
	log := node.log.With("dkgID", dkgID)
	var err error
	var chainID coretypes.ChainID
	if chainID, err = coretypes.NewChainIDFromBase58(dkgID); err != nil {
		return nil, err
	}
	var peerPubs []kyber.Point
	if peerPubs, err = pubsFromBytes(msg.PeerPubs, node.suite); err != nil {
		return nil, err
	}
	var initiatorPub kyber.Point
	if initiatorPub, err = pubFromBytes(msg.InitiatorPub, node.suite); err != nil {
		return nil, err
	}
	var netGroup peering.GroupProvider
	if netGroup, err = node.netProvider.Group(msg.PeerLocs); err != nil {
		return nil, err
	}
	var dkgImpl *rabin_dkg.DistKeyGenerator
	if dkgImpl, err = rabin_dkg.NewDistKeyGenerator(node.suite, node.secKey, peerPubs, int(msg.Threshold)); err != nil {
		return nil, err
	}
	var nodeIndex int
	if nodeIndex, err = netGroup.PeerIndex(node.netProvider.Self()); err != nil {
		return nil, err
	}
	timeout := time.Millisecond * time.Duration(msg.TimeoutMS)
	p := proc{
		dkgID:           dkgID,
		chainID:         chainID,
		node:            node,
		nodeIndex:       nodeIndex,
		nodeLoc:         node.netProvider.Self().Location(),
		peerLocs:        msg.PeerLocs,
		peerPubs:        peerPubs,
		initiatorPub:    initiatorPub,
		threshold:       msg.Threshold,
		version:         msg.Version,
		stepTimeout:     timeout,
		netGroup:        netGroup,
		dkgImpl:         dkgImpl,
		cancelCh:        make(chan bool),
		initiatorStepCh: make(chan initiatorStepCh, 10),
		peerMsgCh:       make(chan peerMsgCh, len(peerPubs)),
		done:            false,
		log:             log,
	}
	go p.processLoop(timeout)
	node.netProvider.Attach(chainID, p.onPeerMessage)
	return &p, nil
}

// Handles a `step` call from the initiator and pass it to the main thread.
func (p *proc) onInitiatorStep(msg *StepReq) error {
	resp := make(chan error)
	p.initiatorStepCh <- initiatorStepCh{
		msg:  msg,
		resp: resp,
	}
	return <-resp
}

// Handles a `pubKey` call from the initiator.
func (p *proc) onInitiatorPubKey() (*PubKeyResp, error) {
	var err error
	if p.dkShare == nil {
		if err = p.onInitiatorStep(&StepReq{Step: rabinStep6R6SendReconstructCommits}); err != nil {
			return nil, err
		}
	}
	var sharedPublicBytes []byte
	if sharedPublicBytes, err = p.dkShare.SharedPublic.MarshalBinary(); err != nil {
		return nil, err
	}
	var publicShareBytes []byte
	if publicShareBytes, err = p.dkShare.PublicShare.MarshalBinary(); err != nil {
		return nil, err
	}
	var signature []byte
	switch p.version {
	case address.VersionED25519:
		if signature, err = schnorr.Sign(p.node.suite, p.dkShare.PrivateShare, publicShareBytes); err != nil {
			return nil, err
		}
	case address.VersionBLS:
		switch pairingSuite := p.node.suite.(type) {
		case pairing.Suite:
			if signature, err = bdn.Sign(pairingSuite, p.dkShare.PrivateShare, publicShareBytes); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("pairing suite is required for address.VersionBLS")
		}
	}
	pubKeyResp := PubKeyResp{
		ChainID:      p.dkShare.ChainID.Bytes(),
		SharedPublic: sharedPublicBytes,
		PublicShare:  publicShareBytes,
		Signature:    signature,
	}
	return &pubKeyResp, nil
}

// Handles a message from a peer and pass it to the main thread.
func (p *proc) onPeerMessage(from peering.PeerSender, msg *peering.PeerMessage) {
	var err error
	var fromIdx int
	if fromIdx, err = p.netGroup.PeerIndex(from); err != nil {
		p.log.Warnf("Dropping message from unexpected peer %v: %v", from.Location(), msg)
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
	acceptPeerMsgType := rabinDealMsgType
	acceptPeerMsgCh := make(chan peerMsgCh, len(p.peerPubs))
	for !done {
		select {
		case initiatorStepCall := <-p.initiatorStepCh:
			switch initiatorStepCall.msg.Step {
			case rabinStep1R21SendDeals:
				go func() {
					res := p.doInitiatorStepSendDeals(acceptPeerMsgCh)
					acceptPeerMsgType = rabinResponseMsgType
					initiatorStepCall.resp <- res
				}()
			case rabinStep2R22SendResponses:
				go func() {
					res := p.doInitiatorStepSendResponses(acceptPeerMsgCh)
					acceptPeerMsgType = rabinJustificationMsgType
					initiatorStepCall.resp <- res
				}()
			case rabinStep3R23SendJustifications:
				go func() {
					res := p.doInitiatorStepSendJustifications(acceptPeerMsgCh)
					acceptPeerMsgType = rabinSecretCommitsMsgType
					initiatorStepCall.resp <- res
				}()
			case rabinStep4R4SendSecretCommits:
				go func() {
					res := p.doInitiatorStepSendSecretCommits(acceptPeerMsgCh)
					acceptPeerMsgType = rabinComplaintCommitsMsgType
					initiatorStepCall.resp <- res
				}()
			case rabinStep5R5SendComplaintCommits:
				go func() {
					res := p.doInitiatorStepSendComplaintCommits(acceptPeerMsgCh)
					acceptPeerMsgType = rabinReconstructCommitsMsgType
					initiatorStepCall.resp <- res
				}()
			case rabinStep6R6SendReconstructCommits: // Invoked from onInitiatorPubKey
				go func() {
					res := p.doInitiatorStepSendReconstructCommits(acceptPeerMsgCh)
					acceptPeerMsgType = 0 // None accepted.
					initiatorStepCall.resp <- res
				}()
			case rabinStep7CommitAndTerminate:
				if err := p.doInitiatorStepCommitAndTerminate(); err != nil {
					initiatorStepCall.resp <- err
					continue
				}
				p.node.dropProcess(p)
				done = true
				initiatorStepCall.resp <- nil
			default:
				p.log.Warnf("Dropping unexpected step message: %v", initiatorStepCall.msg.Step)
				initiatorStepCall.resp <- errors.New("unknown_step")
			}
		case cast := <-p.peerMsgCh:
			if cast.msg.MsgType != acceptPeerMsgType {
				p.log.Warnf("Dropping unexpected peer message: type=%v, expected=%v",
					cast.msg.MsgType,
					acceptPeerMsgType,
				)
				continue
			}
			acceptPeerMsgCh <- cast
		case <-p.cancelCh:
			p.node.dropProcess(p)
			done = true
		case <-timeoutCh:
			if p.node.dropProcess(p) {
				p.log.Debugf("Deleting a DkgProc on timeout.")
			}
			done = true
		}
	}
	p.done = true
}

func (p *proc) doInitiatorStepSendDeals(peerMsgCh chan peerMsgCh) error {
	var err error
	var deals map[int]*rabin_dkg.Deal
	if deals, err = p.dkgImpl.Deals(); err != nil {
		p.log.Errorf("Deals -> %+v", err)
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
	p.log.Debugf("All deals received.")
	return nil
}

func (p *proc) doInitiatorStepSendResponses(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Process the received deals and produce responses.
	ourResponses := []*rabin_dkg.Response{}
	for i := range p.recvDealMsgs {
		var r *rabin_dkg.Response
		if r, err = p.dkgImpl.ProcessDeal(p.recvDealMsgs[i].deal); err != nil {
			p.log.Errorf("ProcessDeal(%v) -> %+v", i, err)
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
	p.log.Debugf("All responses received.")
	return nil
}

func (p *proc) doInitiatorStepSendJustifications(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Process the received responses and produce justifications.
	t0 := time.Now() // TODO: Handle it in a nicer way.
	ourJustifications := []*rabin_dkg.Justification{}
	for i := range p.recvResponseMsgs {
		for _, r := range p.recvResponseMsgs[i].responses {
			var j *rabin_dkg.Justification
			if j, err = p.dkgImpl.ProcessResponse(r); err != nil {
				p.log.Errorf("ProcessResponse(%v) -> %+v", i, err)
				return err
			}
			if j != nil {
				ourJustifications = append(ourJustifications, j)
			}
		}
	}
	//
	// Send our justifications.
	t1 := time.Now()
	for i := range p.recvDealMsgs { // To all other peers.
		p.castPeerByIndex(i, &rabinJustificationMsg{
			justifications: ourJustifications,
		})
	}
	//
	// Receive other's justifications.
	t2 := time.Now()
	var receivedMsgs map[int]*peering.PeerMessage
	if receivedMsgs, err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	//
	// Decode the received justifications.
	t3 := time.Now()
	receivedJustifications := make(map[int]*rabinJustificationMsg, len(receivedMsgs))
	for i := range receivedMsgs {
		peerJustificationMsg := rabinJustificationMsg{}
		if err = peerJustificationMsg.fromBytes(receivedMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		receivedJustifications[i] = &peerJustificationMsg
	}
	p.recvJustificationMsgs = receivedJustifications // Stored for the next step.
	t4 := time.Now()
	p.log.Debugf(
		"doInitiatorStepSendJustifications: All justifications received, timing ProcessResponse=%v Send=%v Wait=%v Decode=%v",
		t1.Sub(t0).Milliseconds(),
		t2.Sub(t1).Milliseconds(),
		t3.Sub(t2).Milliseconds(),
		t4.Sub(t3).Milliseconds(),
	)
	return nil
}

func (p *proc) doInitiatorStepSendSecretCommits(peerMsgCh chan peerMsgCh) error {
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
	p.log.Debugf("All justifications processed.")
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
	p.log.Debugf("All secret commits received, QUAL=%v.", p.dkgImpl.QUAL())
	return nil
}

func (p *proc) doInitiatorStepSendComplaintCommits(peerMsgCh chan peerMsgCh) error {
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
	p.log.Debugf("All complaint commits received.")
	return nil
}

func (p *proc) doInitiatorStepSendReconstructCommits(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Process the received complaint commits.
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
	//
	// Receive other's reconstruct commits.
	var receivedMsgs map[int]*peering.PeerMessage
	if receivedMsgs, err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
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
		return fmt.Errorf("peer %v has not finished the DKG procedure", p.nodeLoc)
	}
	if p.distKeyShare, err = p.dkgImpl.DistKeyShare(); err != nil {
		return err
	}
	publicShare := p.node.suite.Point().Mul(p.distKeyShare.PriShare().V, nil)
	p.dkShare, err = NewDKShare(
		uint32(p.distKeyShare.PriShare().I), // Index
		uint32(len(p.peerPubs)),             // N
		p.threshold,                         // T
		p.distKeyShare.Public(),             // SharedPublic
		publicShare,                         // PublicShare
		p.distKeyShare.PriShare().V,         // PrivateShare
		p.version,                           // version
		p.node.suite,                        // suite
	)
	if err != nil {
		return err
	}
	p.log.Debugf(
		"All reconstruct commits received, shared public: %v.",
		p.distKeyShare.Public(),
	)
	return nil
}

func (p *proc) doInitiatorStepCommitAndTerminate() error {
	return p.node.registry.SaveDKShare(p.dkShare)
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
