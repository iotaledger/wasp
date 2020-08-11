package consensus

import (
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"sort"
	"time"
)

// check if the request message is well formed
func (op *operator) validateRequestBlock(reqRef *committee.RequestMsg) error {
	// TODO check rewards etc
	return nil
}

func (op *operator) newRequest(reqId sctransaction.RequestId) *request {
	reqLog := op.log.Named(reqId.Short())
	ret := &request{
		reqId:         reqId,
		log:           reqLog,
		notifications: make([]bool, op.size()),
	}
	return ret
}

// request record is retrieved by request id.
// If it doesn't exist and is not in the list of processed requests, it is created
func (op *operator) requestFromId(reqId sctransaction.RequestId) (*request, bool) {
	if op.isRequestProcessed(&reqId) {
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
func (op *operator) requestFromMsg(reqMsg *committee.RequestMsg) (*request, bool) {
	reqId := sctransaction.NewRequestId(reqMsg.Transaction.ID(), reqMsg.Index)
	ret, ok := op.requests[reqId]
	msgFirstTime := !ok || ret.reqTx == nil

	if ok {
		if msgFirstTime {
			ret.reqTx = reqMsg.Transaction
			ret.whenMsgReceived = time.Now()
		}
	} else {
		ret = op.newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.reqTx = reqMsg.Transaction
		op.requests[reqId] = ret
	}
	ret.notifications[op.peerIndex()] = true

	nowis := time.Now()
	if msgFirstTime && ret.isTimelocked(nowis) {
		ret.expectTimeUnlockEvent = true
	}
	tl := ""
	if ret.isTimelocked(nowis) {
		tl = fmt.Sprintf(". Time locked until %d (nowis = %d)", ret.timelock(), util.TimeNowUnix())
	}
	ret.log.Infof("NEW REQUEST from msg%s", tl)

	return ret, msgFirstTime
}

func (op *operator) requestCandidateList() []*request {
	ret := make([]*request, 0, len(op.requests))
	nowis := time.Now()
	for _, req := range op.requests {
		if req.reqTx == nil {
			continue
		}
		if req.isTimelocked(nowis) {
			continue
		}
		ret = append(ret, req)
	}
	return ret
}

func (req *request) requestCode() sctransaction.RequestCode {
	return req.reqTx.Requests()[req.reqId.Index()].RequestCode()
}

func (req *request) timelock() uint32 {
	return req.reqTx.Requests()[req.reqId.Index()].Timelock()
}

func (req *request) isTimelocked(nowis time.Time) bool {
	return req.timelock() > uint32(nowis.Unix())
}

// selectRequestsToProcess select requests to process in the batch by counting votes of notification messages
// first it selects candidates with >= quorum 'seen' votes and sorts by num votes
// then it selects maximum number of requests which has been seen by at least quorum of common peers
// the requests are sorted by arrival time
// only requests in "full batches" are selected, it means request is in the selection together with ALL other requests
// from the same request transaction, or it is not selected
func (op *operator) selectRequestsToProcess() []*request {
	candidates := op.requestMessagesSeenQuorumTimes()
	if len(candidates) == 0 {
		return nil
	}
	candidates = op.filterNotReadyYet(candidates)
	if len(candidates) == 0 {
		return nil
	}

	ret := []*request{candidates[0]}
	intersection := make([]bool, op.size())
	copy(intersection, candidates[0].notifications)

	for i := uint16(1); int(i) < len(candidates); i++ {
		for j := range intersection {
			intersection[j] = intersection[j] && candidates[i].notifications[j]
		}
		if numTrue(intersection) < op.quorum() {
			break
		}
		ret = append(ret, candidates[i])
	}
	if ret == nil {
		return nil
	}
	before := idsShortStr(takeIds(ret))
	ret = op.filterNotCompletePackages(ret)

	after := idsShortStr(takeIds(ret))

	if len(after) != len(before) {
		op.log.Debugf("filterNotCompletePackages: %+v --> %+v\nbalances: %s",
			before, after, util.BalancesToString(op.balances))
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].whenMsgReceived.Before(ret[j].whenMsgReceived)
	})
	return ret
}

type requestWithVotes struct {
	*request
	seenTimes uint16
}

func (op *operator) requestMessagesSeenQuorumTimes() []*request {
	ret1 := make([]*requestWithVotes, 0)
	for _, req := range op.requests {
		if req.reqTx == nil {
			continue
		}
		votes := numTrue(req.notifications)
		if votes >= op.quorum() {
			ret1 = append(ret1, &requestWithVotes{
				request:   req,
				seenTimes: votes,
			})
		}
	}
	sort.Slice(ret1, func(i, j int) bool {
		return ret1[i].seenTimes > ret1[j].seenTimes
	})
	ret := make([]*request, len(ret1))
	for i, req := range ret1 {
		ret[i] = req.request
	}
	return ret
}

