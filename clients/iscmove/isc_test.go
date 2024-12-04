package iscmove_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestIscCodec(t *testing.T) {
	type ExampleObj struct {
		A int
	}

	bcs.TestCodec(t, iscmove.RefWithObject[ExampleObj]{
		ObjectRef: *iotatest.RandomObjectRef(),
		Object:    &ExampleObj{A: 42},
		Owner:     iotago.MustAddressFromHex(testcommon.TestAddress),
	})

	anchor := iscmovetest.RandomAnchor()

	var digest iotago.Base58
	_, err := rand.Read(digest)
	require.NoError(t, err)

	anchorRef := iscmove.RefWithObject[iscmove.Anchor]{
		ObjectRef: iotago.ObjectRef{
			ObjectID: &anchor.ID,
			Version:  13,
			Digest:   &digest,
		},
		Object: &anchor,
		Owner:  iotago.MustAddressFromHex(testcommon.TestAddress),
	}

	bcs.TestCodec(t, anchorRef)
}
