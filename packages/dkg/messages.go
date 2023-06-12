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
	"github.com/iotaledger/wasp/packages/isc"
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

func (msg *initiatorInitMsg) MsgType() byte {
	return initiatorInitMsgType
}

func (msg *initiatorInitMsg) Step() byte {
	return msg.step
}

func (msg *initiatorInitMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *initiatorInitMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	ww.WriteString(msg.dkgRef)
	ww.WriteN(msg.peeringID[:])

	ww.WriteSize(len(msg.peerPubs))
	for i := range msg.peerPubs {
		ww.WriteBytes(msg.peerPubs[i].AsBytes())
	}

	ww.WriteBytes(msg.initiatorPub.AsBytes())
	ww.WriteUint16(msg.threshold)
	ww.WriteDuration(msg.timeout)
	ww.WriteDuration(msg.roundRetry)
	return ww.Err
}

func (msg *initiatorInitMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	msg.dkgRef = rr.ReadString()
	rr.ReadN(msg.peeringID[:])

	size := rr.ReadSize()
	msg.peerPubs = make([]*cryptolib.PublicKey, size)
	for i := range msg.peerPubs {
		msg.peerPubs[i] = rwutil.ReadFromBytes(rr, cryptolib.PublicKeyFromBytes)
	}

	msg.initiatorPub = rwutil.ReadFromBytes(rr, cryptolib.PublicKeyFromBytes)
	msg.threshold = rr.ReadUint16()
	msg.timeout = rr.ReadDuration()
	msg.roundRetry = rr.ReadDuration()
	return rr.Err
}

func (msg *initiatorInitMsg) fromBytes(buf []byte) error {
	r := bytes.NewReader(buf)
	return msg.Read(r)
}

func (msg *initiatorInitMsg) Error() error {
	return nil
}

