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
// in the `committeeLog`.
type VarAccessNodeState struct {
	chainID   isc.ChainID
	tipAnchor *isc.StateAnchor
	log       log.Logger
}

func NewVarAccessNodeState(chainID isc.ChainID, log log.Logger) *VarAccessNodeState {
	return &VarAccessNodeState{
		chainID:   chainID,
		tipAnchor: nil,
		log:       log,
	}
}

func (vas *VarAccessNodeState) Tip() *isc.StateAnchor {
	return vas.tipAnchor
}

// BlockProduced
// TODO: Probably this function can be removed at all. This left from the pipelining.
func (vas *VarAccessNodeState) BlockProduced(tx *iotasigner.SignedTransaction) (*isc.StateAnchor, bool, *state.L1Commitment) {
	vas.log.LogDebugf("BlockProduced: tx=%v", tx)
	return vas.tipAnchor, false, nil
}

func (vas *VarAccessNodeState) BlockConfirmed(confirmed *isc.StateAnchor) (*isc.StateAnchor, bool) {
	vas.log.LogDebugf("BlockConfirmed: confirmed=%v", confirmed)
	return vas.outputIfChanged(confirmed)
}

func (vas *VarAccessNodeState) outputIfChanged(newTip *isc.StateAnchor) (*isc.StateAnchor, bool) {
	if vas.tipAnchor == nil && newTip == nil {
		vas.log.LogDebugf("⊳ Tip remains nil.")
		return vas.tipAnchor, false
	}
	if newTip == nil {
		vas.log.LogDebugf("⊳ Tip remains %v, new candidate was nil.", vas.tipAnchor)
		return vas.tipAnchor, false
	}
	if vas.tipAnchor == nil {
		vas.log.LogDebugf("⊳ New tip=%v, was %v", newTip, vas.tipAnchor)
		vas.tipAnchor = newTip
		return vas.tipAnchor, true
	}
	if vas.tipAnchor.Equals(newTip) {
		vas.log.LogDebugf("⊳ Tip remains %v.", vas.tipAnchor)
		return vas.tipAnchor, false
	}
	vas.log.LogDebugf("⊳ New tip=%v, was %v", newTip, vas.tipAnchor)
	vas.tipAnchor = newTip
	return vas.tipAnchor, true
}
