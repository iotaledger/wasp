package consensus

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"time"
)

// check if the request message is well formed
func (op *Operator) validateRequestBlock(reqRef *committee.RequestMsg) error {
	// TODO check rewards etc
	return nil
}

func (op *Operator) newRequest(reqId *sctransaction.RequestId) *request {
	reqLog := log.Named(reqId.Short())
	ret := &request{
		reqId: reqId,
		log:   reqLog,
	}
	reqLog.Info("NEW REQUEST")
	return ret
}

// request record retrieved (or created) by request message
func (op *Operator) requestFromMsg(reqMsg *committee.RequestMsg) *request {
	reqId := reqMsg.Id()
	ret, ok := op.requests[*reqId]
	if ok && ret.reqMsg == nil {
		ret.reqMsg = reqMsg
		ret.whenMsgReceived = time.Now()
		return ret
	}
	if !ok {
		ret = op.newRequest(reqId)
		ret.whenMsgReceived = time.Now()
		ret.reqMsg = reqMsg
		op.requests[*reqId] = ret
	}
	ret.msgCounter++
	return ret
}

// request record is retrieved by request id.
// If it doesn't exist and is not in the list of processed requests, it is created
func (op *Operator) requestFromId(reqId *sctransaction.RequestId) (*request, bool) {
	if op.isRequestProcessed(reqId) {
		return nil, false
	}
	ret, ok := op.requests[*reqId]
	if !ok {
		ret = op.newRequest(reqId)
		op.requests[*reqId] = ret
	}
	ret.msgCounter++
	return ret, true
}

// TODO caching processed requests
// TODO gracefull reaction in DB error
func (op *Operator) isRequestProcessed(reqid *sctransaction.RequestId) bool {
	processed, err := state.IsRequestProcessed(reqid)
	if err != nil {
		panic(err)
	}
	return processed
}

func (op *Operator) removeRequest(reqId *sctransaction.RequestId) bool {
	if _, ok := op.requests[*reqId]; ok {
		delete(op.requests, *reqId)
		return true
	}
	return false
}
