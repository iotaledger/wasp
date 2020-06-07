package consensus

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

type operator struct {
	committee committee.Committee
	dkshare   *tcrypto.DKShare
	stateTx   *sctransaction.Transaction

	balances            map[valuetransaction.ID][]*balance.Balance
	getBalancesDeadline time.Time

	variableState  state.VariableState
	processorReady bool

	// notifications with future state indices
	notificationsBacklog []*committee.NotifyReqMsg

	requests map[sctransaction.RequestId]*request

	peerPermutation           *util.Permutation16
	leaderRotationDeadlineSet bool
	leaderRotationDeadline    time.Time
	// states of requests being processed: as leader and as subordinate

	leaderStatus             *leaderStatus
	currentStateCompRequests []*computationRequest
	nextStateCompRequests    []*computationRequest

	log *logger.Logger
}

type leaderStatus struct {
	reqs          []*request
	batch         state.Batch
	batchHash     hashing.HashValue
	timestamp     time.Time
	balances      map[valuetransaction.ID][]*balance.Balance
	resultTx      *sctransaction.Transaction
	finalized     bool
	signedResults []*signedResult
}

type signedResult struct {
	essenceHash hashing.HashValue
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
	reqId sctransaction.RequestId

	// from request message. nil if request message wasn't received yet
	reqTx *sctransaction.Transaction

	// time when request message was received by the operator
	whenMsgReceived time.Time

	// notification vector for the current state
	notifications []bool

	log *logger.Logger
}

func NewOperator(committee committee.Committee, dkshare *tcrypto.DKShare, log *logger.Logger) *operator {
	defer committee.SetReadyConsensus()

	return &operator{
		committee:       committee,
		dkshare:         dkshare,
		requests:        make(map[sctransaction.RequestId]*request),
		peerPermutation: util.NewPermutation16(committee.Size(), nil),
		log:             log.Named("c"),
	}
}

func (op *operator) peerIndex() uint16 {
	return op.dkshare.Index
}

func (op *operator) quorum() uint16 {
	return op.dkshare.T
}

func (op *operator) size() uint16 {
	return op.dkshare.N
}

func (op *operator) stateIndex() uint32 {
	if op.variableState == nil {
		return 0
	}
	return op.variableState.StateIndex()
}

func (op *operator) getProgramHashStr() (string, bool) {
	if op.variableState == nil {
		return "", false
	}
	p, ok := op.variableState.Variables().Get(origin.VarNameProgramHash)
	if !ok {
		return "", false
	}
	progHashStr, ok := p.(string)
	if !ok {
		return "", false
	}
	return progHashStr, true
}

func (op *operator) MustValidStateIndex(stateIndex uint32) {
	if stateIndex != op.stateIndex() && stateIndex != op.stateIndex()+1 {
		// only tolerated messages from current and next state indices
		// stateManager should not pass other messages
		panic(fmt.Errorf("wrong state index. Current %d, got %d", op.stateIndex(), stateIndex))
	}
}
