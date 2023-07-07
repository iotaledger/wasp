package corecontracts

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetBlobInfo(ch chain.Chain, blobHash hashing.HashValue, blockIndexOrTrieRoot string) (map[string]uint32, bool, error) {
	ret, err := common.CallView(
		ch,
		blob.Contract.Hname(),
		blob.ViewGetBlobInfo.Hname(),
		codec.MakeDict(map[string]interface{}{blob.ParamHash: blobHash.Bytes()}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, false, err
	}

	if ret.IsEmpty() {
		return nil, false, nil
	}

	blobMap, err := blob.DecodeSizesMap(ret)
	if err != nil {
		return nil, false, err
	}

	return blobMap, true, nil
}

func GetBlobValue(ch chain.Chain, blobHash hashing.HashValue, key string, blockIndexOrTrieRoot string) ([]byte, error) {
	ret, err := common.CallView(
		ch,
		blob.Contract.Hname(),
		blob.ViewGetBlobField.Hname(),
		codec.MakeDict(map[string]interface{}{
			blob.ParamHash:  blobHash.Bytes(),
			blob.ParamField: []byte(key),
		}),
		blockIndexOrTrieRoot,
	)
	if err != nil {
		return nil, err
	}

	return ret[blob.ParamBytes], nil
}

func ListBlobs(ch chain.Chain, blockIndexOrTrieRoot string) (map[hashing.HashValue]uint32, error) {
	ret, err := common.CallView(ch, blob.Contract.Hname(), blob.ViewListBlobs.Hname(), nil, blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	return blob.DecodeDirectory(ret)
}
