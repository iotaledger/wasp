// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

//
// This file contains message types, exchanged between the DKG nodes
// via the peering network.
//

import (
	"bytes"
	"errors"
	"io"
	"time"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	rabin_vss "go.dedis.ch/kyber/v3/share/vss/rabin"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const (
	//
	// Initiator <-> Peer node communication.
	//
	// NOTE: initiatorInitMsgType must be unique across all the uses of peering package,
	// because it is used to start new chain, thus peeringID is not used for message recognition.
	initiatorInitMsgType = peering.FirstUserMsgCode + 184 // Initiator -> Peer: init new DKG, reply with initiatorStatusMsgType.
	//
	// Initiator <-> Peer proc communication.
	initiatorMsgBase         = peering.FirstUserMsgCode + 4 // 4 to align with round numbers.
	initiatorStepMsgType     = initiatorMsgBase + 1         // Initiator -> Peer: start new step, reply with initiatorStatusMsgType.
	initiatorDoneMsgType     = initiatorMsgBase + 2         // Initiator -> Peer: finalize the proc, reply with initiatorStatusMsgType.
	initiatorPubShareMsgType = initiatorMsgBase + 3         // Peer -> Initiator; if keys are already generated, that's response to initiatorStepMsgType.
	initiatorStatusMsgType   = initiatorMsgBase + 4         // Peer -> Initiator; in the case of error or void ack.
	initiatorMsgFree         = initiatorMsgBase + 5         // Just a placeholder for first unallocated message type.
	//
	// Peer <-> Peer communication for the Rabin protocol.
	rabinMsgFrom                   = initiatorMsgFree
	rabinDealMsgType               = rabinMsgFrom + 0
	rabinResponseMsgType           = rabinMsgFrom + 1
	rabinJustificationMsgType      = rabinMsgFrom + 2
	rabinSecretCommitsMsgType      = rabinMsgFrom + 3
	rabinComplaintCommitsMsgType   = rabinMsgFrom + 4
	rabinReconstructCommitsMsgType = rabinMsgFrom + 5
	rabinMsgTill                   = rabinMsgFrom + 6 // Just a placeholder for first unallocated message type.
	//
	// Peer <-> Peer communication for the Rabin protocol, messages repeatedly sent
	// in response to duplicated messages from other peers. They should be treated
	// in a special way to avoid infinite message loops.
	rabinEchoFrom = rabinMsgTill
	rabinEchoTill = rabinEchoFrom + (rabinMsgTill - rabinMsgFrom)
	//
	// The Peer<->Peer communication includes a corresponding KeySetType.
	// We encode it to the MsgType. Messages are recognized as follows:
	//  [rabinMsgFrom,        rabinEchoTill)       --> KeySetType = Ed25519
	//  [rabinKeySetTypeFrom, rabinKeySetTypeTill) --> KeySetType = BLS
	// NOTE: There is not enough bits to encode KeySetType and Echo flags as bits.
	rabinKeySetTypeFrom = rabinEchoTill
	rabinKeySetTypeTill = rabinKeySetTypeFrom + (rabinEchoTill - rabinMsgFrom)
)

type keySetType byte

const (
	keySetTypeEd25519 keySetType = iota // Used to produce L1 signatures.
	keySetTypeBLS                       // Used internally only (randomness).
)

var initPeeringID peering.PeeringID

// Check if that's a Initiator -> PeerProc message.
func isDkgInitProcRecvMsg(msgType byte) bool {
	return msgType == initiatorStepMsgType || msgType == initiatorDoneMsgType
}

// isDkgRabinRoundMsg detects, if the received MsgType is RabinDKG Peer <-> Peer message type and splits it into components.
func isDkgRabinRoundMsg(msgType byte) (bool, keySetType, bool, byte) {
	if msgType < rabinMsgFrom || msgType >= rabinKeySetTypeTill {
		return false, keySetTypeEd25519, false, 0
	}
	kst := keySetTypeEd25519
	if msgType >= rabinKeySetTypeFrom {
		kst = keySetTypeBLS
		msgType -= rabinKeySetTypeFrom
	}
	echo := false
	if msgType >= rabinEchoFrom {
		echo = true
		msgType -= rabinEchoFrom
	}
	return true, kst, echo, msgType
}

// makeDkgRabinMsgType creates a peeringMsgType out of components composing it for the Rabin DKG Peer <-> Peer messages.
func makeDkgRabinMsgType(rabinMsgType byte, kst keySetType, echo bool) byte {
	msgType := rabinMsgType
	if echo {
		msgType = msgType - rabinMsgFrom + rabinEchoFrom
	}
	if kst == keySetTypeBLS {
		msgType = msgType - rabinMsgType + rabinKeySetTypeFrom
	}
	return msgType
}

// All the messages exchanged via the Peering subsystem will implement this.
type msgByteCoder interface {
	MsgType() byte
	Step() byte
	SetStep(step byte)
	Write(io.Writer) error
	Read(io.Reader) error
}

func makePeerMessage(peeringID peering.PeeringID, receiver, step byte, msg msgByteCoder) *peering.PeerMessageData {
	msg.SetStep(step)
	return &peering.PeerMessageData{
		PeeringID:   peeringID,
		MsgReceiver: receiver,
		MsgType:     msg.MsgType(),
		MsgData:     rwutil.WriterToBytes(msg),
	}
}

// All the messages in this module have a step as a first byte in the payload.
// This function reads that step without decoding all the data.
func readDkgMessageStep(msgData []byte) byte {
	return msgData[0]
}

type initiatorMsg interface {
	msgByteCoder
	Error() error
	IsResponse() bool
}

func readInitiatorMsg(peerMessage *peering.PeerMessageData, edSuite, blsSuite kyber.Group) (bool, initiatorMsg, error) {
	switch peerMessage.MsgType {
	case initiatorInitMsgType:
		msg := initiatorInitMsg{}
		if err := msg.fromBytes(peerMessage.MsgData); err != nil {
			return true, nil, err
		}
		return true, &msg, nil
	case initiatorStepMsgType:
		msg := initiatorStepMsg{}
		if err := msg.fromBytes(peerMessage.MsgData); err != nil {
			return true, nil, err
		}
		return true, &msg, nil
	case initiatorDoneMsgType:
		msg := initiatorDoneMsg{}
		if err := msg.fromBytes(peerMessage.MsgData, edSuite, blsSuite); err != nil {
			return true, nil, err
		}
		return true, &msg, nil
	case initiatorPubShareMsgType:
		msg := initiatorPubShareMsg{}
		if err := msg.fromBytes(peerMessage.MsgData, edSuite, blsSuite); err != nil {
			return true, nil, err
		}
		return true, &msg, nil
	case initiatorStatusMsgType:
		msg := initiatorStatusMsg{}
		if err := msg.fromBytes(peerMessage.MsgData); err != nil {
			return true, nil, err
		}
		return true, &msg, nil
	default:
		return false, nil, nil
	}
}

// initiatorInitMsg
//
// This is a message sent by the initiator to all the peers to
// initiate the DKG process.
type initiatorInitMsg struct {
	step         byte
	dkgRef       string // Some unique string to identify duplicate initialization.
	peeringID    peering.PeeringID
	peerPubs     []*cryptolib.PublicKey
	initiatorPub *cryptolib.PublicKey
	threshold    uint16
	timeout      time.Duration
	roundRetry   time.Duration
}

type initiatorInitMsgIn struct {
	initiatorInitMsg
	SenderPubKey *cryptolib.PublicKey
}

func (m *initiatorInitMsg) MsgType() byte {
	return initiatorInitMsgType
}

func (m *initiatorInitMsg) Step() byte {
	return m.step
}

func (m *initiatorInitMsg) SetStep(step byte) {
	m.step = step
}

func (m *initiatorInitMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteString(m.dkgRef)
	ww.WriteN(m.peeringID[:])

	ww.WriteSize(len(m.peerPubs))
	for i := range m.peerPubs {
		ww.WriteBytes(m.peerPubs[i].AsBytes())
	}

	ww.WriteBytes(m.initiatorPub.AsBytes())
	ww.WriteUint16(m.threshold)
	ww.WriteDuration(m.timeout)
	ww.WriteDuration(m.roundRetry)
	return ww.Err
}

func (m *initiatorInitMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	m.dkgRef = rr.ReadString()
	rr.ReadN(m.peeringID[:])

	size := rr.ReadSize()
	m.peerPubs = make([]*cryptolib.PublicKey, size)
	for i := range m.peerPubs {
		m.peerPubs[i] = rwutil.ReadFromBytes(rr, cryptolib.NewPublicKeyFromBytes)
	}

	m.initiatorPub = rwutil.ReadFromBytes(rr, cryptolib.NewPublicKeyFromBytes)
	m.threshold = rr.ReadUint16()
	m.timeout = rr.ReadDuration()
	m.roundRetry = rr.ReadDuration()
	return rr.Err
}

func (m *initiatorInitMsg) fromBytes(buf []byte) error {
	r := bytes.NewReader(buf)
	return m.Read(r)
}

func (m *initiatorInitMsg) Error() error {
	return nil
}

func (m *initiatorInitMsg) IsResponse() bool {
	return false
}

// initiatorStepMsg
//
// This is a message used to synchronize the DKG procedure by
// ensuring the lock-step, as required by the DKG algorithm
// assumptions (Rabin as well as Pedersen).
type initiatorStepMsg struct {
	step byte
}

func (m *initiatorStepMsg) MsgType() byte {
	return initiatorStepMsgType
}

func (m *initiatorStepMsg) Step() byte {
	return m.step
}

func (m *initiatorStepMsg) SetStep(step byte) {
	m.step = step
}

func (m *initiatorStepMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	return ww.Err
}

func (m *initiatorStepMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	return rr.Err
}

func (m *initiatorStepMsg) fromBytes(buf []byte) error {
	r := bytes.NewReader(buf)
	return m.Read(r)
}

func (m *initiatorStepMsg) Error() error {
	return nil
}

func (m *initiatorStepMsg) IsResponse() bool {
	return false
}

// initiatorDoneMsg
type initiatorDoneMsg struct {
	step         byte
	edPubShares  []kyber.Point
	edSuite      kyber.Group // Transient, for un-marshaling only.
	blsPubShares []kyber.Point
	blsSuite     kyber.Group // Transient, for un-marshaling only.
}

func (m *initiatorDoneMsg) MsgType() byte {
	return initiatorDoneMsgType
}

func (m *initiatorDoneMsg) Step() byte {
	return m.step
}

func (m *initiatorDoneMsg) SetStep(step byte) {
	m.step = step
}

func (m *initiatorDoneMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)

	ww.WriteSize(len(m.edPubShares))
	for i := range m.edPubShares {
		ww.WriteMarshaled(m.edPubShares[i])
	}

	ww.WriteSize(len(m.blsPubShares))
	for i := range m.blsPubShares {
		ww.WriteMarshaled(m.blsPubShares[i])
	}
	return ww.Err
}

