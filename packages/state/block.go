package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
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
	blockIndex    uint32 // non-persistent
	stateOutputID ledgerstate.OutputID
	stateUpdates  []*stateUpdate
}

// validates, enumerates and creates a block from array of state updates
func NewBlock(blockIndex uint32, stateUpdates ...StateUpdate) *block {
	arr := make([]*stateUpdate, len(stateUpdates)+1)
	for i := 0; i < len(arr)-1; i++ {
		arr[i] = stateUpdates[i].Clone().(*stateUpdate)
	}
	arr[len(arr)-1] = NewStateUpdate()
	arr[len(arr)-1].setBlockIndexMutation(blockIndex)
	return &block{
		blockIndex:   blockIndex,
		stateUpdates: arr,
	}
}

func BlockFromBytes(data []byte) (*block, error) {
	ret := new(block)
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, xerrors.Errorf("BlockFromBytes: %w", err)
	}
	// check if the block index mutation is present in the last stateUpdate
	if len(ret.stateUpdates) == 0 {
		return nil, xerrors.New("BlockFromBytes: state updates not found")
	}
	last := ret.stateUpdates[len(ret.stateUpdates)-1]
	blockIndexBin, exists := last.Mutations().Get(kv.Key(coreutil.StatePrefixBlockIndex))
	if !exists {
		return nil, xerrors.New("BlockFromBytes: block index mutation not found")
	}
	var err error
	if ret.blockIndex, err = util.Uint32From4Bytes(blockIndexBin); err != nil {
		return nil, xerrors.Errorf("BlockFromBytes: %w", err)
	}
	return ret, nil
}

func LoadBlock(partition kvstore.KVStore, stateIndex uint32) (*block, error) {
	data, err := partition.Get(dbkeyBlock(stateIndex))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return BlockFromBytes(data)
}

// block with empty state update and nil state hash
func NewOriginBlock() *block {
	return NewBlock(0)
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
	return b.stateUpdates[len(b.stateUpdates)-1].Timestamp()
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

func dbkeyBlock(stateIndex uint32) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeStateUpdateBatch, util.Uint32To4Bytes(stateIndex))
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
