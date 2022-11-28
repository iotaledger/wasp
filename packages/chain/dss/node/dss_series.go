// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"time"

	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/chain/dss"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

const (
	msgTypeDSS byte = iota
)

type dssInstance struct {
	inst      dss.DSS
	asGPA     gpa.AckHandler
	hadInput  bool
	outPart   []int
	outPartCB func([]int)
	outSig    []byte
	outSigCB  func([]byte)
}

type dssSeriesImpl struct {
	node     *dssNodeImpl
	key      string
	dssInsts map[int]*dssInstance
	dkShare  tcrypto.DKShare
	peerPubs map[gpa.NodeID]*cryptolib.PublicKey
	peerNIDs []gpa.NodeID // Index to NodeID mapping.
}

func IsEnoughQuorum(n, t int) (bool, int) {
	maxF := (n - 1) / 3
	return t >= (n - maxF), maxF
}

func newSeries(node *dssNodeImpl, key string, dkShare tcrypto.DKShare) *dssSeriesImpl {
	dkSharePubKeys := dkShare.GetNodePubKeys()
	nodeIndexToPeerNIDs := make([]gpa.NodeID, len(dkSharePubKeys))
	nodeIDsToPeerPubs := map[gpa.NodeID]*cryptolib.PublicKey{}
	for i := range dkSharePubKeys {
		nodeIndexToPeerNIDs[i] = pubKeyAsNodeID(dkSharePubKeys[i])
		nodeIDsToPeerPubs[nodeIndexToPeerNIDs[i]] = dkSharePubKeys[i]
	}
	cmtN := int(dkShare.GetN())
	dksT := int(dkShare.GetT())
	if ok, maxF := IsEnoughQuorum(cmtN, dksT); !ok {
		panic(xerrors.Errorf("cannot create DSS instance for cmtN=%v, maxF=%v, T=%v", cmtN, maxF, dksT))
	}
	s := &dssSeriesImpl{
		node:     node,
		key:      key,
		dssInsts: map[int]*dssInstance{},
		dkShare:  dkShare,
		peerPubs: nodeIDsToPeerPubs,
		peerNIDs: nodeIndexToPeerNIDs,
	}
	return s
}

func (s *dssSeriesImpl) tick(now time.Time) {
	for i := range s.dssInsts {
		s.sendMessages(s.dssInsts[i].asGPA.Input(s.dssInsts[i].asGPA.MakeTickInput(now)), i)
		s.tryReportOutput(s.dssInsts[i])
	}
}

func (s *dssSeriesImpl) start(index int, partCB func([]int), sigCB func([]byte)) error {
	var err error
	//
	// Create new instance, if not yet created.
	if _, ok := s.dssInsts[index]; !ok {
		var dssInst dss.DSS
		var dssAsGPA gpa.AckHandler
		if dssInst, dssAsGPA, err = s.newDSSImpl(); err != nil {
			return err
		}
		s.dssInsts[index] = &dssInstance{
			inst:     dssInst,
			asGPA:    dssAsGPA,
			hadInput: false,
		}
	}
	//
	// Assign (or reassign) the output callback, invoke callbacks, if data is pending.
	if s.dssInsts[index].outPartCB == nil && s.dssInsts[index].outPart != nil {
		partCB(s.dssInsts[index].outPart)
	}
	if s.dssInsts[index].outSigCB == nil && s.dssInsts[index].outSig != nil {
		sigCB(s.dssInsts[index].outSig)
	}
	s.dssInsts[index].outPartCB = partCB
	s.dssInsts[index].outSigCB = sigCB
	//
	// Start the protocol.
	if !s.dssInsts[index].hadInput {
		s.dssInsts[index].hadInput = true
		s.sendMessages(s.dssInsts[index].asGPA.Input(dss.NewInputStart()), index)
		s.tryReportOutput(s.dssInsts[index])
	}
	return nil
}

