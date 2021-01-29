package sctransaction

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"io"
)

// RequestArgs encodes request parameters taking into account hashes of data blobs
type RequestArgs dict.Dict

// NewRequestArgs constructor
func NewRequestArgs() RequestArgs {
	return RequestArgs(dict.New())
}

func NewRequestArgsFromDict(d dict.Dict) RequestArgs {
	ret := NewRequestArgs()
	for k, v := range d {
		ret.Add(k, v)
	}
	return ret
}

const optimalSize = 32

func NewOptimizedRequestArgs(d dict.Dict, optSize ...int) (RequestArgs, map[kv.Key][]byte) {
	var osize int
	if len(optSize) > 0 {
		osize = optSize[0]
	}
	if osize <= optimalSize {
		osize = optimalSize
	}
	ret := NewRequestArgs()
	retOptimized := make(map[kv.Key][]byte)
	for k, v := range d {
		if len(v) <= osize {
			ret.Add(k, v)
		} else {
			ret.AddAsBlobHash(k, v)
			retOptimized[k] = v
		}
	}
	return ret, retOptimized
}

// Add add new ordinary argument. Encodes the key as "normal"
func (a RequestArgs) Add(name kv.Key, data []byte) {
	a["-"+name] = data
}

// AddAsBlobHash adds argument with the data hash instead of data itself.
// Encodes key as "blob reference"
func (a RequestArgs) AddAsBlobHash(name kv.Key, data []byte) hashing.HashValue {
	h := hashing.HashData(data)
	a.AddBlobRef(name, h)
	return h
}

// AddBlobRef adds hash as data and marks it is a blob reference
func (a RequestArgs) AddBlobRef(name kv.Key, hash hashing.HashValue) {
	a["*"+name] = hash[:]
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

// SolidifyRequestArguments decodes RequestArgs. For each blob reference encoded it
// looks for the data by hash into the registry and replaces dict entry with the data
// It returns ok flag == false if at least one blob hash don't have data in the registry
func (a RequestArgs) SolidifyRequestArguments(reg registry.BlobRegistryProvider) (dict.Dict, bool, error) {
	ret := dict.New()
	ok := true
	var err error
	var data []byte
	var h hashing.HashValue
	reqArgsDict := dict.Dict(a)
	reqArgsDict.ForEach(func(key kv.Key, value []byte) bool {
		d := []byte(key)
		if d[0] == '*' {
			h, err = hashing.HashValueFromBytes(value)
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
		} else {
			ret.Set(kv.Key(d[1:]), value)
		}
		return true
	})
	if err != nil || !ok {
		ret = nil
	}
	return ret, ok, err
}
