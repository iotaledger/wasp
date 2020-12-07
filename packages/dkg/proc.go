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
	"github.com/iotaledger/wasp/packages/dks"
	"github.com/iotaledger/wasp/packages/peering"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/schnorr"
)

const (
	rabinStep0Initialize               = 0
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
	dkgID        *coretypes.ChainID // DKG procedure ID we are participating in.
	dkShare      *dks.DKShare       // This will be generated as a result of this procedure.
	step         string             // The current step.
	node         *Node              // DKG node we are running in.
	nodeIndex    uint16             // Index of this node.
	initiatorPub kyber.Point
	threshold    uint16
	version      address.Version
	stepTimeout  time.Duration
	netGroup     peering.GroupProvider
	dkgImpl      *rabin_dkg.DistKeyGenerator
	attachID     interface{}
	peerMsgCh    chan *peering.RecvEvent                  // A buffer for the received peer messages.
	log          *logger.Logger                           // A logger to use.
	recvMsgs     map[byte]map[uint16]*peering.PeerMessage // Messages received in particular step ([Step][Peer]).
}

func onInitiatorInit(dkgID *coretypes.ChainID, msg *initiatorInitMsg, node *Node) (*proc, error) {
	log := node.log.With("dkgID", dkgID.String())
	var err error

	var netGroup peering.GroupProvider
	if netGroup, err = node.netProvider.Group(msg.peerLocs); err != nil {
		return nil, err
	}
	var dkgImpl *rabin_dkg.DistKeyGenerator
	if dkgImpl, err = rabin_dkg.NewDistKeyGenerator(node.suite, node.secKey, msg.peerPubs, int(msg.threshold)); err != nil {
		return nil, err
	}
	var nodeIndex uint16
	if nodeIndex, err = netGroup.PeerIndex(node.netProvider.Self()); err != nil {
		return nil, err
	}
	timeout := time.Millisecond * time.Duration(msg.timeoutMS)
	p := proc{
		dkgID:        dkgID,
		node:         node,
		nodeIndex:    nodeIndex,
		initiatorPub: msg.initiatorPub,
		threshold:    msg.threshold,
		version:      msg.version,
		stepTimeout:  timeout,
		netGroup:     netGroup,
		dkgImpl:      dkgImpl,
		peerMsgCh:    make(chan *peering.RecvEvent, len(msg.peerPubs)),
		log:          log,
		recvMsgs:     map[byte]map[uint16]*peering.PeerMessage{},
	}
	go p.processLoop(timeout)
	p.attachID = node.netProvider.Attach(dkgID, p.onPeerMessage)
	return &p, nil
}

func (p *proc) stringID() string {
	return p.dkgID.String()
}

// Handles a message from a peer and pass it to the main thread.
func (p *proc) onPeerMessage(recv *peering.RecvEvent) {
	var err error
	var fromIdx uint16
	if fromIdx, err = p.netGroup.PeerIndex(recv.From); err != nil {
		p.log.Warnf("Dropping message from unexpected peer %v: %v", recv.From.Location(), recv.Msg)
		return
	}
	recv.Msg.SenderIndex = uint16(fromIdx)
	p.peerMsgCh <- recv
}

