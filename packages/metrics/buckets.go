package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	postTimeBuckets = prometheus.ExponentialBucketsRange(0.1, 60*60, 17) // Time to confirm/reject a TX in L1 [0.1s - 1h].
	execTimeBuckets = prometheus.ExponentialBucketsRange(0.01, 100, 17)  // Execution of misc functions.
	recCountBuckets = prometheus.ExponentialBucketsRange(1, 1000, 16)
)
