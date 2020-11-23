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
	dkgID           string            // DKG procedure ID we are participating in.
	dkgChainID      coretypes.ChainID // The same as dkgID, just converted to ChainID.
	dkShare         *dks.DKShare      // This will be generated as a result of this procedure.
	step            string            // The current step.
	node            *node             // DKG node we are running in.
	nodeIndex       int               // Index of this node.
	initiatorPub    kyber.Point
	threshold       uint32
	version         address.Version
	stepTimeout     time.Duration
	netGroup        peering.GroupProvider
	dkgImpl         *rabin_dkg.DistKeyGenerator
	cancelCh        chan bool                            // To stop a timer on completion.
	initiatorStepCh chan initiatorStepCh                 // To process all the step requests in a single thread.
	peerMsgCh       chan peerMsgCh                       // A buffer for the received peer messages.
	log             *logger.Logger                       // A logger to use.
	recvMsgs        map[int]map[int]*peering.PeerMessage // Messages received in particular step ([Step][Peer]).
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
		dkgChainID:      chainID,
		node:            node,
		nodeIndex:       nodeIndex,
		initiatorPub:    initiatorPub,
		threshold:       msg.Threshold,
		version:         msg.Version,
		stepTimeout:     timeout,
		netGroup:        netGroup,
		dkgImpl:         dkgImpl,
		cancelCh:        make(chan bool),
		initiatorStepCh: make(chan initiatorStepCh, 10),
		peerMsgCh:       make(chan peerMsgCh, len(peerPubs)),
		log:             log,
		recvMsgs:        map[int]map[int]*peering.PeerMessage{},
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
	acceptPeerMsgCh := make(chan peerMsgCh, len(p.netGroup.AllNodes()))
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
}

func (p *proc) doInitiatorStepSendDeals(peerMsgCh chan peerMsgCh) error {
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
		p.castPeerByIndex(d, &rabinDealMsg{
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

func (p *proc) doInitiatorStepSendResponses(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Decode the received deals.
	recvDealMsgs := p.recvMsgs[rabinStep1R21SendDeals]
	recvDeals := make(map[int]*rabinDealMsg, len(recvDealMsgs))
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

func (p *proc) doInitiatorStepSendJustifications(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Decode the received responces.
	recvResponseMsgs := p.recvMsgs[rabinStep2R22SendResponses]
	recvResponses := make(map[int]*rabinResponseMsg, len(recvResponseMsgs))
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

func (p *proc) doInitiatorStepSendSecretCommits(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Decode the received justifications.
	recvJustificationMsgs := p.recvMsgs[rabinStep3R23SendJustifications]
	recvJustifications := make(map[int]*rabinJustificationMsg, len(recvJustificationMsgs))
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

func (p *proc) doInitiatorStepSendComplaintCommits(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Decode and process the received secret commits.
	recvSecretCommitMsgs := p.recvMsgs[rabinStep4R4SendSecretCommits]
	recvSecretCommits := make(map[int]*rabinSecretCommitsMsg, len(recvSecretCommitMsgs))
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

func (p *proc) doInitiatorStepSendReconstructCommits(peerMsgCh chan peerMsgCh) error {
	var err error
	//
	// Decode and process the received secret commits.
	recvComplaintCommitMsgs := p.recvMsgs[rabinStep5R5SendComplaintCommits]
	recvComplaintCommits := make(map[int]*rabinComplaintCommitsMsg, len(recvComplaintCommitMsgs))
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
		return fmt.Errorf("DKG procedure is not finished")
	}
	var distKeyShare *rabin_dkg.DistKeyShare
	if distKeyShare, err = p.dkgImpl.DistKeyShare(); err != nil {
		return err
	}
	publicShare := p.node.suite.Point().Mul(distKeyShare.PriShare().V, nil)
	p.dkShare, err = dks.NewDKShare(
		uint32(distKeyShare.PriShare().I),  // Index
		uint32(len(p.netGroup.AllNodes())), // N
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

func (p *proc) castPeerByIndex(peerIdx int, msg msgByteCoder) {
	p.netGroup.SendMsgByIndex(peerIdx, &peering.PeerMessage{
		ChainID:     p.dkgChainID,
		SenderIndex: 0, // TODO: Should be resolved on the receiving side.
		Timestamp:   0, // TODO: What to do with this?
		MsgType:     msg.MsgType(),
		MsgData:     util.MustBytes(msg),
	})
}

func (p *proc) waitForOthers(peerMsgCh chan peerMsgCh, timeout time.Duration) (map[int]*peering.PeerMessage, error) {
	flags := make([]bool, len(p.netGroup.AllNodes()))
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
