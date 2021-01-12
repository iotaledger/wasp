package blob

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("blob.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

// storeBlob treats parameters as names of fields and field values
// it stores it in the state in deterministic binary representation
// Returns hash of the blob
func storeBlob(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("blob.storeBlob.begin")
	state := ctx.State()
	params := ctx.Params()
	// calculate a deterministic hash of all blob fields
	blobHash, kSorted, values := mustGetBlobHash(params)

	directory := GetDirectory(state)
	if directory.MustHasAt(blobHash[:]) {
		// blob already exists
		return nil, fmt.Errorf("blob.storeBlob.fail: blob with hash %s already exist", blobHash.String())
	}
	// get a record by blob hash
	blbValues := GetBlobValues(state, blobHash)
	blbSizes := GetBlobSizes(state, blobHash)

	totalSize := uint32(0)

	// save record of the blob. In parallel save record of sizes of blob fields
	sizes := make([]uint32, len(kSorted))
	for i, k := range kSorted {
		size := uint32(len(values[i]))

		blbValues.MustSetAt([]byte(k), values[i])
		blbSizes.MustSetAt([]byte(k), EncodeSize(size))
		sizes[i] = size
		totalSize += size
	}

	ret := dict.New()
	ret.Set(ParamHash, codec.EncodeHashValue(&blobHash))

	directory.MustSetAt(blobHash[:], EncodeSize(totalSize))

	ctx.Event(fmt.Sprintf("[blob] hash: %s, field sizes: %+v", blobHash.String(), sizes))

	ctx.Log().Debugf("blob.storeBlob.success hash = %s", blobHash.String())
	return ret, nil
}

// getBlobInfo return lengths of all fields in the blob
func getBlobInfo(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("blob.getBlobInfo.begin")
	state := ctx.State()
	blobHash, ok, err := codec.DecodeHashValue(ctx.Params().MustGet(ParamHash))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'blob hash' not found")
	}
	blbSizes := GetBlobSizesR(state, *blobHash)
	ret := dict.New()
	blbSizes.MustIterate(func(field []byte, value []byte) bool {
		ret.Set(kv.Key(field), value)
		return true
	})
	return ret, nil
}

func getBlobField(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("blob.getBlobField.begin")
	state := ctx.State()

	blobHash, ok, err := codec.DecodeHashValue(ctx.Params().MustGet(ParamHash))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'blob hash' not found")
	}

	field := ctx.Params().MustGet(ParamField)
	if field == nil {
		return nil, fmt.Errorf("parameter 'blob field' not found")
	}

	blbValues := GetBlobValuesR(state, *blobHash)
	if blbValues.MustLen() == 0 {
		return nil, fmt.Errorf("blob with hash %s has not been found", blobHash.String())
	}
	value := blbValues.MustGetAt(field)
	if value == nil {
		return nil, fmt.Errorf("'blob field %s value not found", string(field))
	}
	ret := dict.New()
	ret.Set(ParamBytes, value)
	return ret, nil
}

func listBlobs(ctx vmtypes.SandboxView) (dict.Dict, error) {
	ctx.Log().Debugf("blob.listBlobs.begin")
	ret := dict.New()
	GetDirectoryR(ctx.State()).MustIterate(func(hash []byte, totalSize []byte) bool {
		ret.Set(kv.Key(hash), totalSize)
		return true
	})
	return ret, nil
}
