// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package commonsubset

import (
	"sort"
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"golang.org/x/xerrors"
)

const (
	futureInstances = 5 // How many future instances to accept.
	pastInstances   = 2 // How many past instance to keep not closed.

	peerMsgTypeBatch = iota
)

// CommonSubsetCoordinator is responsible for maintaining a series of ACS
// instances and to implement the AsynchronousCommonSubsetRunner interface.
//
// The main functions:
//   - Check new CommonSubset instances.
//   - Terminate outdated CommonSubset instances.
//   - Dispatch incoming messages to appropriate instances.
//
// NOTE: On the termination of the ACS instances. The instances are initiated either
// by the current node or by a message from other node.
//
//   - If we have a newer StateOutput, older ACS are irrelevant, because the state
//     is already in the Ledger and is confirmed.
//
//   - It is safe to drop messages for some future ACS instances, because if they
//     are correct ones, they will redeliver the messages until we wil catch up
//     to their StateIndex. So the handling of the future ACS instance is only a
//     performance optimization.
//
// What to do in the case of reset? Set the current state index as the last one
// this node has provided. Some future ACS instances will be left for some time,
// until they will be discarded because of the growing StateIndex in the new branch.
type CommonSubsetCoordinator struct {
	csInsts                     map[uint64]*CommonSubset // The actual instances can be created on request or on peer message.
	csAsked                     map[uint64]bool          // Indicates, which instances are already asked by this nodes.
	currentStateIndex           uint32                   // Last state index passed by this node.
	receivePeerMessagesAttachID interface{}
	lock                        sync.RWMutex

	netGroup peering.GroupProvider
	dkShare  tcrypto.DKShare
	log      *logger.Logger
}

func NewCommonSubsetCoordinator(
	net peering.NetworkProvider,
	netGroup peering.GroupProvider,
	dkShare tcrypto.DKShare,
	log *logger.Logger,
) *CommonSubsetCoordinator {
	ret := &CommonSubsetCoordinator{
		csInsts:  make(map[uint64]*CommonSubset),
		csAsked:  make(map[uint64]bool),
		lock:     sync.RWMutex{},
		netGroup: netGroup,
		dkShare:  dkShare,
		log:      log,
	}
	ret.receivePeerMessagesAttachID = ret.netGroup.Attach(peering.PeerMessageReceiverCommonSubset, ret.receiveCommitteePeerMessages)
	return ret
}

// Close implements the AsynchronousCommonSubsetRunner interface.
func (csc *CommonSubsetCoordinator) Close() {
	csc.netGroup.Detach(csc.receivePeerMessagesAttachID)
	for i := range csc.csInsts {
		csc.csInsts[i].Close()
	}
}

// RunACSConsensus implements the AsynchronousCommonSubsetRunner interface.
// It is possible that the instacne is already created because of the messages
// from other nodes. In such case we will just provide our input and register the callback.
func (csc *CommonSubsetCoordinator) RunACSConsensus(
	value []byte, // Our proposal.
	sessionID uint64, // Consensus to participate in.
	stateIndex uint32, // Monotonic sequence, used to clear old ACS instances.
	callback func(sessionID uint64, acs [][]byte),
) {
	var err error
	var cs *CommonSubset
	if len(csc.netGroup.AllNodes()) == 1 {
		// There is no point to do a consensus for a single node.
		// Moreover, the erasure coding fails for the case of single node.
		go callback(sessionID, [][]byte{value})
		return
	}
	if cs, err = csc.getOrCreateCS(sessionID, stateIndex, callback); err != nil {
		csc.log.Debugf("Unable to get a CommonSubset instance for sessionID=%v, reason=%v", sessionID, err)
		return
	}
	cs.Input(value)
}

