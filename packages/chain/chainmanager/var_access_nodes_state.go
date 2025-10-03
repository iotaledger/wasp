package chainmanager

import (
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
)

// AccessNodeStateVariable tracks the active state at the access nodes. If this node is part of the committee,
// then the tip tracked by this node should be ignored and the state tracked by the
// committee should be used. The algorithm itself is similar to the `localViewVariable`
// in the `committeeLog`.
type AccessNodeStateVariable interface {
	Tip() *isc.StateAnchor
	// Considers the produced (not yet confirmed) block / TX and returns new tip Anchor.
	// The returned bool indicates if the tip has changed because of this call.
	// This function should return L1 commitment, if the corresponding block should be added to the store.
	BlockProduced(tx *iotasigner.SignedTransaction) (*isc.StateAnchor, bool, *state.L1Commitment)
	// Considers a confirmed Anchor and returns new tip Anchor.
	// The returned bool indicates if the tip has changed because of this call.
	BlockConfirmed(ao *isc.StateAnchor) (*isc.StateAnchor, bool)
}

type accessNodeStateVariableImpl struct {
	chainID   isc.ChainID
	tipAnchor *isc.StateAnchor
	log       log.Logger
}

func NewAccessNodeStateVariable(chainID isc.ChainID, log log.Logger) AccessNodeStateVariable {
	return &accessNodeStateVariableImpl{
		chainID:   chainID,
		tipAnchor: nil,
		log:       log,
	}
}

func (v *accessNodeStateVariableImpl) Tip() *isc.StateAnchor {
	return v.tipAnchor
}

// TODO: Probably this function can be removed at all. This left from the pipelining.
func (v *accessNodeStateVariableImpl) BlockProduced(tx *iotasigner.SignedTransaction) (*isc.StateAnchor, bool, *state.L1Commitment) {
	v.log.LogDebugf("BlockProduced: tx=%v", tx)
	return v.tipAnchor, false, nil
}

func (v *accessNodeStateVariableImpl) BlockConfirmed(confirmed *isc.StateAnchor) (*isc.StateAnchor, bool) {
	v.log.LogDebugf("BlockConfirmed: confirmed=%v", confirmed)
	return v.outputIfChanged(confirmed)
}

func (v *accessNodeStateVariableImpl) outputIfChanged(newTip *isc.StateAnchor) (*isc.StateAnchor, bool) {
	if v.tipAnchor == nil && newTip == nil {
		v.log.LogDebugf("⊳ Tip remains nil.")
		return v.tipAnchor, false
	}
	if newTip == nil {
		v.log.LogDebugf("⊳ Tip remains %v, new candidate was nil.", v.tipAnchor)
		return v.tipAnchor, false
	}
	if v.tipAnchor == nil {
		v.log.LogDebugf("⊳ New tip=%v, was %v", newTip, v.tipAnchor)
		v.tipAnchor = newTip
		return v.tipAnchor, true
	}
	if v.tipAnchor.Equals(newTip) {
		v.log.LogDebugf("⊳ Tip remains %v.", v.tipAnchor)
		return v.tipAnchor, false
	}
	v.log.LogDebugf("⊳ New tip=%v, was %v", newTip, v.tipAnchor)
	v.tipAnchor = newTip
	return v.tipAnchor, true
}
