// package solobench provides tools to benchmark contracts running under solo
package solobench

import (
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

type Func func(b *testing.B, chain *solo.Chain, reqs []*solo.CallParams, keyPair *cryptolib.KeyPair)

// RunBenchmarkSync processes requests synchronously, producing 1 block per request
func RunBenchmarkSync(b *testing.B, chain *solo.Chain, reqs []*solo.CallParams, keyPair *cryptolib.KeyPair) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := chain.PostRequestSync(reqs[i], keyPair)
		require.NoError(b, err)
	}
}

// RunBenchmarkAsync processes requests asynchronously, producing 1 block per many requests
func RunBenchmarkAsync(b *testing.B, chain *solo.Chain, reqs []*solo.CallParams, keyPair *cryptolib.KeyPair) {
	txs := make([]*iotago.Transaction, b.N)
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
