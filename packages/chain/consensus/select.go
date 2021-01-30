// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// the file contains functions responsible for the request batch selection logic
package consensus

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"sort"
	"time"
)

// selectRequestsToProcess select requests to process in the batch.
// 1. it filters out candidates which was seen less than quorum times.
// 2. the requests which are not ready yet to process in the current context are filtered out
// 3. selects maximum possible set of those which were seen by same quorum of peers
// only requests in "full batches" are selected, it means request is in the selection together with ALL other requests
// from the same request transaction, or it is not selected
func (op *operator) selectRequestsToProcess() []*request {
	candidates := op.requestCandidateList()
	if len(candidates) == 0 {
		return nil
	}
	if candidates = op.filterRequestsNotSeenQuorumTimes(candidates); len(candidates) == 0 {
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
	if len(ret) == 0 {
		return nil
	}
	op.log.Debugf("requests selected for process: %d out of total %d", len(ret), len(op.requests))
	return ret
}

func (op *operator) allRequests() []*request {
	ret := make([]*request, 0, len(op.requests))
	for _, req := range op.requests {
		ret = append(ret, req)
	}
	return ret
}

// all requests from the backlog which:
// - has known messages
// - has solid arguments
// - are not timelocked
// sort by arrival time
func (op *operator) requestCandidateList() []*request {
	ret := op.allRequests()
	nowis := time.Now()
	ret = filterRequests(ret, func(r *request) bool {
		return r.hasMessage() && !r.isTimeLocked(nowis) && r.hasSolidArgs()
	})
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].whenMsgReceived.Before(ret[j].whenMsgReceived)
	})
	return ret
}

func (op *operator) requestsTimeLocked() []*request {
	ret := make([]*request, 0, len(op.requests))

	nowis := time.Now()
	for _, req := range op.requests {
		if req.reqTx == nil {
			continue
		}
		if !req.isTimeLocked(nowis) {
			continue
		}
		ret = append(ret, req)
	}
	return ret
}

type requestWithVotes struct {
	*request
	seenTimes uint16
}

func (op *operator) filterRequestsNotSeenQuorumTimes(candidates []*request) []*request {
	if len(candidates) == 0 {
		return nil
	}
	ret1 := make([]*requestWithVotes, 0)
	for _, req := range candidates {
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
	ret := candidates[:0] // same underlying array
	for _, req := range ret1 {
		ret = append(ret, req.request)
	}
	return ret
}

func (op *operator) collectProcessableBatch(reqIds []coretypes.RequestID) []*request {
	nowis := time.Now()
	return filterRequests(op.takeFromIds(reqIds), func(r *request) bool {
		return r.hasMessage() && !r.isTimeLocked(nowis) && r.hasSolidArgs()
	})
}

func filterRequests(reqs []*request, fn func(r *request) bool) []*request {
	ret := reqs[:0]
	for _, r := range reqs {
		if fn(r) {
			ret = append(ret, r)
		}
	}
	return ret
}

// filterNotReadyYet checks all ids and returns list of corresponding request records
// return empty list if not all requests in the list can be processed by the node atm
// note, that filter out criteria are temporary, so the same request may be ready next time
// 'not ready yet' requests are:
//  - which has not received message with request transaction yet (the ID is known from peer only)
//  - the user defined request while processor is not ready yet
//  - the request is timelocked yet
//func (op *operator) filterNotReadyYet(reqs []*request) []*request {
//	if len(reqs) == 0 {
//		return nil
//	}
//	ret := reqs[:0] // same underlying array, different slice
//
//	for _, req := range reqs {
//		if req.reqTx == nil {
//			op.log.Debugf("request %s not yet known to the node: can't be processed", req.reqId.Short())
//			continue
//		}
//		ret = append(ret, req)
//	}
//	before := len(ret)
//	ret = filterTimelocked(ret)
//	after := len(ret)
//	op.log.Debugf("Number of time locked requests filtered out: %d", before-after)
//
//	before = len(ret)
//	ret = filterNotSolidArgs(ret)
//	after = len(ret)
//	op.log.Debugf("Number of requests with non-solid args filtered out: %d", before-after)
//	op.log.Debugf("Number of selected requests: %d", len(ret))
//	return ret
//}

// filterOutRequestsWithoutTokens leaves only those first requests
// which has corresponding request tokens.
//// TODO redundant? It is checked by tx.Properties
//func (op *operator) filterOutRequestsWithoutTokens(reqs []*request) []*request {
//	if op.balances == nil {
//		return nil
//	}
//	byColor, _ := txutil.BalancesByColor(op.balances)
//	ret := reqs[:0]
//	for _, req := range reqs {
//		col := balance.Color(*req.reqId.TransactionID())
//		v, _ := byColor[col]
//		if v <= 0 {
//			continue
//		}
//		byColor[col] = v - 1
//		ret = append(ret, req)
//	}
//	return ret
//}

// checkChainToken validates if the NFT, the chain token, is among balances, as expected
//func (op *operator) checkChainToken(balances map[valuetransaction.ID][]*balance.Balance) error {
//	sum := int64(0)
//	color := *op.chain.Color()
//	for _, bals := range balances {
//		sum += txutil.BalanceOfColor(bals, color)
//	}
//	if stateIndex, ok := op.blockIndex(); ok && stateIndex > 0 {
//		if sum != 1 {
//			return fmt.Errorf("must be exactly 1 SC token of color %s. Found %d instead", color.String(), sum)
//		}
//	}
//	return nil
//}

//func filterNotSolidArgs(reqs []*request) []*request {
//	ret := reqs[:0]
//	for _, req := range reqs {
//		if req.reqTx == nil {
//			// just in case??
//			continue
//		}
//		if !req.argsSolid {
//			req.log.Debugf("not solid args: filtered out")
//			continue
//		}
//		ret = append(ret, req)
//	}
//	return ret
//}
//
//func filterTimelocked(reqs []*request) []*request {
//	ret := reqs[:0]
//	nowis := time.Now()
//	for _, req := range reqs {
//		if req.reqTx == nil {
//			// just in case??
//			continue
//		}
//		if req.isTimeLocked(nowis) {
//			if req.timelock() > 0 {
//				req.log.Debugf("timelocked until %d: filtered out. nowis %d", req.timelock(), nowis.Unix())
//			}
//			continue
//		}
//		if req.timelock() > 0 {
//			req.log.Debugf("timelocked until %d: pass. nowis %d", req.timelock(), nowis.Unix())
//		}
//		ret = append(ret, req)
//	}
//	return ret
//}

func numTrue(bs []bool) uint16 {
	ret := uint16(0)
	for _, v := range bs {
		if v {
			ret++
		}
	}
	return ret
}
