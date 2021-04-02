package mempool

import (
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
)

func TestMempool(t *testing.T) {
	m := New(coretypes.NewDummyBlobCache(), testlogger.NewLogger(t))
	time.Sleep(2 * time.Second)
	m.Close()
	time.Sleep(1 * time.Second)
}
