package snapshots

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"fortio.org/safecast"

	"github.com/iotaledger/wasp/v2/packages/state"
)

type snapshotterImpl struct {
	store state.Store
}

var _ snapshotter = &snapshotterImpl{}

const constLengthArrayLength = 4 // bytes

func newSnapshotter(store state.Store) snapshotter {
	return &snapshotterImpl{store: store}
}

func (sn *snapshotterImpl) storeSnapshot(snapshotInfo SnapshotInfo, w io.Writer) error {
	indexArray := make([]byte, 4) // Size of block index, which is of type uint32: 4 bytes
	binary.LittleEndian.PutUint32(indexArray, snapshotInfo.StateIndex())
	err := writeBytes(indexArray, w)
	if err != nil {
		return fmt.Errorf("failed writing block index %v: %w", snapshotInfo.StateIndex(), err)
	}

	trieRootBytes := snapshotInfo.Commitment().Bytes()
	err = writeBytes(trieRootBytes, w)
	if err != nil {
		return fmt.Errorf("failed writing L1 commitment %s: %w", snapshotInfo.Commitment(), err)
	}

	err = sn.store.TakeSnapshot(snapshotInfo.TrieRoot(), w)
	if err != nil {
		return fmt.Errorf("failed to store snapshot: %w", err)
	}
	return nil
}

func (sn *snapshotterImpl) loadSnapshot(snapshotInfo SnapshotInfo, r io.Reader) error {
	readSnapshotInfo, err := readSnapshotInfo(r)
	if err != nil {
		return fmt.Errorf("failed reading snapshot info: %w", err)
	}
	if !readSnapshotInfo.Equals(snapshotInfo) {
		return fmt.Errorf("snapshot read %s is different than expected %v", readSnapshotInfo, snapshotInfo)
	}
	err = sn.store.RestoreSnapshot(readSnapshotInfo.TrieRoot(), r, true)
	if err != nil {
		return fmt.Errorf("failed restoring snapshot: %w", err)
	}
	return nil
}

func readSnapshotInfo(r io.Reader) (SnapshotInfo, error) {
	indexArray, err := readBytes(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read block index: %w", err)
	}
	if len(indexArray) != 4 { // Size of block index, which is of type uint32: 4 bytes
		return nil, fmt.Errorf("block index is %v instead of 4 bytes", len(indexArray))
	}
	index := binary.LittleEndian.Uint32(indexArray)

	trieRootArray, err := readBytes(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read trie root: %w", err)
	}
	commitment, err := state.NewL1CommitmentFromBytes(trieRootArray)
	if err != nil {
		return nil, fmt.Errorf("failed to parse L1 commitment: %w", err)
	}

	return NewSnapshotInfo(index, commitment), nil
}

func writeBytes(bytes []byte, w io.Writer) error {
	lengthArray := make([]byte, constLengthArrayLength)
	bytesLen, err := safecast.Convert[uint32](len(bytes))
	if err != nil {
		return fmt.Errorf("integer overflow in bytes length: %w", err)
	}
	binary.LittleEndian.PutUint32(lengthArray, bytesLen)
	n, err := w.Write(lengthArray)
	if n != constLengthArrayLength {
		return fmt.Errorf("only %v of total %v bytes of length written", n, constLengthArrayLength)
	}
	if err != nil {
		return fmt.Errorf("failed writing length: %w", err)
	}

	n, err = w.Write(bytes)
	if n != len(bytes) {
		return fmt.Errorf("only %v of total %v bytes of array written", n, len(bytes))
	}
	if err != nil {
		return fmt.Errorf("failed writing array: %w", err)
	}

	return nil
}

func readBytes(r io.Reader) ([]byte, error) {
	w := new(bytes.Buffer)
	read, err := io.CopyN(w, r, constLengthArrayLength)
	lengthArray := w.Bytes()
	if err != nil {
		return nil, fmt.Errorf("read only %v bytes out of %v of length, error: %w", read, constLengthArrayLength, err)
	}

	length := int64(binary.LittleEndian.Uint32(lengthArray))
	w = new(bytes.Buffer)
	read, err = io.CopyN(w, r, length)
	array := w.Bytes()
	if err != nil {
		return nil, fmt.Errorf("only %v of %v bytes of array read, error: %w", read, len(array), err)
	}

	return array, nil
}
