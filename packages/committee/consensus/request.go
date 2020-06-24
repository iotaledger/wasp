package consensus

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
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
	reqLog.Info("NEW REQUEST")
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
	}
	return ret, true
}

func (op *operator) requestMsgList() []*request {
	ret := make([]*request, 0, len(op.requests))
	for _, req := range op.requests {
		if req.reqTx != nil {
			ret = append(ret, req)
		}
	}
	return ret
}

// request record retrieved (or created) by request message
func (op *operator) requestFromMsg(reqMsg committee.RequestMsg) (*request, bool) {
	reqId := sctransaction.NewRequestId(reqMsg.Transaction.ID(), reqMsg.Index)
	ret, ok := op.requests[reqId]
	if ok {
		newreq := ret.reqTx == nil
		if newreq {
			ret.reqTx = reqMsg.Transaction
			ret.whenMsgReceived = time.Now()
		}
		return ret, newreq
	}
	if !ok {
		ret = op.newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.reqTx = reqMsg.Transaction
		op.requests[reqId] = ret
	}
	ret.notifications[op.peerIndex()] = true

	return ret, true
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
	ret = op.filterFullRequestTokenConsumers(ret)

	after := idsShortStr(takeIds(ret))

	if len(after) != len(before) {
		op.log.Debugf("filterFullRequestTokenConsumers: %+v --> %+v\nbalances: %s",
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

// selects only those requests which together consume ALL request tokens from current balances
func (op *operator) filterFullRequestTokenConsumers(reqs []*request) []*request {
	if len(reqs) == 0 {
		return nil
	}
	if op.balances == nil {
		return nil
	}
	// count number of requests by request transaction
	reqtxs := make(map[valuetransaction.ID]int)
	for _, req := range reqs {
		txid := req.reqId.TransactionId()
		if _, ok := reqtxs[*txid]; !ok {
			reqtxs[*txid] = 0
		}
		reqtxs[*txid] += 1
	}
	coltxs, _ := util.BalancesByColor(op.balances)

	ret := reqs[:0] // same underlying array, different slice
	for _, req := range reqs {
		txid := *req.reqId.TransactionId()
		isOriginTx := txid == (valuetransaction.ID)(*op.committee.Color())
		sum, ok := coltxs[(balance.Color)(txid)]
		if !ok {
			continue
		}
		if isOriginTx {
			// one token is smart contract token
			sum = sum - 1
		}
		numreq := reqtxs[txid]
		if sum != int64(numreq) {
			continue
		}
		ret = append(ret, req)
	}
	return ret
}

// filterNotReadyYet checks all ids and returns list of corresponding request records
// return empty list if not all requests in the list can be processed by the node atm
// note, that filter out criteria are temporary, so the same request may ready next time
func (op *operator) filterNotReadyYet(reqs []*request) []*request {
	ret := reqs[:0] // same underlying array, different slice

	for _, req := range reqs {
		if req.reqTx == nil {
			op.log.Debugf("request %s not known by the node: can't be processed", req.reqId.Short())
			continue
		}
		reqBlock := req.reqTx.Requests()[req.reqId.Index()]
		if reqBlock.RequestCode().IsUserDefined() && !op.processorReady {
			op.log.Debugf("request %s can't be processed: processor not ready", req.reqId.Short())
			continue
		}
		ret = append(ret, req)
	}
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
