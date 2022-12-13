package corecontracts

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Blob struct {
	vmService interfaces.VMService
}

func NewBlob(vmService interfaces.VMService) *Blob {
	return &Blob{
		vmService: vmService,
	}
}

func (b *Blob) GetBlobInfo(chainID *isc.ChainID, blobHash hashing.HashValue) (map[string]uint32, bool, error) {
	ret, err := b.vmService.CallViewByChainID(chainID, blob.Contract.Hname(), blob.ViewGetBlobInfo.Hname(), codec.MakeDict(map[string]interface{}{
		blob.ParamHash: blobHash[:],
	}))

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

func (b *Blob) GetBlobValue(chainID *isc.ChainID, blobHash hashing.HashValue, key string) ([]byte, error) {
	ret, err := b.vmService.CallViewByChainID(chainID, blob.Contract.Hname(), blob.ViewGetBlobInfo.Hname(), codec.MakeDict(map[string]interface{}{
		blob.ParamHash:  blobHash[:],
		blob.ParamField: []byte(key),
	}))

	if err != nil {
		return nil, err
	}

	return ret[blob.ParamBytes], nil
}

func (b *Blob) ListBlobs(chainID *isc.ChainID) (map[hashing.HashValue]uint32, error) {
	ret, err := b.vmService.CallViewByChainID(chainID, blob.Contract.Hname(), blob.ViewListBlobs.Hname(), nil)

	if err != nil {
		return nil, err
	}

	return blob.DecodeDirectory(ret)
}
