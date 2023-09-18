// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblob"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

// this is the expected blob hash for key0/val0 key1/val1
const expectedBlobHash = "0x54cb8e9c45ca6d368dba92da34cfa47ce617f04807af19f67de333fad0039e6b"

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
	fStore.Params.Blobs().GetBytes("key1").SetValue([]byte("val1"))
	fStore.Func.Post()
	require.NoError(t, ctx.Err)
	require.Equal(t, expectedBlobHash, fStore.Results.Hash().Value().String())

	fList := coreblob.ScFuncs.ListBlobs(ctx)
	fList.Func.Call()
	size := fList.Results.BlobSizes().GetInt32(wasmtypes.HashFromString(expectedBlobHash)).Value()
	// The sum of the size of the value of `key0` and `key1` is len("val0")+len("val1") = 8
	require.Equal(t, int32(8), size)
}
