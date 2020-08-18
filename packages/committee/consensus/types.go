package consensus

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

type operator struct {
	committee committee.Committee
	dkshare   *tcrypto.DKShare
	//currentState
	currentState state.VirtualState
	stateTx      *sctransaction.Transaction
	balances     map[valuetransaction.ID][]*balance.Balance
	synchronized bool

	requestBalancesDeadline time.Time
	processorReady          bool

	// notifications with future currentState indices
	notificationsBacklog []*committee.NotifyReqMsg

	// notifications must be sent in the new cycle
	sendNotificationsScheduled bool

	requests map[sctransaction.RequestId]*request

	peerPermutation           *util.Permutation16
	leaderRotationDeadlineSet bool
	stateTxEvidenced          bool
	leaderRotationDeadline    time.Time

	leaderStatus        *leaderStatus
	sentResultsToLeader map[uint16]*sctransaction.Transaction

	log *logger.Logger
}

type leaderStatus struct {
	reqs          []*request
	batch         state.Batch
	batchHash     hashing.HashValue
	timestamp     int64
	balances      map[valuetransaction.ID][]*balance.Balance
	resultTx      *sctransaction.Transaction
	finalized     bool
	signedResults []*signedResult
}

type signedResult struct {
	essenceHash hashing.HashValue
	sigShare    tbdn.SigShare
}

// keeps stateTx of the request
type request struct {
	// id of the hash of request tx id and request block index
	reqId sctransaction.RequestId
	// from request message. nil if request message wasn't received yet
	reqTx *sctransaction.Transaction
	// time when request message was received by the operator
	whenMsgReceived time.Time
	// notification vector for the current currentState
	notifications []bool
	// initially time locked
	timelocked bool

	log *logger.Logger
}

func NewOperator(committee committee.Committee, dkshare *tcrypto.DKShare, log *logger.Logger) *operator {
	defer committee.SetReadyConsensus()

	return &operator{
		committee:           committee,
		dkshare:             dkshare,
		requests:            make(map[sctransaction.RequestId]*request),
		peerPermutation:     util.NewPermutation16(committee.Size(), nil),
		sentResultsToLeader: make(map[uint16]*sctransaction.Transaction),
		log:                 log.Named("c"),
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

func (op *operator) stateIndex() (uint32, bool) {
	if op.currentState == nil {
		return 0, false
	}
	return op.currentState.StateIndex(), true
}

func (op *operator) mustStateIndex() uint32 {
	ret, ok := op.stateIndex()
	if !ok {
		panic("mustStateIndex")
	}
	return ret
}

func (op *operator) getProgramHash() (*hashing.HashValue, bool) {
	if op.currentState == nil {
		return nil, false
	}
	h, ok, err := op.currentState.Variables().Codec().GetHashValue(vmconst.VarNameProgramHash)
	if !ok || err != nil {
		return nil, false
	}
	return h, true
}

func (op *operator) getRewardAddress() address.Address {
	return registry.GetRewardAddress(op.committee.Address())
}

func (op *operator) getMinimumReward() int64 {
	if _, ok := op.stateIndex(); !ok {
		return 0
	}
	vt, ok, err := op.currentState.Variables().Codec().GetInt64(vmconst.VarNameMinimumReward)
	if err != nil {
		panic(err)
	}
	if !ok {
		return 0
	}
	return vt
}
