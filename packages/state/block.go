package state

import (
	"bytes"
	"fmt"
	"golang.org/x/xerrors"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

type block struct {
	stateOutputID ledgerstate.OutputID
	stateUpdates  []*stateUpdate
	blockIndex    uint32 // not persistent
}

// validates, enumerates and creates a block from array of state updates
func NewBlock(stateUpdates ...StateUpdate) (*block, error) {
	arr := make([]*stateUpdate, len(stateUpdates))
	for i := range arr {
		arr[i] = stateUpdates[i].Clone().(*stateUpdate)
	}
	ret := &block{
		stateUpdates: arr,
	}
	var err error
	if ret.blockIndex, err = findBlockIndexMutation(arr); err != nil {
		return nil, err
	}
	return ret, nil
}

func BlockFromBytes(data []byte) (*block, error) {
	ret := new(block)
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, xerrors.Errorf("BlockFromBytes: %w", err)
	}
	var err error
	if ret.blockIndex, err = findBlockIndexMutation(ret.stateUpdates); err != nil {
		return nil, err
	}
	return ret, nil
}

func LoadBlockBytes(partition kvstore.KVStore, stateIndex uint32) ([]byte, error) {
	data, err := partition.Get(dbprovider.MakeKey(dbprovider.ObjectTypeBlock, util.Uint32To4Bytes(stateIndex)))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func LoadBlock(partition kvstore.KVStore, stateIndex uint32) (*block, error) {
	data, err := LoadBlockBytes(partition, stateIndex)
	if err != nil {
		return nil, err
	}
	return BlockFromBytes(data)
}

// block with empty state update and nil state hash
func NewOriginBlock() *block {
	ret, err := NewBlock(NewStateUpdateWithBlockIndexMutation(0, time.Time{}))
	if err != nil {
		panic(err)
	}
	return ret
}

func (b *block) Bytes() []byte {
	var buf bytes.Buffer
	_ = b.Write(&buf)
	return buf.Bytes()
}

func (b *block) String() string {
	ret := ""
	ret += fmt.Sprintf("Block: state index: %d\n", b.BlockIndex())
	ret += fmt.Sprintf("state txid: %s\n", b.ApprovingOutputID().String())
	ret += fmt.Sprintf("timestamp: %v\n", b.Timestamp())
	ret += fmt.Sprintf("size: %d\n", b.Size())
	ret += fmt.Sprintf("essence: %s\n", b.EssenceHash().String())
	for i, su := range b.stateUpdates {
		ret += fmt.Sprintf("   #%d: %s\n", i, su.String())
	}
	return ret
}

func (b *block) ApprovingOutputID() ledgerstate.OutputID {
	return b.stateOutputID
}

func (b *block) BlockIndex() uint32 {
	return b.blockIndex
}

// Timestamp of the last state update
func (b *block) Timestamp() time.Time {
	ts, err := findTimestampMutation(b.stateUpdates)
	if err != nil {
		panic(err)
	}
	return ts
}

func (b *block) WithApprovingOutputID(vtxid ledgerstate.OutputID) Block {
	b.stateOutputID = vtxid
	return b
}

func (b *block) ForEach(fun func(uint16, StateUpdate) bool) {
	for i, su := range b.stateUpdates {
		if !fun(uint16(i), su) {
			return
		}
	}
}

func (b *block) Size() uint16 {
	return uint16(len(b.stateUpdates))
}

// hash of all data except state transaction hash
func (b *block) EssenceHash() hashing.HashValue {
	var buf bytes.Buffer
	if err := b.writeEssence(&buf); err != nil {
		panic("EssenceHash")
	}
	return hashing.HashData(buf.Bytes())
}

func (b *block) Write(w io.Writer) error {
	if err := b.writeEssence(w); err != nil {
		return err
	}
	if _, err := w.Write(b.stateOutputID.Bytes()); err != nil {
		return err
	}
	return nil
}

func (b *block) writeEssence(w io.Writer) error {
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

func (b *block) Read(r io.Reader) error {
	if err := b.readEssence(r); err != nil {
		return err
	}
	if _, err := r.Read(b.stateOutputID[:]); err != nil {
		return err
	}
	return nil
}

func (b *block) readEssence(r io.Reader) error {
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	b.stateUpdates = make([]*stateUpdate, size)
	var err error
	for i := range b.stateUpdates {
		b.stateUpdates[i], err = newStateUpdateFromReader(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *block) IsApprovedBy(chainOutput *ledgerstate.AliasOutput) bool {
	if chainOutput == nil {
		return false
	}
	if b.BlockIndex() != chainOutput.GetStateIndex() {
		return false
	}
	var nilOID ledgerstate.OutputID
	if b.ApprovingOutputID() != nilOID && b.ApprovingOutputID() != chainOutput.ID() {
		return false
	}
	sh, err := hashing.HashValueFromBytes(chainOutput.GetStateData())
	if err != nil {
		return false
	}
	if b.EssenceHash() != sh {
		return false
	}
	return true
}
