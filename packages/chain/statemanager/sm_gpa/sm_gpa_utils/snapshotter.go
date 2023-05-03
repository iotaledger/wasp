//
//
//
//
//
//

package sm_gpa_utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
)

type snapshotterImpl struct {
	dir            string
	lastIndex      uint32
	lastIndexMutex sync.Mutex
	period         uint32
	store          state.Store
	log            *logger.Logger
}

var _ Snapshotter = &snapshotterImpl{}

const (
	constSnapshotIndexHashFileNameSepparator = "-"
	constSnapshotFileSuffix                  = ".snap"
	constSnapshotTmpFileSuffix               = ".tmp"
	constLengthArrayLength                   = 4 // bytes
)

func NewSnapshotter(log *logger.Logger, baseDir string, chainID isc.ChainID, period uint32, store state.Store) (Snapshotter, error) {
	dir := filepath.Join(baseDir, chainID.String())
	if err := ioutils.CreateDirectory(dir, 0o777); err != nil {
		return nil, fmt.Errorf("Snapshotter cannot create folder %v: %w", dir, err)
	}

	result := &snapshotterImpl{
		dir:            dir,
		lastIndex:      0, // TODO: what about loading snapshots?
		lastIndexMutex: sync.Mutex{},
		period:         period,
		store:          store,
		log:            log,
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
func (sn *snapshotterImpl) BlockCommitted(block state.Block) {
	index := block.StateIndex()
	if (index > sn.lastIndex) && (index%sn.period == 0) { // TODO: what if snapshotted state has been reverted?
		commitment := block.L1Commitment()
		tmpFileName := tempSnapshotFileName(index, commitment.BlockHash())
		tmpFilePath := filepath.Join(sn.dir, tmpFileName)
		exists, _, _ := ioutils.PathExists(tmpFilePath)
		if exists {
			sn.log.Debugf("Skipped making state snapshot on index %v commitment %s as it is already being produced", index, commitment)
			return
		}
		f, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
		if err != nil {
			sn.log.Errorf("Failed to create temporary snapshot file %s: %w", tmpFilePath, err)
			return
		}
		sn.log.Debugf("Starting making state snapshot on index %v commitment %s", index, commitment)
		snapshot := mapdb.NewMapDB()
		err = sn.store.TakeSnapshot(commitment.TrieRoot(), snapshot)
		if err != nil {
			sn.log.Errorf("Failed to obtain snapshot %s: %v", commitment, err)
			return
		}
		go func() {
			sn.log.Debugf("State index %v commitment %s obtained, iterating it and writing to file", index, commitment)
			err := writeSnapshotToFile(commitment.TrieRoot(), snapshot, f)
			if err != nil {
				sn.log.Errorf("Failed to write state index %v commitment %s to temporary snapshot file %s: %w", index, commitment, tmpFilePath, err)
				return
			}

			finalFileName := snapshotFileName(index, commitment.BlockHash())
			finalFilePath := filepath.Join(sn.dir, finalFileName)
			err = os.Rename(tmpFilePath, finalFilePath)
			if err != nil {
				sn.log.Errorf("Failed to move temporary snapshot file %s to permanent location %s: %w", tmpFilePath, finalFilePath, err)
				return
			}

			sn.lastIndexMutex.Lock()
			if index > sn.lastIndex {
				sn.lastIndex = index
			}
			sn.lastIndexMutex.Unlock()
			sn.log.Infof("Snapshot on state index %v commitment %s was created in %s", index, commitment, finalFilePath)
		}()
	}
}

func (sn *snapshotterImpl) cleanTempFiles() {
	tempFileRegExp := tempSnapshotFileNameString("*", "*")
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

func writeSnapshotToFile(trieRoot trie.Hash, snapshot kvstore.KVStore, f *os.File) error {
	defer f.Close()

	trieRootBytes := trieRoot.Bytes()
	err := writeBytes(trieRootBytes, f)
	if err != nil {
		return fmt.Errorf("failed writing trie root %s: %w", trieRoot, err)
	}

	iterErr := snapshot.Iterate(kvstore.EmptyPrefix, func(key kvstore.Key, value kvstore.Value) bool {
		e := writeBytes(key, f)
		if e != nil {
			err = fmt.Errorf("failed writing key %v: %w", key, e)
			return false
		}

		e = writeBytes(value, f)
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

func readSnapshotFromFile(filePath string) (trie.Hash, kvstore.KVStore, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return trie.Hash{}, nil, fmt.Errorf("Failed to open snapshot file %s: %w", filePath, err)
	}
	defer f.Close()

	trieRootArray, err := readBytes(f)
	if err != nil {
		return trie.Hash{}, nil, fmt.Errorf("failed to read trie root: %w", err)
	}
	trieRoot, err := trie.HashFromBytes(trieRootArray)
	if err != nil {
		return trie.Hash{}, nil, fmt.Errorf("failed to read parse trie root: %w", err)
	}

	snapshot := mapdb.NewMapDB()
	for key, err := readBytes(f); !errors.Is(err, io.EOF); key, err = readBytes(f) {
		if err != nil {
			return trie.Hash{}, nil, fmt.Errorf("failed to read key: %w", err)
		}

		value, err := readBytes(f)
		if err != nil {
			return trie.Hash{}, nil, fmt.Errorf("failed to read value of key %v: %w", key, err)
		}

		err = snapshot.Set(key, value)
		if err != nil {
			return trie.Hash{}, nil, fmt.Errorf("failed to set key's %v value %v to snapshot: %w", key, value, err)
		}
	}
	return trieRoot, snapshot, nil
}

func tempSnapshotFileName(index uint32, blockHash state.BlockHash) string {
	return tempSnapshotFileNameString(fmt.Sprint(index), blockHash.String())
}

func tempSnapshotFileNameString(index string, blockHash string) string {
	return snapshotFileNameString(index, blockHash) + constSnapshotTmpFileSuffix
}

func snapshotFileName(index uint32, blockHash state.BlockHash) string {
	return snapshotFileNameString(fmt.Sprint(index), blockHash.String())
}

func snapshotFileNameString(index string, blockHash string) string {
	return index + constSnapshotIndexHashFileNameSepparator + blockHash + constSnapshotFileSuffix
}

func writeBytes(bytes []byte, f *os.File) error {
	n, err := f.Write(arrayLengthToArray(bytes))
	if n != constLengthArrayLength {
		return fmt.Errorf("only %v of total %v bytes of length written", n, constLengthArrayLength)
	}
	if err != nil {
		return fmt.Errorf("failed writing length: %w", err)
	}

	n, err = f.Write(bytes)
	if n != len(bytes) {
		return fmt.Errorf("only %v of total %v bytes of array written", n, len(bytes))
	}
	if err != nil {
		return fmt.Errorf("failed writing array: %w", err)
	}

	return nil
}

func readBytes(f *os.File) ([]byte, error) {
	lenArray := make([]byte, constLengthArrayLength)
	read, err := f.Read(lenArray)
	if err != nil {
		return nil, fmt.Errorf("failed to read length: %w", err)
	}
	if read < constLengthArrayLength {
		return nil, fmt.Errorf("read only %v bytes out of %v of length", read, constLengthArrayLength)
	}

	array, err := arrayToArrayOfLength(lenArray)
	if err != nil {
		return nil, fmt.Errorf("failed to parse length: %w", err)
	}
	read, err = f.Read(array)
	if err != nil {
		return nil, fmt.Errorf("failed to read array: %w", err)
	}
	if read < len(array) {
		return nil, fmt.Errorf("only %v of %v bytes of array read", read, len(array))
	}

	return array, nil
}

func arrayLengthToArray(array []byte) []byte {
	length := uint32(len(array))
	res := make([]byte, constLengthArrayLength)
	binary.LittleEndian.PutUint32(res, length)
	return res
}

func arrayToArrayOfLength(lengthArray []byte) ([]byte, error) {
	if len(lengthArray) != constLengthArrayLength {
		return nil, fmt.Errorf("array length array contains %v bytes instead of %v", len(lengthArray), constLengthArrayLength)
	}
	length := binary.LittleEndian.Uint32(lengthArray)
	return make([]byte, length), nil
}
