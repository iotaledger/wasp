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
	"fmt"
	"io"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	rabin_vss "go.dedis.ch/kyber/v3/share/vss/rabin"
	"golang.org/x/xerrors"
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

// Checks if that's a Initiator -> PeerNode message.
func isDkgInitNodeMsg(msgType byte) bool { //nolint:unused,deadcode
	return msgType == initiatorInitMsgType
}

// Checks if that's a Initiator <-> PeerProc message.
func isDkgInitProcMsg(msgType byte) bool { //nolint:unused,deadcode
	return initiatorMsgBase <= msgType && msgType < initiatorMsgFree
}

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

// Checks if that's a PeerProc <-> PeerProc message.
// func isDkgRabinRoundMsg(msgType byte) bool {
// 	return rabinMsgBase <= msgType && msgType < rabinMsgFree
// }

// Checks if that's a PeerProc <-> PeerProc echoed / repeated message.
// func isDkgRabinEchoMsg(msgType byte) bool {
// 	return rabinEcho <= msgType && msgType < rabinMsgFree-rabinMsgBase+rabinEcho
// }

// func makeDkgRoundEchoMsg(msgType byte) (byte, error) {
// 	if isDkgRabinRoundMsg(msgType) {
// 		return msgType - rabinMsgBase + rabinEcho, nil
// 	}
// 	if isDkgRabinEchoMsg(msgType) {
// 		return msgType, nil
// 	}
// 	return msgType, errors.New("round_msg_type_expected")
// }

// func makeDkgRoundMsg(msgType byte) (byte, error) { //nolint:unused,deadcode
// 	if isDkgRabinRoundMsg(msgType) {
// 		return msgType, nil
// 	}
// 	if isDkgRabinEchoMsg(msgType) {
// 		return msgType - rabinEcho + rabinMsgBase, nil
// 	}
// 	return msgType, errors.New("round_or_echo_msg_type_expected")
// }

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
		MsgData:     util.MustBytes(msg),
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

//nolint:gocritic
func (m *initiatorInitMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	if err = util.WriteString16(w, m.dkgRef); err != nil {
		return err
	}
	if _, err = w.Write(m.peeringID[:]); err != nil {
		return err
	}
	if err = util.WriteUint16(w, uint16(len(m.peerPubs))); err != nil {
		return err
	}
	for i := range m.peerPubs {
		if err = util.WriteBytes16(w, m.peerPubs[i].AsBytes()); err != nil {
			return err
		}
	}
	if err = util.WriteBytes16(w, m.initiatorPub.AsBytes()); err != nil {
		return err
	}
	if err = util.WriteUint16(w, m.threshold); err != nil {
		return err
	}
	if err = util.WriteInt64(w, m.timeout.Milliseconds()); err != nil {
		return err
	}
	return util.WriteInt64(w, m.roundRetry.Milliseconds())
}

//nolint:gocritic
func (m *initiatorInitMsg) Read(r io.Reader) error {
	var err error
	var n int
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	if m.dkgRef, err = util.ReadString16(r); err != nil {
		return err
	}
	if n, err = r.Read(m.peeringID[:]); err != nil {
		return err
	}
	if n != iotago.Ed25519AddressBytesLength {
		return fmt.Errorf("error while reading peering ID: read %v bytes, expected %v bytes",
			n, iotago.Ed25519AddressBytesLength)
	}
	var arrLen uint16
	if err = util.ReadUint16(r, &arrLen); err != nil {
		return err
	}
	m.peerPubs = make([]*cryptolib.PublicKey, arrLen)
	for i := range m.peerPubs {
		var peerPubBytes []byte
		if peerPubBytes, err = util.ReadBytes16(r); err != nil {
			return err
		}
		peerPubKey, err := cryptolib.NewPublicKeyFromBytes(peerPubBytes)
		if err != nil {
			return err
		}
		m.peerPubs[i] = peerPubKey
	}
	var initiatorPubBytes []byte
	if initiatorPubBytes, err = util.ReadBytes16(r); err != nil {
		return err
	}
	initiatorPub, err := cryptolib.NewPublicKeyFromBytes(initiatorPubBytes)
	if err != nil {
		return err
	}
	m.initiatorPub = initiatorPub
	if err = util.ReadUint16(r, &m.threshold); err != nil {
		return err
	}
	var timeoutMS int64
	if err = util.ReadInt64(r, &timeoutMS); err != nil {
		return err
	}
	m.timeout = time.Duration(timeoutMS) * time.Millisecond
	var roundRetryMS int64
	if err = util.ReadInt64(r, &roundRetryMS); err != nil {
		return err
	}
	m.roundRetry = time.Duration(roundRetryMS) * time.Millisecond
	return nil
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
	return util.WriteByte(w, m.step)
}

