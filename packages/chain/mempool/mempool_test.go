package mempool

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/solo"
)

func TestMempool(t *testing.T) {
	m := New(solo.NewDummyBlobCache())
	time.Sleep(2 * time.Second)
	m.Close()
	time.Sleep(1 * time.Second)
}
