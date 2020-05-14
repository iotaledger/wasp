package consensus

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"time"
)

type operator struct {
	committee committee.Committee
	dkshare   *tcrypto.DKShare
	stateTx   *sctransaction.Transaction

	balances            map[valuetransaction.ID][]*balance.Balance
	getBalancesDeadline time.Time

	variableState state.VariableState
	// VM
	processor committee.Processor

	requests map[sctransaction.RequestId]*request

	requestNotificationsCurrentState []*requestNotification
	requestNotificationsNextState    []*requestNotification

	leaderPeerIndexList       []uint16
	currLeaderSeqIndex        uint16
	myLeaderSeqIndex          uint16
	leaderRotationDeadlineSet bool
	leaderRotationDeadline    time.Time
	// states of requests being processed: as leader and as subordinate

	leaderStatus             *leaderStatus
	currentStateCompRequests []*computationRequest
	nextStateCompRequests    []*computationRequest
}

type requestNotification struct {
	reqId     *sctransaction.RequestId
	peerIndex uint16
}

type leaderStatus struct {
	req           *request
	ts            time.Time
	resultTx      *sctransaction.Transaction
	finalized     bool
	signedResults []signedResult
}

type signedResult struct {
	essenceHash *hashing.HashValue
	sigShare    tbdn.SigShare
}

type computationRequest struct {
	ts              time.Time
	leaderPeerIndex uint16
	req             *request
	processed       bool
}

// keeps stateTx of the request
type request struct {

	// id of the hash of request tx id and request block index
	reqId *sctransaction.RequestId

	// request message or nil if wasn't received yet
	reqMsg *committee.RequestMsg

	// time when request message was received by the operator
	whenMsgReceived time.Time

	// request message as received by the operator.
	// Contains parsed SC transaction and the request block index
	//reqRef *sctransaction.RequestRef

	msgCounter int
	log        *logger.Logger
}

func NewOperator(committee committee.Committee, dkshare *tcrypto.DKShare) *operator {
	return &operator{
		committee: committee,
		dkshare:   dkshare,
	}
}

func (op *operator) quorum() uint16 {
	return op.dkshare.T
}

func (op *operator) stateIndex() uint32 {
	if op.variableState == nil {
		return 0
	}
	return op.variableState.StateIndex()
}

func (op *operator) MustValidStateIndex(stateIndex uint32) {
	if stateIndex != op.stateIndex() && stateIndex != op.stateIndex()+1 {
		// only tolerated messages from current and next state indices
		// stateManager should not pass other messages
		panic(fmt.Errorf("wrong state index. Current %d, got %d", op.stateIndex(), stateIndex))
	}
}