func (m *initiatorStepMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	return nil
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

//nolint:gocritic
func (m *initiatorDoneMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	if err = util.WriteUint16(w, uint16(len(m.edPubShares))); err != nil {
		return err
	}
	for i := range m.edPubShares {
		if err = util.WriteMarshaled(w, m.edPubShares[i]); err != nil {
			return err
		}
	}
	if err = util.WriteUint16(w, uint16(len(m.blsPubShares))); err != nil {
		return err
	}
	for i := range m.blsPubShares {
		if err = util.WriteMarshaled(w, m.blsPubShares[i]); err != nil {
			return err
		}
	}
	return nil
}

//nolint:gocritic
func (m *initiatorDoneMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	//
	// edPubShares
	var arrLen uint16
	if err = util.ReadUint16(r, &arrLen); err != nil {
		return err
	}
	m.edPubShares = make([]kyber.Point, arrLen)
	for i := range m.edPubShares {
		m.edPubShares[i] = m.edSuite.Point()
		if err = util.ReadMarshaled(r, m.edPubShares[i]); err != nil {
			return xerrors.Errorf("failed to unmarshal initiatorDoneMsg.edPubShares: %w", err)
		}
	}
	//
	// blsPubShares
	if err = util.ReadUint16(r, &arrLen); err != nil {
		return err
	}
	m.blsPubShares = make([]kyber.Point, arrLen)
	for i := range m.blsPubShares {
		m.blsPubShares[i] = m.blsSuite.Point()
		if err = util.ReadMarshaled(r, m.blsPubShares[i]); err != nil {
			return xerrors.Errorf("failed to unmarshal initiatorDoneMsg.blsPubShares: %w", err)
		}
	}
	return nil
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

//nolint:gocritic
func (m *initiatorPubShareMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, isc.BytesFromAddress(m.sharedAddress)); err != nil {
		return err
	}
	{ // Ed25519 part.
		if err = util.WriteMarshaled(w, m.edSharedPublic); err != nil {
			return err
		}
		if err = util.WriteMarshaled(w, m.edPublicShare); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, m.edSignature); err != nil {
			return err
		}
	}
	{ // BLS part.
		if err = util.WriteMarshaled(w, m.blsSharedPublic); err != nil {
			return err
		}
		if err = util.WriteMarshaled(w, m.blsPublicShare); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, m.blsSignature); err != nil {
			return err
		}
	}
	return nil
}

func (m *initiatorPubShareMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	//
	// SharedAddress.
	var sharedAddressBin []byte
	var sharedAddress iotago.Address
	if sharedAddressBin, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if sharedAddress, _, err = isc.AddressFromBytes(sharedAddressBin); err != nil {
		return err
	}
	m.sharedAddress = sharedAddress
	//
	// Ed25519 part.
	m.edSharedPublic = m.edSuite.Point()
	if err = util.ReadMarshaled(r, m.edSharedPublic); err != nil {
		return xerrors.Errorf("failed to unmarshal initiatorPubShareMsg.edSharedPublic: %w", err)
	}
	m.edPublicShare = m.edSuite.Point()
	if err = util.ReadMarshaled(r, m.edPublicShare); err != nil {
		return xerrors.Errorf("failed to unmarshal initiatorPubShareMsg.edPublicShare: %w", err)
	}
	if m.edSignature, err = util.ReadBytes16(r); err != nil {
		return err
	}
	//
	// BLS part.
	m.blsSharedPublic = m.blsSuite.Point()
	if err = util.ReadMarshaled(r, m.blsSharedPublic); err != nil {
		return xerrors.Errorf("failed to unmarshal initiatorPubShareMsg.blsSharedPublic: %w", err)
	}
	m.blsPublicShare = m.blsSuite.Point()
	if err = util.ReadMarshaled(r, m.blsPublicShare); err != nil {
		return xerrors.Errorf("failed to unmarshal initiatorPubShareMsg.blsPublicShare: %w", err)
	}
	if m.blsSignature, err = util.ReadBytes16(r); err != nil {
		return err
	}
	return nil
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
	if err := util.WriteByte(w, m.step); err != nil {
		return err
	}
	var errMsg string
	if m.error != nil {
		errMsg = m.error.Error()
	}
	return util.WriteString16(w, errMsg)
}