// That's the main thread executing all the procedure steps.
// We use a single process to make all the actions sequential.
func (p *proc) processLoop(timeout time.Duration) {
	timeoutCh := time.After(timeout)
	done := false
	acceptPeerMsgType := rabinDealMsgType
	acceptPeerMsgCh := make(chan *peering.RecvEvent, len(p.netGroup.AllNodes()))
	for !done {
		select {
		case recv := <-p.peerMsgCh:
			switch recv.Msg.MsgType {
			case initiatorStepMsgType:
				stepMsg := initiatorStepMsg{}
				if err := stepMsg.fromBytes(recv.Msg.MsgData, p.node.suite); err != nil {
					p.log.Warnf("Dropping message, failed to decode: %v", recv)
				}
				switch stepMsg.step {
				case rabinStep1R21SendDeals:
					go func() {
						res := p.doInitiatorStepSendDeals(acceptPeerMsgCh)
						acceptPeerMsgType = rabinResponseMsgType
						recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: stepMsg.step, error: res}))
					}()
				case rabinStep2R22SendResponses:
					go func() {
						res := p.doInitiatorStepSendResponses(acceptPeerMsgCh)
						acceptPeerMsgType = rabinJustificationMsgType
						recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: stepMsg.step, error: res}))
					}()
				case rabinStep3R23SendJustifications:
					go func() {
						res := p.doInitiatorStepSendJustifications(acceptPeerMsgCh)
						acceptPeerMsgType = rabinSecretCommitsMsgType
						recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: stepMsg.step, error: res}))
					}()
				case rabinStep4R4SendSecretCommits:
					go func() {
						res := p.doInitiatorStepSendSecretCommits(acceptPeerMsgCh)
						acceptPeerMsgType = rabinComplaintCommitsMsgType
						recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: stepMsg.step, error: res}))
					}()
				case rabinStep5R5SendComplaintCommits:
					go func() {
						res := p.doInitiatorStepSendComplaintCommits(acceptPeerMsgCh)
						acceptPeerMsgType = rabinReconstructCommitsMsgType
						recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: stepMsg.step, error: res}))
					}()
				case rabinStep6R6SendReconstructCommits: // Invoked from onInitiatorPubKey
					go func() {
						res := p.doInitiatorStepSendReconstructCommits(acceptPeerMsgCh)
						acceptPeerMsgType = 0 // None accepted.
						if res == nil {
							if pubShareMsg, err := p.makeInitiatorPubShareMsg(stepMsg.step); err == nil {
								recv.From.SendMsg(makePeerMessage(p.dkgID, pubShareMsg))
							} else {
								recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: stepMsg.step, error: res}))
							}
						} else {
							recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: stepMsg.step, error: res}))
						}
					}()
				default:
					p.log.Warnf("Dropping unexpected step message: %v", stepMsg.step)
					err := errors.New("unknown_step")
					recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: stepMsg.step, error: err}))
				}
			case initiatorDoneMsgType:
				doneMsg := initiatorDoneMsg{}
				if err := doneMsg.fromBytes(recv.Msg.MsgData, p.node.suite); err != nil {
					p.log.Warnf("Dropping message, failed to decode: %v", recv)
				}
				if err := p.doInitiatorStepCommitAndTerminate(); err != nil {
					recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: doneMsg.step, error: err}))
					continue
				}
				p.node.netProvider.Detach(p.attachID)
				p.node.dropProcess(p)
				done = true
				recv.From.SendMsg(makePeerMessage(p.dkgID, &initiatorStatusMsg{step: doneMsg.step, error: nil}))
			case initiatorPubShareMsgType:
				// That's a message for initiator, ignore it here.
			case initiatorStatusMsgType:
				// That's a message for initiator, ignore it here.
			case acceptPeerMsgType: // NOTE: Variable, dinamically selects a message for particular step.
				acceptPeerMsgCh <- recv
			default:
				p.log.Warnf("Dropping unexpected peer message: type=%v, expected=%v",
					recv.Msg.MsgType,
					acceptPeerMsgType,
				)
				continue
			}
		case <-timeoutCh:
			p.node.netProvider.Detach(p.attachID)
			if p.node.dropProcess(p) {
				p.log.Debugf("Deleting a DkgProc on timeout.")
			}
			done = true
		}
	}
}

func (p *proc) doInitiatorStepSendDeals(peerMsgCh chan *peering.RecvEvent) error {
	var err error
	//
	// Create the deals.
	var deals map[int]*rabin_dkg.Deal
	if deals, err = p.dkgImpl.Deals(); err != nil {
		p.log.Errorf("Deals -> %+v", err)
		return err
	}
	//
	// Send own deals.
	for d := range deals {
		p.castPeerByIndex(uint16(d), &rabinDealMsg{
			deal: deals[d],
		})
	}
	//
	// Receive other's deals.
	if p.recvMsgs[rabinStep1R21SendDeals], err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	p.log.Debugf("All deals received.")
	return nil
}