func (m *initiatorDoneMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()

	size := rr.ReadSize()
	m.edPubShares = make([]kyber.Point, size)
	for i := range m.edPubShares {
		m.edPubShares[i] = m.edSuite.Point()
		rr.ReadMarshaled(m.edPubShares[i])
	}

	size = rr.ReadSize()
	m.blsPubShares = make([]kyber.Point, size)
	for i := range m.blsPubShares {
		m.blsPubShares[i] = m.blsSuite.Point()
		rr.ReadMarshaled(m.blsPubShares[i])
	}
	return rr.Err
}

func (m *initiatorDoneMsg) fromBytes(buf []byte, edSuite, blsSuite kyber.Group) error {
	r := bytes.NewReader(buf)
	m.edSuite = edSuite
	m.blsSuite = blsSuite
	return m.Read(r)
}

func (m *initiatorDoneMsg) Error() error {
	return nil
}

func (m *initiatorDoneMsg) IsResponse() bool {
	return false
}

// initiatorPubShareMsg
//
// This is a message responded to the initiator
// by the DKG peers returning the shared public key.
// All the nodes must return the same public key.
type initiatorPubShareMsg struct {
	step            byte
	sharedAddress   iotago.Address
	edSharedPublic  kyber.Point
	edPublicShare   kyber.Point
	edSignature     []byte
	edSuite         kyber.Group // Transient, for un-marshaling only.
	blsSharedPublic kyber.Point
	blsPublicShare  kyber.Point
	blsSignature    []byte
	blsSuite        kyber.Group // Transient, for un-marshaling only.
}

