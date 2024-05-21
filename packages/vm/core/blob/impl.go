package blob

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
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

func (s *StateWriter) SetInitialState() {
	// does not do anything
}

var errBlobAlreadyExists = coreerrors.Register("blob already exists").Create()

// storeBlob treats parameters as names of fields and field values
// it stores it in the state in deterministic binary representation
// Returns hash of the blob
func storeBlob(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("blob.storeBlob.begin")
	state := NewStateWriterFromSandbox(ctx)
	params := ctx.Params()
	// calculate a deterministic hash of all blob fields
	blobHash, fieldsSorted, valuesSorted := mustGetBlobHash(params.Dict)

	directory := state.GetDirectory()
	if directory.HasAt(blobHash[:]) {
		panic(errBlobAlreadyExists)
	}

	// get a record by blob hash
	blbValues := state.GetBlobValues(blobHash)
	blbSizes := state.GetBlobSizes(blobHash)

	totalSize := uint32(0)
	totalSizeWithKeys := uint32(0)

	// save record of the blob.
	for i, k := range fieldsSorted {
		size := uint32(len(valuesSorted[i]))
		blbValues.SetAt([]byte(k), valuesSorted[i])
		blbSizes.SetAt([]byte(k), EncodeSize(size))
		totalSize += size
		totalSizeWithKeys += size + uint32(len(k))
	}
	directory.SetAt(blobHash[:], EncodeSize(totalSize))

	eventStore(ctx, blobHash)

	return dict.Dict{ParamHash: codec.HashValue.Encode(blobHash)}
}

// getBlobInfo return lengths of all fields in the blob
func getBlobInfo(ctx isc.SandboxView, blobHash hashing.HashValue) map[string]uint32 {
	ctx.Log().Debugf("blob.getBlobInfo.begin")
	state := NewStateReaderFromSandbox(ctx)
	ret := map[string]uint32{}
	state.GetBlobSizes(blobHash).Iterate(func(field []byte, value []byte) bool {
		ret[string(field)] = lo.Must(DecodeSize(value))
		return true
	})
	return ret
}

var errNotFound = coreerrors.Register("not found").Create()

func getBlobField(ctx isc.SandboxView, blobHash hashing.HashValue, field []byte) []byte {
	ctx.Log().Debugf("blob.getBlobField.begin")
	state := NewStateReaderFromSandbox(ctx)
	blobValues := state.GetBlobValues(blobHash)
	if blobValues.Len() == 0 {
		panic(errNotFound)
	}
	value := blobValues.GetAt(field)
	if value == nil {
		panic(errNotFound)
	}
	return value
}

func listBlobs(ctx isc.SandboxView) map[hashing.HashValue]uint32 {
	ctx.Log().Debugf("blob.listBlobs.begin")
	ret := map[hashing.HashValue]uint32{}
	NewStateReaderFromSandbox(ctx).GetDirectory().Iterate(func(hash []byte, totalSize []byte) bool {
		ret[lo.Must(codec.HashValue.Decode(hash))] = lo.Must(DecodeSize(totalSize))
		return true
	})
	return ret
}
