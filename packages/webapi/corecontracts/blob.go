package corecontracts

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetBlobInfo(ch chain.Chain, blobHash hashing.HashValue, blockIndexOrTrieRoot string) (map[string]uint32, bool, error) {
	ret, err := common.CallView(ch, blob.ViewGetBlobInfo.Message(blobHash), blockIndexOrTrieRoot)
	if err != nil {
		return nil, false, err
	}
	fields, err := blob.ViewGetBlobInfo.Output.Decode(ret)
	return fields, len(fields) > 0, err
}

func GetBlobValue(ch chain.Chain, blobHash hashing.HashValue, key string, blockIndexOrTrieRoot string) ([]byte, error) {
	ret, err := common.CallView(ch, blob.ViewGetBlobField.Message(blobHash, []byte(key)), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return blob.ViewGetBlobField.Output.Decode(ret)
}
