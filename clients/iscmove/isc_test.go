package iscmove_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/sui/suitest"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestIscCodec(t *testing.T) {
	type ExampleObj struct {
		A int
	}

	bcs.TestCodec(t, iscmove.RefWithObject[ExampleObj]{
		ObjectRef: *suitest.RandomObjectRef(),
		Object:    &ExampleObj{A: 42},
	})

	anchor := iscmovetest.RandomAnchor()

	var digest sui2.Base58
	_, err := rand.Read(digest)
	require.NoError(t, err)

	anchorRef := iscmove.RefWithObject[iscmove.Anchor]{
		ObjectRef: sui2.ObjectRef{
			ObjectID: &anchor.ID,
			Version:  13,
			Digest:   &digest,
		},
		Object: &anchor,
	}

	bcs.TestCodec(t, anchorRef)
}
