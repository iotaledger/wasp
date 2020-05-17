// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"time"
)

type stateManager struct {
	committee committee.Committee

	// pending batches of state updates are candidates to confirmation by the state transaction
	// which leads to the state transition
	// the map key is hash of the variable state which is a result of applying the
	// batch of state updates to the solid variable state
	pendingBatches map[hashing.HashValue]*pendingBatch

	// state transaction with +1 state index from the state index of solid variable state
	// it may be nil if does not exist or not fetched yet
	nextStateTransaction *sctransaction.Transaction

	// last variable state stored in the database
	// it may be nil at bootstrap when origin variable state is calculated
	solidVariableState state.VariableState

	// the timeout deadline for async node request
	nodeRequestDeadline time.Time

	// largest state index evidenced by other messages. If this index is more than 1 step ahead
	// of the solid variable state, it means the state of the smart contract in the current node
	// falls behind the state of the smart contract, i.e. it is not synced
	largestEvidencedStateIndex uint32

	// pseudo-random permutation of peer indices. Serves a sequence in which peers are queried for state updates
	// the permutation is calculated taking last solid variable state hash as a seed
	permutationOfPeers []uint16

	// next peer permutationOfPeers[permutationIndex] is a next peer will be asked for ths state update
	permutationIndex uint16

	// the timeout deadline for sync inquiries
	syncMessageDeadline time.Time

	// current batch being synced
	syncedBatch *syncedBatch

	// logger
	log *logger.Logger
}

type syncedBatch struct {
	msgCounter   uint16
	stateIndex   uint32
	stateUpdates []state.StateUpdate
	statetxId    valuetransaction.ID
}

type pendingBatch struct {
	// batch of state updates, not validated yet
	batch state.Batch
	// resulting variable state after applied the batch to the solidVariableState
	nextVariableState state.VariableState
}

const nodeRequestTimeout = 2 * time.Second

func New(committee committee.Committee, log *logger.Logger) committee.StateManager {
	ret := &stateManager{
		committee:      committee,
		pendingBatches: make(map[hashing.HashValue]*pendingBatch),
		log:            log.Named("smgr"),
	}
	go ret.initLoadState()

	return ret
}

// initial loading of the solid state
func (sm *stateManager) initLoadState() {
	var err error
	var batch state.Batch

	stateExist, err := state.StateExist(sm.committee.Address())
	if err != nil {
		sm.log.Errorf("error occurred. sc addr %s: %v", sm.committee.Address().String(), err)
		sm.committee.Dismiss()
		return
	}

	if stateExist {
		// load last solid state and batch
		sm.solidVariableState, batch, err = state.LoadVariableState(sm.committee.Address())
		if err != nil {
			sm.log.Errorf("can't load variable state for sc addr %s: %v", sm.committee.Address().String(), err)
			sm.committee.Dismiss()
			return
		}
	} else {
		// origin state
		sm.solidVariableState = nil // por las dudas
		par := sm.committee.MetaData()
		batch = state.NewOriginBatch(state.NewOriginBatchParams{
			Address:      &par.Address,
			OwnerAddress: &par.OwnerAddress,
			Description:  par.Description,
			ProgramHash:  &par.ProgramHash,
		})
		sm.log.Infof("using meta data to create origin state update batch. Sc addr %s",
			sm.committee.Address().String())
	}
	// loaded solid variable state and the last batch of state updates
	// it needs to be validated by the state transaction, therefore it is added to the
	// pending batches and transaction is requested from the node
	if !sm.addPendingBatch(batch) {
		sm.log.Errorf("assertion failed: sm.addPendingBatch(stateUpdate)")
		sm.committee.Dismiss()
		return
	}
	if sm.solidVariableState == nil {
		if !(len(sm.pendingBatches) == 1) {
			panic("assertion: len(sm.pendingBatches) == 1")
		}
		// origin state
		for stateHash := range sm.pendingBatches {
			sm.log.Infof("origin state hash %s address %s", stateHash.String(), sm.committee.Address().String())
		}
	}
	// open msg queue for the committee
	sm.committee.OpenQueue()

	// request state transaction to approve last solid state
	txid := batch.StateTransactionId()
	nodeconn.RequestTransactionFromNode(&txid)
	sm.nodeRequestDeadline = time.Now().Add(nodeRequestTimeout)
}
