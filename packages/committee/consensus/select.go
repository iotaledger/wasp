package consensus

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"sort"
	"time"
)

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

func numTrue(bs []bool) uint16 {
	ret := uint16(0)
	for _, v := range bs {
		if v {
			ret++
		}
	}
	return ret
}
