package sctransaction

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

// state block of the SC transaction. Represents SC state update
// previous state block can be determined by the chain transfer of the SC token in the UTXO part of the
// transaction
type StateBlock struct {
	// color of the SC which is updated
	// color contains balance.NEW_COLOR for the origin transaction
	color balance.Color
	// stata index is 0 for the origin transaction
	// consensus maintains incremental sequence of state indexes
	stateIndex uint32
	// timestamp of the transaction. 0 means transaction is not timestamped
	timestamp int64
	// requestId = tx hash + requestId index which originated this state update
	// the list is needed for batches of requests
	// this reference makes requestIds (inputs to state update) immutable part of the state update
	variableStateHash hashing.HashValue
}

type NewStateBlockParams struct {
	Color      balance.Color
	StateIndex uint32
	StateHash  hashing.HashValue
	Timestamp  int64
}

func NewStateBlock(par NewStateBlockParams) *StateBlock {
	return &StateBlock{
		color:             par.Color,
		stateIndex:        par.StateIndex,
		variableStateHash: par.StateHash,
		timestamp:         par.Timestamp,
	}
}

func (sb *StateBlock) Color() balance.Color {
	return sb.color
}

func (sb *StateBlock) StateIndex() uint32 {
	return sb.stateIndex
}

func (sb *StateBlock) Timestamp() int64 {
	return sb.timestamp
}

func (sb *StateBlock) VariableStateHash() hashing.HashValue {
	return sb.variableStateHash
}

func (sb *StateBlock) WithTimestamp(ts int64) *StateBlock {
	sb.timestamp = ts
	return sb
}

func (sb *StateBlock) WithVariableStateHash(h *hashing.HashValue) *StateBlock {
	sb.variableStateHash = *h
	return sb
}

// encoding
// important: each block starts with 65 bytes of scid

func (sb *StateBlock) Write(w io.Writer) error {
	if _, err := w.Write(sb.color[:]); err != nil {
		return err
	}
	if err := util.WriteUint32(w, sb.stateIndex); err != nil {
		return err
	}
	if err := util.WriteUint64(w, uint64(sb.timestamp)); err != nil {
		return err
	}
	if err := sb.variableStateHash.Write(w); err != nil {
		return err
	}
	return nil
}

func (sb *StateBlock) Read(r io.Reader) error {
	if n, err := r.Read(sb.color[:]); err != nil || n != balance.ColorLength {
		return fmt.Errorf("error while reading color: %v", err)
	}
	if err := util.ReadUint32(r, &sb.stateIndex); err != nil {
		return err
	}
	var timestamp uint64
	if err := util.ReadUint64(r, &timestamp); err != nil {
		return err
	}
	sb.timestamp = int64(timestamp)
	if err := sb.variableStateHash.Read(r); err != nil {
		return err
	}
	return nil
}
