// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil // not `..._test` because it uses peeringMsg.

import (
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/stretchr/testify/require"
)

func TestPeeringNetDynamicReliable(t *testing.T) {
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
	behavior = NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false))
	behavior.AddLink(inCh, outCh, "dst")
	for i := 0; i < 10; i++ {
		inCh <- &peeringMsg{from: &someNode}
	}
	<-doneCh
	behavior.Close()
}

func TestPeeringNetDynamicUnreliable(t *testing.T) {
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
	behavior = NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false)).
		WithLosingChannel(nil, 50).
		WithRepeatingChannel(nil, 50).
		WithDelayingChannel(nil, 50*time.Millisecond, 100*time.Millisecond)
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

func TestPeeringNetDynamicGoodQuality(t *testing.T) {
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
	behavior = NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false)).
		WithLosingChannel(nil, 100).
		WithRepeatingChannel(nil, 0).
		WithDelayingChannel(nil, 0*time.Millisecond, 0*time.Millisecond)
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

func TestPeeringNetDynamicChanging(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg)
	//
	// Receiver process.
	stopCh := make(chan bool)
	durations := make([]time.Duration, 0)
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case msg := <-outCh:
				durations = append(durations, time.Since(time.Unix(0, msg.msg.Timestamp)))
			}
		}
	}()
	var someNode = peeringNode{netID: "src"}
	sendFun := func() {
		inCh <- &peeringMsg{from: &someNode, msg: peering.PeerMessage{Timestamp: time.Now().UnixNano()}}
	}
	averageDurationFun := func() int64 {
		var result int64 = 0
		for _, d := range durations {
			result += d.Milliseconds()
		}
		return result / int64(len(durations))
	}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false))
	behavior.AddLink(inCh, outCh, "dst")
	for i := 0; i < 100; i++ {
		sendFun()
	}
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 100, len(durations))
	require.Less(t, averageDurationFun(), int64(20))
	durations = durations[:0]

	deliver40Name := "Deliver40"
	deliver70Name := "Deliver70"
	behavior.WithLosingChannel(&deliver70Name, 70).WithLosingChannel(&deliver40Name, 40) // 70% * 40% = 28% delivery probability
	for i := 0; i < 1000; i++ {
		sendFun()
	}
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 280, len(durations), 90)
	require.Less(t, averageDurationFun(), int64(20))
	durations = durations[:0]

	delayName := "Delay"
	behavior.WithDelayingChannel(&delayName, 20*time.Millisecond, 70*time.Millisecond) // 28% delivery probability and 20-70 ms delay
	for i := 0; i < 1000; i++ {
		sendFun()
	}
	time.Sleep(150 * time.Millisecond)
	require.InDelta(t, 280, len(durations), 90)
	require.InDelta(t, 45, averageDurationFun(), 20)
	durations = durations[:0]

	behavior.RemoveHandler(deliver40Name) // 70% delivery probability and 20-70 ms delay
	for i := 0; i < 1000; i++ {
		sendFun()
	}
	time.Sleep(150 * time.Millisecond)
	require.InDelta(t, 700, len(durations), 90)
	require.InDelta(t, 45, averageDurationFun(), 20)
	durations = durations[:0]

	behavior.RemoveHandler(delayName) // 70% delivery probability without a delay
	for i := 0; i < 1000; i++ {
		sendFun()
	}
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 700, len(durations), 90)
	require.Less(t, averageDurationFun(), int64(20))
	durations = durations[:0]

	// Stop the test.
	stopCh <- true
	behavior.Close()
}
