package snapshot

import (
	"fmt"
	"io"
	"path"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type ConsoleReportParams struct {
	Console           io.Writer
	StatsEveryKVPairs int
}

func FileName(chainID *isc.ChainID, stateIndex uint32) string {
	return fmt.Sprintf("%s.%d.snapshot", chainID, stateIndex)
}

// WriteKVToStream dumps k/v pairs of the state into the
// file. Keys are not sorted, so the result in general is not deterministic
func WriteKVToStream(store kv.KVIterator, stream kv.StreamWriter, p ...ConsoleReportParams) error {
	par := ConsoleReportParams{
		Console:           io.Discard,
		StatsEveryKVPairs: 100,
	}
	if len(p) > 0 {
		par = p[0]
	}
	var err, errW error
	err = store.Iterate("", func(k kv.Key, v []byte) bool {
		if errW = stream.Write([]byte(k), v); errW != nil {
			return false
		}
		if par.StatsEveryKVPairs > 0 {
			kvCount, bCount := stream.Stats()
			if kvCount%par.StatsEveryKVPairs == 0 {
				fmt.Fprintf(par.Console, "[WriteKVToStream] k/v pairs: %d, bytes: %d\n", kvCount, bCount)
			}
		}
		return errW == nil
	})
	if err != nil {
		fmt.Fprintf(par.Console, "[WriteKVToStream] error while reading: %v\n", err)
		return err
	}
	if errW != nil {
		fmt.Fprintf(par.Console, "[WriteKVToStream] error while writing: %v\n", err)
		return errW
	}
	return nil
}

func WriteSnapshot(ordr state.OptimisticStateReader, dir string, p ...ConsoleReportParams) error {
	par := ConsoleReportParams{
		Console:           io.Discard,
		StatsEveryKVPairs: 100,
	}
	if len(p) > 0 {
		par = p[0]
	}
	chainID, err := ordr.ChainID()
	if err != nil {
		return err
	}
	stateIndex, err := ordr.BlockIndex()
	if err != nil {
		return err
	}
	timestamp, err := ordr.Timestamp()
	if err != nil {
		return err
	}
	fmt.Fprintf(par.Console, "[WriteSnapshot] chainID:     %s\n", chainID)
	fmt.Fprintf(par.Console, "[WriteSnapshot] state index: %d\n", stateIndex)
	fmt.Fprintf(par.Console, "[WriteSnapshot] timestamp: %v\n", timestamp)
	fname := path.Join(dir, FileName(chainID, stateIndex))
	fmt.Fprintf(par.Console, "[WriteSnapshot] will be writing to file: %s\n", fname)

	fstream, err := kv.CreateKVStreamFile(fname)
	if err != nil {
		return err
	}
	defer fstream.File.Close()

	fmt.Printf("[WriteSnapshot] writing to file ")
	if err := WriteKVToStream(ordr.KVStoreReader(), fstream, par); err != nil {
		return err
	}
	tKV, tBytes := fstream.Stats()
	fmt.Fprintf(par.Console, "[WriteSnapshot] TOTAL: kv records: %d, bytes: %d\n", tKV, tBytes)
	return nil
}

type FileProperties struct {
	FileName   string
	ChainID    *isc.ChainID
	StateIndex uint32
	TimeStamp  time.Time
	NumRecords int
	MaxKeyLen  int
	Bytes      int
}

func Scan(rdr kv.StreamIterator) (*FileProperties, error) {
	ret := &FileProperties{}
	var chainIDFound, stateIndexFound, timestampFound bool
	var errR error

	err := rdr.Iterate(func(k, v []byte) bool {
		if len(k) == 0 {
			if chainIDFound {
				errR = xerrors.New("duplicate record with chainID")
				return false
			}
			if ret.ChainID, errR = isc.ChainIDFromBytes(v); errR != nil {
				return false
			}
			chainIDFound = true
		}
		if string(k) == coreutil.StatePrefixBlockIndex {
			if stateIndexFound {
				errR = xerrors.New("duplicate record with state index")
				return false
			}
			if ret.StateIndex, errR = util.Uint32From4Bytes(v); errR != nil {
				return false
			}
			stateIndexFound = true
		}
		if string(k) == coreutil.StatePrefixTimestamp {
			if timestampFound {
				errR = xerrors.New("duplicate record with timestamp")
				return false
			}
			if ret.TimeStamp, errR = codec.DecodeTime(v); errR != nil {
				return false
			}
			timestampFound = true
		}
		if len(v) == 0 {
			errR = xerrors.New("empty value encountered")
			return false
		}
		ret.NumRecords++
		if len(k) > ret.MaxKeyLen {
			ret.MaxKeyLen = len(k)
		}
		ret.Bytes += len(k) + len(v) + 6
		return true
	})
	if err != nil {
		return nil, err
	}
	if errR != nil {
		return nil, errR
	}
	return ret, nil
}

func ScanFile(fname string) (*FileProperties, error) {
	stream, err := kv.OpenKVStreamFile(fname)
	if err != nil {
		return nil, err
	}
	defer stream.File.Close()

	ret, err := Scan(stream)
	if err != nil {
		return nil, err
	}
	ret.FileName = fname
	return ret, nil
}
