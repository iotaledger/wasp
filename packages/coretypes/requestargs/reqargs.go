package requestargs

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"io"
)

// RequestArgs encodes request parameters taking into account hashes of data blobs
type RequestArgs dict.Dict

// New makes new object taking 'as is', without encoding
func New(d ...dict.Dict) RequestArgs {
	if len(d) == 0 || len(d[0]) == 0 {
		return RequestArgs(dict.New())
	}
	if len(d) > 1 {
		panic("len(d) > 1")
	}
	return RequestArgs(d[0].Clone())
}

const optimalSize = 32

// NewOptimizedRequestArgs takes dictionary and encodes it
func NewOptimizedRequestArgs(d dict.Dict, optSize ...int) (RequestArgs, map[kv.Key][]byte) {
	var osize int
	if len(optSize) > 0 {
		osize = optSize[0]
	}
	if osize <= optimalSize {
		osize = optimalSize
	}
	ret := New(nil)
	retOptimized := make(map[kv.Key][]byte)
	for k, v := range d {
		if len(v) <= osize {
			ret.AddEncodeSimple(k, v)
		} else {
			ret.AddAsBlobRef(k, v)
			retOptimized[k] = v
		}
	}
	return ret, retOptimized
}

// AddEncodeSimple add new ordinary argument. Encodes the key as "normal"
func (a RequestArgs) AddEncodeSimple(name kv.Key, data []byte) RequestArgs {
	a["-"+name] = data
	return a
}

// AddEncodeBlobRef adds hash as data and marks it is a blob reference
func (a RequestArgs) AddEncodeBlobRef(name kv.Key, hash hashing.HashValue) RequestArgs {
	a["*"+name] = hash[:]
	return a
}

// AddAsBlobRef adds argument with the data hash instead of data itself.
// Encodes key as "blob reference"
func (a RequestArgs) AddAsBlobRef(name kv.Key, data []byte) hashing.HashValue {
	h := hashing.HashData(data)
	a.AddEncodeBlobRef(name, h)
	return h
}

func (a RequestArgs) AddEncodeSimpleMany(d dict.Dict) RequestArgs {
	for k, v := range d {
		a.AddEncodeSimple(k, v)
	}
	return a
}

// HasBlobRef return if request arguments contain at least one blob reference
func (a RequestArgs) HasBlobRef() bool {
	var ret bool
	(dict.Dict(a)).ForEach(func(key kv.Key, _ []byte) bool {
		ret = []byte(key)[0] == '*'
		if ret {
			return false
		}
		return true
	})
	return ret
}

func (a RequestArgs) String() string {
	return (dict.Dict(a)).String()
}

func (a RequestArgs) Clone() RequestArgs {
	return RequestArgs((dict.Dict(a)).Clone())
}

func (a RequestArgs) Write(w io.Writer) error {
	return (dict.Dict(a)).Write(w)
}

func (a RequestArgs) Read(r io.Reader) error {
	return (dict.Dict(a)).Read(r)
}

// SolidifyRequestArguments decodes RequestArgs.
// each value treated according to the value of the first byte:
//  - if the value is '*' the data is a content reference. First 32 bytes always treated as data hash.
//    The rest (if any) is a content address. It will be treated by a downloader
//  - otherwise it is a raw data
func (a RequestArgs) SolidifyRequestArguments(reg coretypes.BlobCache) (dict.Dict, bool, error) {
	ret := dict.New()
	ok := true
	var err error
	var data []byte
	var h hashing.HashValue
	reqArgsDict := dict.Dict(a)
	reqArgsDict.ForEach(func(key kv.Key, value []byte) bool {
		d := []byte(key)
		if len(d) == 0 {
			err = fmt.Errorf("wrong request argument key '%s'", key)
			return false
		}
		if d[0] != '*' {
			ret.Set(kv.Key(d[1:]), value)
			return true
		}
		// d[0] == '*'
		if len(value) < hashing.HashSize {
			err = fmt.Errorf("wrong request argument '%s'", key)
			return false
		}
		h, err = hashing.HashValueFromBytes(value[:hashing.HashSize])
		if err != nil {
			ok = false
			return false
		}
		data, ok, err = reg.GetBlob(h)
		if err != nil {
			ok = false
			return false
		}
		if !ok {
			ok = false
			return false
		}
		ret.Set(kv.Key(d[1:]), data)
		return true
	})
	if err != nil || !ok {
		ret = nil
	}
	return ret, ok, err
}