// TryHandleMessage implements the AsynchronousCommonSubsetRunner interface.
// It handles the network messages, if they are of correct type.
func (csc *CommonSubsetCoordinator) receiveCommitteePeerMessages(peerMsg *peering.PeerMessageGroupIn) {
	if peerMsg.MsgType != peerMsgTypeBatch {
		csc.log.Warnf("Wrong type of committee message: %v, ignoring it", peerMsg.MsgType)
		return
	}
	mb, err := newMsgBatch(peerMsg.MsgData)
	if err != nil {
		csc.log.Error(err)
		return
	}
	csc.log.Debugf("ACS::IO - Received a msgBatch=%+v", *mb)
	var cs *CommonSubset
	if cs, err = csc.getOrCreateCS(mb.sessionID, mb.stateIndex, nil); err != nil {
		csc.log.Debugf("Unable to get a CommonSubset instance for sessionID=%v, reason=%v", mb.sessionID, err)
		return
	}
	cs.HandleMsgBatch(mb)
}

func (csc *CommonSubsetCoordinator) getOrCreateCS(
	sessionID uint64,
	stateIndex uint32,
	callback func(sessionID uint64, acs [][]byte),
) (*CommonSubset, error) {
	csc.lock.Lock()
	defer csc.lock.Unlock()
	ownCall := callback != nil
	//
	// Reject duplicate calls from this node.
	if _, ok := csc.csAsked[sessionID]; ok && ownCall {
		return nil, xerrors.Errorf("duplicate acs request")
	}
	//
	// Record the current state index and drop the outdated instances.
	if ownCall && csc.currentStateIndex != stateIndex {
		csc.currentStateIndex = stateIndex
		for i := range csc.csInsts {
			if csc.csInsts[i].stateIndex < csc.currentStateIndex && !csc.inRange(csc.csInsts[i].stateIndex) {
				// We close only the past instances, because they are not needed anymore.
				// The future instances cannot be deleted in this way (e.g. in case of chain reset)
				// because some of the messages are probably already received and acknowledged, thus
				// will not be resent in the future, if we would recreate them.
				csc.csInsts[i].Close()
				delete(csc.csInsts, i)
			}
		}
	}
	//
	// Return the existing instance, if any. Register the callback if passed.
	if cs, ok := csc.csInsts[sessionID]; ok {
		if ownCall {
			csc.csAsked[sessionID] = true
			go csc.callbackOnEvent(sessionID, cs.OutputCh(), callback)
		}
		return cs, nil
	}
	//
	// Otherwise create new instance, and register the callback if passed.
	if ownCall || csc.inRange(stateIndex) {
		var err error
		var newCS *CommonSubset
		outCh := make(chan map[uint16][]byte, 1)
		if newCS, err = NewCommonSubset(sessionID, stateIndex, csc.netGroup, csc.dkShare, false, outCh, csc.log); err != nil {
			return nil, err
		}
		csc.csInsts[sessionID] = newCS
		if ownCall {
			csc.csAsked[sessionID] = true
			go csc.callbackOnEvent(sessionID, newCS.OutputCh(), callback)
		}
		return newCS, nil
	}
	return nil, xerrors.Errorf("stateIndex %v out of range (current=%v)", stateIndex, csc.currentStateIndex)
}

// This should be run in a separate thread.
func (csc *CommonSubsetCoordinator) callbackOnEvent(
	sessionID uint64,
	outCh chan map[uint16][]byte,
	callback func(sessionID uint64, acs [][]byte),
) {
	out, ok := <-outCh
	if !ok {
		// The ACS was interrupted / canceled.
		// We will not invoke the callback in this case.
		return
	}
	keys := make([]int, 0)
	for i := range out {
		keys = append(keys, int(i))
	}
	sort.Ints(keys)
	values := make([][]byte, 0)
	for _, i := range keys {
		values = append(values, out[uint16(i)])
	}
	callback(sessionID, values)
}

func (csc *CommonSubsetCoordinator) inRange(stateIndex uint32) bool {
	minRange := uint32(0)
	if csc.currentStateIndex > pastInstances {
		minRange = csc.currentStateIndex - pastInstances
	}
	return minRange <= stateIndex && stateIndex <= csc.currentStateIndex+futureInstances
}
