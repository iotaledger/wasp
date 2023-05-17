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
	"path/filepath"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
)

type snapshotterImpl struct {
	dir   string
	store state.Store
	log   *logger.Logger
}

var _ snapshotter = &snapshotterImpl{}

const (
	constSnapshotFileSuffix    = ".snap"
	constSnapshotTmpFileSuffix = ".tmp"
	constLengthArrayLength     = 4 // bytes
)

func newSnapshotter(log *logger.Logger, baseDir string, chainID isc.ChainID, store state.Store) (snapshotter, error) {
	dir := filepath.Join(baseDir, chainID.String())
	if err := ioutils.CreateDirectory(dir, 0o777); err != nil {
		return nil, fmt.Errorf("Snapshotter cannot create folder %v: %w", dir, err)
	}

	result := &snapshotterImpl{
		dir:   dir,
		store: store,
		log:   log,
	}
	result.cleanTempFiles() // To be able to make snapshots, which were not finished. See comment in `BlockCommitted` function
	log.Debugf("Snapshotter created folder %v for snapshots", dir)
	return result, nil
}

// Snapshotter makes snapshot of every `period`th state only, if this state hasn't
// been snapshotted before. The snapshot file name includes state index and state hash.
// Snapshotter first writes the state to temporary file and only then moves it to
// permanent location. Writing is done in separate thread to not interfere with
// normal State manager routine, as it may be lengthy. If snapshotter detects that
// the temporary file, needed to create a snapshot, already exists, it assumes
// that another go routine is already making a snapshot and returns. For this reason
// it is important to delete all temporary files on snapshotter start.
func (sn *snapshotterImpl) createSnapshotAsync(stateIndex uint32, commitment *state.L1Commitment, doneCallback func()) {
	sn.log.Debugf("Creating snapshot %v %s...", stateIndex, commitment)
	tmpFileName := tempSnapshotFileName(commitment.BlockHash())
	tmpFilePath := filepath.Join(sn.dir, tmpFileName)
	exists, _, _ := ioutils.PathExists(tmpFilePath)
	if exists {
		sn.log.Debugf("Creating snapshot %v %s: skipped making snapshot as it is already being produced", stateIndex, commitment)
		return
	}
	f, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		sn.log.Errorf("Creating snapshot %v %s: failed to create temporary snapshot file %s: %w", stateIndex, commitment, tmpFilePath, err)
		f.Close()
		return
	}
	sn.log.Debugf("Creating snapshot %v %s: reading store...", stateIndex, commitment)
	snapshot := mapdb.NewMapDB()
	err = sn.store.TakeSnapshot(commitment.TrieRoot(), snapshot)
	if err != nil {
		sn.log.Errorf("Creating snapshot %v %s: failed to read store: %w", stateIndex, commitment, err)
		f.Close()
		return
	}
	go func() {
		sn.log.Debugf("Creating snapshot %v %s: store read, iterating snapshot and writing it to file", stateIndex, commitment)
		err := writeSnapshot(stateIndex, commitment.TrieRoot(), snapshot, f)
		f.Close()
		if err != nil {
			sn.log.Errorf("Creating snapshot %v %s: filed to write snaphost to temporary file %s: %w", stateIndex, commitment, tmpFilePath, err)
			return
		}

		finalFileName := snapshotFileName(commitment.BlockHash())
		finalFilePath := filepath.Join(sn.dir, finalFileName)
		err = os.Rename(tmpFilePath, finalFilePath)
		if err != nil {
			sn.log.Errorf("Creating snapshot %v %s: failed to move temporary snapshot file %s to permanent location %s: %w",
				stateIndex, commitment, tmpFilePath, finalFilePath, err)
			return
		}

		doneCallback()
		sn.log.Infof("Creating snapshot %v %s: snapshot created in %s", stateIndex, commitment, finalFilePath)
	}()
}

// func (sn *snapshotterImpl) loadSnapshot(r io.Reader) error {
func (sn *snapshotterImpl) loadSnapshot(s string) error {
	return nil
}

func (sn *snapshotterImpl) cleanTempFiles() {
	tempFileRegExp := tempSnapshotFileNameString("*")
	tempFileRegExpWithPath := filepath.Join(sn.dir, tempFileRegExp)
	tempFiles, err := filepath.Glob(tempFileRegExpWithPath)
	if err != nil {
		sn.log.Errorf("Failed to obtain temporary snapshot file list: %v", err)
		return
	}

	removed := 0
	for _, tempFile := range tempFiles {
		err = os.Remove(tempFile)
		if err != nil {
			sn.log.Warnf("Failed to remove temporary snapshot file %s: %v", tempFile, err)
		} else {
			removed++
		}
	}
	sn.log.Debugf("Removed %v out of %v temporary snapshot files", removed, len(tempFiles))
}

func writeSnapshot(index uint32, trieRoot trie.Hash, snapshot kvstore.KVStore, w io.Writer) error {
	indexArray := make([]byte, 4) // Size of block index, which is of type uint32: 4 bytes
	binary.LittleEndian.PutUint32(indexArray, index)
	err := writeBytes(indexArray, w)
	if err != nil {
		return fmt.Errorf("failed writing block index %v: %w", index, err)
	}

	trieRootBytes := trieRoot.Bytes()
	err = writeBytes(trieRootBytes, w)
	if err != nil {
		return fmt.Errorf("failed writing trie root %s: %w", trieRoot, err)
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
