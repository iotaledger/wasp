//
//
//
//
//
//

package sm_snapshots

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
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
	err := sn.store.TakeSnapshot(snapshotInfo.GetTrieRoot(), snapshot)
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
	stateIndex, trieRoot, snapshot, err := readSnapshot(r)
	if err != nil {
		return fmt.Errorf("failed reading snapshot: %w", err)
	}
	if stateIndex != snapshotInfo.GetStateIndex() {
		return fmt.Errorf("state index read %v is different than expected %v", stateIndex, snapshotInfo.GetStateIndex())
	}
	if !trieRoot.Equals(snapshotInfo.GetTrieRoot()) {
		return fmt.Errorf("trie root read %s is different than expected %s", trieRoot, snapshotInfo.GetTrieRoot())
	}
	err = sn.store.RestoreSnapshot(trieRoot, snapshot)
	if err != nil {
		return fmt.Errorf("failed restoring snapshot: %w", err)
	}
	return nil
}

func writeSnapshot(snapshotInfo SnapshotInfo, snapshot kvstore.KVStore, w io.Writer) error {
	indexArray := make([]byte, 4) // Size of block index, which is of type uint32: 4 bytes
	binary.LittleEndian.PutUint32(indexArray, snapshotInfo.GetStateIndex())
	err := writeBytes(indexArray, w)
	if err != nil {
		return fmt.Errorf("failed writing block index %v: %w", snapshotInfo.GetStateIndex(), err)
	}

	trieRootBytes := snapshotInfo.GetTrieRoot().Bytes()
	err = writeBytes(trieRootBytes, w)
	if err != nil {
		return fmt.Errorf("failed writing trie root %s: %w", snapshotInfo.GetTrieRoot(), err)
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

func readSnapshotFromFile(filePath string) (uint32, trie.Hash, kvstore.KVStore, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, trie.Hash{}, nil, fmt.Errorf("failed to open snapshot file %s: %w", filePath, err)
	}
	defer f.Close()

	return readSnapshot(f)
}

func readSnapshot(r io.Reader) (uint32, trie.Hash, kvstore.KVStore, error) {
	indexArray, err := readBytes(r)
	if err != nil {
		return 0, trie.Hash{}, nil, fmt.Errorf("failed to read block index: %w", err)
	}
	if len(indexArray) != 4 { // Size of block index, which is of type uint32: 4 bytes
		return 0, trie.Hash{}, nil, fmt.Errorf("block index is %v instead of 4 bytes", len(indexArray))
	}
	index := binary.LittleEndian.Uint32(indexArray)

	trieRootArray, err := readBytes(r)
	if err != nil {
		return 0, trie.Hash{}, nil, fmt.Errorf("failed to read trie root: %w", err)
	}
	trieRoot, err := trie.HashFromBytes(trieRootArray)
	if err != nil {
		return 0, trie.Hash{}, nil, fmt.Errorf("failed to read parse trie root: %w", err)
	}

	snapshot := mapdb.NewMapDB()
	for key, err := readBytes(r); !errors.Is(err, io.EOF); key, err = readBytes(r) {
		if err != nil {
			return 0, trie.Hash{}, nil, fmt.Errorf("failed to read key: %w", err)
		}

		value, err := readBytes(r)
		if err != nil {
			return 0, trie.Hash{}, nil, fmt.Errorf("failed to read value of key %v: %w", key, err)
		}

		err = snapshot.Set(key, value)
		if err != nil {
			return 0, trie.Hash{}, nil, fmt.Errorf("failed to set key's %v value %v to snapshot: %w", key, value, err)
		}
	}
	return index, trieRoot, snapshot, nil
}

func tempSnapshotFileName(blockHash state.BlockHash) string {
	return tempSnapshotFileNameString(blockHash.String())
}

func tempSnapshotFileNameString(blockHash string) string {
	return snapshotFileNameString(blockHash) + constSnapshotTmpFileSuffix
}

func snapshotFileName(blockHash state.BlockHash) string {
	return snapshotFileNameString(blockHash.String())
}

func snapshotFileNameString(blockHash string) string {
	return blockHash + constSnapshotFileSuffix
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
	lengthArray := make([]byte, constLengthArrayLength)
	read, err := r.Read(lengthArray)
	if err != nil {
		return nil, fmt.Errorf("failed to read length: %w", err)
	}
	if read < constLengthArrayLength {
		return nil, fmt.Errorf("read only %v bytes out of %v of length", read, constLengthArrayLength)
	}

	length := binary.LittleEndian.Uint32(lengthArray)
	array := make([]byte, length)
	if err != nil {
		return nil, fmt.Errorf("failed to parse length: %w", err)
	}
	read, err = r.Read(array)
	if err != nil {
		return nil, fmt.Errorf("failed to read array: %w", err)
	}
	if read < len(array) {
		return nil, fmt.Errorf("only %v of %v bytes of array read", read, len(array))
	}

	return array, nil
}