func (op *operator) isRequestProcessed(reqid *sctransaction.RequestId) bool {
	addr := op.committee.Address()
	processed, err := state.IsRequestCompleted(addr, reqid)
	if err != nil {
		panic(err)
	}
	return processed
}

// deleteCompletedRequests deletes requests which were successfully processed or failed more than maximum retry limit
func (op *operator) deleteCompletedRequests() error {
	toDelete := make([]*sctransaction.RequestId, 0)

	for _, req := range op.requests {
		if completed, err := state.IsRequestCompleted(op.committee.Address(), &req.reqId); err != nil {
			return err
		} else {
			if completed {
				toDelete = append(toDelete, &req.reqId)
			}
		}
	}
	for _, rid := range toDelete {
		delete(op.requests, *rid)
		op.log.Debugf("removed from backlog: processed request %s", rid.String())
	}
	return nil
}

type txReqNums struct {
	totalNumOfRequestsInTx int
	numOfRequestsInTheList int
}

// ensure that ether ALL user-defined requests to this smart contract are in the batch or none
func (op *operator) filterNotCompletePackages(reqs []*request) []*request {
	if len(reqs) == 0 {
		return nil
	}
	if op.balances == nil {
		return nil
	}
	// count number of user-defined requests by request transaction
	reqstats := make(map[valuetransaction.ID]*txReqNums)
	for _, req := range reqs {
		if !req.requestCode().IsUserDefined() {
			continue
		}
		txid := req.reqTx.ID()
		if _, ok := reqstats[txid]; !ok {
			reqstats[txid] = &txReqNums{
				totalNumOfRequestsInTx: req.reqTx.NumRequestsToAddress(op.committee.Address()),
				numOfRequestsInTheList: 0,
			}
		}
		reqstats[txid].numOfRequestsInTheList += 1
	}
	if len(reqstats) == 0 {
		// no user defined-requests
		return reqs
	}
	ret := reqs[:0] // same underlying array, different slice
	for _, req := range reqs {
		st := reqstats[req.reqTx.ID()]
		if st.numOfRequestsInTheList != st.totalNumOfRequestsInTx {
			continue
		}
		ret = append(ret, req)
	}
	return ret
}

func filterTimelocked(reqs []*request) []*request {
	ret := reqs[:0]
	nowis := time.Now()
	for _, req := range reqs {
		if req.reqTx == nil {
			// just in case??
			continue
		}
		if req.isTimelocked(nowis) {
			if req.timelock() > 0 {
				req.log.Debugf("timelocked until %d: filtered out. nowis %d", req.timelock(), nowis.Unix())
			}
			continue
		}
		if req.timelock() > 0 {
			req.log.Debugf("timelocked until %d: pass. nowis %d", req.timelock(), nowis.Unix())
		}
		ret = append(ret, req)
	}
	return ret
}

// filterNotReadyYet checks all ids and returns list of corresponding request records
// return empty list if not all requests in the list can be processed by the node atm
// note, that filter out criteria are temporary, so the same request may be ready next time
func (op *operator) filterNotReadyYet(reqs []*request) []*request {
	ret := reqs[:0] // same underlying array, different slice

	for _, req := range reqs {
		if req.reqTx == nil {
			op.log.Debugf("request %s not known to the node: can't be processed", req.reqId.Short())
			continue
		}
		if req.requestCode().IsUserDefined() && !op.processorReady {
			op.log.Debugf("request %s can't be processed: processor not ready", req.reqId.Short())
			continue
		}
		ret = append(ret, req)
	}
	before := len(ret)
	ret = filterTimelocked(ret)
	after := len(ret)

	op.log.Debugf("Number of timelocked requests filtered out: %d", before-after)

	return ret
}

func setAllFalse(bs []bool) {
	for i := range bs {
		bs[i] = false
	}
}

func numTrue(bs []bool) uint16 {
	ret := uint16(0)
	for _, v := range bs {
		if v {
			ret++
		}
	}
	return ret
}

func idsShortStr(ids []sctransaction.RequestId) []string {
	ret := make([]string, len(ids))
	for i := range ret {
		ret[i] = ids[i].Short()
	}
	return ret
}

func (op *operator) takeFromIds(reqIds []sctransaction.RequestId) []*request {
	ret := make([]*request, 0, len(reqIds))
	for _, reqId := range reqIds {
		req, _ := op.requestFromId(reqId)
		if req == nil {
			continue
		}
		ret = append(ret, req)
	}
	return ret
}

func takeIds(reqs []*request) []sctransaction.RequestId {
	ret := make([]sctransaction.RequestId, len(reqs))
	for i := range ret {
		ret[i] = reqs[i].reqId
	}
	return ret
}

func takeRefs(reqs []*request) []sctransaction.RequestRef {
	ret := make([]sctransaction.RequestRef, len(reqs))
	for i := range ret {
		ret[i] = sctransaction.RequestRef{
			Tx:    reqs[i].reqTx,
			Index: reqs[i].reqId.Index(),
		}
	}
	return ret
}
