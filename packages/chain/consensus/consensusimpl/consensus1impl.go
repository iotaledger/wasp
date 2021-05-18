package consensusimpl

import (
	"time"

	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/state"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"go.uber.org/atomic"
)

type consensusImpl struct {
	isReady                    atomic.Bool
	chain                      chain.ChainCore
	committee                  chain.Committee
	mempool                    chain.Mempool
	nodeConn                   chain.NodeConnection
	vmRunner                   vm.VMRunner
	currentState               state.VirtualState
	stateOutput                *ledgerstate.AliasOutput
	stateTimestamp             time.Time
	acsSessionID               uint64
	consensusBatch             *batchProposal
	iAmContributor             bool
	myContributionSeqNumber    uint16
	contributors               []uint16
	workflow                   workflowFlags
	delayBatchProposalUntil    time.Time
	delayRunVMUntil            time.Time
	resultTxEssence            *ledgerstate.TransactionEssence
	resultState                state.VirtualState
	resultSignatures           []*chain.SignedResultMsg
	finalTx                    *ledgerstate.Transaction
	approvingOutputID          ledgerstate.OutputID
	postTxDeadline             time.Time
	pullInclusionStateDeadline time.Time
	lastTimerTick              atomic.Int64
	consensusInfoSnapshot      atomic.Value
	log                        *logger.Logger
	eventStateTransitionMsgCh  chan *chain.StateTransitionMsg
	eventResultCalculatedMsgCh chan *chain.VMResultMsg
	eventSignedResultMsgCh     chan *chain.SignedResultMsg
	eventInclusionStateMsgCh   chan *chain.InclusionStateMsg
	eventACSMsgCh              chan *chain.AsynchronousCommonSubsetMsg
	eventVMResultMsgCh         chan *chain.VMResultMsg
	eventTimerMsgCh            chan chain.TimerTick
	closeCh                    chan struct{}
}

type workflowFlags struct {
	stateReceived                bool
	batchProposalSent            bool
	consensusBatchKnown          bool
	vmStarted                    bool
	vmResultSignedAndBroadcasted bool
	transactionFinalized         bool
	transactionPosted            bool
	transactionSeen              bool
	finished                     bool
}

var _ chain.Consensus = &consensusImpl{}

func New(chainCore chain.ChainCore, mempool chain.Mempool, committee chain.Committee, nodeConn chain.NodeConnection, log *logger.Logger) *consensusImpl {
	ret := &consensusImpl{
		chain:                      chainCore,
		committee:                  committee,
		mempool:                    mempool,
		nodeConn:                   nodeConn,
		vmRunner:                   runvm.NewVMRunner(),
		resultSignatures:           make([]*chain.SignedResultMsg, committee.Size()),
		log:                        log.Named("c"),
		eventStateTransitionMsgCh:  make(chan *chain.StateTransitionMsg),
		eventResultCalculatedMsgCh: make(chan *chain.VMResultMsg),
		eventSignedResultMsgCh:     make(chan *chain.SignedResultMsg),
		eventInclusionStateMsgCh:   make(chan *chain.InclusionStateMsg),
		eventACSMsgCh:              make(chan *chain.AsynchronousCommonSubsetMsg),
		eventVMResultMsgCh:         make(chan *chain.VMResultMsg),
		eventTimerMsgCh:            make(chan chain.TimerTick),
		closeCh:                    make(chan struct{}),
	}
	ret.refreshConsensusInfo()
	go ret.recvLoop()
	return ret
}

func (c *consensusImpl) IsReady() bool {
	return c.isReady.Load()
}

func (c *consensusImpl) Close() {
	close(c.closeCh)
}

func (c *consensusImpl) recvLoop() {
	// wait at startup
	for !c.committee.IsReady() {
		select {
		case <-time.After(100 * time.Millisecond):
		case <-c.closeCh:
			return
		}
	}
	c.log.Debugf("consensus object is ready")
	c.isReady.Store(true)
	for {
		select {
		case msg, ok := <-c.eventStateTransitionMsgCh:
			if ok {
				c.eventStateTransitionMsg(msg)
			}
		case msg, ok := <-c.eventResultCalculatedMsgCh:
			if ok {
				c.eventResultCalculated(msg)
			}
		case msg, ok := <-c.eventSignedResultMsgCh:
			if ok {
				c.eventSignedResult(msg)
			}
		case msg, ok := <-c.eventInclusionStateMsgCh:
			if ok {
				c.eventInclusionState(msg)
			}
		case msg, ok := <-c.eventACSMsgCh:
			if ok {
				c.eventAsynchronousCommonSubset(msg)
			}
		case msg, ok := <-c.eventVMResultMsgCh:
			if ok {
				c.eventVMResultMsg(msg)
			}
		case msg, ok := <-c.eventTimerMsgCh:
			if ok {
				c.eventTimerMsg(msg)
			}
		case <-c.closeCh:
			return
		}
	}
}

func (c *consensusImpl) refreshConsensusInfo() {
	index := uint32(0)
	if c.currentState != nil {
		index = c.currentState.BlockIndex()
	}
	c.consensusInfoSnapshot.Store(&chain.ConsensusInfo{
		StateIndex: index,
		Mempool:    c.mempool.Stats(),
		TimerTick:  int(c.lastTimerTick.Load()),
	})
}

func (c *consensusImpl) GetStatusSnapshot() *chain.ConsensusInfo {
	ret := c.consensusInfoSnapshot.Load()
	if ret == nil {
		return nil
	}
	return ret.(*chain.ConsensusInfo)
}
