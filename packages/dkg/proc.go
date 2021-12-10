// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

// TODO: [KP] Check, if error responses are considered gracefully at the initiator.

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/mr-tron/base58"
	"go.dedis.ch/kyber/v3"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/util/key"
	"golang.org/x/xerrors"
)

const (
	rabinStep0Initialize               = byte(0)
	rabinStep1R21SendDeals             = byte(1)
	rabinStep2R22SendResponses         = byte(2)
	rabinStep3R23SendJustifications    = byte(3)
	rabinStep4R4SendSecretCommits      = byte(4)
	rabinStep5R5SendComplaintCommits   = byte(5)
	rabinStep6R6SendReconstructCommits = byte(6)
	rabinStep7CommitAndTerminate       = byte(7)
)

//
// Stands for a DKG procedure instance on a particular node.
//
type proc struct {
	dkgRef       string            // User supplied unique ID for this instance.
	dkgID        peering.PeeringID // DKG procedure ID we are participating in.
	dkShare      *tcrypto.DKShare  // This will be generated as a result of this procedure.
	node         *Node             // DKG node we are running in.
	nodeIndex    uint16            // Index of this node.
	initiatorPub ed25519.PublicKey
	threshold    uint16
	roundRetry   time.Duration                    // Retry period for the Peer <-> Peer communication.
	netGroup     peering.GroupProvider            // A group for which the distributed key is generated.
	dkgImpl      *rabin_dkg.DistKeyGenerator      // The cryptographic implementation to use.
	dkgLock      *sync.RWMutex                    // Guard access to dkgImpl
	attachID     interface{}                      // We keep it here to be able to detach from the network.
	peerMsgCh    chan *peering.PeerMessageGroupIn // A buffer for the received peer messages.
	log          *logger.Logger                   // A logger to use.
	myNetID      string                           // Just to make logging easier.
	steps        map[byte]*procStep               // All the steps for the procedure.
}

