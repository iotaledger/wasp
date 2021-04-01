package mempool

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/solo"
	"testing"
	"time"
)

func TestMempool(t *testing.T) {
	m := chain.NewMempool(solo.NewDummyBlobCache())
	time.Sleep(2 * time.Second)
	m.Close()
	time.Sleep(1 * time.Second)
}
