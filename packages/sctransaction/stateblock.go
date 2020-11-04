package sctransaction

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

// StateSection of the SC transaction. Represents SC state update
// previous state block can be determined by the chain transfer of the SC token in the UTXO part of the
// transaction
type StateSection struct {
	// color of the chain which is updated
	// color contains balance.NEW_COLOR for the origin transaction
	color balance.Color
	// blockIndex is 0 for the origin transaction
	// consensus maintains incremental sequence of state indexes
	blockIndex uint32
	// timestamp of the transaction. 0 means transaction is not timestamped
	timestamp int64
	// stateHash is hash of the state it is locked in the transaction
	stateHash hashing.HashValue
}

type NewStateSectionParams struct {
	Color      balance.Color
	BlockIndex uint32
	StateHash  hashing.HashValue
	Timestamp  int64
}

func NewStateSection(par NewStateSectionParams) *StateSection {
	return &StateSection{
		color:      par.Color,
		blockIndex: par.BlockIndex,
		stateHash:  par.StateHash,
		timestamp:  par.Timestamp,
	}
}

func (sb *StateSection) Clone() *StateSection {
	if sb == nil {
		return nil
	}
	return NewStateSection(NewStateSectionParams{
		Color:      sb.color,
		BlockIndex: sb.blockIndex,
		StateHash:  sb.stateHash,
		Timestamp:  sb.timestamp,
	})
}

func (sb *StateSection) Color() balance.Color {
	return sb.color
}

func (sb *StateSection) BlockIndex() uint32 {
	return sb.blockIndex
}

func (sb *StateSection) Timestamp() int64 {
	return sb.timestamp
}

func (sb *StateSection) StateHash() hashing.HashValue {
	return sb.stateHash
}

func (sb *StateSection) WithStateParams(stateIndex uint32, h *hashing.HashValue, ts int64) *StateSection {
	sb.blockIndex = stateIndex
	sb.stateHash = *h
	sb.timestamp = ts
	return sb
}

// encoding

func (sb *StateSection) Write(w io.Writer) error {
	if _, err := w.Write(sb.color[:]); err != nil {
		return err
	}
	if err := util.WriteUint32(w, sb.blockIndex); err != nil {
		return err
	}
	if err := util.WriteUint64(w, uint64(sb.timestamp)); err != nil {
		return err
	}
	if err := sb.stateHash.Write(w); err != nil {
		return err
	}
	return nil
}

func (sb *StateSection) Read(r io.Reader) error {
	if n, err := r.Read(sb.color[:]); err != nil || n != balance.ColorLength {
		return fmt.Errorf("error while reading color: %v", err)
	}
	if err := util.ReadUint32(r, &sb.blockIndex); err != nil {
		return err
	}
	var timestamp uint64
	if err := util.ReadUint64(r, &timestamp); err != nil {
		return err
	}
	sb.timestamp = int64(timestamp)
	if err := sb.stateHash.Read(r); err != nil {
		return err
	}
	return nil
}