func onInitiatorInit(dkgID peering.PeeringID, msg *initiatorInitMsg, node *Node) (*proc, error) {
	log := node.log.With("dkgID", dkgID.String())
	var err error

	var netGroup peering.GroupProvider
	if netGroup, err = node.netProvider.PeerGroup(dkgID, msg.peerNetIDs); err != nil {
		return nil, err
	}
	var dkgImpl *rabin_dkg.DistKeyGenerator
	if len(msg.peerPubs) >= 2 {
		// We use real DKG only if N >= 2. Otherwise we just generate key pair, and that's all.
		kyberPeerPubs := make([]kyber.Point, len(msg.peerPubs))
		for i := range kyberPeerPubs {
			kyberPeerPubs[i] = node.edSuite.Point()
			if err := kyberPeerPubs[i].UnmarshalBinary(msg.peerPubs[i].Bytes()); err != nil {
				return nil, err
			}
		}
		if dkgImpl, err = rabin_dkg.NewDistKeyGenerator(node.blsSuite, node.edSuite, node.secKey, kyberPeerPubs, int(msg.threshold)); err != nil {
			return nil, xerrors.Errorf("failed to instantiate DistKeyGenerator: %w", err)
		}
	}
	p := proc{
		dkgRef:       msg.dkgRef,
		dkgID:        dkgID,
		node:         node,
		nodeIndex:    netGroup.SelfIndex(),
		initiatorPub: msg.initiatorPub,
		threshold:    msg.threshold,
		roundRetry:   msg.roundRetry,
		netGroup:     netGroup,
		dkgImpl:      dkgImpl,
		dkgLock:      &sync.RWMutex{},
		peerMsgCh:    make(chan *peering.PeerMessageGroupIn, len(msg.peerPubs)),
		log:          log,
		myNetID:      node.netProvider.Self().NetID(),
	}
	p.log.Infof("Starting DKG Peer process at %v for DkgID=%v", p.myNetID, p.dkgID.String())
	stepsStart := make(chan map[uint16]*peering.PeerMessageData)
	p.steps = make(map[byte]*procStep)
	if p.dkgImpl == nil {
		p.steps[rabinStep6R6SendReconstructCommits] = newProcStep(rabinStep6R6SendReconstructCommits, &p,
			stepsStart,
			p.rabinStep6R6SendReconstructCommitsMakeSent,
			p.rabinStep6R6SendReconstructCommitsMakeResp,
		)
		p.steps[rabinStep7CommitAndTerminate] = newProcStep(rabinStep7CommitAndTerminate, &p,
			p.steps[rabinStep6R6SendReconstructCommits].doneCh,
			p.rabinStep7CommitAndTerminateMakeSent,
			p.rabinStep7CommitAndTerminateMakeResp,
		)
	} else {
		p.steps[rabinStep1R21SendDeals] = newProcStep(rabinStep1R21SendDeals, &p,
			stepsStart,
			p.rabinStep1R21SendDealsMakeSent,
			p.rabinStep1R21SendDealsMakeResp,
		)
		p.steps[rabinStep2R22SendResponses] = newProcStep(rabinStep2R22SendResponses, &p,
			p.steps[rabinStep1R21SendDeals].doneCh,
			p.rabinStep2R22SendResponsesMakeSent,
			p.rabinStep2R22SendResponsesMakeResp,
		)
		p.steps[rabinStep3R23SendJustifications] = newProcStep(rabinStep3R23SendJustifications, &p,
			p.steps[rabinStep2R22SendResponses].doneCh,
			p.rabinStep3R23SendJustificationsMakeSent,
			p.rabinStep3R23SendJustificationsMakeResp,
		)
		p.steps[rabinStep4R4SendSecretCommits] = newProcStep(rabinStep4R4SendSecretCommits, &p,
			p.steps[rabinStep3R23SendJustifications].doneCh,
			p.rabinStep4R4SendSecretCommitsMakeSent,
			p.rabinStep4R4SendSecretCommitsMakeResp,
		)
		p.steps[rabinStep5R5SendComplaintCommits] = newProcStep(rabinStep5R5SendComplaintCommits, &p,
			p.steps[rabinStep4R4SendSecretCommits].doneCh,
			p.rabinStep5R5SendComplaintCommitsMakeSent,
			p.rabinStep5R5SendComplaintCommitsMakeResp,
		)
		p.steps[rabinStep6R6SendReconstructCommits] = newProcStep(rabinStep6R6SendReconstructCommits, &p,
			p.steps[rabinStep5R5SendComplaintCommits].doneCh,
			p.rabinStep6R6SendReconstructCommitsMakeSent,
			p.rabinStep6R6SendReconstructCommitsMakeResp,
		)
		p.steps[rabinStep7CommitAndTerminate] = newProcStep(rabinStep7CommitAndTerminate, &p,
			p.steps[rabinStep6R6SendReconstructCommits].doneCh,
			p.rabinStep7CommitAndTerminateMakeSent,
			p.rabinStep7CommitAndTerminateMakeResp,
		)
	}
	go p.processLoop(msg.timeout, p.steps[rabinStep7CommitAndTerminate].doneCh)
	p.attachID = p.netGroup.Attach(peering.PeerMessageReceiverDkg, p.onPeerMessage)
	stepsStart <- make(map[uint16]*peering.PeerMessageData)
	return &p, nil
}

// Handles a message from a peer and pass it to the main thread.
func (p *proc) onPeerMessage(peerMsg *peering.PeerMessageGroupIn) {
	p.peerMsgCh <- peerMsg
}

// That's the main thread executing all the procedure steps.
// We use a single process to make all the actions sequential.
func (p *proc) processLoop(timeout time.Duration, doneCh chan map[uint16]*peering.PeerMessageData) {
	done := false
	timeoutCh := time.After(timeout)
	for {
		select {
		case recv := <-p.peerMsgCh:
			if isDkgInitProcRecvMsg(recv.MsgType) || isDkgRabinRoundMsg(recv.MsgType) || isDkgRabinEchoMsg(recv.MsgType) {
				step := readDkgMessageStep(recv.MsgData)
				if s := p.steps[step]; s != nil {
					s.recv(recv)
				} else {
					p.log.Warnf("Dropping message with unexpected step=%v", step)
				}
			}
			continue // Drop messages sent for the node or the initiator.
		case <-doneCh:
			// We cannot terminate the process here, because other peers can still request
			// to resend some messages. We will wait until the timeout.
			done = true
		case <-timeoutCh:
			p.netGroup.Detach(p.attachID)
			for i := range p.steps {
				p.steps[i].close()
			}
			if p.node.dropProcess(p) {
				if done {
					p.log.Debugf("Deleting completed DkgProc.")
				} else {
					p.log.Warnf("Deleting non-completed a DkgProc on timeout.")
				}
			}
			return
		}
	}
}