func (m *initiatorPubShareMsg) MsgType() byte {
	return initiatorPubShareMsgType
}

func (m *initiatorPubShareMsg) Step() byte {
	return m.step
}

func (m *initiatorPubShareMsg) SetStep(step byte) {
	m.step = step
}

func (m *initiatorPubShareMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteAddress(m.sharedAddress)

	ww.WriteMarshaled(m.edSharedPublic)
	ww.WriteMarshaled(m.edPublicShare)
	ww.WriteBytes(m.edSignature)

	ww.WriteMarshaled(m.blsSharedPublic)
	ww.WriteMarshaled(m.blsPublicShare)
	ww.WriteBytes(m.blsSignature)
	return ww.Err
}

func (m *initiatorPubShareMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	m.sharedAddress = rr.ReadAddress()

	m.edSharedPublic = m.edSuite.Point()
	rr.ReadMarshaled(m.edSharedPublic)
	m.edPublicShare = m.edSuite.Point()
	rr.ReadMarshaled(m.edPublicShare)
	m.edSignature = rr.ReadBytes()

	m.blsSharedPublic = m.blsSuite.Point()
	rr.ReadMarshaled(m.blsSharedPublic)
	m.blsPublicShare = m.blsSuite.Point()
	rr.ReadMarshaled(m.blsPublicShare)
	m.blsSignature = rr.ReadBytes()
	return rr.Err
}

