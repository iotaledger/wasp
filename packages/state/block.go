package state

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/util"
)

type blockImpl struct {
	stateOutputID *iotago.UTXOInput
	stateUpdate   *stateUpdateImpl
	blockIndex    uint32 // not persistent
}

var _ Block = &blockImpl{}

// validates, enumerates and creates a block from array of state updates
func newBlock(muts *buffered.Mutations) (Block, error) {
	ret := &blockImpl{
		stateUpdate: &stateUpdateImpl{
			mutations: muts.Clone(),
		},
	}
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

func (b *blockImpl) Bytes() []byte {
	return util.MustBytes(b)
}

func (b *blockImpl) String() string {
	ret := ""
	ret += fmt.Sprintf("Block: state index: %d\n", b.BlockIndex())
	ret += fmt.Sprintf("state txid: %s\n", isc.OID(b.ApprovingOutputID()))
	ret += fmt.Sprintf("timestamp: %v\n", b.Timestamp())
	ret += fmt.Sprintf("state update: %s\n", (*b.stateUpdate).String())
	return ret
}

func (b *blockImpl) ApprovingOutputID() *iotago.UTXOInput {
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
func (b *blockImpl) PreviousL1Commitment() *L1Commitment {
	ph, err := findPrevStateCommitmentMutation(b.stateUpdate)
	if err != nil {
		panic(err)
	}
	return ph
}

func (b *blockImpl) SetApprovingOutputID(oid *iotago.UTXOInput) {
	b.stateOutputID = oid
}

func (b *blockImpl) essenceBytes() []byte {
	var buf bytes.Buffer
	if err := b.writeEssence(&buf); err != nil {
		panic("essenceBytes")
	}
	return buf.Bytes()
}

func (b *blockImpl) Write(w io.Writer) error {
	if err := b.writeEssence(w); err != nil {
		return err
	}
	if err := b.writeOutputID(w); err != nil {
		return err
	}
	return nil
}

func (b *blockImpl) writeEssence(w io.Writer) error {
	return b.stateUpdate.Write(w)
}

func (b *blockImpl) writeOutputID(w io.Writer) error {
	if err := util.WriteBoolByte(w, b.stateOutputID != nil); err != nil {
		return err
	}
	if b.stateOutputID == nil {
		return nil
	}
	serialized, err := b.stateOutputID.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return err
	}
	_, err = w.Write(serialized)
	return err
}

func (b *blockImpl) Read(r io.Reader) error {
	if err := b.readEssence(r); err != nil {
		return err
	}
	if err := b.readOutputID(r); err != nil {
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

func (b *blockImpl) readOutputID(r io.Reader) error {
	var oidPresent bool
	if err := util.ReadBoolByte(r, &oidPresent); err != nil {
		return err
	}
	if !oidPresent {
		return nil
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}
	b.stateOutputID = &iotago.UTXOInput{}
	_, err := b.stateOutputID.Deserialize(buf.Bytes(), serializer.DeSeriModeNoValidation, nil)
	return err
}

func (b *blockImpl) GetHash() (ret BlockHash) {
	r := blake2b.Sum256(b.essenceBytes())
	copy(ret[:BlockHashSize], r[:BlockHashSize])
	return
}

func (b *blockImpl) Equals(other Block) bool {
	return b.GetHash().Equals(other.GetHash())
}