//
// rabinStep1R21SendDeals
//
func (p *proc) rabinStep1R21SendDealsMakeSent(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error) {
	var err error
	if p.dkgImpl == nil {
		return nil, errors.New("unexpected step for n=1")
	}
	p.dkgLock.Lock()
	var deals map[int]*rabin_dkg.Deal
	if deals, err = p.dkgImpl.Deals(); err != nil {
		p.dkgLock.Unlock()
		p.log.Errorf("Deals -> %+v", err)
		return nil, err
	}
	p.dkgLock.Unlock()
	sentMsgs := make(map[uint16]*peering.PeerMessageData)
	for i := range deals {
		sentMsgs[uint16(i)] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinDealMsg{
			deal: deals[i],
		})
	}
	return sentMsgs, nil
}

func (p *proc) rabinStep1R21SendDealsMakeResp(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error) {
	return makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &initiatorStatusMsg{error: nil}), nil
}

//
// rabinStep2R22SendResponses
//
func (p *proc) rabinStep2R22SendResponsesMakeSent(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error) {
	var err error
	if p.dkgImpl == nil {
		return nil, errors.New("unexpected step for n=1")
	}
	//
	// Decode the received deals, avoid nested locks.
	recvDeals := make(map[uint16]*rabinDealMsg, len(prevMsgs))
	for i := range prevMsgs {
		peerDealMsg := rabinDealMsg{}
		if err := peerDealMsg.fromBytes(prevMsgs[i].MsgData, p.node.edSuite); err != nil {
			return nil, err
		}
		recvDeals[i] = &peerDealMsg
	}
	//
	// Process the received deals and produce responses.
	ourResponses := []*rabin_dkg.Response{}
	for i := range recvDeals {
		var r *rabin_dkg.Response
		p.dkgLock.Lock()
		if r, err = p.dkgImpl.ProcessDeal(recvDeals[i].deal); err != nil {
			p.dkgLock.Unlock()
			p.log.Errorf("ProcessDeal(%v) -> %+v", i, err)
			return nil, err
		}
		p.dkgLock.Unlock()
		p.log.Debugf("RabinDKG[%v] DealResponse[%v|%v]=%v", p.myNetID, r.Index, r.Response.Index, base58.Encode(r.Response.SessionID))
		ourResponses = append(ourResponses, r)
	}
	//
	// Produce the sent messages.
	sentMsgs := make(map[uint16]*peering.PeerMessageData)
	for i := range prevMsgs { // Use peerIdx from the previous round.
		sentMsgs[i] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinResponseMsg{
			responses: ourResponses,
		})
	}
	return sentMsgs, nil
}

func (p *proc) rabinStep2R22SendResponsesMakeResp(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error) {
	return makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &initiatorStatusMsg{error: nil}), nil
}

//
// rabinStep3R23SendJustifications
//
func (p *proc) rabinStep3R23SendJustificationsMakeSent(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error) {
	var err error
	if p.dkgImpl == nil {
		return nil, errors.New("unexpected step for n=1")
	}
	//
	// Decode the received response.
	recvResponses := make(map[uint16]*rabinResponseMsg)
	for i := range prevMsgs {
		peerResponseMsg := rabinResponseMsg{}
		if err = peerResponseMsg.fromBytes(prevMsgs[i].MsgData); err != nil {
			err = fmt.Errorf("Response: decoding failed: %v", err)
			return nil, err
		}
		recvResponses[i] = &peerResponseMsg
	}
	//
	// Process the received responses and produce justifications.
	ourJustifications := []*rabin_dkg.Justification{}
	for i := range recvResponses {
		for _, r := range recvResponses[i].responses {
			p.dkgLock.Lock()
			var j *rabin_dkg.Justification
			p.log.Debugf("RabinDKG[%v] ProcResponse[%v|%v]=%v", p.myNetID, r.Index, r.Response.Index, base58.Encode(r.Response.SessionID))
			if j, err = p.dkgImpl.ProcessResponse(r); err != nil {
				p.dkgLock.Unlock()
				p.log.Errorf("ProcessResponse(%v) -> %+v, resp.SessionID=%v", i, err, base58.Encode(r.Response.SessionID))
				return nil, err
			}
			p.dkgLock.Unlock()
			if j != nil {
				ourJustifications = append(ourJustifications, j)
			}
		}
	}
	//
	// Produce the sent messages.
	sentMsgs := make(map[uint16]*peering.PeerMessageData)
	for i := range prevMsgs { // Use peerIdx from the previous round.
		sentMsgs[i] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinJustificationMsg{
			justifications: ourJustifications,
		})
	}
	return sentMsgs, nil
}