func (p *proc) doInitiatorStepSendResponses(peerMsgCh chan *peering.RecvEvent) error {
	var err error
	//
	// Decode the received deals.
	recvDealMsgs := p.recvMsgs[rabinStep1R21SendDeals]
	recvDeals := make(map[uint16]*rabinDealMsg, len(recvDealMsgs))
	for i := range recvDealMsgs {
		peerDealMsg := rabinDealMsg{}
		if err = peerDealMsg.fromBytes(recvDealMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		recvDeals[i] = &peerDealMsg
	}
	//
	// Process the received deals and produce responses.
	ourResponses := []*rabin_dkg.Response{}
	for i := range recvDeals {
		var r *rabin_dkg.Response
		if r, err = p.dkgImpl.ProcessDeal(recvDeals[i].deal); err != nil {
			p.log.Errorf("ProcessDeal(%v) -> %+v", i, err)
			return err
		}
		ourResponses = append(ourResponses, r)
	}
	//
	// Send our responses.
	for i := range recvDealMsgs { // To all other peers.
		p.castPeerByIndex(i, &rabinResponseMsg{
			responses: ourResponses,
		})
	}
	//
	// Receive other's responses.
	if p.recvMsgs[rabinStep2R22SendResponses], err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	p.log.Debugf("All responses received.")
	return nil
}

func (p *proc) doInitiatorStepSendJustifications(peerMsgCh chan *peering.RecvEvent) error {
	var err error
	//
	// Decode the received responces.
	recvResponseMsgs := p.recvMsgs[rabinStep2R22SendResponses]
	recvResponses := make(map[uint16]*rabinResponseMsg, len(recvResponseMsgs))
	for i := range recvResponseMsgs {
		peerResponseMsg := rabinResponseMsg{}
		if err = peerResponseMsg.fromBytes(recvResponseMsgs[i].MsgData); err != nil {
			return err
		}
		recvResponses[i] = &peerResponseMsg
	}
	//
	// Process the received responses and produce justifications.
	ourJustifications := []*rabin_dkg.Justification{}
	for i := range recvResponses {
		for _, r := range recvResponses[i].responses {
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
	for i := range recvResponseMsgs { // To all other peers.
		p.castPeerByIndex(i, &rabinJustificationMsg{
			justifications: ourJustifications,
		})
	}
	//
	// Receive other's justifications.
	if p.recvMsgs[rabinStep3R23SendJustifications], err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	return nil
}

func (p *proc) doInitiatorStepSendSecretCommits(peerMsgCh chan *peering.RecvEvent) error {
	var err error
	//
	// Decode the received justifications.
	recvJustificationMsgs := p.recvMsgs[rabinStep3R23SendJustifications]
	recvJustifications := make(map[uint16]*rabinJustificationMsg, len(recvJustificationMsgs))
	for i := range recvJustificationMsgs {
		peerJustificationMsg := rabinJustificationMsg{}
		if err = peerJustificationMsg.fromBytes(recvJustificationMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		recvJustifications[i] = &peerJustificationMsg
	}
	//
	// Process the received justifications.
	for i := range recvJustifications {
		for _, j := range recvJustifications[i].justifications {
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
		return fmt.Errorf("node not certified")
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
	for i := range recvJustificationMsgs { // To all other peers.
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
	if p.recvMsgs[rabinStep4R4SendSecretCommits], err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	p.log.Debugf("All secret commits received, QUAL=%v.", p.dkgImpl.QUAL())
	return nil
}

func (p *proc) doInitiatorStepSendComplaintCommits(peerMsgCh chan *peering.RecvEvent) error {
	var err error
	//
	// Decode and process the received secret commits.
	recvSecretCommitMsgs := p.recvMsgs[rabinStep4R4SendSecretCommits]
	recvSecretCommits := make(map[uint16]*rabinSecretCommitsMsg, len(recvSecretCommitMsgs))
	for i := range recvSecretCommitMsgs {
		peerSecretCommitsMsg := rabinSecretCommitsMsg{}
		if err = peerSecretCommitsMsg.fromBytes(recvSecretCommitMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		recvSecretCommits[i] = &peerSecretCommitsMsg
	}
	//
	// Process the received secret commits.
	ourComplaintCommits := []*rabin_dkg.ComplaintCommits{}
	if p.nodeInQUAL(p.nodeIndex) {
		for i := range recvSecretCommits {
			sc := recvSecretCommits[i].secretCommits
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
	for i := range recvSecretCommitMsgs { // To all other peers.
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
	if p.recvMsgs[rabinStep5R5SendComplaintCommits], err = p.waitForOthers(peerMsgCh, p.stepTimeout); err != nil {
		return err
	}
	p.log.Debugf("All complaint commits received.")
	return nil
}

func (p *proc) doInitiatorStepSendReconstructCommits(peerMsgCh chan *peering.RecvEvent) error {
	var err error
	//
	// Decode and process the received secret commits.
	recvComplaintCommitMsgs := p.recvMsgs[rabinStep5R5SendComplaintCommits]
	recvComplaintCommits := make(map[uint16]*rabinComplaintCommitsMsg, len(recvComplaintCommitMsgs))
	for i := range recvComplaintCommitMsgs {
		peerComplaintCommitsMsg := rabinComplaintCommitsMsg{}
		if err = peerComplaintCommitsMsg.fromBytes(recvComplaintCommitMsgs[i].MsgData, p.node.suite); err != nil {
			return err
		}
		recvComplaintCommits[i] = &peerComplaintCommitsMsg
	}
	//
	// Process the received complaint commits.
	ourReconstructCommits := []*rabin_dkg.ReconstructCommits{}
	if p.nodeInQUAL(p.nodeIndex) {
		for i := range recvComplaintCommits {
			for _, cc := range recvComplaintCommits[i].complaintCommits {
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
	for i := range recvComplaintCommitMsgs { // To all other peers.
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
	var receivedMsgs map[uint16]*peering.PeerMessage
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
		return fmt.Errorf("DKG procedure is not finished")
	}
	var distKeyShare *rabin_dkg.DistKeyShare
	if distKeyShare, err = p.dkgImpl.DistKeyShare(); err != nil {
		return err
	}
	publicShare := p.node.suite.Point().Mul(distKeyShare.PriShare().V, nil)
	p.dkShare, err = dks.NewDKShare(
		uint16(distKeyShare.PriShare().I),  // Index
		uint16(len(p.netGroup.AllNodes())), // N
		p.threshold,                        // T
		distKeyShare.Public(),              // SharedPublic
		publicShare,                        // PublicShare
		distKeyShare.PriShare().V,          // PrivateShare
		p.version,                          // version
		p.node.suite,                       // suite
	)
	if err != nil {
		return err
	}
	p.log.Debugf(
		"All reconstruct commits received, shared public: %v.",
		distKeyShare.Public(),
	)
	return nil
}

func (p *proc) doInitiatorStepCommitAndTerminate() error {
	return p.node.registry.SaveDKShare(p.dkShare)
}

func (p *proc) castPeerByIndex(peerIdx uint16, msg msgByteCoder) {
	p.netGroup.SendMsgByIndex(peerIdx, makePeerMessage(p.dkgID, msg))
}

func (p *proc) waitForOthers(peerMsgCh chan *peering.RecvEvent, timeout time.Duration) (map[uint16]*peering.PeerMessage, error) {
	flags := make([]bool, len(p.netGroup.AllNodes()))
	flags[p.nodeIndex] = true // Assume our message is already present.
	msgs := make(map[uint16]*peering.PeerMessage)
	timeoutCh := time.After(timeout)
	for !haveAll(flags) {
		select {
		case recv := <-peerMsgCh:
			msgs[recv.Msg.SenderIndex] = recv.Msg
			flags[recv.Msg.SenderIndex] = true
		case <-timeoutCh:
			return nil, errors.New("step_timeout")
		}
	}
	return msgs, nil
}

func (p *proc) nodeInQUAL(nodeIdx uint16) bool {
	for _, q := range p.dkgImpl.QUAL() {
		if uint16(q) == nodeIdx {
			return true
		}
	}
	return false
}

func (p *proc) makeInitiatorPubShareMsg(step byte) (*initiatorPubShareMsg, error) {
	var err error
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
	return &initiatorPubShareMsg{
		step:         step,
		chainID:      p.dkShare.ChainID,
		sharedPublic: p.dkShare.SharedPublic,
		publicShare:  p.dkShare.PublicShare,
		signature:    signature,
	}, nil
}