func (m *initiatorStatusMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	var errMsg string
	if errMsg, err = util.ReadString16(r); err != nil {
		return err
	}
	if errMsg != "" {
		m.error = errors.New(errMsg)
	} else {
		m.error = nil
	}
	return nil
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

//nolint:gocritic
func (m *rabinDealMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	if err = util.WriteUint32(w, m.deal.Index); err != nil {
		return err
	}
	if err = util.WriteMarshaled(w, m.deal.Deal.DHKey); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, m.deal.Deal.Signature); err != nil {
		return err
	}
	if err = util.WriteBytes16(w, m.deal.Deal.Nonce); err != nil {
		return err
	}
	return util.WriteBytes16(w, m.deal.Deal.Cipher)
}

//nolint:gocritic
func (m *rabinDealMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	if err = util.ReadUint32(r, &m.deal.Index); err != nil {
		return err
	}
	if err = util.ReadMarshaled(r, m.deal.Deal.DHKey); err != nil {
		return err
	}
	if m.deal.Deal.Signature, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if m.deal.Deal.Nonce, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if m.deal.Deal.Cipher, err = util.ReadBytes16(r); err != nil {
		return err
	}
	return nil
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

//nolint:gocritic
func (m *rabinResponseMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	listLen := uint32(len(m.responses))
	if err = util.WriteUint32(w, listLen); err != nil {
		return err
	}
	for _, r := range m.responses {
		if err = util.WriteUint32(w, r.Index); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, r.Response.SessionID); err != nil {
			return err
		}
		if err = util.WriteUint32(w, r.Response.Index); err != nil {
			return err
		}
		if err = util.WriteBoolByte(w, r.Response.Approved); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, r.Response.Signature); err != nil {
			return err
		}
	}
	return nil
}

//nolint:gocritic
func (m *rabinResponseMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	var listLen uint32
	if err = util.ReadUint32(r, &listLen); err != nil {
		return err
	}
	m.responses = make([]*rabin_dkg.Response, int(listLen))
	for i := range m.responses {
		response := rabin_dkg.Response{
			Response: &rabin_vss.Response{},
		}
		m.responses[i] = &response
		if err = util.ReadUint32(r, &response.Index); err != nil {
			return err
		}
		if response.Response.SessionID, err = util.ReadBytes16(r); err != nil {
			return err
		}
		if err = util.ReadUint32(r, &response.Response.Index); err != nil {
			return err
		}
		if err = util.ReadBoolByte(r, &response.Response.Approved); err != nil {
			return err
		}
		if response.Response.Signature, err = util.ReadBytes16(r); err != nil {
			return err
		}
	}
	return nil
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

//nolint:gocritic
func (m *rabinJustificationMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	jLen := uint32(len(m.justifications))
	if err = util.WriteUint32(w, jLen); err != nil {
		return err
	}
	for _, j := range m.justifications {
		if err = util.WriteUint32(w, j.Index); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, j.Justification.SessionID); err != nil {
			return err
		}
		if err = util.WriteUint32(w, j.Justification.Index); err != nil {
			return err
		}
		if err = writeVssDeal(w, j.Justification.Deal); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, j.Justification.Signature); err != nil {
			return err
		}
	}
	return nil
}