func (p *proc) rabinStep3R23SendJustificationsMakeResp(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error) {
	return makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &initiatorStatusMsg{error: nil}), nil
}

//
// rabinStep4R4SendSecretCommits
//
func (p *proc) rabinStep4R4SendSecretCommitsMakeSent(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error) {
	var err error
	if p.dkgImpl == nil {
		return nil, errors.New("unexpected step for n=1")
	}
	//
	// Decode the received justifications.
	recvJustifications := make(map[uint16]*rabinJustificationMsg)
	for i := range prevMsgs {
		peerJustificationMsg := rabinJustificationMsg{}
		if err = peerJustificationMsg.fromBytes(prevMsgs[i].MsgData, p.node.blsSuite); err != nil {
			return nil, fmt.Errorf("Justification: decoding failed: %v", err)
		}
		recvJustifications[i] = &peerJustificationMsg
	}
	//
	// Process the received justifications.
	p.dkgLock.Lock()
	for i := range recvJustifications {
		for _, j := range recvJustifications[i].justifications {
			if err = p.dkgImpl.ProcessJustification(j); err != nil {
				p.dkgLock.Unlock()
				return nil, fmt.Errorf("Justification: processing failed: %v", err)
			}
		}
	}
	p.dkgLock.Unlock()
	p.log.Debugf("All justifications processed.")
	//
	// Take the QUAL set.
	p.dkgLock.Lock()
	p.dkgImpl.SetTimeout()
	if !p.dkgImpl.Certified() {
		p.dkgLock.Unlock()
		return nil, fmt.Errorf("node not certified")
	}
	p.dkgLock.Unlock()
	thisInQual := p.nodeInQUAL(p.nodeIndex)
	var ourSecretCommits *rabin_dkg.SecretCommits // Will be nil, if we are not in QUAL.
	if thisInQual {
		p.dkgLock.Lock()
		if ourSecretCommits, err = p.dkgImpl.SecretCommits(); err != nil {
			p.dkgLock.Unlock()
			return nil, fmt.Errorf("SecretCommits: generation failed: %v", err)
		}
		p.dkgLock.Unlock()
	}
	//
	// Produce the sent messages.
	sentMsgs := make(map[uint16]*peering.PeerMessageData)
	for i := range prevMsgs { // Use peerIdx from the previous round.
		if thisInQual && p.nodeInQUAL(i) {
			sentMsgs[i] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinSecretCommitsMsg{
				secretCommits: ourSecretCommits,
			})
		} else {
			sentMsgs[i] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinSecretCommitsMsg{
				secretCommits: nil,
			})
		}
	}
	return sentMsgs, nil
}

func (p *proc) rabinStep4R4SendSecretCommitsMakeResp(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error) {
	return makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &initiatorStatusMsg{error: nil}), nil
}

//
// rabinStep5R5SendComplaintCommits
//
func (p *proc) rabinStep5R5SendComplaintCommitsMakeSent(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error) {
	var err error
	if p.dkgImpl == nil {
		return nil, errors.New("unexpected step for n=1")
	}
	//
	// Decode and process the received secret commits.
	recvSecretCommits := make(map[uint16]*rabinSecretCommitsMsg)
	for i := range prevMsgs {
		peerSecretCommitsMsg := rabinSecretCommitsMsg{}
		if err := peerSecretCommitsMsg.fromBytes(prevMsgs[i].MsgData, p.node.blsSuite); err != nil {
			return nil, err
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
				p.dkgLock.Lock()
				var cc *rabin_dkg.ComplaintCommits
				if cc, err = p.dkgImpl.ProcessSecretCommits(sc); err != nil {
					p.dkgLock.Unlock()
					return nil, err
				}
				p.dkgLock.Unlock()
				if cc != nil {
					ourComplaintCommits = append(ourComplaintCommits, cc)
				}
			}
		}
	}
	//
	// Produce the sent messages.
	sentMsgs := make(map[uint16]*peering.PeerMessageData)
	for i := range prevMsgs { // Use peerIdx from the previous round.
		if p.nodeInQUAL(i) {
			sentMsgs[i] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinComplaintCommitsMsg{
				complaintCommits: ourComplaintCommits,
			})
		} else {
			sentMsgs[i] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinComplaintCommitsMsg{
				complaintCommits: []*rabin_dkg.ComplaintCommits{},
			})
		}
	}
	return sentMsgs, nil
}

