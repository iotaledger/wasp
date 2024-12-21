package bcs_test

import (
	"testing"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestOption(t *testing.T) {
	bcs.TestCodecAndBytes(t, &bcs.Option[uint32]{None: true}, []byte{0x0})
	bcs.TestCodecAndBytes(t, &bcs.Option[uint32]{Some: 10}, []byte{0x1, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, &bcs.Option[*uint32]{Some: lo.ToPtr[uint32](10)}, []byte{0x1, 0xa, 0x0, 0x0, 0x0})
}
