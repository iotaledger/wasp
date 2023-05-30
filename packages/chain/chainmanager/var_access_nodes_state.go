package chainmanager

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
)

// Tracks the active state at the access nodes. If this node is part of the committee,
// then the tip tracked by this node should be ignored and the state tracked by the
// committee should be used. The algorithm itself is similar to the `varLocalView`
// in the `cmtLog`.
type VarAccessNodeState interface {
	Tip() *isc.AliasOutputWithID
	// Considers the produced (not yet confirmed) block / TX and returns new tip AO.
	// The returned bool indicates if the tip has changed because of this call.
	// This function should return L1 commitment, if the corresponding block should be added to the store.
	BlockProduced(tx *iotago.Transaction) (*isc.AliasOutputWithID, bool, *state.L1Commitment)
	// Considers a confirmed AO and returns new tip AO.
	// The returned bool indicates if the tip has changed because of this call.
	BlockConfirmed(ao *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool)
}

type varAccessNodeStateImpl struct {
	chainID   isc.ChainID
	tipAO     *isc.AliasOutputWithID                                         // Will point to the latest known good state while the chain don't have the current good state.
	confirmed *isc.AliasOutputWithID                                         // Latest known confirmed AO.
	pending   *shrinkingmap.ShrinkingMap[uint32, []*varAccessNodeStateEntry] // A set of unconfirmed outputs (StateIndex => TX).
	log       *logger.Logger                                                 // Will write this just for the alignment.
}

type varAccessNodeStateEntry struct {
	output   *isc.AliasOutputWithID // The published AO.
	consumed iotago.OutputID        // The AO used as an input for the TX.
}

func NewVarAccessNodeState(chainID isc.ChainID, log *logger.Logger) VarAccessNodeState {
	return &varAccessNodeStateImpl{
		chainID:   chainID,
		tipAO:     nil,
		confirmed: nil,
		pending:   shrinkingmap.New[uint32, []*varAccessNodeStateEntry](),
		log:       log,
	}
}

func (vas *varAccessNodeStateImpl) Tip() *isc.AliasOutputWithID {
	return vas.tipAO
}

func (vas *varAccessNodeStateImpl) BlockProduced(tx *iotago.Transaction) (*isc.AliasOutputWithID, bool, *state.L1Commitment) {
	txID, err := tx.ID()
	if err != nil {
		vas.log.Debugf("BlockProduced: Ignoring, cannot extract txID: %v", err)
		return vas.tipAO, false, nil
	}
	consumed, published, err := vas.extractConsumedPublished(tx)
	if err != nil {
		vas.log.Debugf("BlockProduced(tx.ID=%v): Ignoring because of %v", txID, err)
		return vas.tipAO, false, nil
	}
	//
	vas.log.Debugf("BlockProduced: consumed.ID=%v, published=%v", consumed.ToHex(), published)
	stateIndex := published.GetStateIndex()
	//
	// Add it to the pending list.
	var entries []*varAccessNodeStateEntry
	entries, ok := vas.pending.Get(stateIndex)
	if !ok {
		entries = []*varAccessNodeStateEntry{}
	}
	publishedL1Commitment, err := transaction.L1CommitmentFromAliasOutput(published.GetAliasOutput())
	if err != nil {
		vas.log.Warnf("Cannot extract L1Commitment from the published AO: %v", err)
		publishedL1Commitment = nil // Will ignore it.
	}
	if lo.ContainsBy(entries, func(e *varAccessNodeStateEntry) bool { return e.output.Equals(published) }) {
		vas.log.Debugf("⊳ Ignoring it, duplicate.")
		return vas.tipAO, false, nil
	}
	entries = append(entries, &varAccessNodeStateEntry{
		output:   published,
		consumed: consumed,
	})
	vas.pending.Set(stateIndex, entries)
	//
	// Check, if the added AO is a new tip for the chain.
	if published.Equals(vas.findLatestPending()) {
		vas.log.Debugf("⊳ Will consider consensusOutput=%v as a tip, the current confirmed=%v.", published, vas.confirmed)
		changedAO, changed := vas.outputIfChanged(published)
		return changedAO, changed, publishedL1Commitment
	}
	vas.log.Debugf("⊳ That's not a tip.")
	return vas.tipAO, false, publishedL1Commitment
}

func (vas *varAccessNodeStateImpl) BlockConfirmed(confirmed *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool) {
	vas.log.Debugf("BlockConfirmed: confirmed=%v", confirmed)
	stateIndex := confirmed.GetStateIndex()
	vas.confirmed = confirmed
	if vas.isAliasOutputPending(confirmed) {
		vas.pending.ForEach(func(si uint32, es []*varAccessNodeStateEntry) bool {
			if si <= stateIndex {
				for _, e := range es {
					vas.log.Debugf("⊳ Removing[%v≤%v] %v", si, stateIndex, e.output)
				}
				vas.pending.Delete(si)
			}
			return true
		})
	} else {
		vas.pending.ForEach(func(si uint32, es []*varAccessNodeStateEntry) bool {
			for _, e := range es {
				vas.log.Debugf("⊳ Removing[all] %v", e.output)
			}
			vas.pending.Delete(si)
			return true
		})
	}
	return vas.outputIfChanged(vas.findLatestPending())
}

