// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblob"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

// this is the expected blob hash for key0/val0 key1/val1
const expectedBlobHash = "0x5fec3bfc701d80bdf75e337cb3dcb401c2423d15fc17a74d5b644dae143118b1"

func setupBlob(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreblob.ScName, coreblob.OnDispatch)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestStoreBlob(t *testing.T) {
	ctx := setupBlob(t)
	require.NoError(t, ctx.Err)

	f := coreblob.ScFuncs.StoreBlob(ctx)
	f.Params.Blobs().GetBytes("key0").SetValue([]byte("val0"))
	f.Params.Blobs().GetBytes("key1").SetValue([]byte("val1"))
	f.Func.Post()
	require.NoError(t, ctx.Err)
	require.Equal(t, expectedBlobHash, f.Results.Hash().Value().String())
}

func TestGetBlobInfo(t *testing.T) {
	ctx := setupBlob(t)
	require.NoError(t, ctx.Err)

	fStore := coreblob.ScFuncs.StoreBlob(ctx)
	fStore.Params.Blobs().GetBytes("key0").SetValue([]byte("val0"))
	fStore.Params.Blobs().GetBytes("key1").SetValue([]byte("val1"))
	fStore.Func.Post()
	require.NoError(t, ctx.Err)
	require.Equal(t, expectedBlobHash, fStore.Results.Hash().Value().String())

	fList := coreblob.ScFuncs.GetBlobInfo(ctx)
	fList.Params.Hash().SetValue(wasmtypes.HashFromString(expectedBlobHash))
	fList.Func.Call()
	size := fList.Results.BlobSizes().GetInt32("key0").Value()
	require.Equal(t, int32(4), size)
}

func TestGetBlobField(t *testing.T) {
	ctx := setupBlob(t)
	require.NoError(t, ctx.Err)

	fStore := coreblob.ScFuncs.StoreBlob(ctx)
	fStore.Params.Blobs().GetBytes("key0").SetValue([]byte("val0"))
	fStore.Params.Blobs().GetBytes("key1").SetValue([]byte("val1"))
	fStore.Func.Post()
	require.NoError(t, ctx.Err)
	require.Equal(t, expectedBlobHash, fStore.Results.Hash().Value().String())

	fList := coreblob.ScFuncs.GetBlobField(ctx)
	fList.Params.Field().SetValue("key0")
	fList.Params.Hash().SetValue(wasmtypes.HashFromString(expectedBlobHash))
	fList.Func.Call()
	stored := fList.Results.Bytes().Value()
	require.Equal(t, []byte("val0"), stored)
}

func TestListBlobs(t *testing.T) {
	ctx := setupBlob(t)
	require.NoError(t, ctx.Err)

	fStore := coreblob.ScFuncs.StoreBlob(ctx)
	fStore.Params.Blobs().GetBytes("key0").SetValue([]byte("val0"))
	fStore.Params.Blobs().GetBytes("key1").SetValue([]byte("_val1"))
	fStore.Func.Post()
	require.NoError(t, ctx.Err)
	expectedHash := "0x462af4abe5977f4dd985a0a097705925b9fa6c033c9d931c1e2171f710693462"
	require.Equal(t, expectedHash, fStore.Results.Hash().Value().String())

	fList := coreblob.ScFuncs.ListBlobs(ctx)
	fList.Func.Call()
	size := fList.Results.BlobSizes().GetInt32(wasmtypes.HashFromString(expectedHash)).Value()
	// The sum of the size of the value of `key0` and `key1` is len("val0")+len("_val1") = 9
	require.Equal(t, int32(9), size)
}