func (p *proc) rabinStep5R5SendComplaintCommitsMakeResp(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error) {
	return makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &initiatorStatusMsg{error: nil}), nil
}

//
// rabinStep6R6SendReconstructCommits
//
func (p *proc) rabinStep6R6SendReconstructCommitsMakeSent(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error) {
	var err error
	if p.dkgImpl == nil {
		// Nothing to exchange in the round, if N=1
		return make(map[uint16]*peering.PeerMessageData), nil
	}
	//
	// Decode and process the received secret commits.
	recvComplaintCommits := make(map[uint16]*rabinComplaintCommitsMsg)
	for i := range prevMsgs {
		peerComplaintCommitsMsg := rabinComplaintCommitsMsg{}
		if err := peerComplaintCommitsMsg.fromBytes(prevMsgs[i].MsgData, p.node.blsSuite); err != nil {
			return nil, err
		}
		recvComplaintCommits[i] = &peerComplaintCommitsMsg
	}
	//
	// Process the received complaint commits.
	ourReconstructCommits := []*rabin_dkg.ReconstructCommits{}
	if p.nodeInQUAL(p.nodeIndex) {
		for i := range recvComplaintCommits {
			for _, cc := range recvComplaintCommits[i].complaintCommits {
				p.dkgLock.Lock()
				var rc *rabin_dkg.ReconstructCommits
				if rc, err = p.dkgImpl.ProcessComplaintCommits(cc); err != nil {
					p.dkgLock.Unlock()
					return nil, err
				}
				p.dkgLock.Unlock()
				if rc != nil {
					ourReconstructCommits = append(ourReconstructCommits, rc)
				}
			}
		}
	}
	//
	// Produce the sent messages.
	sentMsgs := make(map[uint16]*peering.PeerMessageData)
	for i := range prevMsgs { // Use peerIdx from the previous round.
		if p.nodeInQUAL(i) {
			sentMsgs[i] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinReconstructCommitsMsg{
				reconstructCommits: ourReconstructCommits,
			})
		} else {
			sentMsgs[i] = makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &rabinReconstructCommitsMsg{
				reconstructCommits: []*rabin_dkg.ReconstructCommits{},
			})
		}
	}
	return sentMsgs, nil
}

func (p *proc) rabinStep6R6SendReconstructCommitsMakeResp(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error) {
	var err error
	if p.dkgImpl == nil {
		// This is the case for N=1, just use simple BLS key pair.
		keyPair := key.NewKeyPair(p.node.blsSuite)
		p.dkShare, err = tcrypto.NewDKShare(
			0,                             // Index
			1,                             // N
			1,                             // T
			keyPair.Public,                // SharedPublic
			make([]kyber.Point, 0),        // PublicCommits
			[]kyber.Point{keyPair.Public}, // PublicShares
			keyPair.Private,               // PrivateShare
		)
		if err != nil {
			return nil, err
		}
	} else {
		//
		// Process the received reconstruct commits.
		for i := range recvMsgs {
			peerReconstructCommitsMsg := rabinReconstructCommitsMsg{}
			if err := peerReconstructCommitsMsg.fromBytes(recvMsgs[i].MsgData); err != nil {
				return nil, err
			}
			p.dkgLock.Lock()
			for _, rc := range peerReconstructCommitsMsg.reconstructCommits {
				if err = p.dkgImpl.ProcessReconstructCommits(rc); err != nil {
					p.dkgLock.Unlock()
					return nil, err
				}
			}
			p.dkgLock.Unlock()
		}
		//
		// Retrieve the generated DistKeyShare.
		p.dkgLock.Lock()
		if !p.dkgImpl.Finished() {
			p.dkgLock.Unlock()
			return nil, fmt.Errorf("DKG procedure is not finished")
		}
		var distKeyShare *rabin_dkg.DistKeyShare
		if distKeyShare, err = p.dkgImpl.DistKeyShare(); err != nil {
			p.dkgLock.Unlock()
			return nil, err
		}
		p.dkgLock.Unlock()
		//
		// Save the needed info.
		groupSize := uint16(len(p.netGroup.AllNodes()))
		ownIndex := uint16(distKeyShare.PriShare().I)
		publicShares := make([]kyber.Point, groupSize)
		publicShares[ownIndex] = p.node.blsSuite.Point().Mul(distKeyShare.PriShare().V, nil)
		p.dkShare, err = tcrypto.NewDKShare(
			ownIndex,                  // Index
			groupSize,                 // N
			p.threshold,               // T
			distKeyShare.Public(),     // SharedPublic
			distKeyShare.Commits,      // PublicCommits
			publicShares,              // PublicShares
			distKeyShare.PriShare().V, // PrivateShare
		)
		if err != nil {
			return nil, err
		}
	}
	p.log.Debugf(
		"All reconstruct commits received, shared public: %v.",
		p.dkShare.SharedPublic,
	)
	var pubShareMsg *initiatorPubShareMsg
	if pubShareMsg, err = p.makeInitiatorPubShareMsg(step); err != nil {
		return nil, err
	}
	return makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, pubShareMsg), nil
}

