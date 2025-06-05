// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package distsync implements distributed synchronization mechanisms for the mempool.
package distsync

import (
	"context"
	"fmt"
	"math/rand"
	"slices"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	maxTTL byte = 1
)

// The implementation is trivial and naive for now. A proper gossip or structured
// broadcast (possibly a mix of both) should be implemented here.
//
// In the current algorithm, for sharing a message:
//   - Just send a message to all the committee nodes.
//
// For querying a message:
//   - First ask all the committee for the message.
//   - If response not received, ask random subsets of server nodes.
//
// TODO: For the future releases: Implement proper dissemination algorithm.
type distSyncImpl struct {
	me                gpa.NodeID
	serverNodes       []gpa.NodeID // Should be used to push and query for requests.
	accessNodes       []gpa.NodeID // Maybe is not needed? Lets keep it until the redesign.
	committeeNodes    []gpa.NodeID // Subset of serverNodes and accessNodes.
	requestNeededCB   func(*isc.RequestRef) isc.Request
	requestReceivedCB func(isc.Request) bool
	nodeCountToShare  int // Number of nodes to share a request per iteration.
	maxMsgsPerTick    int
	needed            *shrinkingmap.ShrinkingMap[isc.RequestRefKey, *distSyncReqNeeded]
	missingReqsMetric func(count int)
	rnd               *rand.Rand
	log               log.Logger
}

var _ gpa.GPA = &distSyncImpl{}

type distSyncReqNeeded struct {
	reqRef  *isc.RequestRef
	waiters []context.Context
}

func New(
	me gpa.NodeID,
	requestNeededCB func(*isc.RequestRef) isc.Request,
	requestReceivedCB func(isc.Request) bool,
	maxMsgsPerTick int,
	missingReqsMetric func(count int),
	log log.Logger,
) gpa.GPA {
	return &distSyncImpl{
		me:                me,
		serverNodes:       []gpa.NodeID{},
		accessNodes:       []gpa.NodeID{},
		committeeNodes:    []gpa.NodeID{},
		requestNeededCB:   requestNeededCB,
		requestReceivedCB: requestReceivedCB,
		nodeCountToShare:  0,
		maxMsgsPerTick:    maxMsgsPerTick,
		needed:            shrinkingmap.New[isc.RequestRefKey, *distSyncReqNeeded](),
		missingReqsMetric: missingReqsMetric,
		rnd:               util.NewPseudoRand(),
		log:               log,
	}
}

func (dsi *distSyncImpl) Input(input gpa.Input) gpa.OutMessages {
	dsi.log.LogDebugf("Input %T: %+v", input, input)
	switch input := input.(type) {
	case *inputServerNodes:
		return dsi.handleInputServerNodes(input)
	case *inputAccessNodes:
		return dsi.handleInputAccessNodes(input)
	case *inputPublishRequest:
		return dsi.handleInputPublishRequest(input)
	case *inputRequestNeeded:
		return dsi.handleInputRequestNeeded(input)
	case *inputTimeTick:
		return dsi.handleInputTimeTick()
	}
	panic(fmt.Errorf("unexpected input type %T: %+v", input, input))
}

func (dsi *distSyncImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msg := msg.(type) {
	case *msgMissingRequest:
		return dsi.handleMsgMissingRequest(msg)
	case *msgShareRequest:
		return dsi.handleMsgShareRequest(msg)
	}
	dsi.log.LogWarnf("unexpected message %T: %+v", msg, msg)
	return nil
}

func (dsi *distSyncImpl) Output() gpa.Output {
	return nil // Output is provided via callbacks.
}

func (dsi *distSyncImpl) StatusString() string {
	return fmt.Sprintf("{MP, neededReqs=%v, nodeCountToShare=%v}", dsi.needed.Size(), dsi.nodeCountToShare)
}

func (dsi *distSyncImpl) handleInputServerNodes(input *inputServerNodes) gpa.OutMessages {
	dsi.log.LogDebugf("handleInputServerNodes: %v", input)
	dsi.handleCommitteeNodes(input.committeeNodes)
	dsi.serverNodes = input.serverNodes
	for i := range dsi.committeeNodes { // Ensure server nodes contain the committee nodes.
		if slices.Index(dsi.serverNodes, dsi.committeeNodes[i]) == -1 {
			dsi.serverNodes = append(dsi.serverNodes, dsi.committeeNodes[i])
		}
	}
	return dsi.handleInputTimeTick() // Re-send requests if node set has changed.
}

func (dsi *distSyncImpl) handleInputAccessNodes(input *inputAccessNodes) gpa.OutMessages {
	dsi.log.LogDebugf("handleInputAccessNodes: %v", input)
	dsi.handleCommitteeNodes(input.committeeNodes)
	dsi.accessNodes = input.accessNodes
	for i := range dsi.committeeNodes { // Ensure access nodes contain the committee nodes.
		if slices.Index(dsi.accessNodes, dsi.committeeNodes[i]) == -1 {
			dsi.accessNodes = append(dsi.accessNodes, dsi.committeeNodes[i])
		}
	}
	return dsi.handleInputTimeTick() // Re-send requests if node set has changed.
}

func (dsi *distSyncImpl) handleCommitteeNodes(committeeNodes []gpa.NodeID) {
	dsi.committeeNodes = committeeNodes
	dsi.nodeCountToShare = (len(dsi.committeeNodes)-1)/3 + 1 // F+1
	if dsi.nodeCountToShare < 2 {
		dsi.nodeCountToShare = 2
	}
	if dsi.nodeCountToShare > len(dsi.committeeNodes) {
		dsi.nodeCountToShare = len(dsi.committeeNodes)
	}
}

