package processor

import "time"

// each processor is locked during run. This is the timeout to acquire lock on
// particular processor
const processorAcquireTimeout = 2 * time.Second
