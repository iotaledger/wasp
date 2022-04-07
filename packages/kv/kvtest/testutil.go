// Package kvtest contains key/value related functions used for testing (otherwise of general nature)
package kvtest

import (
	"io"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

const nilprefix = ""

func ByteSize(s kv.KVStoreReader) int {
	accLen := 0
	err := s.Iterate(nilprefix, func(k kv.Key, v []byte) bool {
		accLen += len([]byte(k)) + len(v)
		return true
	})
	if err != nil {
		return 0
	}
	return accLen
}

func DumpToFile(r kv.KVStoreReader, fname string) (int, error) {
	file, err := os.Create(fname)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var bytesTotal int
	err = r.Iterate("", func(k kv.Key, v []byte) bool {
		n, errw := writeKV(file, []byte(k), v)
		if errw != nil {
			err = errw
			return false
		}
		bytesTotal += n
		return true
	})
	return bytesTotal, err
}

func UnDumpFromFile(w kv.KVWriter, fname string) (int, error) {
	file, err := os.Open(fname)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var k, v []byte
	var exit bool
	n := 0
	for {
		if k, v, exit = readKV(file); exit {
			break
		}
		n += len(k) + len(v) + 6
		w.Set(kv.Key(k), v)
	}
	return n, nil
}

func writeKV(w io.Writer, k, v []byte) (int, error) {
	if err := util.WriteBytes16(w, k); err != nil {
		return 0, err
	}
	if err := util.WriteBytes32(w, v); err != nil {
		return len(k) + 2, err
	}
	return len(k) + len(v) + 6, nil
}

func readKV(r io.Reader) ([]byte, []byte, bool) {
	k, err := util.ReadBytes16(r)
	if xerrors.Is(err, io.EOF) {
		return nil, nil, true
	}
	v, err := util.ReadBytes32(r)
	if err != nil {
		panic(err)
	}
	return k, v, false
}

type randStreamIterator struct {
	rnd   *rand.Rand
	par   RandStreamParams
	count int
}

type RandStreamParams struct {
	Seed       int64
	NumKVPairs int // 0 means infinite
	MaxKey     int // max length of key
	MaxValue   int // max length of value
}

func NewRandStreamIterator(p ...RandStreamParams) *randStreamIterator {
	ret := &randStreamIterator{
		par: RandStreamParams{
			Seed:       time.Now().UnixNano(),
			NumKVPairs: 0, // infinite
			MaxKey:     64,
			MaxValue:   128,
		},
	}
	if len(p) > 0 {
		ret.par = p[0]
	}
	ret.rnd = rand.New(rand.NewSource(ret.par.Seed))
	return ret
}

func (r *randStreamIterator) Iterate(fun func(k []byte, v []byte) bool) error {
	max := r.par.NumKVPairs
	if max <= 0 {
		max = math.MaxInt
	}
	for r.count < max {
		k := make([]byte, r.rnd.Intn(r.par.MaxKey-1)+1)
		r.rnd.Read(k)
		v := make([]byte, r.rnd.Intn(r.par.MaxValue-1)+1)
		r.rnd.Read(v)
		if !fun(k, v) {
			return nil
		}
		r.count++
	}
	return nil
}
