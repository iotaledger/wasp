package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
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

// request record retrieved (or created) by request message
func (op *operator) requestFromMsg(reqMsg *committee.RequestMsg) *request {
	reqId := sctransaction.NewRequestId(reqMsg.Transaction.ID(), reqMsg.Index)
	ret, ok := op.requests[reqId]
	if ok && ret.reqTx == nil {
		ret.reqTx = reqMsg.Transaction
		ret.whenMsgReceived = time.Now()
		return ret
	}
	if !ok {
		ret = op.newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.reqTx = reqMsg.Transaction
		op.requests[reqId] = ret
	}
	ret.notifications[op.peerIndex()] = true

	return ret
}

// TODO caching processed requests
// TODO gracefull reaction in DB error
func (op *operator) isRequestProcessed(reqid *sctransaction.RequestId) bool {
	addr := op.committee.Address()
	processed, err := state.IsRequestCompleted(&addr, reqid)
	if err != nil {
		panic(err)
	}
	return processed
}

// deleteCompletedRequests deletes requests which were successfully processed or failed more than maximum retry limit
func (op *operator) deleteCompletedRequests() error {
	toDelete := make([]*sctransaction.RequestId, 0)

	addr := op.committee.Address()
	for _, req := range op.requests {
		if completed, err := state.IsRequestCompleted(&addr, &req.reqId); err != nil {
			return err
		} else {
			if completed {
				toDelete = append(toDelete, &req.reqId)
			}
		}
	}
	for _, rid := range toDelete {
		delete(op.requests, *rid)
	}
	return nil
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

func takeIds(reqs []*request) []sctransaction.RequestId {
	ret := make([]sctransaction.RequestId, len(reqs))
	for i := range ret {
		ret[i] = reqs[i].reqId
	}
	return ret
}

func takeRefs(reqs []*request) ([]sctransaction.RequestRef, bool) {
	ret := make([]sctransaction.RequestRef, len(reqs))
	for i := range ret {
		if reqs[i].reqTx == nil {
			return nil, false
		}
		ret[i] = sctransaction.RequestRef{
			Tx:    reqs[i].reqTx,
			Index: reqs[i].reqId.Index(),
		}
	}
	return ret, true
}