//
// rabinStep7CommitAndTerminate
//
func (p *proc) rabinStep7CommitAndTerminateMakeSent(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error) {
	var err error
	doneMsg := initiatorDoneMsg{}
	if err = doneMsg.fromBytes(initRecv.MsgData, p.node.blsSuite); err != nil {
		p.log.Warnf("Dropping message, failed to decode: %v", initRecv)
		return nil, err
	}
	if p.dkShare == nil {
		return nil, errors.New("there is no dkShare to commit")
	}
	p.dkShare.PublicShares = doneMsg.pubShares // Store public shares of all the other peers.
	if err := p.node.registry.SaveDKShare(p.dkShare); err != nil {
		return nil, err
	}
	return make(map[uint16]*peering.PeerMessageData), nil
}

func (p *proc) rabinStep7CommitAndTerminateMakeResp(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error) {
	return makePeerMessage(p.dkgID, peering.PeerMessageReceiverDkg, step, &initiatorStatusMsg{error: nil}), nil
}

func (p *proc) nodeInQUAL(nodeIdx uint16) bool {
	if nodeIdx == 0 && p.dkgImpl == nil {
		return true // If N=1, Idx=0 is in QUAL.
	}
	p.dkgLock.RLock()
	for _, q := range p.dkgImpl.QUAL() {
		if uint16(q) == nodeIdx {
			p.dkgLock.RUnlock()
			return true
		}
	}
	p.dkgLock.RUnlock()
	return false
}

func (p *proc) makeInitiatorPubShareMsg(step byte) (*initiatorPubShareMsg, error) {
	var err error
	var publicShareBytes []byte
	if publicShareBytes, err = p.dkShare.PublicShares[*p.dkShare.Index].MarshalBinary(); err != nil {
		return nil, err
	}
	var signature []byte
	if signature, err = bdn.Sign(p.node.blsSuite, p.dkShare.PrivateShare, publicShareBytes); err != nil {
		return nil, err
	}
	return &initiatorPubShareMsg{
		step:          step,
		sharedAddress: p.dkShare.Address,
		sharedPublic:  p.dkShare.SharedPublic,
		publicShare:   p.dkShare.PublicShares[*p.dkShare.Index],
		signature:     signature,
	}, nil
}

type procStep struct {
	step     byte
	startCh  <-chan map[uint16]*peering.PeerMessageData // Gives a signal to start the current step.
	prevMsgs map[uint16]*peering.PeerMessageData        // Messages received from other peers in the previous step.
	sentMsgs map[uint16]*peering.PeerMessageData        // Messages produced by this peer in this step and sent to others.
	recvMsgs map[uint16]*peering.PeerMessageData        // Messages received from other peers in this step.
	initRecv *peering.PeerMessageGroupIn                // Initiator that activated this step.
	initResp *peering.PeerMessageData                   // Step response to the initiator.
	recvCh   chan *peering.PeerMessageGroupIn           // Channel to receive messages for this step (from initiator and peers).
	doneCh   chan map[uint16]*peering.PeerMessageData   // Indicates, that this step is done.
	closeCh  chan bool                                  // For terminating this process.
	makeSent func(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error)
	onceSent *sync.Once
	makeResp func(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error)
	onceResp *sync.Once
	retryCh  <-chan time.Time
	proc     *proc
	log      *logger.Logger
}