// In the current algorithm, for sharing a message:
//   - Just send a message to all the committee nodes (or server nodes, if committee is not known).
func (dsi *distSyncImpl) handleInputPublishRequest(input *inputPublishRequest) gpa.OutMessages {
	msgs := dsi.propagateRequest(input.request)
	//
	// Delete the it from the "needed" list, if any.
	// This node has the request, if it tries to publish it.
	reqRef := isc.RequestRefFromRequest(input.request)
	if dsi.needed.Delete(reqRef.AsKey()) {
		dsi.missingReqsMetric(dsi.needed.Size())
	}
	return msgs
}

func (dsi *distSyncImpl) propagateRequest(request isc.Request) gpa.OutMessages {
	msgs := gpa.NoMessages()
	var publishToNodes []gpa.NodeID
	if len(dsi.committeeNodes) > 0 {
		publishToNodes = dsi.committeeNodes
		dsi.log.LogDebugf("Forwarding request %v to committee nodes: %v", request.ID(), dsi.committeeNodes)
	} else {
		dsi.log.LogDebugf("Forwarding request %v to server nodes: %v", request.ID(), dsi.serverNodes)
		publishToNodes = dsi.serverNodes
	}
	for i := range publishToNodes {
		msgs.Add(newMsgShareRequest(request, 0, publishToNodes[i]))
	}
	return msgs
}

// For querying a message:
//   - First ask all the committee for the message.
//   - ...
func (dsi *distSyncImpl) handleInputRequestNeeded(input *inputRequestNeeded) gpa.OutMessages {
	reqRefKey := input.requestRef.AsKey()
	reqNeeded, have := dsi.needed.Get(reqRefKey)
	if have {
		if lo.Contains(reqNeeded.waiters, input.ctx) {
			return nil // Duplicate call, ignore it.
		}
		reqNeeded.waiters = append(reqNeeded.waiters, input.ctx)
	} else {
		reqNeeded = &distSyncReqNeeded{
			reqRef:  input.requestRef,
			waiters: []context.Context{input.ctx},
		}
	}
	if dsi.needed.Set(reqRefKey, reqNeeded) {
		dsi.missingReqsMetric(dsi.needed.Size())
	}
	msgs := gpa.NoMessages()
	for _, nid := range dsi.committeeNodes {
		msgs.Add(newMsgMissingRequest(input.requestRef, nid))
	}
	return msgs
}

// For querying a message:
//   - ...
//   - If response not received, ask random subsets of server nodes.
func (dsi *distSyncImpl) handleInputTimeTick() gpa.OutMessages {
	if dsi.needed.Size() == 0 {
		return nil
	}
	nodeCount := len(dsi.serverNodes)
	if nodeCount == 0 {
		return nil
	}
	msgs := gpa.NoMessages()
	nodePerm := dsi.rnd.Perm(nodeCount)
	counter := 0
	dsi.needed.ForEach(func(reqRefKey isc.RequestRefKey, reqNeeded *distSyncReqNeeded) bool { // Access is randomized.
		stillNeeded := lo.ContainsBy(reqNeeded.waiters, func(ctx context.Context) bool { return ctx.Err() == nil })
		if !stillNeeded {
			dsi.needed.Delete(reqRefKey)
			dsi.log.LogDebugf("Clearing MsgMissingRequest, not needed anymore: %v", reqNeeded.reqRef)
			return true
		}
		recipient := dsi.serverNodes[nodePerm[counter%nodeCount]]
		dsi.log.LogDebugf("Sending MsgMissingRequest for %v to %v", reqNeeded.reqRef, recipient)
		msgs.Add(newMsgMissingRequest(reqNeeded.reqRef, recipient))
		counter++
		return counter <= dsi.maxMsgsPerTick
	})
	dsi.missingReqsMetric(dsi.needed.Size())
	return msgs
}

func (dsi *distSyncImpl) handleMsgMissingRequest(msg *msgMissingRequest) gpa.OutMessages {
	req := dsi.requestNeededCB(msg.requestRef)
	if req != nil {
		msgs := gpa.NoMessages()
		msgs.Add(newMsgShareRequest(req, 0, msg.Sender()))
		return msgs
	}
	return nil
}

func (dsi *distSyncImpl) handleMsgShareRequest(msg *msgShareRequest) gpa.OutMessages {
	msgs := gpa.NoMessages()
	reqRefKey := isc.RequestRefFromRequest(msg.request).AsKey()
	added := dsi.requestReceivedCB(msg.request)
	if dsi.needed.Delete(reqRefKey) {
		dsi.missingReqsMetric(dsi.needed.Size())
	}
	//
	// Propagate the message, if it was new to us, and was received from outside of the committee.
	// The "outside of the committee" condition is used here to decrease echo-factor of the synchronization.
	// Each fair committee will send the request to all the committee nodes, thus we can avoid repeating it.
	// Follow the logic as if the message is received via the API.
	if added && !lo.Contains(dsi.committeeNodes, msg.Sender()) {
		msgs.AddAll(dsi.propagateRequest(msg.request))
	}
	//
	// The following is de-factor unused, as TTL is always 0 currently.
	if msg.ttl > 0 {
		ttl := msg.ttl
		if ttl > maxTTL {
			ttl = maxTTL
		}
		perm := dsi.rnd.Perm(len(dsi.committeeNodes))
		for i := 0; i < dsi.nodeCountToShare; i++ {
			msgs.Add(newMsgShareRequest(msg.request, ttl-1, dsi.committeeNodes[perm[i]]))
		}
		return msgs
	}
	return msgs
}
