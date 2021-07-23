package requestargs

import (
	"fmt"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/downloader"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
)

// TODO: extend '*' option in RequestArgs with download options (web, IPFS)

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

func (a RequestArgs) String() string {
	return (dict.Dict(a)).String()
}

func (a RequestArgs) Clone() RequestArgs {
	return RequestArgs((dict.Dict(a)).Clone())
}

func (a RequestArgs) Bytes() []byte {
	return (dict.Dict(a)).Bytes()
}

func (a RequestArgs) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	(dict.Dict(a)).WriteToMarshalUtil(mu)
}

func (a RequestArgs) ReadFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	return (dict.Dict(a)).ReadFromMarshalUtil(mu)
}

// SolidifyRequestArguments decodes RequestArgs.
// each key-value pair ir treated according to the first byte of the key:
//  - if the key starts with '*' the value is a content reference.
//    First 32 bytes of the value are always treated as data hash.
//    The rest (if any) is a content address. It will be treated by a downloader
//  - otherwise, value is treated a raw data and the first byte of the key is ignored
func (a RequestArgs) SolidifyRequestArguments(reg registry.BlobCache, downloaderOpt ...*downloader.Downloader) (dict.Dict, bool, error) {
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
		if ok {
			ret.Set(kv.Key(d[1:]), data)
			return true
		}
		contentAddr := value[hashing.HashSize:]
		if len(contentAddr) > 0 {
			downloaderObj := downloader.GetDefaultDownloader()
			if len(downloaderOpt) > 0 {
				downloaderObj = downloaderOpt[0]
			}
			if downloaderObj != nil {
				err = downloaderObj.DownloadAndStore(h, string(contentAddr), reg)
			}
		}
		return false
	})
	if err != nil || !ok {
		ret = nil
	}
	return ret, ok, err
}
