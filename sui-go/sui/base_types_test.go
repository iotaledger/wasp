package sui_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func TestObjectRefReadWrite(t *testing.T) {
	ref := sui.RandomObjectRef()
	rwutil.ReadWriteTest[*sui.ObjectRef](t, ref, &sui.ObjectRef{})
}
