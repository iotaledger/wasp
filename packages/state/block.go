package state

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type BlockImpl struct {
	stateOutputID ledgerstate.OutputID
	stateUpdates  []*StateUpdateImpl
	blockIndex    uint32 // not persistent
}

// validates, enumerates and creates a block from array of state updates
func newBlock(stateUpdates ...StateUpdate) (*BlockImpl, error) {
	arr := make([]*StateUpdateImpl, len(stateUpdates))
	for i := range arr {
		arr[i] = stateUpdates[i].(*StateUpdateImpl) // do not clone
	}
	ret := &BlockImpl{
		stateUpdates: arr,
	}
	var err error
	if ret.blockIndex, err = findBlockIndexMutation(arr); err != nil {
		return nil, err
	}
	return ret, nil
}

func BlockFromBytes(data []byte) (*BlockImpl, error) {
	ret := new(BlockImpl)
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
func newOriginBlock() *BlockImpl {
	ret, err := newBlock(NewStateUpdateWithBlockIndexMutation(0, time.Time{}))
	if err != nil {
		panic(err)
	}
	return ret
}

func (b *BlockImpl) Bytes() []byte {
	var buf bytes.Buffer
	_ = b.Write(&buf)
	return buf.Bytes()
}

func (b *BlockImpl) String() string {
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

func (b *BlockImpl) ApprovingOutputID() ledgerstate.OutputID {
	return b.stateOutputID
}

func (b *BlockImpl) BlockIndex() uint32 {
	return b.blockIndex
}

// Timestamp of the last state update
func (b *BlockImpl) Timestamp() time.Time {
	ts, err := findTimestampMutation(b.stateUpdates)
	if err != nil {
		panic(err)
	}
	return ts
}

func (b *BlockImpl) SetApprovingOutputID(oid ledgerstate.OutputID) {
	b.stateOutputID = oid
}

func (b *BlockImpl) Size() uint16 {
	return uint16(len(b.stateUpdates))
}

// hash of all data except state transaction hash
func (b *BlockImpl) EssenceBytes() []byte {
	var buf bytes.Buffer
	if err := b.writeEssence(&buf); err != nil {
		panic("EssenceBytes")
	}
	return buf.Bytes()
}

func (b *BlockImpl) Write(w io.Writer) error {
	if err := b.writeEssence(w); err != nil {
		return err
	}
	if _, err := w.Write(b.stateOutputID.Bytes()); err != nil {
		return err
	}
	return nil
}

func (b *BlockImpl) writeEssence(w io.Writer) error {
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

func (b *BlockImpl) Read(r io.Reader) error {
	if err := b.readEssence(r); err != nil {
		return err
	}
	if _, err := r.Read(b.stateOutputID[:]); err != nil {
		return err
	}
	return nil
}

func (b *BlockImpl) readEssence(r io.Reader) error {
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	b.stateUpdates = make([]*StateUpdateImpl, size)
	var err error
	for i := range b.stateUpdates {
		b.stateUpdates[i], err = newStateUpdateFromReader(r)
		if err != nil {
			return err
		}
	}
	return nil
}
