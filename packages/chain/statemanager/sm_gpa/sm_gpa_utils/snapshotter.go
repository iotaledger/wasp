//
//
//
//
//
//

package sm_gpa_utils

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/state"
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
		state, err := sn.store.StateByTrieRoot(commitment.TrieRoot())
		if err != nil {
			sn.log.Errorf("Failed to obtain state %s: %v", commitment, err)
			return
		}
		go func() {
			sn.log.Debugf("State index %v commitment %s obtained, iterating it and writing to file", index, commitment)
			err := writeStateToFile(state, tmpFilePath, f)
			if err != nil {
				sn.log.Errorf("Failed to write state index %v commitment %s to temporary snapshot file: %w", index, commitment, err)
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

func writeStateToFile(state state.State, filePath string, f *os.File) error {
	defer f.Close()

	var err error
	err = nil
	state.Iterate(kv.EmptyPrefix, func(key kv.Key, value []byte) bool {
		keyBytes := []byte(key)
		n, e := f.Write(arrayLengthToArray(keyBytes))
		if n != constLengthArrayLength {
			err = fmt.Errorf("only %v of total %v bytes of key %s length were written to file %s", n, constLengthArrayLength, key, filePath)
			return false
		}
		if e != nil {
			err = fmt.Errorf("failed writing key %s length to file %s: %w", key, filePath, e)
			return false
		}

		n, e = f.Write(keyBytes)
		if n != len(keyBytes) {
			err = fmt.Errorf("only %v of total %v bytes of key %s were written to file %s", n, len(keyBytes), key, filePath)
			return false
		}
		if e != nil {
			err = fmt.Errorf("failed writing key %s to file %s: %w", key, filePath, e)
			return false
		}

		n, e = f.Write(arrayLengthToArray(value))
		if n != constLengthArrayLength {
			err = fmt.Errorf("only %v of total %v bytes of value of key %s length were written to file %s", n, constLengthArrayLength, key, filePath)
			return false
		}
		if e != nil {
			err = fmt.Errorf("failed writing value of key %s length to file %s: %w", key, filePath, e)
			return false
		}

		n, e = f.Write(value)
		if n != len(value) {
			err = fmt.Errorf("only %v of total %v bytes of value of key %s were written to file %s", n, len(value), key, filePath)
			return false
		}
		if e != nil {
			err = fmt.Errorf("failed writing value of key %s to file %s: %w", key, filePath, e)
			return false
		}

		return true
	})

	return err
}

func readStateFromFile(filePath string) (*buffered.Mutations, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open snapshot file %s: %w", filePath, err)
	}
	defer f.Close()

	mutations := buffered.NewMutations()
	lenArray := make([]byte, constLengthArrayLength)
	for read, err := f.Read(lenArray); err != io.EOF; read, err = f.Read(lenArray) {
		if err != nil {
			return nil, fmt.Errorf("failed to read key length: %w", err)
		}
		if read < constLengthArrayLength {
			return nil, fmt.Errorf("read only %v bytes out of %v of key length", read, constLengthArrayLength)
		}

		keyArray, err := arrayToArrayOfLength(lenArray)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key length: %w", err)
		}
		read, err = f.Read(keyArray)
		if err != nil {
			return nil, fmt.Errorf("failed to read key: %w", err)
		}
		if read < len(keyArray) {
			return nil, fmt.Errorf("read only %v bytes out of %v of key", read, len(keyArray))
		}
		key := kv.Key(keyArray)

		read, err := f.Read(lenArray)
		if err != nil {
			return nil, fmt.Errorf("failed to read value length of key %s: %w", key, err)
		}
		if read < constLengthArrayLength {
			return nil, fmt.Errorf("read only %v bytes out of %v of value length of key %s", read, constLengthArrayLength, key)
		}

		value, err := arrayToArrayOfLength(lenArray)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value length of key %s: %w", key, err)
		}
		read, err = f.Read(value)
		if err != nil {
			return nil, fmt.Errorf("failed to read value of key %s: %w", key, err)
		}
		if read < len(value) {
			return nil, fmt.Errorf("read only %v bytes out of %v of value of key %s", read, len(value), key)
		}
		mutations.Set(key, value)
	}
	return mutations, nil
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
