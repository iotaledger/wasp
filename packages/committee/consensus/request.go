package consensus

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/publisher"
	"time"
)

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

	publish := false
	if ok {
		if msgFirstTime {
			ret.reqTx = reqMsg.Transaction
			ret.whenMsgReceived = time.Now()
			publish = true
		}
	} else {
		ret = op.newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.reqTx = reqMsg.Transaction
		op.requests[reqId] = ret
		op.addRequestIdConcurrent(&reqId)
		publish = true
	}
	if publish {
		publisher.Publish("request_in",
			op.committee.Address().String(),
			reqMsg.Transaction.ID().String(),
			fmt.Sprintf("%d", reqMsg.Index),
		)
	}

	ret.notifications[op.peerIndex()] = true

	tl := ""
	if msgFirstTime && ret.isTimelocked(time.Now()) {
		tl = fmt.Sprintf(". Time locked until %d (nowis = %d)", ret.timelock(), util.TimeNowUnix())
	}
	ret.log.Infof("NEW REQUEST from msg%s", tl)

	return ret, msgFirstTime
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
		op.removeRequestIdConcurrent(rid)
		op.log.Debugf("removed from backlog: processed request %s", rid.String())
	}
	return nil
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

func (op *operator) addRequestIdConcurrent(reqId *sctransaction.RequestId) {
	op.concurrentAccessMutex.Lock()
	defer op.concurrentAccessMutex.Unlock()

	op.requestIdsProtected[*reqId] = true
}

func (op *operator) removeRequestIdConcurrent(reqId *sctransaction.RequestId) {
	op.concurrentAccessMutex.Lock()
	defer op.concurrentAccessMutex.Unlock()

	delete(op.requestIdsProtected, *reqId)
}

func (op *operator) hasRequestIdConcurrent(reqId *sctransaction.RequestId) bool {
	op.concurrentAccessMutex.RLock()
	defer op.concurrentAccessMutex.RUnlock()

	_, ok := op.requestIdsProtected[*reqId]
	return ok
}

func (op *operator) IsRequestInBacklog(reqId *sctransaction.RequestId) bool {
	return op.hasRequestIdConcurrent(reqId)
}
