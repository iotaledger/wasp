package state

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
)

type blockImpl struct {
	stateOutputID ledgerstate.OutputID
	stateUpdate   *stateUpdateImpl
	blockIndex    uint32 // not persistent
}

// validates, enumerates and creates a block from array of state updates
func newBlock(stateUpdate StateUpdate) (Block, error) {
	ret := &blockImpl{stateUpdate: &stateUpdateImpl{
		mutations: stateUpdate.Mutations(),
	}}
	var err error
	if ret.blockIndex, err = findBlockIndexMutation(ret.stateUpdate); err != nil {
		return nil, err
	}
	return ret, nil
}

func BlockFromBytes(data []byte) (Block, error) {
	ret := new(blockImpl)
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, xerrors.Errorf("BlockFromBytes: %w", err)
	}
	var err error
	if ret.blockIndex, err = findBlockIndexMutation(ret.stateUpdate); err != nil {
		return nil, err
	}
	return ret, nil
}

// block with empty state update and nil state hash
func newOriginBlock() Block {
	ret, err := newBlock(NewStateUpdateWithBlocklogValues(0, time.Time{}, hashing.NilHash))
	if err != nil {
		panic(err)
	}
	return ret
}

func (b *blockImpl) Bytes() []byte {
	var buf bytes.Buffer
	_ = b.Write(&buf)
	return buf.Bytes()
}

func (b *blockImpl) String() string {
	ret := ""
	ret += fmt.Sprintf("Block: state index: %d\n", b.BlockIndex())
	ret += fmt.Sprintf("state txid: %s\n", b.ApprovingOutputID().String())
	ret += fmt.Sprintf("timestamp: %v\n", b.Timestamp())
	ret += fmt.Sprintf("state update: %s\n", (*b.stateUpdate).String())
	return ret
}

func (b *blockImpl) ApprovingOutputID() ledgerstate.OutputID {
	return b.stateOutputID
}

func (b *blockImpl) BlockIndex() uint32 {
	return b.blockIndex
}

// Timestamp of the last state update
func (b *blockImpl) Timestamp() time.Time {
	ts, err := findTimestampMutation(b.stateUpdate)
	if err != nil {
		panic(err)
	}
	return ts
}

// PreviousStateHash of the last state update
func (b *blockImpl) PreviousStateHash() hashing.HashValue {
	ph, err := findPrevStateHashMutation(b.stateUpdate)
	if err != nil {
		panic(err)
	}
	return ph
}

func (b *blockImpl) SetApprovingOutputID(oid ledgerstate.OutputID) {
	b.stateOutputID = oid
}

// hash of all data except state transaction hash
func (b *blockImpl) EssenceBytes() []byte {
	var buf bytes.Buffer
	if err := b.writeEssence(&buf); err != nil {
		panic("EssenceBytes")
	}
	return buf.Bytes()
}

func (b *blockImpl) Write(w io.Writer) error {
	if err := b.writeEssence(w); err != nil {
		return err
	}
	if _, err := w.Write(b.stateOutputID.Bytes()); err != nil {
		return err
	}
	return nil
}

func (b *blockImpl) writeEssence(w io.Writer) error {
	return b.stateUpdate.Write(w)
}

func (b *blockImpl) Read(r io.Reader) error {
	if err := b.readEssence(r); err != nil {
		return err
	}
	if _, err := r.Read(b.stateOutputID[:]); err != nil {
		return err
	}
	return nil
}

func (b *blockImpl) readEssence(r io.Reader) error {
	var err error
	b.stateUpdate, err = newStateUpdateFromReader(r)
	if err != nil {
		return err
	}
	return nil
}
