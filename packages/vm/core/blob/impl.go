package blob

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var Processor = Contract.Processor(nil,
	FuncStoreBlob.WithHandler(storeBlob),
	ViewGetBlobField.WithHandler(getBlobField),
	ViewGetBlobInfo.WithHandler(getBlobInfo),
	ViewListBlobs.WithHandler(listBlobs),
)

func SetInitialState(state kv.KVStore) {
	// does not do anything
}

var errBlobAlreadyExists = coreerrors.Register("blob already exists").Create()

// storeBlob treats parameters as names of fields and field values
// it stores it in the state in deterministic binary representation
// Returns hash of the blob
func storeBlob(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("blob.storeBlob.begin")
	state := ctx.State()
	params := ctx.Params()
	// calculate a deterministic hash of all blob fields
	blobHash, kSorted, values := mustGetBlobHash(params.Dict)

	directory := GetDirectory(state)
	if directory.HasAt(blobHash[:]) {
		panic(errBlobAlreadyExists)
	}

	// get a record by blob hash
	blbValues := GetBlobValues(state, blobHash)
	blbSizes := GetBlobSizes(state, blobHash)

	totalSize := uint32(0)
	totalSizeWithKeys := uint32(0)

	// save record of the blob.
	for i, k := range kSorted {
		size := uint32(len(values[i]))
		blbValues.SetAt([]byte(k), values[i])
		blbSizes.SetAt([]byte(k), EncodeSize(size))
		totalSize += size
		totalSizeWithKeys += size + uint32(len(k))
	}

	ret := dict.New()
	ret.Set(ParamHash, codec.EncodeHashValue(blobHash))

	directory.SetAt(blobHash[:], EncodeSize(totalSize))

	eventStore(ctx, blobHash)
	return ret
}

// getBlobInfo return lengths of all fields in the blob
func getBlobInfo(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("blob.getBlobInfo.begin")

	blobHash := ctx.Params().MustGetHashValue(ParamHash)

	blbSizes := GetBlobSizesR(ctx.StateR(), blobHash)
	ret := dict.New()
	blbSizes.Iterate(func(field []byte, value []byte) bool {
		ret.Set(kv.Key(field), value)
		return true
	})
	return ret
}

var errNotFound = coreerrors.Register("not found").Create()

func getBlobField(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("blob.getBlobField.begin")
	state := ctx.StateR()

	params := ctx.Params()
	blobHash := params.MustGetHashValue(ParamHash)
	field := params.MustGetBytes(ParamField)

	blobValues := GetBlobValuesR(state, blobHash)
	if blobValues.Len() == 0 {
		panic(errNotFound)
	}
	value := blobValues.GetAt(field)
	if value == nil {
		panic(errNotFound)
	}
	ret := dict.New()
	ret.Set(ParamBytes, value)
	return ret
}

func listBlobs(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("blob.listBlobs.begin")
	ret := dict.New()
	GetDirectoryR(ctx.StateR()).Iterate(func(hash []byte, totalSize []byte) bool {
		ret.Set(kv.Key(hash), totalSize)
		return true
	})
	return ret
}
