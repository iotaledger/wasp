// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil // not `..._test` because it uses peeringMsg.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeeringNetReliable(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg)
	doneCh := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			<-outCh
		}
		doneCh <- true
	}()
	var someNode = peeringNode{netID: "src"}
	var behavior PeeringNetBehavior
	behavior = NewPeeringNetReliable()
	behavior.AddLink(inCh, outCh, "dst")
	for i := 0; i < 10; i++ {
		inCh <- &peeringMsg{from: &someNode}
	}
	<-doneCh
	behavior.Close()
}

func TestPeeringNetUnreliable(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg)
	//
	// Receiver process.
	stopCh := make(chan bool)
	startTime := time.Now()
	durations := make([]time.Duration, 0)
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-outCh:
				durations = append(durations, time.Since(startTime))
			}
		}
	}()
	//
	// Run the test.
	var someNode = peeringNode{netID: "src"}
	var behavior PeeringNetBehavior
	behavior = NewPeeringNetUnreliable(50, 50, 50*time.Millisecond, 100*time.Millisecond, NewLogger(t))
	behavior.AddLink(inCh, outCh, "dst")
	for i := 0; i < 1000; i++ {
		inCh <- &peeringMsg{from: &someNode}
	}
	time.Sleep(500 * time.Millisecond)
	//
	// Verify the results (with some tolerance for randomness).
	{ // 50% of messages dropped + 50% duplicated -> delivered ~75%
		require.Greater(t, len(durations), 500)
		require.Less(t, len(durations), 900)
	}
	{ // Average should be between the specified boundaries.
		var avgDuration int64 = 0
		for _, d := range durations {
			avgDuration += d.Milliseconds()
		}
		avgDuration = avgDuration / int64(len(durations))
		require.Greater(t, avgDuration, int64(50))
		require.Less(t, avgDuration, int64(100))
	}
	//
	// Stop the test.
	stopCh <- true
	behavior.Close()
}

func TestPeeringNetGoodQuality(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg)
	//
	// Receiver process.
	stopCh := make(chan bool)
	startTime := time.Now()
	durations := make([]time.Duration, 0)
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-outCh:
				durations = append(durations, time.Since(startTime))
			}
		}
	}()
	//
	// Run the test.
	var someNode = peeringNode{netID: "src"}
	var behavior PeeringNetBehavior
	behavior = NewPeeringNetUnreliable(100, 0, 0*time.Microsecond, 0*time.Millisecond, NewLogger(t)) // NOTE: No drops, duplicates, delays.
	behavior.AddLink(inCh, outCh, "dst")
	for i := 0; i < 1000; i++ {
		inCh <- &peeringMsg{from: &someNode}
	}
	time.Sleep(500 * time.Millisecond)
	//
	// Verify the results (with some tolerance for randomness).
	{ // All messages should be delivered.
		require.Equal(t, 1000, len(durations))
	}
	{ // Average should be small enough.
		var avgDuration int64 = 0
		for _, d := range durations {
			avgDuration += d.Milliseconds()
		}
		avgDuration = avgDuration / int64(len(durations))
		require.Less(t, avgDuration, int64(100))
	}
	//
	// Stop the test.
	stopCh <- true
	behavior.Close()
}