func newProcStep(
	step byte,
	proc *proc,
	startCh <-chan map[uint16]*peering.PeerMessageData,
	makeSent func(step byte, initRecv *peering.PeerMessageGroupIn, prevMsgs map[uint16]*peering.PeerMessageData) (map[uint16]*peering.PeerMessageData, error),
	makeResp func(step byte, initRecv *peering.PeerMessageGroupIn, recvMsgs map[uint16]*peering.PeerMessageData) (*peering.PeerMessageData, error),
) *procStep {
	s := procStep{
		step:     step,
		startCh:  startCh,
		prevMsgs: nil,
		sentMsgs: nil,
		recvMsgs: make(map[uint16]*peering.PeerMessageData),
		initResp: nil,
		recvCh:   make(chan *peering.PeerMessageGroupIn, 1000), // NOTE: The channel depth is not necessary, just for performance.
		doneCh:   make(chan map[uint16]*peering.PeerMessageData),
		closeCh:  make(chan bool),
		makeSent: makeSent,
		onceSent: &sync.Once{},
		makeResp: makeResp,
		onceResp: &sync.Once{},
		retryCh:  nil,
		proc:     proc,
		log:      proc.log.Named(strconv.Itoa(int(step))),
	}
	go s.run()
	return &s
}

func (s *procStep) close() {
	close(s.closeCh)
}

func (s *procStep) recv(msg *peering.PeerMessageGroupIn) {
	s.recvCh <- msg
}

func (s *procStep) run() {
	var err error
	for {
		select {
		case prevMsgs, ok := <-s.startCh:
			if !ok {
				return
			}
			if s.prevMsgs == nil {
				// Only take the first version of the previous messages, just in case.
				s.prevMsgs = prevMsgs
			}
		case recv, ok := <-s.recvCh:
			if !ok {
				return
			}
			if s.prevMsgs == nil {
				continue // Drop early messages.
			}
			//
			// The following is for the case, when we already completed our step, but receiving
			// messages from others. Maybe our messages were lost, so we just resend the same messages.
			if s.initResp != nil {
				if isDkgInitProcRecvMsg(recv.MsgType) {
					s.log.Debugf("[%v -%v-> %v] Resending initiator response.", s.proc.myNetID, s.initResp.MsgType, recv.SenderNetID)
					s.proc.netGroup.SendMsgByIndex(recv.SenderIndex, s.initResp.MsgReceiver, s.initResp.MsgType, s.initResp.MsgData)
					continue
				}
				if isDkgRabinEchoMsg(recv.MsgType) {
					// Do not respond to echo messages, a resend loop will be initiated otherwise.
					continue
				}
				if isDkgRabinRoundMsg(recv.MsgType) {
					// Resend the peer messages as echo messages, because we don't need the responses anymore.
					s.sendEcho(recv)
					continue
				}
				s.log.Warnf("[%v -%v-> %v] Dropping unknown message.", recv.SenderNetID, recv.MsgType, s.proc.myNetID)
				continue
			}
			//
			// The following processes te messages while this step is active.
			if isDkgInitProcRecvMsg(recv.MsgType) {
				s.onceSent.Do(func() {
					s.initRecv = recv
					s.retryCh = time.After(s.proc.roundRetry) // Check the retries.
					if s.sentMsgs, err = s.makeSent(s.step, s.initRecv, s.prevMsgs); err != nil {
						s.log.Errorf("Step %v failed to make round messages, reason=%v", s.step, err)
						s.sentMsgs = make(map[uint16]*peering.PeerMessageData) // No messages will be sent on error.
						s.markDone(makePeerMessage(s.proc.dkgID, peering.PeerMessageReceiverDkg, s.step, &initiatorStatusMsg{error: err}))
					}
					for i := range s.sentMsgs {
						netID, _ := s.proc.netGroup.NetIDByIndex(i)
						s.log.Debugf("[%v -%v-> %v] Sending peer message (first).", s.proc.myNetID, s.sentMsgs[i].MsgType, netID)
						s.proc.netGroup.SendMsgByIndex(i, s.sentMsgs[i].MsgReceiver, s.sentMsgs[i].MsgType, s.sentMsgs[i].MsgData)
					}
					if s.haveAll() {
						s.makeDone()
					}
				})
				continue
			}
			if isDkgRabinRoundMsg(recv.MsgType) || isDkgRabinEchoMsg(recv.MsgType) {
				// in the current step we consider echo messages as ordinary round messages,
				// because it is possible that we have requested for them.
				if s.recvMsgs[recv.SenderIndex] == nil {
					s.recvMsgs[recv.SenderIndex] = &recv.PeerMessageData
				} else if s.sentMsgs != nil && isDkgRabinRoundMsg(recv.MsgType) {
					// If that's a repeated message from the peer, maybe our message has been
					// lost, so we repeat it as an echo, to avoid resend loops.
					s.sendEcho(recv)
				}
				if s.initRecv != nil && s.sentMsgs != nil && s.haveAll() {
					s.makeDone()
				}
				continue
			}
			s.log.Warnf("[%v -%v-> %v] Dropping unknown message.", recv.SenderNetID, recv.MsgType, s.proc.myNetID)
			continue
		case <-s.retryCh:
			// Resend all the messages, from who we haven't received.
			s.retryCh = time.After(s.proc.roundRetry) // Repeat the timer.
			for i := range s.sentMsgs {
				if s.recvMsgs[i] == nil {
					netID, _ := s.proc.netGroup.NetIDByIndex(i)
					s.log.Debugf("[%v -%v-> %v] Resending peer message (retry).", s.proc.myNetID, s.sentMsgs[i].MsgType, netID)
					s.proc.netGroup.SendMsgByIndex(i, s.sentMsgs[i].MsgReceiver, s.sentMsgs[i].MsgType, s.sentMsgs[i].MsgData)
				}
			}
			continue
		case <-s.closeCh:
			return
		}
	}
}

