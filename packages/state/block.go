package state

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/iotaledger/wasp/packages/hashing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type blockImpl struct {
	stateOutputID ledgerstate.OutputID
	stateUpdates  []*stateUpdateImpl
	blockIndex    uint32 // not persistent
}

// validates, enumerates and creates a block from array of state updates
func newBlock(stateUpdates ...StateUpdate) (*blockImpl, error) {
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

func BlockFromBytes(data []byte) (*blockImpl, error) {
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
func newOriginBlock() *blockImpl {
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
	ret += fmt.Sprintf("size: %d\n", b.Size())
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

func (b *blockImpl) Size() uint16 {
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

func (b *blockImpl) writeEssence(w io.Writer) error {
	if err := util.WriteUint16(w, uint16(len(b.stateUpdates))); err != nil {
		return err
	}
	for _, su := range b.stateUpdates {
		if err := su.Write(w); err != nil {
			return err
		}
	}
	return nil
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
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	b.stateUpdates = make([]*stateUpdateImpl, size)
	var err error
	for i := range b.stateUpdates {
		b.stateUpdates[i], err = newStateUpdateFromReader(r)
		if err != nil {
			return err
		}
	}
	return nil
}
