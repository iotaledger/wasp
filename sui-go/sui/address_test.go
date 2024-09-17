package sui_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func TestAddressReadWrite(t *testing.T) {
	address := sui.RandomAddress()
	rwutil.ReadWriteTest[*sui.Address](t, address, &sui.Address{})
}
