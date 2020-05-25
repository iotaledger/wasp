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

func (op *operator) newRequest(reqId *sctransaction.RequestId) *request {
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
func (op *operator) requestFromId(reqId *sctransaction.RequestId) (*request, bool) {
	if op.isRequestProcessed(reqId) {
		return nil, false
	}
	ret, ok := op.requests[*reqId]
	if !ok {
		ret = op.newRequest(reqId)
		op.requests[*reqId] = ret
	}
	return ret, true
}

// request record retrieved (or created) by request message
func (op *operator) requestFromMsg(reqMsg *committee.RequestMsg) *request {
	reqId := reqMsg.RequestId()
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
	ret.notifications[op.peerIndex()] = true

	return ret
}

// TODO caching processed requests
// TODO gracefull reaction in DB error
func (op *operator) isRequestProcessed(reqid *sctransaction.RequestId) bool {
	processed, err := state.IsRequestProcessed(reqid)
	if err != nil {
		panic(err)
	}
	return processed
}

func (op *operator) removeRequest(reqId *sctransaction.RequestId) bool {
	if _, ok := op.requests[*reqId]; ok {
		delete(op.requests, *reqId)
		return true
	}
	return false
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
		ret[i] = *reqs[i].reqId
	}
	return ret
}

func takeMsgs(reqs []*request) ([]*committee.RequestMsg, bool) {
	ret := make([]*committee.RequestMsg, len(reqs))
	for i := range ret {
		if reqs[i].reqMsg == nil {
			return nil, false
		}
		ret[i] = reqs[i].reqMsg
	}
	return ret, true
}
