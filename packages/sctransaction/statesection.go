package sctransaction

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

// StateSection of the SC transaction. Represents SC state update
// previous state block can be determined by the chain transfer of the SC token in the UTXO part of the
// transaction
type StateSection struct {
	// color of the chain. Will be changed in the future
	color ledgerstate.Color
	// blockIndex is 0 for the origin transaction
	// consensus maintains incremental sequence of state indexes
	blockIndex uint32
	// stateHash is hash of the state it is locked in the transaction
	stateHash hashing.HashValue
}

func NewStateSection(color ledgerstate.Color, blockIndex uint32, stateHash hashing.HashValue) *StateSection {
	return &StateSection{
		color:      color,
		blockIndex: blockIndex,
		stateHash:  stateHash,
	}
}

func (sb *StateSection) Clone() *StateSection {
	if sb == nil {
		return nil
	}
	return NewStateSection(sb.color, sb.blockIndex, sb.stateHash)
}

func (sb *StateSection) String() string {
	return fmt.Sprintf("[[color: %s block #: %d]]", sb.color.String(), sb.blockIndex)
}

func (sb *StateSection) Color() ledgerstate.Color {
	return sb.color
}

func (sb *StateSection) BlockIndex() uint32 {
	return sb.blockIndex
}

func (sb *StateSection) StateHash() hashing.HashValue {
	return sb.stateHash
}

func (sb *StateSection) WithStateParams(stateIndex uint32, h hashing.HashValue) *StateSection {
	sb.blockIndex = stateIndex
	sb.stateHash = h
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
	if err := sb.stateHash.Write(w); err != nil {
		return err
	}
	return nil
}

func (sb *StateSection) Read(r io.Reader) error {
	if err := util.ReadColor(r, &sb.color); err != nil {
		return fmt.Errorf("error while reading color: %v", err)
	}
	if err := util.ReadUint32(r, &sb.blockIndex); err != nil {
		return err
	}
	var timestamp uint64
	if err := util.ReadUint64(r, &timestamp); err != nil {
		return err
	}
	if err := sb.stateHash.Read(r); err != nil {
		return err
	}
	return nil
}
