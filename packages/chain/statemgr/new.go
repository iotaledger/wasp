// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"time"

	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type stateManager struct {
	committee chain.Chain

	// becomes true after initially loaded state is validated.
	// after that it is always true
	solidStateValid bool

	// flag pingPong[idx] if ping-pong message was received from the peer idx
	pingPong              []bool
	deadlineForPongQuorum time.Time

	// pending batches of state updates are candidates to confirmation by the state transaction
	// which leads to the state transition
	// the map key is hash of the variable state which is a result of applying the
	// batch of state updates to the solid variable state
	pendingBatches map[hashing.HashValue]*pendingBatch

	// last variable state stored in the database
	// it may be nil at bootstrap when origin variable state is calculated
	solidState state.VirtualState

	// state transaction with +1 state index from the state index of solid variable state
	// it may be nil if does not exist or not fetched yet
	nextStateTransaction *sctransaction.Transaction

	// state transaction which approves current state
	approvingTransaction *sctransaction.Transaction

	// was state transition message of the current state sent to the consensus operator
	consensusNotifiedOnStateTransition bool

	// largest state index evidenced by other messages. If this index is more than 1 step ahead
	// of the solid variable state, it means the state of the smart contract in the current node
	// falls behind the state of the smart contract, i.e. it is not synced
	largestEvidencedStateIndex uint32

	// the timeout deadline for sync inquiries
	syncMessageDeadline time.Time

	// current batch being synced
	syncedBatch *syncedBatch

	// for the pseudo-random sequence of peers
	permutation *util.Permutation16

	// logger
	log *logger.Logger
}

type syncedBatch struct {
	msgCounter   uint16
	stateIndex   uint32
	stateUpdates []state.StateUpdate
	stateTxId    valuetransaction.ID
}

type pendingBatch struct {
	// batch of state updates, not validated yet
	batch state.Batch
	// resulting variable state after applied the batch to the solidState
	nextState state.VirtualState
	// state transaction request deadline. For committed batches only
	stateTransactionRequestDeadline time.Time
}

func New(c chain.Chain, log *logger.Logger) chain.StateManager {
	ret := &stateManager{
		committee:      c,
		pingPong:       make([]bool, c.Size()),
		pendingBatches: make(map[hashing.HashValue]*pendingBatch),
		permutation:    util.NewPermutation16(c.NumPeers(), nil),
		log:            log.Named("s"),
	}
	go ret.initLoadState()

	return ret
}

// initial loading of the solid state
func (sm *stateManager) initLoadState() {
	var err error
	var batch state.Batch
	var stateExists bool

	sm.solidState, batch, stateExists, err = state.LoadSolidState(sm.committee.Address())
	if err != nil {
		sm.log.Errorf("initLoadState: %v", err)
		sm.committee.Dismiss()
		return
	}

	if stateExists {
		// state loaded, will be waiting for it to be confirmed from the tangle
		sm.addPendingBatch(batch)
		sm.largestEvidencedStateIndex = sm.solidState.StateIndex()

		h := sm.solidState.Hash()
		txh := batch.StateTransactionId()
		sm.log.Debugw("solid state has been loaded",
			"state index", sm.solidState.StateIndex(),
			"state hash", h.String(),
			"approving tx", txh.String(),
		)
	} else {
		// pre-origin state. Origin batch is emty batch.
		// Will be waiting for the origin transaction to arrive
		sm.addPendingBatch(state.MustNewOriginBatch(sm.committee.Color()))

		sm.log.Info("solid state does not exist: WAITING FOR THE ORIGIN TRANSACTION")
	}

	// open msg queue for the committee
	sm.committee.SetReadyStateManager()
}