func (m *initiatorPubShareMsg) fromBytes(buf []byte, edSuite, blsSuite kyber.Group) error {
	r := bytes.NewReader(buf)
	m.edSuite = edSuite
	m.blsSuite = blsSuite
	return m.Read(r)
}

func (m *initiatorPubShareMsg) Error() error {
	return nil
}

func (m *initiatorPubShareMsg) IsResponse() bool {
	return true
}

// initiatorStatusMsg
type initiatorStatusMsg struct {
	step  byte
	error error
}

func (m *initiatorStatusMsg) MsgType() byte {
	return initiatorStatusMsgType
}

func (m *initiatorStatusMsg) Step() byte {
	return m.step
}

func (m *initiatorStatusMsg) SetStep(step byte) {
	m.step = step
}

func (m *initiatorStatusMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	var errMsg string
	if m.error != nil {
		errMsg = m.error.Error()
	}
	ww.WriteString(errMsg)
	return ww.Err
}

func (m *initiatorStatusMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	errMsg := rr.ReadString()
	m.error = nil
	if errMsg != "" {
		m.error = errors.New(errMsg)
	}
	return rr.Err
}

func (m *initiatorStatusMsg) fromBytes(buf []byte) error {
	r := bytes.NewReader(buf)
	return m.Read(r)
}

func (m *initiatorStatusMsg) Error() error {
	return m.error
}

