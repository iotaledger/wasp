package consensus

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"go.uber.org/atomic"
)

type Consensus struct {
	isReady                          atomic.Bool
	chain                            chain.ChainCore
	committee                        chain.Committee
	mempool                          chain.Mempool
	nodeConn                         chain.NodeConnection
	vmRunner                         vm.VMRunner
	currentState                     state.VirtualState
	stateOutput                      *ledgerstate.AliasOutput
	stateTimestamp                   time.Time
	acsSessionID                     uint64
	consensusBatch                   *BatchProposal
	consensusEntropy                 hashing.HashValue
	iAmContributor                   bool
	myContributionSeqNumber          uint16
	contributors                     []uint16
	workflow                         workflowFlags
	delayBatchProposalUntil          time.Time
	delayRunVMUntil                  time.Time
	delaySendingSignedResult         time.Time
	resultTxEssence                  *ledgerstate.TransactionEssence
	resultState                      state.VirtualState
	resultSignatures                 []*messages.SignedResultMsg
	finalTx                          *ledgerstate.Transaction
	postTxDeadline                   time.Time
	pullInclusionStateDeadline       time.Time
	lastTimerTick                    atomic.Int64
	consensusInfoSnapshot            atomic.Value
	timers                           ConsensusTimers
	log                              *logger.Logger
	eventStateTransitionMsgCh        chan *messages.StateTransitionMsg
	eventSignedResultMsgCh           chan *messages.SignedResultMsg
	eventInclusionStateMsgCh         chan *messages.InclusionStateMsg
	eventACSMsgCh                    chan *messages.AsynchronousCommonSubsetMsg
	eventVMResultMsgCh               chan *messages.VMResultMsg
	eventTimerMsgCh                  chan messages.TimerTick
	closeCh                          chan struct{}
	assert                           assert.Assert
	missingRequestsFromBatch         map[coretypes.RequestID][32]byte
	missingRequestsMutex             sync.Mutex
	pullMissingRequestsFromCommittee bool
}

type workflowFlags struct {
	stateReceived        bool
	batchProposalSent    bool
	consensusBatchKnown  bool
	vmStarted            bool
	vmResultSigned       bool
	transactionFinalized bool
	transactionPosted    bool
	transactionSeen      bool
	finished             bool
}

var _ chain.Consensus = &Consensus{}

func New(chainCore chain.ChainCore, mempool chain.Mempool, committee chain.Committee, nodeConn chain.NodeConnection, pullMissingRequestsFromCommittee bool, timersOpt ...ConsensusTimers) *Consensus {
	var timers ConsensusTimers
	if len(timersOpt) > 0 {
		timers = timersOpt[0]
	} else {
		timers = NewConsensusTimers()
	}
	log := chainCore.Log().Named("c")
	ret := &Consensus{
		chain:                            chainCore,
		committee:                        committee,
		mempool:                          mempool,
		nodeConn:                         nodeConn,
		vmRunner:                         runvm.NewVMRunner(),
		resultSignatures:                 make([]*messages.SignedResultMsg, committee.Size()),
		timers:                           timers,
		log:                              log,
		eventStateTransitionMsgCh:        make(chan *messages.StateTransitionMsg),
		eventSignedResultMsgCh:           make(chan *messages.SignedResultMsg),
		eventInclusionStateMsgCh:         make(chan *messages.InclusionStateMsg),
		eventACSMsgCh:                    make(chan *messages.AsynchronousCommonSubsetMsg),
		eventVMResultMsgCh:               make(chan *messages.VMResultMsg),
		eventTimerMsgCh:                  make(chan messages.TimerTick),
		closeCh:                          make(chan struct{}),
		assert:                           assert.NewAssert(log),
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
	}
	ret.refreshConsensusInfo()
	go ret.recvLoop()
	return ret
}

func (c *Consensus) IsReady() bool {
	return c.isReady.Load()
}

func (c *Consensus) Close() {
	close(c.closeCh)
}

func (c *Consensus) recvLoop() {
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

func (c *Consensus) refreshConsensusInfo() {
	index := uint32(0)
	if c.currentState != nil {
		index = c.currentState.BlockIndex()
	}
	consensusInfo := &chain.ConsensusInfo{
		StateIndex: index,
		Mempool:    c.mempool.Info(),
		TimerTick:  int(c.lastTimerTick.Load()),
	}
	c.log.Debugf("Refreshing consensus info: index=%v, timerTick=%v, "+
		"totalPool=%v, mempoolReady=%v, inBufCounter=%v, outBufCounter=%v, "+
		"inPoolCounter=%v, outPoolCounter=%v",
		consensusInfo.StateIndex, consensusInfo.TimerTick,
		consensusInfo.Mempool.TotalPool, consensusInfo.Mempool.ReadyCounter,
		consensusInfo.Mempool.InBufCounter, consensusInfo.Mempool.OutBufCounter,
		consensusInfo.Mempool.InPoolCounter, consensusInfo.Mempool.OutPoolCounter,
	)
	c.consensusInfoSnapshot.Store(consensusInfo)
}

func (c *Consensus) GetStatusSnapshot() *chain.ConsensusInfo {
	ret := c.consensusInfoSnapshot.Load()
	if ret == nil {
		return nil
	}
	return ret.(*chain.ConsensusInfo)
}
