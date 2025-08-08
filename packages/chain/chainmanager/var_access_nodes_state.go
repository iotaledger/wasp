package chainmanager

import (
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
)

// VarAccessNodeState tracks the active state at the access nodes. If this node is part of the committee,
// then the tip tracked by this node should be ignored and the state tracked by the
// committee should be used. The algorithm itself is similar to the `varLocalView`
// in the `cmtLog`.
type VarAccessNodeState interface {
	Tip() *isc.StateAnchor
	// Considers the produced (not yet confirmed) block / TX and returns new tip AO.
	// The returned bool indicates if the tip has changed because of this call.
	// This function should return L1 commitment, if the corresponding block should be added to the store.
	BlockProduced(tx *iotasigner.SignedTransaction) (*isc.StateAnchor, bool, *state.L1Commitment)
	// Considers a confirmed AO and returns new tip AO.
	// The returned bool indicates if the tip has changed because of this call.
	BlockConfirmed(ao *isc.StateAnchor) (*isc.StateAnchor, bool)
}

type varAccessNodeStateImpl struct {
	tipAO *isc.StateAnchor
	log   log.Logger
}

func NewVarAccessNodeState(log log.Logger) VarAccessNodeState {
	return &varAccessNodeStateImpl{
		tipAO: nil,
		log:   log,
	}
}

func (vas *varAccessNodeStateImpl) Tip() *isc.StateAnchor {
	return vas.tipAO
}

// TODO: Probably this function can be removed at all. This left from the pipelining.
func (vas *varAccessNodeStateImpl) BlockProduced(tx *iotasigner.SignedTransaction) (*isc.StateAnchor, bool, *state.L1Commitment) {
	vas.log.LogDebugf("BlockProduced: tx=%v", tx)
	return vas.tipAO, false, nil
}

func (vas *varAccessNodeStateImpl) BlockConfirmed(confirmed *isc.StateAnchor) (*isc.StateAnchor, bool) {
	vas.log.LogDebugf("BlockConfirmed: confirmed=%v", confirmed)
	return vas.outputIfChanged(confirmed)
}

func (vas *varAccessNodeStateImpl) outputIfChanged(newTip *isc.StateAnchor) (*isc.StateAnchor, bool) {
	if vas.tipAO == nil && newTip == nil {
		vas.log.LogDebugf("⊳ Tip remains nil.")
		return vas.tipAO, false
	}
	if newTip == nil {
		vas.log.LogDebugf("⊳ Tip remains %v, new candidate was nil.", vas.tipAO)
		return vas.tipAO, false
	}
	if vas.tipAO == nil {
		vas.log.LogDebugf("⊳ New tip=%v, was %v", newTip, vas.tipAO)
		vas.tipAO = newTip
		return vas.tipAO, true
	}
	if vas.tipAO.Equals(newTip) {
		vas.log.LogDebugf("⊳ Tip remains %v.", vas.tipAO)
		return vas.tipAO, false
	}
	vas.log.LogDebugf("⊳ New tip=%v, was %v", newTip, vas.tipAO)
	vas.tipAO = newTip
	return vas.tipAO, true
}
