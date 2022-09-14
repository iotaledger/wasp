// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util"
)

func TestWaitChan(t *testing.T) {
	wc := util.NewWaitChan()
	// Done should not block.
	wc.Done()
	wc.Done()
	wc.Done()
	// Wait does not block after Done.
	wc.Wait()
	wc.Wait()
	wc.Wait()
	wc.Wait()
	// Reset does not block.
	wc.Reset()
	wc.Reset()
	// Wait for event wo timeout,
	wc.Reset()
	waitStart := time.Now()
	go func() {
		time.Sleep(100 * time.Millisecond)
		wc.Done()
	}()
	wc.Wait()
	waitDuration := time.Since(waitStart)
	require.Greater(t, waitDuration.Milliseconds(), int64(50))
	require.Greater(t, int64(500), waitDuration.Milliseconds())
	// Wait for event with a timeout,
	wc.Reset()
	waitTimeoutStart := time.Now()
	require.False(t, wc.WaitTimeout(100*time.Millisecond))
	waitTimeoutDuration := time.Since(waitTimeoutStart)
	require.Greater(t, waitTimeoutDuration.Milliseconds(), int64(50))
	require.Greater(t, int64(500), waitTimeoutDuration.Milliseconds())
	// Wait for an event with a timeout, immediate,
	wc.Done()
	waitTimeoutImmediateStart := time.Now()
	require.True(t, wc.WaitTimeout(100*time.Millisecond))
	waitTimeoutImmediateDuration := time.Since(waitTimeoutImmediateStart)
	require.Greater(t, int64(50), waitTimeoutImmediateDuration.Milliseconds())
}
