// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil // not `..._test` because it uses peeringMsg.

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func TestPeeringNetReliable(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	doneCh := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			<-outCh
		}
		doneCh <- true
	}()
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	behavior := NewPeeringNetReliable(testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false))
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 10; i++ {
		inCh <- &peeringMsg{from: someNode.identity.GetPublicKey()}
	}
	<-doneCh
	behavior.Close()
}

func TestPeeringNetUnreliable(t *testing.T) {
	inCh := make(chan *peeringMsg, 2000)
	outCh := make(chan *peeringMsg, 2000)
	//
	// Receiver process.
	stopCh := make(chan bool)
	doneCh := make(chan []time.Duration, 1)
	startTime := time.Now()
	var recvCount atomic.Int64
	go func() {
		durations := make([]time.Duration, 0)
		for {
			select {
			case <-stopCh:
				doneCh <- durations
				return
			case <-outCh:
				durations = append(durations, time.Since(startTime))
				recvCount.Add(1)
			}
		}
	}()
	//
	// Run the test.
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	behavior := NewPeeringNetUnreliable(50, 50, 50*time.Millisecond, 100*time.Millisecond, testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false))
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 2000; i++ {
		inCh <- &peeringMsg{from: someNode.identity.GetPublicKey()}
	}
	require.Eventually(t, func() bool {
		return recvCount.Load() >= 1000
	}, 5*time.Second, 100*time.Millisecond, "expected at least 1000 messages to be processed")
	stopCh <- true
	durations := <-doneCh

	//
	// Validate the results (with some tolerance for randomness).
	{ // 50% of messages dropped + 50% duplicated -> delivered ~75%
		require.Greater(t, len(durations), 1000)
		require.Less(t, len(durations), 1800)
	}
	{ // Average should be between the specified boundaries.
		var avgDuration int64 = 0
		for _, d := range durations {
			avgDuration += d.Milliseconds()
		}
		avgDuration /= int64(len(durations))
		require.GreaterOrEqual(t, avgDuration, int64(50))
		require.LessOrEqual(t, avgDuration, int64(200))
	}

	behavior.Close()
}

func TestPeeringNetGoodQuality(t *testing.T) {
	inCh := make(chan *peeringMsg, 1000)
	outCh := make(chan *peeringMsg, 1000)
	//
	// Receiver process.
	stopCh := make(chan bool)
	doneCh := make(chan []time.Duration, 1)
	startTime := time.Now()
	var recvCount atomic.Int64
	go func() {
		durations := make([]time.Duration, 0)
		for {
			select {
			case <-stopCh:
				doneCh <- durations
				return
			case <-outCh:
				durations = append(durations, time.Since(startTime))
				recvCount.Add(1)
			}
		}
	}()
	//
	// Run the test.
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	behavior := NewPeeringNetUnreliable(100, 0, 0*time.Microsecond, 0*time.Millisecond, testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false)) // NOTE: No drops, duplicates, delays.
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 1000; i++ {
		inCh <- &peeringMsg{from: someNode.identity.GetPublicKey()}
	}
	require.Eventually(t, func() bool {
		return recvCount.Load() >= 1000
	}, 5*time.Second, 50*time.Millisecond, "expected all 1000 messages to be processed")
	stopCh <- true
	durations := <-doneCh

	//
	// Validate the results (with some tolerance for randomness).
	{ // All messages should be delivered.
		require.Equal(t, 1000, len(durations))
	}
	{ // Average should be small enough.
		var avgDuration int64 = 0
		for _, d := range durations {
			avgDuration += d.Milliseconds()
		}
		avgDuration /= int64(len(durations))
		require.Less(t, avgDuration, int64(100))
	}

	behavior.Close()
}
