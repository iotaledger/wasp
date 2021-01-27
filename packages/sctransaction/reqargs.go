package sctransaction

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// RequestArguments encodes request parameters taking into account hashes of data blobs
type RequestArguments dict.Dict

// NewRequestArguments constructor
func NewRequestArguments() RequestArguments {
	return RequestArguments(dict.New())
}

// AddArgument add new ordinary argument. Encodes the key as "normal"
func (a RequestArguments) AddArgument(name string, data []byte) {
	a[kv.Key("-"+name)] = data
}

// AddArgumentAsBlobHash adds argument with the data hash instead of data itself.
// Encodes key as "blob reference"
func (a RequestArguments) AddArgumentAsBlobHash(name string, data []byte) {
	h := hashing.HashData(data)
	a[kv.Key("*"+name)] = h[:]
}

// HasBlogRef return if request arguments contain at least one blog reference
func (a RequestArguments) HasBlogRef() bool {
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
