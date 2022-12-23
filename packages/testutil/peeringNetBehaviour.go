// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

// An interface providing some network behavior.
// It is used for testing network protocols in more realistic settings.

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// An interface for all the network behaviors.
type PeeringNetBehavior interface {
	AddLink(inCh, outCh chan *peeringMsg, dstPubKey *cryptolib.PublicKey)
	Close()
}

// peeringNetReliable provides a behavior of a reliable network.
// That's for basic tests.
type peeringNetReliable struct {
	closeChs []chan bool
	log      *logger.Logger
}

// NewPeeringNetReliable constructs the PeeringNetBehavior.
func NewPeeringNetReliable(log *logger.Logger) PeeringNetBehavior {
	return &peeringNetReliable{
		closeChs: make([]chan bool, 0),
		log:      log,
	}
}

// Run implements PeeringNetBehavior.
func (n *peeringNetReliable) AddLink(inCh, outCh chan *peeringMsg, dstPubKey *cryptolib.PublicKey) {
	closeCh := make(chan bool)
	n.closeChs = append(n.closeChs, closeCh)
	go n.recvLoop(inCh, outCh, closeCh)
}

// Close implements PeeringNetBehavior.
func (n *peeringNetReliable) Close() {
	for i := range n.closeChs {
		close(n.closeChs[i])
	}
}

func (n *peeringNetReliable) recvLoop(inCh, outCh chan *peeringMsg, closeCh chan bool) {
	for {
		select {
		case <-closeCh:
			return
		case recv := <-inCh:
			safeSendPeeringMsg(outCh, recv, n.log)
		}
	}
}

// peeringNetUnreliable simulates unreliable network by droppin, repeating, delaying and reordering messages.
type peeringNetUnreliable struct {
	deliverPct int // probability to deliver a message (in percents)
	repeatPct  int // Probability to repeat a message (in percents, if delivered)
	delayFrom  time.Duration
	delayTill  time.Duration
	closeChs   []chan bool
	log        *logger.Logger
}

// NewPeeringNetReliable constructs the PeeringNetBehavior.
func NewPeeringNetUnreliable(deliverPct, repeatPct int, delayFrom, delayTill time.Duration, log *logger.Logger) PeeringNetBehavior {
	return &peeringNetUnreliable{
		deliverPct: deliverPct,
		repeatPct:  repeatPct,
		delayFrom:  delayFrom,
		delayTill:  delayTill,
		closeChs:   make([]chan bool, 0),
		log:        log,
	}
}

// Run implements PeeringNetBehavior.
func (n *peeringNetUnreliable) AddLink(inCh, outCh chan *peeringMsg, dstPubKey *cryptolib.PublicKey) {
	closeCh := make(chan bool)
	n.closeChs = append(n.closeChs, closeCh)
	go n.recvLoop(inCh, outCh, closeCh, dstPubKey)
}

// Close implements PeeringNetBehavior.
func (n *peeringNetUnreliable) Close() {
	for i := range n.closeChs {
		close(n.closeChs[i])
	}
}

func (n *peeringNetUnreliable) recvLoop(inCh, outCh chan *peeringMsg, closeCh chan bool, dstPubKey *cryptolib.PublicKey) {
	for {
		select {
		case <-closeCh:
			return
		case recv, ok := <-inCh:
			if !ok {
				return
			}
			if rand.Intn(100) > n.deliverPct {
				n.log.Debugf("Network dropped message %v -%v-> %v", recv.from.String(), recv.msg.MsgType, dstPubKey.String())
				continue // Drop the message.
			}
			//
			// Let's assume repeatPct can be > 100 meaning
			// the messages will be repeated more than twice.
			numRepeat := 1 + n.repeatPct/100
			if rand.Intn(100) < n.repeatPct%100 {
				numRepeat++
			}
			for i := 0; i < numRepeat; i++ {
				go n.sendDelayed(recv, outCh, dstPubKey, i+1, numRepeat)
			}
		}
	}
}

func (n *peeringNetUnreliable) sendDelayed(recv *peeringMsg, outCh chan *peeringMsg, dstPubKey *cryptolib.PublicKey, dupNum, dupCount int) {
	fromMS := int(n.delayFrom.Milliseconds())
	tillMS := int(n.delayTill.Milliseconds())
	var delay time.Duration
	if tillMS > 0 {
		if fromMS < tillMS {
			delay = time.Duration(rand.Intn(tillMS-fromMS)+fromMS) * time.Millisecond
		} else {
			delay = time.Duration(fromMS) * time.Millisecond
		}
		<-time.After(delay)
	}
	n.log.Debugf(
		"Network delivers message %v -%v-> %v (duplicate %v/%v, delay=%vms)",
		recv.from.String(), recv.msg.MsgType, dstPubKey.String(), dupNum, dupCount, delay.Milliseconds(),
	)
	safeSendPeeringMsg(outCh, recv, n.log)
}

// To avoid panics when tests are being stopped.
func safeSendPeeringMsg(outCh chan *peeringMsg, recv *peeringMsg, log *logger.Logger) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("NOTE: peeringNetReliable dropping message: %v\n", err)
		}
	}()
	select {
	case outCh <- recv:
		return
	default:
		log.Warnf("Dropping message, because outCh is overflown.")
	}
}
