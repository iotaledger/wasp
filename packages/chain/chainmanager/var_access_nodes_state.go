package chainmanager

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

// Tracks the active state at the access nodes. If this node is part of the committee,
// then the tip tracked by this node should be ignored and the state tracked by the
// committee should be used. The algorithm itself is similar to the `varLocalView`
// in the `cmtLog`.
type VarAccessNodeState interface {
	Tip() *isc.StateAnchor
	// Considers the produced (not yet confirmed) block / TX and returns new tip AO.
	// The returned bool indicates if the tip has changed because of this call.
	// This function should return L1 commitment, if the corresponding block should be added to the store.
	BlockProduced(tx *suisigner.SignedTransaction) (*isc.StateAnchor, bool, *state.L1Commitment)
	// Considers a confirmed AO and returns new tip AO.
	// The returned bool indicates if the tip has changed because of this call.
	BlockConfirmed(ao *isc.StateAnchor) (*isc.StateAnchor, bool)
}

type varAccessNodeStateImpl struct {
	chainID isc.ChainID
	tipAO   *isc.StateAnchor
	log     *logger.Logger
}

type varAccessNodeStateEntry struct {
	output   *isc.StateAnchor // The published AO.
	consumed sui.ObjectID     // The AO used as an input for the TX.
}

func NewVarAccessNodeState(chainID isc.ChainID, log *logger.Logger) VarAccessNodeState {
	return &varAccessNodeStateImpl{
		chainID: chainID,
		tipAO:   nil,
		log:     log,
	}
}

func (vas *varAccessNodeStateImpl) Tip() *isc.StateAnchor {
	return vas.tipAO
}

// TODO: Probably this function can be removed at all. This left from the pipelining.
func (vas *varAccessNodeStateImpl) BlockProduced(tx *suisigner.SignedTransaction) (*isc.StateAnchor, bool, *state.L1Commitment) {
	vas.log.Debugf("BlockProduced: tx=%v", tx)
	return vas.tipAO, false, nil
}

func (vas *varAccessNodeStateImpl) BlockConfirmed(confirmed *isc.StateAnchor) (*isc.StateAnchor, bool) {
	vas.log.Debugf("BlockConfirmed: confirmed=%v", confirmed)
	return vas.outputIfChanged(confirmed)
}

func (vas *varAccessNodeStateImpl) outputIfChanged(newTip *isc.StateAnchor) (*isc.StateAnchor, bool) {
	if vas.tipAO == nil && newTip == nil {
		vas.log.Debugf("⊳ Tip remains nil.")
		return vas.tipAO, false
	}
	if newTip == nil {
		vas.log.Debugf("⊳ Tip remains %v, new candidate was nil.", vas.tipAO)
		return vas.tipAO, false
	}
	if vas.tipAO == nil {
		vas.log.Debugf("⊳ New tip=%v, was %v", newTip, vas.tipAO)
		vas.tipAO = newTip
		return vas.tipAO, true
	}
	if vas.tipAO.Equals(newTip) {
		vas.log.Debugf("⊳ Tip remains %v.", vas.tipAO)
		return vas.tipAO, false
	}
	vas.log.Debugf("⊳ New tip=%v, was %v", newTip, vas.tipAO)
	vas.tipAO = newTip
	return vas.tipAO, true
}