//nolint:gocritic
func (m *rabinJustificationMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	var jLen uint32
	if err = util.ReadUint32(r, &jLen); err != nil {
		return err
	}
	m.justifications = make([]*rabin_dkg.Justification, int(jLen))
	for i := range m.justifications {
		j := rabin_dkg.Justification{
			Justification: &rabin_vss.Justification{},
		}
		m.justifications[i] = &j
		if err = util.ReadUint32(r, &j.Index); err != nil {
			return err
		}
		if j.Justification.SessionID, err = util.ReadBytes16(r); err != nil {
			return err
		}
		if err = util.ReadUint32(r, &j.Justification.Index); err != nil {
			return err
		}
		if err = readVssDeal(r, &j.Justification.Deal, m.blsSuite); err != nil {
			return err
		}
		if j.Justification.Signature, err = util.ReadBytes16(r); err != nil {
			return err
		}
	}
	return nil
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

//nolint:gocritic
func (m *rabinSecretCommitsMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	if err = util.WriteBoolByte(w, m.secretCommits == nil); err != nil {
		return err
	}
	if m.secretCommits == nil {
		return nil
	}
	if err = util.WriteUint32(w, m.secretCommits.Index); err != nil {
		return err
	}
	if err = util.WriteUint32(w, uint32(len(m.secretCommits.Commitments))); err != nil {
		return err
	}
	for i := range m.secretCommits.Commitments {
		if err = util.WriteMarshaled(w, m.secretCommits.Commitments[i]); err != nil {
			return err
		}
	}
	if err = util.WriteBytes16(w, m.secretCommits.SessionID); err != nil {
		return err
	}
	return util.WriteBytes16(w, m.secretCommits.Signature)
}

//nolint:gocritic
func (m *rabinSecretCommitsMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	var isNil bool
	if err = util.ReadBoolByte(r, &isNil); err != nil {
		return err
	}
	if isNil {
		m.secretCommits = nil
		return nil
	}
	m.secretCommits = &rabin_dkg.SecretCommits{}
	if err = util.ReadUint32(r, &m.secretCommits.Index); err != nil {
		return err
	}
	var cLen uint32
	if err = util.ReadUint32(r, &cLen); err != nil {
		return err
	}
	m.secretCommits.Commitments = make([]kyber.Point, cLen)
	for i := range m.secretCommits.Commitments {
		m.secretCommits.Commitments[i] = m.blsSuite.Point()
		if err = util.ReadMarshaled(r, m.secretCommits.Commitments[i]); err != nil {
			return err
		}
	}
	if m.secretCommits.SessionID, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if m.secretCommits.Signature, err = util.ReadBytes16(r); err != nil {
		return err
	}
	return nil
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

//nolint:gocritic
func (m *rabinComplaintCommitsMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	if err = util.WriteUint32(w, uint32(len(m.complaintCommits))); err != nil {
		return err
	}
	for i := range m.complaintCommits {
		if err = util.WriteUint32(w, m.complaintCommits[i].Index); err != nil {
			return err
		}
		if err = util.WriteUint32(w, m.complaintCommits[i].DealerIndex); err != nil {
			return err
		}
		if err = writeVssDeal(w, m.complaintCommits[i].Deal); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, m.complaintCommits[i].Signature); err != nil {
			return err
		}
	}
	return nil
}

//nolint:gocritic
func (m *rabinComplaintCommitsMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	var ccLen uint32
	if err = util.ReadUint32(r, &ccLen); err != nil {
		return err
	}
	m.complaintCommits = make([]*rabin_dkg.ComplaintCommits, ccLen)
	for i := range m.complaintCommits {
		m.complaintCommits[i] = &rabin_dkg.ComplaintCommits{}
		if err = util.ReadUint32(r, &m.complaintCommits[i].Index); err != nil {
			return err
		}
		if err = util.ReadUint32(r, &m.complaintCommits[i].DealerIndex); err != nil {
			return err
		}
		if err = readVssDeal(r, &m.complaintCommits[i].Deal, m.blsSuite); err != nil {
			return err
		}
		if m.complaintCommits[i].Signature, err = util.ReadBytes16(r); err != nil {
			return err
		}
	}
	return nil
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

//nolint:gocritic
func (m *rabinReconstructCommitsMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteByte(w, m.step); err != nil {
		return err
	}
	if err = util.WriteUint32(w, uint32(len(m.reconstructCommits))); err != nil {
		return err
	}
	for i := range m.reconstructCommits {
		if err = util.WriteBytes16(w, m.reconstructCommits[i].SessionID); err != nil {
			return err
		}
		if err = util.WriteUint32(w, m.reconstructCommits[i].Index); err != nil {
			return err
		}
		if err = util.WriteUint32(w, m.reconstructCommits[i].DealerIndex); err != nil {
			return err
		}
		if err = writePriShare(w, m.reconstructCommits[i].Share); err != nil {
			return err
		}
		if err = util.WriteBytes16(w, m.reconstructCommits[i].Signature); err != nil {
			return err
		}
	}
	return nil
}

//nolint:gocritic
func (m *rabinReconstructCommitsMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	var ccLen uint32
	if err = util.ReadUint32(r, &ccLen); err != nil {
		return err
	}
	m.reconstructCommits = make([]*rabin_dkg.ReconstructCommits, ccLen)
	for i := range m.reconstructCommits {
		m.reconstructCommits[i] = &rabin_dkg.ReconstructCommits{}
		if m.reconstructCommits[i].SessionID, err = util.ReadBytes16(r); err != nil {
			return err
		}
		if err = util.ReadUint32(r, &m.reconstructCommits[i].Index); err != nil {
			return err
		}
		if err = util.ReadUint32(r, &m.reconstructCommits[i].DealerIndex); err != nil {
			return err
		}
		if err = readPriShare(r, &m.reconstructCommits[i].Share); err != nil {
			return err
		}
		if m.reconstructCommits[i].Signature, err = util.ReadBytes16(r); err != nil {
			return err
		}
	}
	return nil
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
	if err := util.WriteByte(w, m.step); err != nil {
		return err
	}
	if err := util.WriteBytes16(w, m.edMsg.MsgData); err != nil {
		return err
	}
	if err := util.WriteBytes16(w, m.blsMsg.MsgData); err != nil {
		return err
	}
	return nil
}

func (m *multiKeySetMsg) Read(r io.Reader) error {
	var err error
	if m.step, err = util.ReadByte(r); err != nil {
		return err
	}
	m.edMsg = &peering.PeerMessageData{
		PeeringID:   m.peeringID,
		MsgReceiver: m.receiver,
		MsgType:     m.msgType,
		MsgData:     nil, // Assigned below.
	}
	if m.edMsg.MsgData, err = util.ReadBytes16(r); err != nil {
		return err
	}
	m.blsMsg = &peering.PeerMessageData{
		PeeringID:   m.peeringID,
		MsgReceiver: m.receiver,
		MsgType:     m.msgType,
		MsgData:     nil, // Assigned below.
	}
	if m.blsMsg.MsgData, err = util.ReadBytes16(r); err != nil {
		return err
	}
	return nil
}

func (m *multiKeySetMsg) fromBytes(buf []byte, peeringID peering.PeeringID, receiver, msgType byte) error {
	rdr := bytes.NewReader(buf)
	m.peeringID = peeringID
	m.receiver = receiver
	m.msgType = msgType
	return m.Read(rdr)
}

func (m *multiKeySetMsg) mustDataBytes() []byte {
	buf := bytes.Buffer{}
	if err := m.Write(&buf); err != nil {
		panic(err)
	}
	return buf.Bytes()
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
//
//nolint:gocritic
func writePriShare(w io.Writer, val *share.PriShare) error {
	var err error
	if err = util.WriteBoolByte(w, val == nil); err != nil {
		return err
	}
	if val == nil {
		return nil
	}
	if err = util.WriteUint32(w, uint32(val.I)); err != nil {
		return err
	}
	return util.WriteMarshaled(w, val.V)
}

//nolint:gocritic
func readPriShare(r io.Reader, val **share.PriShare) error {
	var err error
	var valNil bool
	if err = util.ReadBoolByte(r, &valNil); err != nil {
		return err
	}
	if valNil {
		*val = nil
	}
	var i uint32
	if err = util.ReadUint32(r, &i); err != nil {
		return err
	}
	(*val).I = int(i)
	return util.ReadMarshaled(r, (*val).V)
}

//	type rabin_vvs.Deal struct {
//		SessionID []byte			// Unique session identifier for this protocol run
//		SecShare *share.PriShare	// Private share generated by the dealer
//		RndShare *share.PriShare	// Random share generated by the dealer
//		T uint32					// Threshold used for this secret sharing run
//		Commitments []kyber.Point	// Commitments are the coefficients used to verify the shares against
//	}
//
//nolint:gocritic
func writeVssDeal(w io.Writer, d *rabin_vss.Deal) error {
	var err error
	if err = util.WriteBytes16(w, d.SessionID); err != nil {
		return err
	}
	if err = writePriShare(w, d.SecShare); err != nil {
		return err
	}
	if err = writePriShare(w, d.RndShare); err != nil {
		return err
	}
	if err = util.WriteUint32(w, d.T); err != nil {
		return err
	}
	if err = util.WriteUint32(w, uint32(len(d.Commitments))); err != nil {
		return err
	}
	for i := range d.Commitments {
		if err = util.WriteMarshaled(w, d.Commitments[i]); err != nil {
			return err
		}
	}
	return nil
}

func readVssDeal(r io.Reader, d **rabin_vss.Deal, blsSuite kyber.Group) error {
	var err error
	dd := rabin_vss.Deal{}
	if dd.SessionID, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if err := readPriShare(r, &dd.SecShare); err != nil {
		return err
	}
	if err := readPriShare(r, &dd.RndShare); err != nil {
		return err
	}
	if err := util.ReadUint32(r, &dd.T); err != nil {
		return err
	}
	var commitmentCount uint32
	if err := util.ReadUint32(r, &commitmentCount); err != nil {
		return err
	}
	dd.Commitments = make([]kyber.Point, int(commitmentCount))
	for i := range dd.Commitments {
		dd.Commitments[i] = blsSuite.Point()
		if err := util.ReadMarshaled(r, dd.Commitments[i]); err != nil {
			return err
		}
	}
	*d = &dd
	return nil
}