func (msg *initiatorInitMsg) IsResponse() bool {
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

func (msg *initiatorStepMsg) MsgType() byte {
	return initiatorStepMsgType
}

func (msg *initiatorStepMsg) Step() byte {
	return msg.step
}

func (msg *initiatorStepMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *initiatorStepMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	return ww.Err
}

func (msg *initiatorStepMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	return rr.Err
}

func (msg *initiatorStepMsg) fromBytes(buf []byte) error {
	r := bytes.NewReader(buf)
	return msg.Read(r)
}

func (msg *initiatorStepMsg) Error() error {
	return nil
}

func (msg *initiatorStepMsg) IsResponse() bool {
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

func (msg *initiatorDoneMsg) MsgType() byte {
	return initiatorDoneMsgType
}

func (msg *initiatorDoneMsg) Step() byte {
	return msg.step
}

func (msg *initiatorDoneMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *initiatorDoneMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)

	ww.WriteSize(len(msg.edPubShares))
	for i := range msg.edPubShares {
		ww.WriteMarshaled(msg.edPubShares[i])
	}

	ww.WriteSize(len(msg.blsPubShares))
	for i := range msg.blsPubShares {
		ww.WriteMarshaled(msg.blsPubShares[i])
	}
	return ww.Err
}

func (msg *initiatorDoneMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()

	size := rr.ReadSize()
	msg.edPubShares = make([]kyber.Point, size)
	for i := range msg.edPubShares {
		msg.edPubShares[i] = msg.edSuite.Point()
		rr.ReadMarshaled(msg.edPubShares[i])
	}

	size = rr.ReadSize()
	msg.blsPubShares = make([]kyber.Point, size)
	for i := range msg.blsPubShares {
		msg.blsPubShares[i] = msg.blsSuite.Point()
		rr.ReadMarshaled(msg.blsPubShares[i])
	}
	return rr.Err
}

func (msg *initiatorDoneMsg) fromBytes(buf []byte, edSuite, blsSuite kyber.Group) error {
	r := bytes.NewReader(buf)
	msg.edSuite = edSuite
	msg.blsSuite = blsSuite
	return msg.Read(r)
}

func (msg *initiatorDoneMsg) Error() error {
	return nil
}

func (msg *initiatorDoneMsg) IsResponse() bool {
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

func (msg *initiatorPubShareMsg) MsgType() byte {
	return initiatorPubShareMsgType
}

func (msg *initiatorPubShareMsg) Step() byte {
	return msg.step
}

func (msg *initiatorPubShareMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *initiatorPubShareMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	isc.AddressToWriter(ww, msg.sharedAddress)

	ww.WriteMarshaled(msg.edSharedPublic)
	ww.WriteMarshaled(msg.edPublicShare)
	ww.WriteBytes(msg.edSignature)

	ww.WriteMarshaled(msg.blsSharedPublic)
	ww.WriteMarshaled(msg.blsPublicShare)
	ww.WriteBytes(msg.blsSignature)
	return ww.Err
}

func (msg *initiatorPubShareMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	msg.sharedAddress = isc.AddressFromReader(rr)

	msg.edSharedPublic = msg.edSuite.Point()
	rr.ReadMarshaled(msg.edSharedPublic)
	msg.edPublicShare = msg.edSuite.Point()
	rr.ReadMarshaled(msg.edPublicShare)
	msg.edSignature = rr.ReadBytes()

	msg.blsSharedPublic = msg.blsSuite.Point()
	rr.ReadMarshaled(msg.blsSharedPublic)
	msg.blsPublicShare = msg.blsSuite.Point()
	rr.ReadMarshaled(msg.blsPublicShare)
	msg.blsSignature = rr.ReadBytes()
	return rr.Err
}

func (msg *initiatorPubShareMsg) fromBytes(buf []byte, edSuite, blsSuite kyber.Group) error {
	r := bytes.NewReader(buf)
	msg.edSuite = edSuite
	msg.blsSuite = blsSuite
	return msg.Read(r)
}

func (msg *initiatorPubShareMsg) Error() error {
	return nil
}

func (msg *initiatorPubShareMsg) IsResponse() bool {
	return true
}

// initiatorStatusMsg
type initiatorStatusMsg struct {
	step  byte
	error error
}

func (msg *initiatorStatusMsg) MsgType() byte {
	return initiatorStatusMsgType
}

func (msg *initiatorStatusMsg) Step() byte {
	return msg.step
}

func (msg *initiatorStatusMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *initiatorStatusMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	var errMsg string
	if msg.error != nil {
		errMsg = msg.error.Error()
	}
	ww.WriteString(errMsg)
	return ww.Err
}

func (msg *initiatorStatusMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	errMsg := rr.ReadString()
	msg.error = nil
	if errMsg != "" {
		msg.error = errors.New(errMsg)
	}
	return rr.Err
}

func (msg *initiatorStatusMsg) fromBytes(buf []byte) error {
	r := bytes.NewReader(buf)
	return msg.Read(r)
}

func (msg *initiatorStatusMsg) Error() error {
	return msg.error
}

func (msg *initiatorStatusMsg) IsResponse() bool {
	return true
}

// rabin_dkg.Deal
type rabinDealMsg struct {
	step byte
	deal *rabin_dkg.Deal
}

func (msg *rabinDealMsg) MsgType() byte {
	return rabinDealMsgType
}

func (msg *rabinDealMsg) Step() byte {
	return msg.step
}

func (msg *rabinDealMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *rabinDealMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	ww.WriteUint32(msg.deal.Index)
	ww.WriteMarshaled(msg.deal.Deal.DHKey)
	ww.WriteBytes(msg.deal.Deal.Signature)
	ww.WriteBytes(msg.deal.Deal.Nonce)
	ww.WriteBytes(msg.deal.Deal.Cipher)
	return ww.Err
}

func (msg *rabinDealMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	msg.deal.Index = rr.ReadUint32()
	rr.ReadMarshaled(msg.deal.Deal.DHKey)
	msg.deal.Deal.Signature = rr.ReadBytes()
	msg.deal.Deal.Nonce = rr.ReadBytes()
	msg.deal.Deal.Cipher = rr.ReadBytes()
	return rr.Err
}

func (msg *rabinDealMsg) fromBytes(buf []byte, edSuite kyber.Group) error {
	msg.deal = &rabin_dkg.Deal{
		Deal: &rabin_vss.EncryptedDeal{
			DHKey: edSuite.Point(),
		},
	}
	rdr := bytes.NewReader(buf)
	return msg.Read(rdr)
}

// rabin_dkg.Response
type rabinResponseMsg struct {
	step      byte
	responses []*rabin_dkg.Response
}

func (msg *rabinResponseMsg) MsgType() byte {
	return rabinResponseMsgType
}

func (msg *rabinResponseMsg) Step() byte {
	return msg.step
}

func (msg *rabinResponseMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *rabinResponseMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	ww.WriteSize(len(msg.responses))
	for _, response := range msg.responses {
		ww.WriteUint32(response.Index)
		ww.WriteBytes(response.Response.SessionID)
		ww.WriteUint32(response.Response.Index)
		ww.WriteBool(response.Response.Approved)
		ww.WriteBytes(response.Response.Signature)
	}
	return ww.Err
}

func (msg *rabinResponseMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	size := rr.ReadSize()
	msg.responses = make([]*rabin_dkg.Response, size)
	for i := range msg.responses {
		response := rabin_dkg.Response{
			Response: &rabin_vss.Response{},
		}
		msg.responses[i] = &response
		response.Index = rr.ReadUint32()
		response.Response.SessionID = rr.ReadBytes()
		response.Response.Index = rr.ReadUint32()
		response.Response.Approved = rr.ReadBool()
		response.Response.Signature = rr.ReadBytes()
	}
	return rr.Err
}

func (msg *rabinResponseMsg) fromBytes(buf []byte) error {
	rdr := bytes.NewReader(buf)
	return msg.Read(rdr)
}

// rabin_dkg.Justification
type rabinJustificationMsg struct {
	step           byte
	justifications []*rabin_dkg.Justification
	blsSuite       kyber.Group // Just for un-marshaling.
}

func (msg *rabinJustificationMsg) MsgType() byte {
	return rabinJustificationMsgType
}

func (msg *rabinJustificationMsg) Step() byte {
	return msg.step
}

func (msg *rabinJustificationMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *rabinJustificationMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	ww.WriteSize(len(msg.justifications))
	for _, j := range msg.justifications {
		ww.WriteUint32(j.Index)
		ww.WriteBytes(j.Justification.SessionID)
		ww.WriteUint32(j.Justification.Index)
		writeVssDeal(ww, j.Justification.Deal)
		ww.WriteBytes(j.Justification.Signature)
	}
	return ww.Err
}

func (msg *rabinJustificationMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	size := rr.ReadSize()
	msg.justifications = make([]*rabin_dkg.Justification, size)
	for i := range msg.justifications {
		j := &rabin_dkg.Justification{
			Justification: &rabin_vss.Justification{},
		}
		msg.justifications[i] = j
		j.Index = rr.ReadUint32()
		j.Justification.SessionID = rr.ReadBytes()
		j.Justification.Index = rr.ReadUint32()
		j.Justification.Deal = readVssDeal(rr, msg.blsSuite)
		j.Justification.Signature = rr.ReadBytes()
	}
	return rr.Err
}

func (msg *rabinJustificationMsg) fromBytes(buf []byte, blsSuite kyber.Group) error {
	msg.blsSuite = blsSuite
	rdr := bytes.NewReader(buf)
	return msg.Read(rdr)
}

// rabin_dkg.SecretCommits
type rabinSecretCommitsMsg struct {
	step          byte
	secretCommits *rabin_dkg.SecretCommits
	blsSuite      kyber.Group // Just for un-marshaling.
}

func (msg *rabinSecretCommitsMsg) MsgType() byte {
	return rabinSecretCommitsMsgType
}

func (msg *rabinSecretCommitsMsg) Step() byte {
	return msg.step
}

func (msg *rabinSecretCommitsMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *rabinSecretCommitsMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	ww.WriteBool(msg.secretCommits == nil)
	if msg.secretCommits == nil {
		return ww.Err
	}

	ww.WriteUint32(msg.secretCommits.Index)

	ww.WriteSize(len(msg.secretCommits.Commitments))
	for i := range msg.secretCommits.Commitments {
		ww.WriteMarshaled(msg.secretCommits.Commitments[i])
	}

	ww.WriteBytes(msg.secretCommits.SessionID)
	ww.WriteBytes(msg.secretCommits.Signature)
	return ww.Err
}

func (msg *rabinSecretCommitsMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	isNil := rr.ReadBool()
	if isNil {
		msg.secretCommits = nil
		return rr.Err
	}

	msg.secretCommits = &rabin_dkg.SecretCommits{}
	msg.secretCommits.Index = rr.ReadUint32()

	size := rr.ReadSize()
	msg.secretCommits.Commitments = make([]kyber.Point, size)
	for i := range msg.secretCommits.Commitments {
		msg.secretCommits.Commitments[i] = msg.blsSuite.Point()
		rr.ReadMarshaled(msg.secretCommits.Commitments[i])
	}

	msg.secretCommits.SessionID = rr.ReadBytes()
	msg.secretCommits.Signature = rr.ReadBytes()
	return rr.Err
}

func (msg *rabinSecretCommitsMsg) fromBytes(buf []byte, blsSuite kyber.Group) error {
	msg.blsSuite = blsSuite
	rdr := bytes.NewReader(buf)
	return msg.Read(rdr)
}

// rabin_dkg.ComplaintCommits
type rabinComplaintCommitsMsg struct {
	step             byte
	complaintCommits []*rabin_dkg.ComplaintCommits
	blsSuite         kyber.Group // Just for un-marshaling.
}

func (msg *rabinComplaintCommitsMsg) MsgType() byte {
	return rabinComplaintCommitsMsgType
}

func (msg *rabinComplaintCommitsMsg) Step() byte {
	return msg.step
}

func (msg *rabinComplaintCommitsMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *rabinComplaintCommitsMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	ww.WriteSize(len(msg.complaintCommits))
	for i := range msg.complaintCommits {
		ww.WriteUint32(msg.complaintCommits[i].Index)
		ww.WriteUint32(msg.complaintCommits[i].DealerIndex)
		writeVssDeal(ww, msg.complaintCommits[i].Deal)
		ww.WriteBytes(msg.complaintCommits[i].Signature)
	}
	return ww.Err
}

func (msg *rabinComplaintCommitsMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	size := rr.ReadSize()
	msg.complaintCommits = make([]*rabin_dkg.ComplaintCommits, size)
	for i := range msg.complaintCommits {
		msg.complaintCommits[i] = &rabin_dkg.ComplaintCommits{}
		msg.complaintCommits[i].Index = rr.ReadUint32()
		msg.complaintCommits[i].DealerIndex = rr.ReadUint32()
		msg.complaintCommits[i].Deal = readVssDeal(rr, msg.blsSuite)
		msg.complaintCommits[i].Signature = rr.ReadBytes()
	}
	return rr.Err
}

func (msg *rabinComplaintCommitsMsg) fromBytes(buf []byte, blsSuite kyber.Group) error {
	msg.blsSuite = blsSuite
	rdr := bytes.NewReader(buf)
	return msg.Read(rdr)
}

// rabin_dkg.ReconstructCommits
type rabinReconstructCommitsMsg struct {
	step               byte
	reconstructCommits []*rabin_dkg.ReconstructCommits
}

func (msg *rabinReconstructCommitsMsg) MsgType() byte {
	return rabinReconstructCommitsMsgType
}

func (msg *rabinReconstructCommitsMsg) Step() byte {
	return msg.step
}

func (msg *rabinReconstructCommitsMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *rabinReconstructCommitsMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	ww.WriteSize(len(msg.reconstructCommits))
	for i := range msg.reconstructCommits {
		ww.WriteBytes(msg.reconstructCommits[i].SessionID)
		ww.WriteUint32(msg.reconstructCommits[i].Index)
		ww.WriteUint32(msg.reconstructCommits[i].DealerIndex)
		writePriShare(ww, msg.reconstructCommits[i].Share)
		ww.WriteBytes(msg.reconstructCommits[i].Signature)
	}
	return ww.Err
}

func (msg *rabinReconstructCommitsMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	size := rr.ReadSize()
	msg.reconstructCommits = make([]*rabin_dkg.ReconstructCommits, size)
	for i := range msg.reconstructCommits {
		msg.reconstructCommits[i] = &rabin_dkg.ReconstructCommits{}
		msg.reconstructCommits[i].SessionID = rr.ReadBytes()
		msg.reconstructCommits[i].Index = rr.ReadUint32()
		msg.reconstructCommits[i].DealerIndex = rr.ReadUint32()
		msg.reconstructCommits[i].Share = readPriShare(rr)
		msg.reconstructCommits[i].Signature = rr.ReadBytes()
	}
	return rr.Err
}

func (msg *rabinReconstructCommitsMsg) fromBytes(buf []byte) error {
	rdr := bytes.NewReader(buf)
	return msg.Read(rdr)
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

func (msg *multiKeySetMsg) MsgType() byte {
	return msg.msgType
}

func (msg *multiKeySetMsg) Step() byte {
	return msg.step
}

func (msg *multiKeySetMsg) SetStep(step byte) {
	msg.step = step
}

func (msg *multiKeySetMsg) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(msg.step)
	ww.WriteBytes(msg.edMsg.MsgData)
	ww.WriteBytes(msg.blsMsg.MsgData)
	return ww.Err
}

func (msg *multiKeySetMsg) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.step = rr.ReadByte()
	msg.edMsg = &peering.PeerMessageData{
		PeeringID:   msg.peeringID,
		MsgReceiver: msg.receiver,
		MsgType:     msg.msgType,
		MsgData:     rr.ReadBytes(),
	}
	msg.blsMsg = &peering.PeerMessageData{
		PeeringID:   msg.peeringID,
		MsgReceiver: msg.receiver,
		MsgType:     msg.msgType,
		MsgData:     rr.ReadBytes(),
	}
	return rr.Err
}

func (msg *multiKeySetMsg) fromBytes(buf []byte, peeringID peering.PeeringID, receiver, msgType byte) error {
	rdr := bytes.NewReader(buf)
	msg.peeringID = peeringID
	msg.receiver = receiver
	msg.msgType = msgType
	return msg.Read(rdr)
}

func (msg *multiKeySetMsg) mustDataBytes() []byte {
	w := new(bytes.Buffer)
	if err := msg.Write(w); err != nil {
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