func (m *initiatorStatusMsg) IsResponse() bool {
	return true
}

// rabin_dkg.Deal
type rabinDealMsg struct {
	step byte
	deal *rabin_dkg.Deal
}

func (m *rabinDealMsg) MsgType() byte {
	return rabinDealMsgType
}

func (m *rabinDealMsg) Step() byte {
	return m.step
}

func (m *rabinDealMsg) SetStep(step byte) {
	m.step = step
}

func (m *rabinDealMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteUint32(m.deal.Index)
	ww.WriteMarshaled(m.deal.Deal.DHKey)
	ww.WriteBytes(m.deal.Deal.Signature)
	ww.WriteBytes(m.deal.Deal.Nonce)
	ww.WriteBytes(m.deal.Deal.Cipher)
	return ww.Err
}

func (m *rabinDealMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	m.deal.Index = rr.ReadUint32()
	rr.ReadMarshaled(m.deal.Deal.DHKey)
	m.deal.Deal.Signature = rr.ReadBytes()
	m.deal.Deal.Nonce = rr.ReadBytes()
	m.deal.Deal.Cipher = rr.ReadBytes()
	return rr.Err
}

func (m *rabinDealMsg) fromBytes(buf []byte, edSuite kyber.Group) error {
	m.deal = &rabin_dkg.Deal{
		Deal: &rabin_vss.EncryptedDeal{
			DHKey: edSuite.Point(),
		},
	}
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

// rabin_dkg.Response
type rabinResponseMsg struct {
	step      byte
	responses []*rabin_dkg.Response
}

func (m *rabinResponseMsg) MsgType() byte {
	return rabinResponseMsgType
}

func (m *rabinResponseMsg) Step() byte {
	return m.step
}

func (m *rabinResponseMsg) SetStep(step byte) {
	m.step = step
}

func (m *rabinResponseMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteSize(len(m.responses))
	for _, response := range m.responses {
		ww.WriteUint32(response.Index)
		ww.WriteBytes(response.Response.SessionID)
		ww.WriteUint32(response.Response.Index)
		ww.WriteBool(response.Response.Approved)
		ww.WriteBytes(response.Response.Signature)
	}
	return ww.Err
}

func (m *rabinResponseMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	size := rr.ReadSize()
	m.responses = make([]*rabin_dkg.Response, size)
	for i := range m.responses {
		response := rabin_dkg.Response{
			Response: &rabin_vss.Response{},
		}
		m.responses[i] = &response
		response.Index = rr.ReadUint32()
		response.Response.SessionID = rr.ReadBytes()
		response.Response.Index = rr.ReadUint32()
		response.Response.Approved = rr.ReadBool()
		response.Response.Signature = rr.ReadBytes()
	}
	return rr.Err
}

func (m *rabinResponseMsg) fromBytes(buf []byte) error {
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

// rabin_dkg.Justification
type rabinJustificationMsg struct {
	step           byte
	justifications []*rabin_dkg.Justification
	blsSuite       kyber.Group // Just for un-marshaling.
}

func (m *rabinJustificationMsg) MsgType() byte {
	return rabinJustificationMsgType
}

func (m *rabinJustificationMsg) Step() byte {
	return m.step
}

func (m *rabinJustificationMsg) SetStep(step byte) {
	m.step = step
}

func (m *rabinJustificationMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteSize(len(m.justifications))
	for _, j := range m.justifications {
		ww.WriteUint32(j.Index)
		ww.WriteBytes(j.Justification.SessionID)
		ww.WriteUint32(j.Justification.Index)
		writeVssDeal(ww, j.Justification.Deal)
		ww.WriteBytes(j.Justification.Signature)
	}
	return ww.Err
}

func (m *rabinJustificationMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	size := rr.ReadSize()
	m.justifications = make([]*rabin_dkg.Justification, size)
	for i := range m.justifications {
		j := &rabin_dkg.Justification{
			Justification: &rabin_vss.Justification{},
		}
		m.justifications[i] = j
		j.Index = rr.ReadUint32()
		j.Justification.SessionID = rr.ReadBytes()
		j.Justification.Index = rr.ReadUint32()
		j.Justification.Deal = readVssDeal(rr, m.blsSuite)
		j.Justification.Signature = rr.ReadBytes()
	}
	return rr.Err
}

func (m *rabinJustificationMsg) fromBytes(buf []byte, blsSuite kyber.Group) error {
	m.blsSuite = blsSuite
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

// rabin_dkg.SecretCommits
type rabinSecretCommitsMsg struct {
	step          byte
	secretCommits *rabin_dkg.SecretCommits
	blsSuite      kyber.Group // Just for un-marshaling.
}

func (m *rabinSecretCommitsMsg) MsgType() byte {
	return rabinSecretCommitsMsgType
}

func (m *rabinSecretCommitsMsg) Step() byte {
	return m.step
}

func (m *rabinSecretCommitsMsg) SetStep(step byte) {
	m.step = step
}

func (m *rabinSecretCommitsMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteBool(m.secretCommits == nil)
	if m.secretCommits == nil {
		return ww.Err
	}

	ww.WriteUint32(m.secretCommits.Index)

	ww.WriteSize(len(m.secretCommits.Commitments))
	for i := range m.secretCommits.Commitments {
		ww.WriteMarshaled(m.secretCommits.Commitments[i])
	}

	ww.WriteBytes(m.secretCommits.SessionID)
	ww.WriteBytes(m.secretCommits.Signature)
	return ww.Err
}

func (m *rabinSecretCommitsMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	isNil := rr.ReadBool()
	if isNil {
		m.secretCommits = nil
		return rr.Err
	}

	m.secretCommits = &rabin_dkg.SecretCommits{}
	m.secretCommits.Index = rr.ReadUint32()

	size := rr.ReadSize()
	m.secretCommits.Commitments = make([]kyber.Point, size)
	for i := range m.secretCommits.Commitments {
		m.secretCommits.Commitments[i] = m.blsSuite.Point()
		rr.ReadMarshaled(m.secretCommits.Commitments[i])
	}

	m.secretCommits.SessionID = rr.ReadBytes()
	m.secretCommits.Signature = rr.ReadBytes()
	return rr.Err
}

func (m *rabinSecretCommitsMsg) fromBytes(buf []byte, blsSuite kyber.Group) error {
	m.blsSuite = blsSuite
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

// rabin_dkg.ComplaintCommits
type rabinComplaintCommitsMsg struct {
	step             byte
	complaintCommits []*rabin_dkg.ComplaintCommits
	blsSuite         kyber.Group // Just for un-marshaling.
}

func (m *rabinComplaintCommitsMsg) MsgType() byte {
	return rabinComplaintCommitsMsgType
}

func (m *rabinComplaintCommitsMsg) Step() byte {
	return m.step
}

func (m *rabinComplaintCommitsMsg) SetStep(step byte) {
	m.step = step
}

func (m *rabinComplaintCommitsMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteSize(len(m.complaintCommits))
	for i := range m.complaintCommits {
		ww.WriteUint32(m.complaintCommits[i].Index)
		ww.WriteUint32(m.complaintCommits[i].DealerIndex)
		writeVssDeal(ww, m.complaintCommits[i].Deal)
		ww.WriteBytes(m.complaintCommits[i].Signature)
	}
	return ww.Err
}

func (m *rabinComplaintCommitsMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	size := rr.ReadSize()
	m.complaintCommits = make([]*rabin_dkg.ComplaintCommits, size)
	for i := range m.complaintCommits {
		m.complaintCommits[i] = &rabin_dkg.ComplaintCommits{}
		m.complaintCommits[i].Index = rr.ReadUint32()
		m.complaintCommits[i].DealerIndex = rr.ReadUint32()
		m.complaintCommits[i].Deal = readVssDeal(rr, m.blsSuite)
		m.complaintCommits[i].Signature = rr.ReadBytes()
	}
	return rr.Err
}

func (m *rabinComplaintCommitsMsg) fromBytes(buf []byte, blsSuite kyber.Group) error {
	m.blsSuite = blsSuite
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

// rabin_dkg.ReconstructCommits
type rabinReconstructCommitsMsg struct {
	step               byte
	reconstructCommits []*rabin_dkg.ReconstructCommits
}

func (m *rabinReconstructCommitsMsg) MsgType() byte {
	return rabinReconstructCommitsMsgType
}

func (m *rabinReconstructCommitsMsg) Step() byte {
	return m.step
}

func (m *rabinReconstructCommitsMsg) SetStep(step byte) {
	m.step = step
}

func (m *rabinReconstructCommitsMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteSize(len(m.reconstructCommits))
	for i := range m.reconstructCommits {
		ww.WriteBytes(m.reconstructCommits[i].SessionID)
		ww.WriteUint32(m.reconstructCommits[i].Index)
		ww.WriteUint32(m.reconstructCommits[i].DealerIndex)
		writePriShare(ww, m.reconstructCommits[i].Share)
		ww.WriteBytes(m.reconstructCommits[i].Signature)
	}
	return ww.Err
}

func (m *rabinReconstructCommitsMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	size := rr.ReadSize()
	m.reconstructCommits = make([]*rabin_dkg.ReconstructCommits, size)
	for i := range m.reconstructCommits {
		m.reconstructCommits[i] = &rabin_dkg.ReconstructCommits{}
		m.reconstructCommits[i].SessionID = rr.ReadBytes()
		m.reconstructCommits[i].Index = rr.ReadUint32()
		m.reconstructCommits[i].DealerIndex = rr.ReadUint32()
		m.reconstructCommits[i].Share = readPriShare(rr)
		m.reconstructCommits[i].Signature = rr.ReadBytes()
	}
	return rr.Err
}

func (m *rabinReconstructCommitsMsg) fromBytes(buf []byte) error {
	rdr := bytes.NewReader(buf)
	return m.Read(rdr)
}

// multiKeySetMsg wraps messages of different protocol instances (for different key set types).
// It is needed to cope with the round synchronization.
type multiKeySetMsg struct {
	step      byte
	edMsg     *peering.PeerMessageData
	blsMsg    *peering.PeerMessageData
	peeringID peering.PeeringID // Transient.
	receiver  byte              // Transient.
	msgType   byte              // Transient.
}

func (m *multiKeySetMsg) MsgType() byte {
	return m.msgType
}

func (m *multiKeySetMsg) Step() byte {
	return m.step
}

func (m *multiKeySetMsg) SetStep(step byte) {
	m.step = step
}

func (m *multiKeySetMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.step)
	ww.WriteBytes(m.edMsg.MsgData)
	ww.WriteBytes(m.blsMsg.MsgData)
	return ww.Err
}

func (m *multiKeySetMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.step = rr.ReadByte()
	m.edMsg = &peering.PeerMessageData{
		PeeringID:   m.peeringID,
		MsgReceiver: m.receiver,
		MsgType:     m.msgType,
		MsgData:     rr.ReadBytes(),
	}
	m.blsMsg = &peering.PeerMessageData{
		PeeringID:   m.peeringID,
		MsgReceiver: m.receiver,
		MsgType:     m.msgType,
		MsgData:     rr.ReadBytes(),
	}
	return rr.Err
}

func (m *multiKeySetMsg) fromBytes(buf []byte, peeringID peering.PeeringID, receiver, msgType byte) error {
	rdr := bytes.NewReader(buf)
	m.peeringID = peeringID
	m.receiver = receiver
	m.msgType = msgType
	return m.Read(rdr)
}

func (m *multiKeySetMsg) mustDataBytes() []byte {
	w := new(bytes.Buffer)
	if err := m.Write(w); err != nil {
		panic(err)
	}
	return w.Bytes()
}

type multiKeySetMsgs map[uint16]*multiKeySetMsg

func (m multiKeySetMsgs) GetEdMsgs() map[uint16]*peering.PeerMessageData {
	res := make(map[uint16]*peering.PeerMessageData)
	for i := range m {
		res[i] = m[i].edMsg
	}
	return res
}

func (m multiKeySetMsgs) GetBLSMsgs() map[uint16]*peering.PeerMessageData {
	res := make(map[uint16]*peering.PeerMessageData)
	for i := range m {
		res[i] = m[i].blsMsg
	}
	return res
}

func (m multiKeySetMsgs) AddDSSMsgs(msgs map[uint16]*peering.PeerMessageData, step byte) {
	for i := range msgs {
		if msg, ok := m[i]; ok {
			msg.edMsg = msgs[i]
		} else {
			m[i] = &multiKeySetMsg{
				step:      step,
				peeringID: msgs[i].PeeringID,
				receiver:  msgs[i].MsgReceiver,
				msgType:   msgs[i].MsgType,
				edMsg:     msgs[i],
			}
		}
	}
}

func (m multiKeySetMsgs) AddBLSMsgs(msgs map[uint16]*peering.PeerMessageData, step byte) {
	for i := range msgs {
		if msg, ok := m[i]; ok {
			msg.blsMsg = msgs[i]
		} else {
			m[i] = &multiKeySetMsg{
				step:      step,
				peeringID: msgs[i].PeeringID,
				receiver:  msgs[i].MsgReceiver,
				msgType:   msgs[i].MsgType,
				blsMsg:    msgs[i],
			}
		}
	}
}

//	type PriShare struct {
//		I int          // Index of the private share
//		V kyber.Scalar // Value of the private share
//	}
func writePriShare(ww *rwutil.Writer, val *share.PriShare) {
	ww.WriteBool(val != nil)
	if val != nil {
		ww.WriteUint32(uint32(val.I))
		ww.WriteMarshaled(val.V)
	}
}

func readPriShare(rr *rwutil.Reader) (ret *share.PriShare) {
	hasPriShare := rr.ReadBool()
	if hasPriShare {
		ret = new(share.PriShare)
		ret.I = int(rr.ReadUint32())
		rr.ReadMarshaled(ret.V)
	}
	return ret
}

//	type rabin_vvs.Deal struct {
//		SessionID []byte			// Unique session identifier for this protocol run
//		SecShare *share.PriShare	// Private share generated by the dealer
//		RndShare *share.PriShare	// Random share generated by the dealer
//		T uint32					// Threshold used for this secret sharing run
//		Commitments []kyber.Point	// Commitments are the coefficients used to verify the shares against
//	}
func writeVssDeal(ww *rwutil.Writer, d *rabin_vss.Deal) {
	ww.WriteBytes(d.SessionID)
	writePriShare(ww, d.SecShare)
	writePriShare(ww, d.RndShare)
	ww.WriteUint32(d.T)
	ww.WriteSize(len(d.Commitments))
	for i := range d.Commitments {
		ww.WriteMarshaled(d.Commitments[i])
	}
}

func readVssDeal(rr *rwutil.Reader, blsSuite kyber.Group) (ret *rabin_vss.Deal) {
	ret = new(rabin_vss.Deal)
	ret.SessionID = rr.ReadBytes()
	ret.SecShare = readPriShare(rr)
	ret.RndShare = readPriShare(rr)
	ret.T = rr.ReadUint32()
	size := rr.ReadSize()
	ret.Commitments = make([]kyber.Point, size)
	for i := range ret.Commitments {
		ret.Commitments[i] = blsSuite.Point()
		rr.ReadMarshaled(ret.Commitments[i])
	}
	return ret
}
