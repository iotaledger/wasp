package consensus

import "sort"

// selectRequestsToProcess select requests to process in the batch by counting votes of notification messages
// first it selects candidates with >= quorum 'seen' votes and sorts by num votes
// then it selects maximum number of requests which has been seen by at least quorum of common peers
// the requests are sorted by arrival time
func (op *operator) selectRequestsToProcess() []*request {
	candidates := op.requestsSeenQuorumTimes()
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
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].whenMsgReceived.Before(ret[j].whenMsgReceived)
	})
	return ret
}

type requestWithVotes struct {
	req       *request
	seenTimes uint16
}

func (op *operator) requestsSeenQuorumTimes() []*request {
	ret1 := make([]*requestWithVotes, 0)
	for _, req := range op.requests {
		votes := numTrue(req.notifications)
		if votes >= op.quorum() {
			ret1 = append(ret1, &requestWithVotes{
				req:       req,
				seenTimes: votes,
			})
		}
	}
	sort.Slice(ret1, func(i, j int) bool {
		return ret1[i].seenTimes > ret1[j].seenTimes
	})
	ret := make([]*request, len(ret1))
	for i, req := range ret1 {
		ret[i] = req.req
	}
	return ret
}
