// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package commoncoin implements a common coin abstraction needed by
// the HoneyBadgerBFT for synchronization and randomness.
//
// See A. Miller et al. The honey badger of bft protocols. 2016.
// <https://eprint.iacr.org/2016/199.pdf>, Appendix C.
package commoncoin

import (
	"bytes"
	"io"
	"time"

	"github.com/iotaledger/hive.go/crypto/bls"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

const (
	commonCoinShareMsgType = 50 + peering.FirstUserMsgCode
	retryTimeout           = 1 * time.Second
	giveUpTimeout          = 1 * time.Hour
)

// Type CommonCoinProvider is used decouple implementation of the coin from the
// implementation of the consensus algorithm using it.
type Provider interface {
	GetCoin(sid []byte) ([]byte, error)
	io.Closer
}

// region commonCoinNode ///////////////////////////////////////////////////////

// This type represents a node, that is responsible for participate in
// in the process of producing common coins. It acts as a server and as
// a client at the same time.
//
// All the operations are performed in a single thread here. The
// client and network peers communicates with it via channels only.
type commonCoinNode struct {
	coins     map[string]*commonCoin
	dkShare   *tcrypto.DKShare
	peeringID peering.PeeringID
	group     peering.GroupProvider
	attachID  interface{}             // Can be nil, if channel was passed via constructor.
	recvCh    chan *peering.RecvEvent // We are receiving the network messages via this.
	coinCh    chan *commonCoinReq     // We are getting the local coin requests via this.
	stopCh    chan bool
	log       *logger.Logger
}

// NewCommonCoinNode creates new common coin generator.
// It is expected than one such node will exist for each chain.
//
// The parameter recvCh is optional here. If not provided, this
// node will subscribe to the PeeringGroup to listen for messages.
func NewCommonCoinNode(
	recvCh chan *peering.RecvEvent,
	dkShare *tcrypto.DKShare,
	peeringID peering.PeeringID,
	group peering.GroupProvider,
	log *logger.Logger,
) Provider {
	var ccn = commonCoinNode{
		coins:     make(map[string]*commonCoin),
		dkShare:   dkShare,
		peeringID: peeringID,
		group:     group,
		coinCh:    make(chan *commonCoinReq),
		stopCh:    make(chan bool),
		log:       log,
	}
	if recvCh == nil {
		ccn.recvCh = make(chan *peering.RecvEvent)
		ccn.attachID = group.Attach(&peeringID, func(recvEvent *peering.RecvEvent) {
			if recvEvent.Msg.MsgType == commonCoinShareMsgType {
				ccn.recvCh <- recvEvent
			}
		})
	} else {
		ccn.recvCh = recvCh
	}
	go ccn.recvLoop()
	return &ccn
}

// GetCoin returns an instance of a common coin.
// It can block for some time, if other peers are lagging.
func (ccn *commonCoinNode) GetCoin(sid []byte) ([]byte, error) {
	waitCh := make(chan []byte)
	ccn.coinCh <- &commonCoinReq{sid: sid, waitCh: waitCh}
	if coin, ok := <-waitCh; ok {
		return coin, nil
	}
	return nil, xerrors.New("timeout waiting for a common coin")
}

// Close terminates the CommonCoin peer.
func (ccn *commonCoinNode) Close() error {
	if ccn.attachID != nil {
		ccn.group.Detach(ccn.attachID)
		close(ccn.recvCh)
	}
	close(ccn.coinCh)
	close(ccn.stopCh)
	return nil
}

func (ccn *commonCoinNode) recvLoop() {
	timerCh := time.After(retryTimeout)
loop:
	for {
		select {
		case <-ccn.stopCh:
			break loop
		case cReq, ok := <-ccn.coinCh:
			if ok {
				ccn.onCoinReq(cReq)
			}
		case recv, ok := <-ccn.recvCh:
			if ok {
				ccn.onCoinShare(recv)
			}
		case <-timerCh:
			ccn.onTimerTick()
			timerCh = time.After(retryTimeout)
		}
	}
	// Cancel all the waiting clients.
	for _, c := range ccn.coins {
		if c.waitCh != nil {
			close(c.waitCh)
			c.waitCh = nil
		}
	}
}

// onCoinReq handles the requests from the local users.
func (ccn *commonCoinNode) onCoinReq(req *commonCoinReq) {
	cc := ccn.getCoinObj(req.sid)
	cc.getCoin(req.waitCh)
}

// onCoinShare handles the coin share received from our peers.
func (ccn *commonCoinNode) onCoinShare(recvEvent *peering.RecvEvent) {
	var err error

	msg := commonCoinMsg{}
	if err = msg.FromBytes(recvEvent.Msg.MsgData); err != nil {
		ccn.log.Errorf("Unable to parse a commonCoinMsg, error=%+v", err)
		return
	}
	var sigShare tbdn.SigShare = msg.coinShare
	if err = ccn.dkShare.VerifySigShare(msg.sid, sigShare); err != nil {
		ccn.log.Errorf("Invalid signature share in %+v, error=%+v", msg, err)
		return
	}
	var peerIndex int
	if peerIndex, err = sigShare.Index(); err != nil {
		ccn.log.Errorf("Invalid peerIndex in %+v, error=%+v", sigShare, err)
		return
	}
	var cc = ccn.getCoinObj(msg.sid)
	var reconstructed bool
	reconstructed, err = cc.acceptCoinShare(uint16(peerIndex), msg.coinShare, msg.needReply, recvEvent.From)
	if err != nil && err.Error() != "threshold not reached" && !reconstructed {
		ccn.log.Debugf("The coin is not revealed yet, reason=%+v", err)
	}
}

// onTimerTick performs periodic tasks: cleanup, message resends.
func (ccn *commonCoinNode) onTimerTick() {
	for sid, cc := range ccn.coins {
		if !cc.onTimerTick() {
			delete(ccn.coins, sid)
		}
	}
}

// getCoinObj returns an object responsible for the specified coin.
// It creates one, if there is no such.
func (ccn *commonCoinNode) getCoinObj(sid []byte) *commonCoin {
	var sidStr = string(sid)
	if _, ok := ccn.coins[sidStr]; !ok {
		ccn.coins[sidStr] = newCommonCoin(sid, ccn)
	}
	return ccn.coins[sidStr]
}

// endregion ///////////////////////////////////////////////////////////////////

// region commonCoin ///////////////////////////////////////////////////////////

// Type commonCoin represents information needed for construct single common coin.
type commonCoin struct {
	sid     []byte          // The original sid that we are signing.
	shares  [][]byte        // Shares of the signature to construct value from.
	missing []bool          // Peers who asked us for a share, but we had no such yet.
	value   []byte          // Resulting value of the coin.
	created time.Time       // For cleanup.
	node    *commonCoinNode // Node maintaining this coin.
	sent    bool            // Indicates, if we have sent our share at least once.
	waitCh  chan []byte     // Channel where the local client waits for the result.
}

// newCommonCoin stands for a constructor.
func newCommonCoin(sid []byte, node *commonCoinNode) *commonCoin {
	return &commonCoin{
		sid:     sid,
		shares:  make([][]byte, node.dkShare.N),
		missing: make([]bool, node.dkShare.N),
		value:   nil,
		created: time.Now(),
		node:    node,
	}
}

// getCoin is used by the local user to request the coin.
func (cc *commonCoin) getCoin(waitCh chan []byte) {
	cc.cancelReq() // Cancel the previous request, if any.
	cc.waitCh = waitCh
	if done, err := cc.addOurShare(); !done || err != nil {
		// On the first iteration we don't ask for resends,
		// because it highly possible that the messages are on the way.
		cc.broadcastCoinShare(false)
	}
}

func (cc *commonCoin) addOurShare() (bool, error) {
	var err error
	var signed tbdn.SigShare
	if signed, err = cc.node.dkShare.SignShare(cc.sid); err != nil {
		return false, xerrors.Errorf("failed to sign our share: %w", err)
	}
	return cc.acceptCoinShare(*cc.node.dkShare.Index, signed, false, nil)
}

func (cc *commonCoin) acceptCoinShare(peerIndex uint16, coinShare []byte, needReply bool, peer peering.PeerSender) (bool, error) {
	var err error
	if len(cc.shares) <= int(peerIndex) {
		return false, xerrors.New("peerIndex out of range")
	}
	if needReply && peer != nil {
		cc.resendCoinShare(peerIndex, peer)
	}
	cc.missing[peerIndex] = needReply
	if cc.shares[peerIndex] == nil {
		// Ignore duplicated messages.
		cc.shares[peerIndex] = coinShare
	}
	var receivedShares = make([][]byte, 0)
	for _, share := range cc.shares {
		if share != nil {
			receivedShares = append(receivedShares, share)
		}
	}
	if uint16(len(receivedShares)) < cc.node.dkShare.T {
		return false, xerrors.New("threshold not reached")
	}
	var sig *bls.SignatureWithPublicKey
	if sig, err = cc.node.dkShare.RecoverFullSignature(receivedShares, cc.sid); err != nil {
		return false, xerrors.Errorf("unable to reconstruct the signature: %w", err)
	}
	var value = sig.Signature.Bytes()
	if err = cc.node.dkShare.VerifyMasterSignature(cc.sid, value); err != nil {
		return false, xerrors.Errorf("unable to verify the master signature: %w", err)
	}
	cc.value = value
	if cc.waitCh != nil {
		cc.waitCh <- value
		close(cc.waitCh)
		cc.waitCh = nil
	}
	return true, nil
}

func (cc *commonCoin) onTimerTick() bool {
	now := time.Now()
	if cc.created.Add(giveUpTimeout).Before(now) {
		cc.cancelReq()
		return false
	}
	cc.broadcastCoinShare(true)
	return true
}

// broadcastCoinShare sends out share to all the peers,
// from which we haven't received a share.
func (cc *commonCoin) broadcastCoinShare(needReply bool) {
	now := time.Now()
	ourIndex := *cc.node.dkShare.Index
	ourShare := cc.shares[ourIndex]
	if ourShare == nil || cc.value != nil {
		return
	}
	for peerIndex, peer := range cc.node.group.OtherNodes() {
		if peerIndex == ourIndex || (cc.shares[peerIndex] != nil && cc.sent && !cc.missing[peerIndex]) {
			continue
		}
		cc.missing[peerIndex] = false
		msg := commonCoinMsg{
			sid:       cc.sid,
			coinShare: ourShare,
			needReply: needReply,
		}
		peer.SendMsg(&peering.PeerMessage{
			PeeringID:   cc.node.peeringID,
			SenderIndex: ourIndex,
			Timestamp:   now.UnixNano(),
			MsgType:     commonCoinShareMsgType,
			MsgData:     msg.Bytes(),
		})
	}
	cc.sent = true
}

func (cc *commonCoin) resendCoinShare(peerIndex uint16, peer peering.PeerSender) {
	now := time.Now()
	ourIndex := *cc.node.dkShare.Index
	ourShare := cc.shares[ourIndex]
	if ourShare == nil {
		return
	}
	msg := commonCoinMsg{
		sid:       cc.sid,
		coinShare: ourShare,
		needReply: cc.shares[peerIndex] == nil,
	}
	peer.SendMsg(&peering.PeerMessage{
		PeeringID:   cc.node.peeringID,
		SenderIndex: ourIndex,
		Timestamp:   now.UnixNano(),
		MsgType:     commonCoinShareMsgType,
		MsgData:     msg.Bytes(),
	})
}

func (cc *commonCoin) cancelReq() {
	if cc.waitCh != nil {
		close(cc.waitCh)
		cc.waitCh = nil
	}
}

// endregion ///////////////////////////////////////////////////////////////////

// region commonCoinMsg ////////////////////////////////////////////////////////
type commonCoinReq struct {
	sid    []byte      // What we are asking for.
	waitCh chan []byte // Channel where the caller is waiting for a response.
}

// endregion ///////////////////////////////////////////////////////////////////

// region commonCoinMsg ////////////////////////////////////////////////////////

// commonCoinMsg represents a messages the peers exchange
// to agree on a common coin.
type commonCoinMsg struct {
	sid       []byte
	coinShare []byte
	needReply bool // Indicates, if sender is still waiting for a share from the receiver.
}

func (m *commonCoinMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteBytes16(w, m.sid); err != nil {
		return xerrors.Errorf("failed to write commonCoinMsg.sid: %w", err)
	}
	if err = util.WriteBytes16(w, m.coinShare); err != nil {
		return xerrors.Errorf("failed to write commonCoinMsg.coinShare: %w", err)
	}
	if err = util.WriteBoolByte(w, m.needReply); err != nil {
		return xerrors.Errorf("failed to write commonCoinMsg.needReply: %w", err)
	}
	return nil
}

func (m *commonCoinMsg) Read(r io.Reader) error {
	var err error
	if m.sid, err = util.ReadBytes16(r); err != nil {
		return xerrors.Errorf("failed to read commonCoinMsg.sid: %w", err)
	}
	if m.coinShare, err = util.ReadBytes16(r); err != nil {
		return xerrors.Errorf("failed to read commonCoinMsg.coinShare: %w", err)
	}
	if err = util.ReadBoolByte(r, &m.needReply); err != nil {
		return xerrors.Errorf("failed to read commonCoinMsg.needReply: %w", err)
	}
	return nil
}

func (m *commonCoinMsg) FromBytes(buf []byte) error {
	r := bytes.NewReader(buf)
	return m.Read(r)
}

func (m *commonCoinMsg) Bytes() []byte {
	var buf bytes.Buffer
	_ = m.Write(&buf)
	return buf.Bytes()
}

// endregion ///////////////////////////////////////////////////////////////////
