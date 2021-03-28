// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

func (op *operator) newRequest(reqId coretypes.RequestID) *request {
	ret := &request{
		log:           op.log.Named(reqId.Short()),
		notifications: make([]bool, op.size()),
	}
	return ret
}

// request record is retrieved by request id.
// If it doesn't exist and is not in the list of processed requests, it is created
func (op *operator) requestFromId(reqId coretypes.RequestID) (*request, bool) {
	if op.isRequestProcessed(reqId) {
		return nil, false
	}
	ret, ok := op.requests[reqId]
	if !ok {
		ret = op.newRequest(reqId)
		op.requests[reqId] = ret
		ret.log.Info("NEW REQUEST from id")
	}
	return ret, true
}

// request record retrieved (or created) by request message
func (op *operator) requestFromMsg(req coretypes.Request) (*request, bool) {
	if op.isRequestProcessed(req.ID()) {
		return nil, false
	}
	ret, ok := op.requests[req.ID()]
	msgFirstTime := !ok || ret.req == nil

	newMsg := false
	if ok {
		if msgFirstTime {
			ret.req = req
			ret.whenMsgReceived = time.Now()
			newMsg = true
		}
	} else {
		ret = op.newRequest(req.ID())
		ret.whenMsgReceived = time.Now()
		ret.req = req
		op.requests[req.ID()] = ret
		op.addRequestIdConcurrent(req.ID())
		newMsg = true
	}
	if newMsg {
		// solidify arguments by resolving blob references from the registry
		// the request will not be selected for processing until ret.argsSolid == true
		if reqOnLedger, ok := req.(*sctransaction.RequestOnLedger); ok {
			ok, err := reqOnLedger.SolidifyArgs(op.chain.BlobCache())
			if err != nil {
				ret.log.Errorf("inconsistency: can't solidify args: %v", err)
			} else {
				ret.argsSolid = ok
			}
		}
	}
	if newMsg {
		publisher.Publish("request_in", op.chain.ID().String(), req.ID().String())
	}

	ret.notifications[op.peerIndex()] = true

	tl := ""
	if msgFirstTime && ret.isTimeLocked(time.Now()) {
		tl = fmt.Sprintf(". Time locked until %d (nowis = %d)", ret.timelock(), util.TimeNowUnix())
	}
	ret.log.Infof("NEW REQUEST from msg%s", tl)

	return ret, msgFirstTime
}

func (req *request) timelock() uint32 {
	reqOnLedger, ok := req.req.(*sctransaction.RequestOnLedger)
	if !ok {
		return 0
	}
	return uint32(reqOnLedger.TimeLock().Unix())
}

func (req *request) isTimeLocked(nowis time.Time) bool {
	return req.timelock() > uint32(nowis.Unix())
}

func (req *request) hasMessage() bool {
	return req.req != nil
}

func (req *request) hasSolidArgs() bool {
	return req.argsSolid
}

func (op *operator) isRequestProcessed(reqid coretypes.RequestID) bool {
	processed, err := state.IsRequestCompleted(op.chain.ID(), reqid)
	if err != nil {
		panic(err)
	}
	return processed
}

// deleteCompletedRequests deletes requests which were successfully processed or failed more than maximum retry limit
func (op *operator) deleteCompletedRequests() error {
	toDelete := make([]coretypes.RequestID, 0)

	for _, req := range op.requests {
		if completed, err := state.IsRequestCompleted(op.chain.ID(), req.req.ID()); err != nil {
			return err
		} else {
			if completed {
				toDelete = append(toDelete, req.req.ID())
			}
		}
	}
	for _, rid := range toDelete {
		delete(op.requests, rid)
		op.removeRequestIdConcurrent(&rid)
		op.log.Debugf("removed from backlog: processed request %s", rid.String())
	}
	return nil
}

func idsShortStr(ids []coretypes.RequestID) []string {
	ret := make([]string, len(ids))
	for i := range ret {
		ret[i] = ids[i].Short()
	}
	return ret
}

func (op *operator) takeFromIds(reqIds []coretypes.RequestID) []*request {
	ret := make([]*request, 0, len(reqIds))
	for _, reqId := range reqIds {
		req, ok := op.requestFromId(reqId)
		if !ok || req == nil {
			continue
		}
		ret = append(ret, req)
	}
	return ret
}

func takeIds(reqs []*request) []coretypes.RequestID {
	ret := make([]coretypes.RequestID, len(reqs))
	for i := range ret {
		ret[i] = reqs[i].req.ID()
	}
	return ret
}

func (op *operator) addRequestIdConcurrent(reqId coretypes.RequestID) {
	op.concurrentAccessMutex.Lock()
	defer op.concurrentAccessMutex.Unlock()

	op.requestIdsProtected[reqId] = true
}

func (op *operator) removeRequestIdConcurrent(reqId *coretypes.RequestID) {
	op.concurrentAccessMutex.Lock()
	defer op.concurrentAccessMutex.Unlock()

	delete(op.requestIdsProtected, *reqId)
}

func (op *operator) hasRequestIdConcurrent(reqId *coretypes.RequestID) bool {
	op.concurrentAccessMutex.RLock()
	defer op.concurrentAccessMutex.RUnlock()

	_, ok := op.requestIdsProtected[*reqId]
	return ok
}

func (op *operator) IsRequestInBacklog(reqId *coretypes.RequestID) bool {
	return op.hasRequestIdConcurrent(reqId)
}
