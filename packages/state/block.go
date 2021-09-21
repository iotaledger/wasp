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
	stateUpdates  []*stateUpdateImpl
	blockIndex    uint32 // not persistent
}

// validates, enumerates and creates a block from array of state updates
func newBlock(stateUpdates ...StateUpdate) (Block, error) {
	arr := make([]*stateUpdateImpl, len(stateUpdates))
	for i := range arr {
		arr[i] = stateUpdates[i].(*stateUpdateImpl) // do not clone
	}
	ret := &blockImpl{
		stateUpdates: arr,
	}
	var err error
	if ret.blockIndex, err = findBlockIndexMutation(arr); err != nil {
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
	if ret.blockIndex, err = findBlockIndexMutation(ret.stateUpdates); err != nil {
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
	ret += fmt.Sprintf("size: %d\n", b.size())
	for i, su := range b.stateUpdates {
		ret += fmt.Sprintf("   #%d: %s\n", i, su.String())
	}
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
	ts, err := findTimestampMutation(b.stateUpdates)
	if err != nil {
		panic(err)
	}
	return ts
}

// PreviousStateHash of the last state update
func (b *blockImpl) PreviousStateHash() hashing.HashValue {
	ph, err := findPrevStateHashMutation(b.stateUpdates)
	if err != nil {
		panic(err)
	}
	return ph
}

func (b *blockImpl) SetApprovingOutputID(oid ledgerstate.OutputID) {
	b.stateOutputID = oid
}

func (b *blockImpl) size() uint16 {
	return uint16(len(b.stateUpdates))
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

// compactStateUpdate returns a stateUpdate that contains all mutations in the block
// (this is useful to trim down the binary representation of the block, because mutations to the same key on
// different stateUpdates will be collapsed into a single one)
func (b *blockImpl) compactStateUpdate() *stateUpdateImpl {
	ret := NewStateUpdate()
	for _, su := range b.stateUpdates {
		for k, v := range su.mutations.Sets {
			ret.mutations.Set(k, v)
		}
		for k := range su.mutations.Dels {
			ret.mutations.Del(k)
		}
	}
	return ret
}

func (b *blockImpl) writeEssence(w io.Writer) error {
	return b.compactStateUpdate().Write(w)
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
	b.stateUpdates = make([]*stateUpdateImpl, 1)
	var err error
	b.stateUpdates[0], err = newStateUpdateFromReader(r)
	if err != nil {
		return err
	}
	return nil
}