func (s *dssSeriesImpl) newDSSImpl() (dss.DSS, gpa.AckHandler, error) {
	nodePKs := s.dkShare.GetNodePubKeys()
	nodeIDs := make([]gpa.NodeID, len(nodePKs))
	for i := range nodeIDs {
		nodeIDs[i] = pubKeyAsNodeID(nodePKs[i])
	}
	myPK := nodePKs[*s.dkShare.GetIndex()]
	mySK := s.node.nid.GetPrivateKey()
	me := pubKeyAsNodeID(myPK)
	//
	// CryptoLib -> Kyber.
	kyberMySK, err := mySK.AsKyberKeyPair()
	if err != nil {
		return nil, nil, err
	}
	kyberNodePKs := make(map[gpa.NodeID]kyber.Point, len(nodePKs))
	for i, nid := range nodeIDs {
		kyberNodePKs[nid], err = nodePKs[i].AsKyberPoint()
		if err != nil {
			return nil, nil, err
		}
	}
	//
	// Construct the DSS protocol instance.
	n := len(nodeIDs)
	f := n - int(s.dkShare.GetT())
	s.node.log.Debugf("Constructing DSS instance, CommitteeAddress=%v, n=%v, f=%v, nodeIDs=%v", s.dkShare.DSSSharedPublic(), n, f, nodeIDs)
	d := dss.New(
		tcrypto.DefaultEd25519Suite(), // suite
		nodeIDs,                       // nodeIDs
		kyberNodePKs,                  // nodePKs
		f,                             // f
		me,                            // me
		kyberMySK.Private,             // mySK
		s.dkShare.DSSSecretShare(),    // longTermSecretShare
		s.node.log,
	)
	dAsGPA := gpa.NewAckHandler(
		me,            // me
		d.AsGPA(),     // nested
		1*time.Second, // resendPeriod
	)
	return d, dAsGPA, nil
}

func (s *dssSeriesImpl) sendMessages(outMsgs gpa.OutMessages, index int) {
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(m gpa.Message) {
		msgData, err := m.MarshalBinary()
		if err != nil {
			s.node.log.Warnf("Failed to send a message: %v", err)
			return
		}
		msgPayload, err := makeMsgData(s.key, index, msgData)
		if err != nil {
			s.node.log.Warnf("Failed to send a message: %v", err)
			return
		}
		pm := &peering.PeerMessageData{
			PeeringID:   *s.node.peeringID,
			MsgReceiver: peering.PeerMessageReceiverChainDSS,
			MsgType:     msgTypeDSS,
			MsgData:     msgPayload,
		}
		s.node.net.SendMsgByPubKey(s.peerPubs[m.Recipient()], pm)
	})
}

func (s *dssSeriesImpl) recvMessage(index int, msgData []byte, sender gpa.NodeID) {
	var err error
	dssInst, ok := s.dssInsts[index]
	if !ok {
		dssInst = &dssInstance{}
		dssInst.inst, dssInst.asGPA, err = s.newDSSImpl()
		if err != nil {
			s.node.log.Errorf("Cannot create DSS instance: %v", err)
			return
		}
		s.dssInsts[index] = dssInst // the callbacks can be left unset here, because this node hasn't asked for the instance yet.
	}
	msg, err := dssInst.asGPA.UnmarshalMessage(msgData)
	if err != nil {
		s.node.log.Errorf("Cannot parse DSS message: %v", err)
		return
	}
	msg.SetSender(sender)
	s.sendMessages(dssInst.asGPA.Message(msg), index)
	s.tryReportOutput(dssInst)
}

// Check, maybe an output is already provided, invoke the appropriate callbacks.
func (s *dssSeriesImpl) tryReportOutput(dssInst *dssInstance) {
	if dssOutput := dssInst.asGPA.Output(); dssOutput != nil {
		dssOutputT := dssOutput.(*dss.Output)
		if dssInst.outPart == nil && dssOutputT.ProposedIndexes != nil {
			dssInst.outPart = dssOutputT.ProposedIndexes
			if dssInst.outPartCB != nil {
				dssInst.outPartCB(dssInst.outPart)
			}
		}
		if dssInst.outSig == nil && dssOutputT.Signature != nil {
			dssInst.outSig = dssOutputT.Signature
			if dssInst.outSigCB != nil {
				dssInst.outSigCB(dssInst.outSig)
			}
		}
	}
}

func (s *dssSeriesImpl) decidedIndexProposals(index int, decidedIndexProposals [][]int, messageToSign []byte) error {
	if len(decidedIndexProposals) != len(s.peerNIDs) {
		return xerrors.Errorf("DSS expected len(decidedIndexProposals)=len(committee)")
	}
	if dssInst, ok := s.dssInsts[index]; ok {
		mappedIndexProposals := map[gpa.NodeID][]int{}
		for i := range decidedIndexProposals {
			if decidedIndexProposals[i] != nil {
				mappedIndexProposals[s.peerNIDs[i]] = decidedIndexProposals[i]
			}
		}
		s.sendMessages(dssInst.asGPA.Input(dss.NewInputDecided(mappedIndexProposals, messageToSign)), index)
		s.tryReportOutput(dssInst)
		return nil
	}
	return xerrors.Errorf("DSS instance for index=%v not found", index)
}

func (s *dssSeriesImpl) statusString(index int) string {
	return s.dssInsts[index].asGPA.StatusString()
}
