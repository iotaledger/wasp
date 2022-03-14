// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil // not `..._test` because it uses peeringMsg.

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
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
	srcPeerIdentity := ed25519.GenerateKeyPair()
	dstPeerIdentity := ed25519.GenerateKeyPair()
	someNode := peeringNode{netID: "src", identity: &srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false))
	behavior.AddLink(inCh, outCh, &dstPeerIdentity.PublicKey)
	for i := 0; i < 10; i++ {
		sendMessage(&someNode, inCh)
	}
	//
	// Stop the test.
	<-doneCh
	behavior.Close()
}

func TestPeeringNetDynamicUnreliable(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	stopCh := make(chan bool)
	durations := make([]time.Duration, 0)
	go testRecvLoop(outCh, &durations, stopCh)
	srcPeerIdentity := ed25519.GenerateKeyPair()
	dstPeerIdentity := ed25519.GenerateKeyPair()
	someNode := peeringNode{netID: "src", identity: &srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false)).
		WithLosingChannel(nil, 50).
		WithRepeatingChannel(nil, 50).
		WithDelayingChannel(nil, 50*time.Millisecond, 100*time.Millisecond)
	behavior.AddLink(inCh, outCh, &dstPeerIdentity.PublicKey)
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(500 * time.Millisecond)
	//
	// Verify the results (with some tolerance for randomness).
	{ // 50% of messages dropped + 50% duplicated -> delivered ~75%
		require.Greater(t, len(durations), 500)
		require.Less(t, len(durations), 900)
	}
	{ // Average should be between the specified boundaries.
		avgDuration := averageDuration(durations)
		require.Greater(t, avgDuration, int64(50))
		require.Less(t, avgDuration, int64(100))
	}
	//
	// Stop the test.
	stopCh <- true
	behavior.Close()
}

func TestPeeringNetDynamicChanging(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	stopCh := make(chan bool)
	durations := make([]time.Duration, 0)
	go testRecvLoop(outCh, &durations, stopCh)
	srcPeerIdentity := ed25519.GenerateKeyPair()
	dstPeerIdentity := ed25519.GenerateKeyPair()
	someNode := peeringNode{netID: "src", identity: &srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false))
	behavior.AddLink(inCh, outCh, &dstPeerIdentity.PublicKey)
	for i := 0; i < 100; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 100, len(durations))
	require.Less(t, averageDuration(durations), int64(20))
	durations = durations[:0]

	deliver40Name := "Deliver40"
	deliver70Name := "Deliver70"
	behavior.WithLosingChannel(&deliver70Name, 70).WithLosingChannel(&deliver40Name, 40) // 70% * 40% = 28% delivery probability
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 280, len(durations), 90)
	require.Less(t, averageDuration(durations), int64(20))
	durations = durations[:0]

	delayName := "Delay"
	behavior.WithDelayingChannel(&delayName, 20*time.Millisecond, 70*time.Millisecond) // 28% delivery probability and 20-70 ms delay
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(150 * time.Millisecond)
	require.InDelta(t, 280, len(durations), 90)
	require.InDelta(t, 45, averageDuration(durations), 20)
	durations = durations[:0]

	behavior.RemoveHandler(deliver40Name) // 70% delivery probability and 20-70 ms delay
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(150 * time.Millisecond)
	require.InDelta(t, 700, len(durations), 90)
	require.InDelta(t, 45, averageDuration(durations), 20)
	durations = durations[:0]

	behavior.RemoveHandler(delayName) // 70% delivery probability without a delay
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 700, len(durations), 90)
	require.Less(t, averageDuration(durations), int64(20))
	durations = durations[:0]

	// Stop the test.
	stopCh <- true
	behavior.Close()
}

func TestPeeringNetDynamicLosingChannel(t *testing.T) { //nolint:dupl
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	stopCh := make(chan bool)
	durations := make([]time.Duration, 0)
	go testRecvLoop(outCh, &durations, stopCh)
	srcPeerIdentity := ed25519.GenerateKeyPair()
	dstPeerIdentity := ed25519.GenerateKeyPair()
	someNode := peeringNode{netID: "src", identity: &srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false)).WithLosingChannel(nil, 50)
	behavior.AddLink(inCh, outCh, &dstPeerIdentity.PublicKey)
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 500, len(durations), 90)
	require.Less(t, averageDuration(durations), int64(20))

	// Stop the test.
	stopCh <- true
	behavior.Close()
}

