package sm_snapshots

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/state"
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
	snapshot := mapdb.NewMapDB()
	err := sn.store.TakeSnapshot(snapshotInfo.TrieRoot(), snapshot)
	if err != nil {
		return fmt.Errorf("failed to read store: %w", err)
	}
	err = writeSnapshot(snapshotInfo, snapshot, w)
	if err != nil {
		return fmt.Errorf("failed writing snapshot: %w", err)
	}
	return nil
}

func (sn *snapshotterImpl) loadSnapshot(snapshotInfo SnapshotInfo, r io.Reader) error {
	readSnapshotInfo, snapshot, err := readSnapshot(r)
	if err != nil {
		return fmt.Errorf("failed reading snapshot: %w", err)
	}
	if !readSnapshotInfo.Equals(snapshotInfo) {
		return fmt.Errorf("snapshot read %s is different than expected %v", readSnapshotInfo, snapshotInfo)
	}
	err = sn.store.RestoreSnapshot(readSnapshotInfo.TrieRoot(), snapshot)
	if err != nil {
		return fmt.Errorf("failed restoring snapshot: %w", err)
	}
	return nil
}

func writeSnapshot(snapshotInfo SnapshotInfo, snapshot kvstore.KVStore, w io.Writer) error {
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

	iterErr := snapshot.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		e := writeBytes(key, w)
		if e != nil {
			err = fmt.Errorf("failed writing key %v: %w", key, e)
			return false
		}

		e = writeBytes(value, w)
		if e != nil {
			err = fmt.Errorf("failed writing key's %v value %v: %w", key, value, e)
			return false
		}

		return true
	})

	if iterErr != nil {
		return iterErr
	}

	return err
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
	commitment, err := state.L1CommitmentFromBytes(trieRootArray)
	if err != nil {
		return nil, fmt.Errorf("failed to parse L1 commitment: %w", err)
	}

	return NewSnapshotInfo(index, commitment), nil
}

func readSnapshot(r io.Reader) (SnapshotInfo, kvstore.KVStore, error) {
	snapshotInfo, err := readSnapshotInfo(r)
	if err != nil {
		return nil, nil, err
	}
	snapshot := mapdb.NewMapDB()
	for key, err := readBytes(r); !errors.Is(err, io.EOF); key, err = readBytes(r) {
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read key: %w", err)
		}

		value, err := readBytes(r)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read value of key %v: %w", key, err)
		}

		err = snapshot.Set(key, value)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to set key's %v value %v to snapshot: %w", key, value, err)
		}
	}
	return snapshotInfo, snapshot, nil
}

func writeBytes(bytes []byte, w io.Writer) error {
	lengthArray := make([]byte, constLengthArrayLength)
	binary.LittleEndian.PutUint32(lengthArray, uint32(len(bytes)))
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