func (s *procStep) sendEcho(recv *peering.PeerMessageGroupIn) {
	var err error
	if sentMsg, sentMsgOK := s.sentMsgs[recv.SenderIndex]; sentMsgOK {
		echoMsg := *sentMsg // Make a copy.
		if echoMsg.MsgType, err = makeDkgRoundEchoMsg(echoMsg.MsgType); err != nil {
			s.log.Warnf("[%v -%v-> %v] Unable to send echo message, reason=%v", s.proc.myNetID, recv.MsgType, recv.SenderNetID, err)
			return
		}
		s.log.Debugf("[%v -%v-> %v] Resending peer message (echo).", s.proc.myNetID, echoMsg.MsgType, recv.SenderNetID)
		s.proc.netGroup.SendMsgByIndex(recv.SenderIndex, echoMsg.MsgReceiver, echoMsg.MsgType, echoMsg.MsgData)
		return
	}
	s.log.Warnf("[%v -%v-> %v] Unable to send echo message, is was not produced yet.", s.proc.myNetID, recv.MsgType, recv.SenderNetID)
}

func (s *procStep) haveAll() bool {
	for i := range s.sentMsgs {
		if s.recvMsgs[i] == nil {
			return false
		}
	}
	return true
}

func (s *procStep) makeDone() {
	var err error
	s.onceResp.Do(func() {
		var initResp *peering.PeerMessageData
		if initResp, err = s.makeResp(s.step, s.initRecv, s.recvMsgs); err != nil {
			s.log.Errorf("Step failed to make round response, reason=%v", err)
			s.markDone(makePeerMessage(s.proc.dkgID, peering.PeerMessageReceiverDkg, s.step, &initiatorStatusMsg{error: err}))
		} else {
			s.markDone(initResp)
		}
	})
}

func (s *procStep) markDone(initResp *peering.PeerMessageData) {
	s.doneCh <- s.recvMsgs // Activate the next step.
	s.initResp = initResp  // Store the response for later resends.
	if s.initRecv != nil {
		s.proc.netGroup.SendMsgByIndex(s.initRecv.SenderIndex, initResp.MsgReceiver, initResp.MsgType, initResp.MsgData) // Send response to the initiator.
	} else {
		s.log.Panicf("Step %v/%v closed with no initiator message.", s.proc.myNetID, s.step)
	}
	s.retryCh = nil // Cancel the retry timer.
	s.log.Debugf("Step %v/%v marked as completed.", s.proc.myNetID, s.step)
}
