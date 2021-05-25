// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package commonsubset

import (
	"sort"
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/peering"
	"golang.org/x/xerrors"
)

// TODO: Handle outdated instances in some way.

// CommonSubsetCoordinator is responsible for maintaining a series of ACS
// instances and to implement the AsynchronousCommonSubsetRunner interface.
//
// The main functions:
//   - Start new CommonSubset instances.
//   - Terminate outdated CommonSubset instances.
//   - Dispatch incoming messages to appropriate instances.
//
type CommonSubsetCoordinator struct {
	csInsts map[uint64]*CommonSubset // The actual instances can be created on request or on peer message.
	csAsked map[uint64]bool          // Indicates, which instances are already asked by this nodes.
	lock    sync.RWMutex

	peeringID  peering.PeeringID
	net        peering.NetworkProvider
	netGroup   peering.GroupProvider
	threshold  uint16
	commonCoin commoncoin.Provider
	log        *logger.Logger
}

func NewCommonSubsetCoordinator(
	peeringID peering.PeeringID,
	net peering.NetworkProvider,
	netGroup peering.GroupProvider,
	threshold uint16,
	commonCoin commoncoin.Provider,
	log *logger.Logger,
) *CommonSubsetCoordinator {
	return &CommonSubsetCoordinator{
		csInsts:    make(map[uint64]*CommonSubset),
		csAsked:    make(map[uint64]bool),
		lock:       sync.RWMutex{},
		peeringID:  peeringID,
		net:        net, // TODO: Avoid using it here.
		netGroup:   netGroup,
		threshold:  threshold,
		commonCoin: commonCoin,
		log:        log,
	}
}

// Close implements the AsynchronousCommonSubsetRunner interface.
func (csc *CommonSubsetCoordinator) Close() {
	for i := range csc.csInsts {
		csc.csInsts[i].Close()
	}
}

// RunACSConsensus implements the AsynchronousCommonSubsetRunner interface.
// It is possible that the instacne is already created because of the messages
// from other nodes. In such case we will just provide our input and register the callback.
func (csc *CommonSubsetCoordinator) RunACSConsensus(
	value []byte,
	sessionID uint64,
	callback func(sessionID uint64, acs [][]byte),
) {
	var err error
	var cs *CommonSubset
	if cs, err = csc.getOrCreateCS(sessionID, callback); err != nil {
		csc.log.Errorf("Unable to get a CommonSubset instance for sessionID=%v, reason=%v", sessionID, err)
		return
	}
	cs.Input(value)
}

// TryHandleMessage implements the AsynchronousCommonSubsetRunner interface.
// It handles the network messages, if they are of correct type.
func (csc *CommonSubsetCoordinator) TryHandleMessage(recv *peering.RecvEvent) bool {
	if recv.Msg.MsgType != acsMsgType {
		return false
	}
	var sessionID uint64
	var err error
	if sessionID, err = sessionIDFromMsgBytes(recv.Msg.MsgData); err != nil {
		csc.log.Warnf("Unable to extract a sessionID from the message, err=%v", err)
		return true
	}
	var cs *CommonSubset
	if cs, err = csc.getOrCreateCS(sessionID, nil); err != nil {
		csc.log.Errorf("Unable to get a CommonSubset instance for sessionID=%v, reason=%v", sessionID, err)
		return true
	}
	return cs.TryHandleMessage(recv)
}

func (csc *CommonSubsetCoordinator) getOrCreateCS(sessionID uint64, callback func(sessionID uint64, acs [][]byte)) (*CommonSubset, error) {
	csc.lock.Lock()
	defer csc.lock.Unlock()
	//
	// Reject duplicate calls from this node.
	if _, ok := csc.csAsked[sessionID]; ok && callback != nil {
		return nil, xerrors.Errorf("duplicate acs request")
	}
	//
	// Return the existing instance, if any. Register the callback if passed.
	if cs, ok := csc.csInsts[sessionID]; ok {
		if callback != nil {
			csc.csAsked[sessionID] = true
			go csc.callbackOnEvent(sessionID, cs.OutputCh(), callback)
		}
		return cs, nil
	}
	//
	// Otherwise create new instance, and register the callback if passed.
	var err error
	var newCS *CommonSubset
	var outCh chan map[uint16][]byte = make(chan map[uint16][]byte)
	if newCS, err = NewCommonSubset(sessionID, csc.peeringID, csc.net, csc.netGroup, csc.threshold, csc.commonCoin, outCh, csc.log); err != nil {
		return nil, err
	}
	csc.csInsts[sessionID] = newCS
	if callback != nil {
		csc.csAsked[sessionID] = true
		go csc.callbackOnEvent(sessionID, newCS.OutputCh(), callback)
	}
	return newCS, nil
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