func TestPeeringNetDynamicRepeatingChannel(t *testing.T) { //nolint:dupl
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 10000)
	stopCh := make(chan bool)
	durations := make([]time.Duration, 0)
	go testRecvLoop(outCh, &durations, stopCh)
	srcPeerIdentity := ed25519.GenerateKeyPair()
	dstPeerIdentity := ed25519.GenerateKeyPair()
	someNode := peeringNode{netID: "src", identity: &srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false)).WithRepeatingChannel(nil, 150)
	behavior.AddLink(inCh, outCh, &dstPeerIdentity.PublicKey)
	for i := 0; i < 1000; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(100 * time.Millisecond)
	require.InDelta(t, 2500, len(durations), 90)
	require.Less(t, averageDuration(durations), int64(20))

	// Stop the test.
	stopCh <- true
	behavior.Close()
}

func TestPeeringNetDynamicDelayingChannel(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	stopCh := make(chan bool)
	durations := make([]time.Duration, 0)
	go testRecvLoop(outCh, &durations, stopCh)
	srcPeerIdentity := ed25519.GenerateKeyPair()
	dstPeerIdentity := ed25519.GenerateKeyPair()
	someNode := peeringNode{netID: "src", identity: &srcPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false)).WithDelayingChannel(nil, 25*time.Millisecond, 75*time.Millisecond)
	behavior.AddLink(inCh, outCh, &dstPeerIdentity.PublicKey)
	for i := 0; i < 100; i++ {
		sendMessage(&someNode, inCh)
	}
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 100, len(durations))
	require.InDelta(t, 50, averageDuration(durations), 20)

	// Stop the test.
	stopCh <- true
	behavior.Close()
}

func TestPeeringNetDynamicPeerDisconnected(t *testing.T) {
	inCh := make(chan *peeringMsg)
	outCh := make(chan *peeringMsg, 1000)
	inChD := make(chan *peeringMsg)
	outChD := make(chan *peeringMsg)
	stopCh := make(chan bool)
	durations := make([]time.Duration, 0)
	durationsD := make([]time.Duration, 0)
	go testRecvLoop(outCh, &durations, stopCh)
	go testRecvLoop(outChD, &durationsD, stopCh)
	srcPeerIdentity := ed25519.GenerateKeyPair()
	dstPeerIdentity := ed25519.GenerateKeyPair()
	disPeerIdentity := ed25519.GenerateKeyPair()
	connectedNode := peeringNode{netID: "src", identity: &srcPeerIdentity}
	disconnectedNode := peeringNode{netID: "disconnected", identity: &disPeerIdentity}
	//
	// Run the test.
	behavior := NewPeeringNetDynamic(testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelError, false)).WithPeerDisconnected(nil, &disPeerIdentity.PublicKey)
	behavior.AddLink(inCh, outCh, &dstPeerIdentity.PublicKey)
	behavior.AddLink(inChD, outChD, &disPeerIdentity.PublicKey)
	for i := 0; i < 100; i++ {
		sendMessage(&connectedNode, inCh)    // Will be received
		sendMessage(&connectedNode, inChD)   // Won't be received - destination is disconnected
		sendMessage(&disconnectedNode, inCh) // Won't be received - source is disconnected
	}
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 100, len(durations))
	require.Less(t, averageDuration(durations), int64(20))
	require.Equal(t, 0, len(durationsD))

	// Stop the test.
	stopCh <- true
	behavior.Close()
}

func testRecvLoop(outCh chan *peeringMsg, durations *[]time.Duration, stopCh chan bool) {
	for {
		select {
		case <-stopCh:
			return
		case msg := <-outCh:
			*durations = append(*durations, time.Since(time.Unix(0, msg.timestamp)))
		}
	}
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
		from:      &from.identity.PublicKey,
		timestamp: time.Now().UnixNano(),
	}
}
