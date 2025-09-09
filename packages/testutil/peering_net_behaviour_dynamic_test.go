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
	time.Sleep(500 * time.Millisecond)
	//
	// Validate the results (with some tolerance for randomness).
	{ // 50% of messages dropped + 50% duplicated -> delivered ~75%
		require.Greater(t, recvLoop.ReceivedCount(), 500)
		require.Less(t, recvLoop.ReceivedCount(), 900)
	}
	{ // Average should be between the specified boundaries.
		avgDuration := recvLoop.AverageDuration()
		require.Greater(t, avgDuration, int64(50))
		require.Less(t, avgDuration, int64(100))
	}
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
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 100, recvLoop.ReceivedCount())
	require.Less(t, recvLoop.AverageDuration(), int64(20))
	recvLoop.Reset()

	deliver40Name := "Deliver40"
	deliver70Name := "Deliver70"
	behavior.WithLosingChannel(&deliver70Name, 70).WithLosingChannel(&deliver40Name, 40) // 70% * 40% = 28% delivery probability
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 280, recvLoop.ReceivedCount(), 90)
	require.Less(t, recvLoop.AverageDuration(), int64(20))
	recvLoop.Reset()

	delayName := "Delay"
	behavior.WithDelayingChannel(&delayName, 20*time.Millisecond, 70*time.Millisecond) // 28% delivery probability and 20-70 ms delay
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(150 * time.Millisecond)
	require.InDelta(t, 280, recvLoop.ReceivedCount(), 90)
	require.InDelta(t, 45, recvLoop.AverageDuration(), 20)
	recvLoop.Reset()

	behavior.RemoveHandler(deliver40Name) // 70% delivery probability and 20-70 ms delay
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(150 * time.Millisecond)
	require.InDelta(t, 700, recvLoop.ReceivedCount(), 90)
	require.InDelta(t, 45, recvLoop.AverageDuration(), 20)
	recvLoop.Reset()

	behavior.RemoveHandler(delayName) // 70% delivery probability without a delay
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 700, recvLoop.ReceivedCount(), 90)
	require.Less(t, recvLoop.AverageDuration(), int64(20))
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
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 500, recvLoop.ReceivedCount(), 90)
	require.Less(t, recvLoop.AverageDuration(), int64(20))

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
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 2500, recvLoop.ReceivedCount(), 90)
	require.Less(t, recvLoop.AverageDuration(), int64(20))

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
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 100, recvLoop.ReceivedCount())
	require.InDelta(t, 50, recvLoop.AverageDuration(), 20)

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
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 100, recvLoop.ReceivedCount())
	require.Less(t, recvLoop.AverageDuration(), int64(20))
	require.Equal(t, 0, recvLoopD.ReceivedCount())

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
