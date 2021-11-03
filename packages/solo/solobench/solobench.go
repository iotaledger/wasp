// package solobench provides tools to benchmark contracts running under solo
package solobench

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

type Func func(b *testing.B, chain *solo.Chain, reqs []*solo.CallParams, keyPair *ed25519.KeyPair)

// RunBenchmarkSync processes requests synchronously, producing 1 block per request
func RunBenchmarkSync(b *testing.B, chain *solo.Chain, reqs []*solo.CallParams, keyPair *ed25519.KeyPair) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := chain.PostRequestSync(reqs[i], keyPair)
		require.NoError(b, err)
	}
}

// RunBenchmarkAsync processes requests asynchronously, producing 1 block per many requests
func RunBenchmarkAsync(b *testing.B, chain *solo.Chain, reqs []*solo.CallParams, keyPair *ed25519.KeyPair) {
	txs := make([]*ledgerstate.Transaction, b.N)
	for i := 0; i < b.N; i++ {
		var err error
		txs[i], _, err = chain.RequestFromParamsToLedger(reqs[i], nil)
		require.NoError(b, err)
	}

	nreq := chain.MempoolInfo().InBufCounter

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go chain.Env.EnqueueRequests(txs[i])
	}
	require.True(b, chain.WaitForRequestsThrough(nreq+b.N, 20*time.Second))
}
