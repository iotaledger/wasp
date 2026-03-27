// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil // not `..._test` because it uses peeringMsg.

import (
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func TestPeeringNetDynamicReliable(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	doneCh := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			<-outCh
		}
		doneCh <- true
	}()
	// peerNetI, peerIdentities := testpeers.SetupKeys(2)
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false))
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 10; i++ {
		sendMessage(&someNode, inCh)
	}
	//
	// Stop the test.
	<-doneCh
	behavior.Close()
}

func TestPeeringNetDynamicUnreliable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	recvLoop := runTestRecvLoop(outCh)
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false)).
		WithLosingChannel(nil, 50).
		WithRepeatingChannel(nil, 50).
		WithDelayingChannel(nil, 50*time.Millisecond, 100*time.Millisecond)
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	//
	// Validate the results (with some tolerance for randomness).
	// 50% of messages dropped + 50% duplicated -> delivered ~75%
	// Average should be between the specified boundaries.
	require.Eventually(t, func() bool {
		count := recvLoop.ReceivedCount()
		avgDur := recvLoop.AverageDuration()
		return count > 500 && count < 900 && avgDur > 50 && avgDur < 100
	}, 5*time.Second, 10*time.Millisecond)
	//
	// Stop the test.
	recvLoop.Stop()
	behavior.Close()
}

func TestPeeringNetDynamicChanging(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	recvLoop := runTestRecvLoop(outCh)
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false))
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 100; i++ {
		sendMessage(&someNode, inCh)
	}
	require.Eventually(t, func() bool {
		return recvLoop.ReceivedCount() == 100 && recvLoop.AverageDuration() < int64(20)
	}, 5*time.Second, 10*time.Millisecond, "expected 100 messages with avg duration < 20ms")
	recvLoop.Reset()

	deliver40Name := "Deliver40"
	deliver70Name := "Deliver70"
	behavior.WithLosingChannel(&deliver70Name, 70).WithLosingChannel(&deliver40Name, 40) // 70% * 40% = 28% delivery probability
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	require.Eventually(t, func() bool {
		return math.Abs(float64(recvLoop.ReceivedCount())-280) <= 90 && recvLoop.AverageDuration() < int64(20)
	}, 5*time.Second, 10*time.Millisecond, "expected ~280 messages (±90) with avg duration < 20ms")
	recvLoop.Reset()

	delayName := "Delay"
	behavior.WithDelayingChannel(&delayName, 20*time.Millisecond, 70*time.Millisecond) // 28% delivery probability and 20-70 ms delay
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	require.Eventually(t, func() bool {
		return math.Abs(float64(recvLoop.ReceivedCount())-280) <= 90 && math.Abs(float64(recvLoop.AverageDuration())-45) <= 20
	}, 5*time.Second, 10*time.Millisecond, "expected ~280 messages (±90) with avg duration ~45ms (±20ms)")
	recvLoop.Reset()

	behavior.RemoveHandler(deliver40Name) // 70% delivery probability and 20-70 ms delay
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	require.Eventually(t, func() bool {
		return math.Abs(float64(recvLoop.ReceivedCount())-700) <= 90 && math.Abs(float64(recvLoop.AverageDuration())-45) <= 20
	}, 5*time.Second, 10*time.Millisecond, "expected ~700 messages (±90) with avg duration ~45ms (±20ms)")

	behavior.RemoveHandler(delayName) // 70% delivery probability without a delay
	// Let any in-flight delayed messages from the previous batch drain before
	// resetting stats, so they don't pollute the next section's average duration.
	time.Sleep(100 * time.Millisecond)
	recvLoop.Reset()
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	require.Eventually(t, func() bool {
		return math.Abs(float64(recvLoop.ReceivedCount())-700) <= 90 && recvLoop.AverageDuration() < int64(20)
	}, 5*time.Second, 10*time.Millisecond, "expected ~700 messages (±90) with avg duration < 20ms")
	recvLoop.Reset()

	// Stop the test.
	recvLoop.Stop()
	behavior.Close()
}

func TestPeeringNetDynamicLosingChannel(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	recvLoop := runTestRecvLoop(outCh)
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false)).WithLosingChannel(nil, 50)
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	require.Eventually(t, func() bool {
		return math.Abs(float64(recvLoop.ReceivedCount())-500) <= 90 && recvLoop.AverageDuration() < int64(20)
	}, 5*time.Second, 10*time.Millisecond, "expected ~500 messages (±90) with avg duration < 20ms")

	// Stop the test.
	recvLoop.Stop()
	behavior.Close()
}

func TestPeeringNetDynamicRepeatingChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 10000)
	recvLoop := runTestRecvLoop(outCh)
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false)).WithRepeatingChannel(nil, 150)
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	require.Eventually(t, func() bool {
		return math.Abs(float64(recvLoop.ReceivedCount())-2500) <= 90 && recvLoop.AverageDuration() < int64(20)
	}, 5*time.Second, 10*time.Millisecond, "expected ~2500 messages (±90) with avg duration < 20ms")

	// Stop the test.
	recvLoop.Stop()
	behavior.Close()
}

func TestPeeringNetDynamicDelayingChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	recvLoop := runTestRecvLoop(outCh)
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	someNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false)).WithDelayingChannel(nil, 25*time.Millisecond, 75*time.Millisecond)
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	for i := 0; i < 100; i++ {
		sendMessage(&someNode, inCh)
	}
	require.Eventually(t, func() bool {
		return recvLoop.ReceivedCount() == 100 && math.Abs(float64(recvLoop.AverageDuration())-50) <= 20
	}, 5*time.Second, 10*time.Millisecond, "expected 100 messages with avg duration ~50ms (±20ms)")

	// Stop the test.
	recvLoop.Stop()
	behavior.Close()
}

func TestPeeringNetDynamicPeerDisconnected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	inChD := make(chan *peeringMsg)
	outChD := make(chan *peeringMsg)
	recvLoop := runTestRecvLoop(outCh)
	recvLoopD := runTestRecvLoop(outChD)
	srcPeerIdentity := cryptolib.NewKeyPair()
	dstPeerIdentity := cryptolib.NewKeyPair()
	disPeerIdentity := cryptolib.NewKeyPair()
	connectedNode := peeringNode{peeringURL: "src", identity: srcPeerIdentity}
	disconnectedNode := peeringNode{peeringURL: "disconnected", identity: disPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), log.LevelError, false)).WithPeerDisconnected(nil, disPeerIdentity.GetPublicKey())
	behavior.AddLink(inCh, outCh, dstPeerIdentity.GetPublicKey())
	behavior.AddLink(inChD, outChD, disPeerIdentity.GetPublicKey())
	for i := 0; i < 100; i++ {
		sendMessage(&connectedNode, inCh)    // Will be received
		sendMessage(&connectedNode, inChD)   // Won't be received - destination is disconnected
		sendMessage(&disconnectedNode, inCh) // Won't be received - source is disconnected
	}
	require.Eventually(t, func() bool {
		return recvLoop.ReceivedCount() == 100 && recvLoop.AverageDuration() < int64(20) && recvLoopD.ReceivedCount() == 0
	}, 5*time.Second, 10*time.Millisecond, "expected 100 messages on connected node and 0 on disconnected node with avg duration < 20ms")

	// Stop the test.
	recvLoop.Stop()
	recvLoopD.Stop()
	behavior.Close()
}

type testLoopStats struct {
	receivedCount   atomic.Int32
	averageDuration atomic.Int64
	resetCh         chan struct{}
	stopCh          chan struct{}
}

func (s *testLoopStats) ReceivedCount() int {
	return int(s.receivedCount.Load())
}

func (s *testLoopStats) AverageDuration() int64 {
	return s.averageDuration.Load()
}

func (s *testLoopStats) Reset() {
	s.resetCh <- struct{}{}
}

func (s *testLoopStats) Stop() {
	s.stopCh <- struct{}{}
}

func runTestRecvLoop(outCh chan *peeringMsg) *testLoopStats {
	stats := testLoopStats{
		resetCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
	}

	var durations []time.Duration

	go func() {
		for {
			select {
			case <-stats.stopCh:
				return
			case <-stats.resetCh:
				durations = durations[:0]
				stats.averageDuration.Store(0)
				stats.receivedCount.Store(0)
			case msg := <-outCh:
				durations = append(durations, time.Since(time.Unix(0, msg.timestamp)))
				stats.receivedCount.Add(1)
				stats.averageDuration.Store(averageDuration(durations))
			}
		}
	}()

	return &stats
}

func averageDuration(durations []time.Duration) int64 {
	result := int64(0)
	for _, d := range durations {
		result += d.Milliseconds()
	}
	return result / int64(len(durations))
}

func sendMessage(from *peeringNode, inCh chan *peeringMsg) {
	inCh <- &peeringMsg{
		from:      from.identity.GetPublicKey(),
		timestamp: time.Now().UnixNano(),
	}
}