func (vas *varAccessNodeStateImpl) outputIfChanged(newTip *isc.AliasOutputWithID) (*isc.AliasOutputWithID, bool) {
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

func (vas *varAccessNodeStateImpl) isAliasOutputPending(ao *isc.AliasOutputWithID) bool {
	found := false
	vas.pending.ForEach(func(si uint32, es []*varAccessNodeStateEntry) bool {
		found = lo.ContainsBy(es, func(e *varAccessNodeStateEntry) bool {
			return e.output.Equals(ao)
		})
		return !found
	})
	return found
}

func (vas *varAccessNodeStateImpl) findLatestPending() *isc.AliasOutputWithID {
	if vas.confirmed == nil {
		return nil
	}
	latest := vas.confirmed
	confirmedSI := vas.confirmed.GetStateIndex()
	pendingSICount := uint32(vas.pending.Size())
	for i := uint32(0); i < pendingSICount; i++ {
		entries, ok := vas.pending.Get(confirmedSI + i + 1)
		if !ok {
			return nil // That's a gap.
		}
		if len(entries) != 1 {
			return nil // Alternatives exist.
		}
		if latest.OutputID() != entries[0].consumed {
			return nil // Don't form a chain.
		}
		latest = entries[0].output
	}
	return latest
}

func (vas *varAccessNodeStateImpl) extractConsumedPublished(tx *iotago.Transaction) (iotago.OutputID, *isc.AliasOutputWithID, error) {
	var consumed iotago.OutputID
	var published *isc.AliasOutputWithID
	var err error
	if vas.confirmed == nil {
		return consumed, nil, fmt.Errorf("don't have the confirmed AO")
	}
	if err = vas.verifyTxSignature(tx, vas.confirmed.GetStateAddress()); err != nil {
		return consumed, nil, fmt.Errorf("cannot validate tx: %v", err)
	}
	//
	// Validate the TX:
	//   - Signature is valid and is by the latest known confirmed state controller.
	//   - Previous known AO is among the TX inputs.
	published, err = isc.AliasOutputWithIDFromTx(tx, vas.chainID.AsAddress())
	if err != nil {
		return consumed, nil, fmt.Errorf("cannot extract alias output from the block: %v", err)
	}
	if published == nil {
		return consumed, nil, fmt.Errorf("extracted nil AO from the TX, something wrong")
	}
	//
	// Get potential inputs.
	publishedSI := published.GetStateIndex()
	confirmedSI := vas.confirmed.GetStateIndex()
	if publishedSI <= confirmedSI {
		return consumed, nil, fmt.Errorf("outdated, confirmedSI=%v, received %v", publishedSI, publishedSI)
	}
	haveOutputs := map[iotago.OutputID]struct{}{}
	if publishedSI == confirmedSI+1 {
		haveOutputs[vas.confirmed.OutputID()] = struct{}{}
	} else {
		entries, found := vas.pending.Get(publishedSI - 1)
		if !found {
			return consumed, nil, fmt.Errorf("there is no outputs with prev SI")
		}
		for _, entry := range entries {
			haveOutputs[entry.output.OutputID()] = struct{}{}
		}
	}
	//
	// Check if we have TX input corresponding to some candidates we already know.
	consumedFound := false
	for _, input := range tx.Essence.Inputs {
		if input.Type() != iotago.InputUTXO {
			continue
		}
		utxoInp, ok := input.(*iotago.UTXOInput)
		if !ok {
			continue
		}
		utxoInpOID := utxoInp.ID()
		if _, ok := haveOutputs[utxoInpOID]; ok {
			if consumedFound {
				return consumed, nil, fmt.Errorf("found more that 1 output that is consumed")
			}
			consumed = utxoInpOID
			consumedFound = true
		}
	}
	if !consumedFound {
		return consumed, nil, fmt.Errorf("found no known outputs as consumed")
	}
	return consumed, published, nil
}

func (vas *varAccessNodeStateImpl) verifyTxSignature(tx *iotago.Transaction, stateController iotago.Address) error {
	signingMessage, err := tx.Essence.SigningMessage()
	if err != nil {
		return fmt.Errorf("cannot extract signing message: %w", err)
	}

	for _, unlock := range tx.Unlocks {
		signatureUnlock, ok := unlock.(*iotago.SignatureUnlock)
		if !ok {
			continue
		}

		ed25519Signature, ok := signatureUnlock.Signature.(*iotago.Ed25519Signature)
		if !ok {
			continue
		}

		ed25519SignatureBy := iotago.Ed25519AddressFromPubKey(ed25519Signature.PublicKey[:])

		if !ed25519SignatureBy.Equal(stateController) {
			continue
		}

		if err := ed25519Signature.Valid(signingMessage, &ed25519SignatureBy); err != nil {
			return fmt.Errorf("signature by stateController invalid: %w", err)
		}
		return nil
	}

	return fmt.Errorf("signature by stateController %v not found", stateController)
}
