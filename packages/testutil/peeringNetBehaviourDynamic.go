// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

// An interface providing some network behavior.
// It is used for testing network protocols in more realistic settings.

import (
	"math/rand"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// peeringNetDynamic provides a behavior of a network with dynamically
// changeable working conditions. It works as peeringNetReliable without any
// handlers set.
type PeeringNetDynamic struct {
	closeChs []chan bool
	handlers []peeringNetDynamicHandlerEntry
	mutex    sync.RWMutex
	log      *logger.Logger
}

var _ PeeringNetBehavior = &PeeringNetDynamic{}

type peeringNetDynamicHandlerEntry struct {
	id *string
	peeringNetDynamicHandler
}

type peeringNetDynamicHandler interface {
	handleSendMessage(
		msg *peeringMsg,
		dstPubKey *cryptolib.PublicKey,
		nextHandlers []peeringNetDynamicHandlerEntry,
		callHandlersAndSendFun func(nextHandlers []peeringNetDynamicHandlerEntry),
		log *logger.Logger,
	)
}

// NewPeeringNetDynamic constructs the PeeringNetBehavior.
func NewPeeringNetDynamic(log *logger.Logger) *PeeringNetDynamic {
	return &PeeringNetDynamic{
		closeChs: make([]chan bool, 0),
		handlers: make([]peeringNetDynamicHandlerEntry, 0),
		log:      log,
	}
}

func (pndT *PeeringNetDynamic) WithLosingChannel(id *string, deliveryProbability int) *PeeringNetDynamic {
	pndT.addHandlerEntry(peeringNetDynamicHandlerEntry{
		id,
		&peeringNetDynamicHandlerLosingChannel{
			probability: deliveryProbability,
		},
	})
	return pndT
}

func (pndT *PeeringNetDynamic) WithRepeatingChannel(id *string, repeatProbability int) *PeeringNetDynamic {
	pndT.addHandlerEntry(peeringNetDynamicHandlerEntry{
		id,
		&peeringNetDynamicHandlerRepeatingChannel{
			probability: repeatProbability,
		},
	})
	return pndT
}

func (pndT *PeeringNetDynamic) WithDelayingChannel(id *string, delayFrom, delayTill time.Duration) *PeeringNetDynamic {
	pndT.addHandlerEntry(peeringNetDynamicHandlerEntry{
		id,
		&peeringNetDynamicHandlerDelayingChannel{
			from: delayFrom,
			till: delayTill,
		},
	})
	return pndT
}

func (pndT *PeeringNetDynamic) WithPeerDisconnected(id *string, peerPubKey *cryptolib.PublicKey) *PeeringNetDynamic {
	pndT.addHandlerEntry(peeringNetDynamicHandlerEntry{
		id,
		&peeringNetDynamicHandlerPeerDisconnected{
			peerPubKey: peerPubKey,
		},
	})
	return pndT
}

func (pndT *PeeringNetDynamic) addHandlerEntry(handler peeringNetDynamicHandlerEntry) {
	pndT.mutex.Lock()
	defer pndT.mutex.Unlock()

	pndT.handlers = append(pndT.handlers, handler)
}

func (pndT *PeeringNetDynamic) RemoveHandler(id string) bool {
	pndT.mutex.Lock()
	defer pndT.mutex.Unlock()

	var i int
	for i = 0; i < len(pndT.handlers); i++ {
		currentHandlerID := pndT.handlers[i].getID()
		if (currentHandlerID != nil) && (*currentHandlerID == id) {
			pndT.handlers = append(pndT.handlers[:i], pndT.handlers[i+1:]...)
			return true
		}
	}
	return false
}

// AddLink implements PeeringNetBehavior.
func (pndT *PeeringNetDynamic) AddLink(inCh, outCh chan *peeringMsg, dstPubKey *cryptolib.PublicKey) {
	closeCh := make(chan bool)
	pndT.closeChs = append(pndT.closeChs, closeCh)
	go pndT.recvLoop(inCh, outCh, closeCh, dstPubKey)
}

// Close implements PeeringNetBehavior.
func (pndT *PeeringNetDynamic) Close() {
	for i := range pndT.closeChs {
		close(pndT.closeChs[i])
	}
}

func (pndT *PeeringNetDynamic) recvLoop(inCh, outCh chan *peeringMsg, closeCh chan bool, dstPubKey *cryptolib.PublicKey) {
	for {
		select {
		case <-closeCh:
			return
		case recv, ok := <-inCh:
			if !ok {
				return
			}
			var callHandlersAndSendFun func(nextHandlers []peeringNetDynamicHandlerEntry)
			callHandlersAndSendFun = func(nextHandlers []peeringNetDynamicHandlerEntry) {
				if len(nextHandlers) > 0 {
					nextHandlers[0].handleSendMessage(recv, dstPubKey, nextHandlers[1:], callHandlersAndSendFun, pndT.log)
				} else {
					pndT.log.Debugf("Network delivers message %v -%v-> %v", recv.from.String(), recv.msg.MsgType, dstPubKey.String())
					safeSendPeeringMsg(outCh, recv, pndT.log)
				}
			}
			pndT.mutex.RLock()
			handlers := make([]peeringNetDynamicHandlerEntry, len(pndT.handlers))
			copy(handlers, pndT.handlers)
			pndT.mutex.RUnlock()
			callHandlersAndSendFun(handlers)
		}
	}
}

func (pndheT *peeringNetDynamicHandlerEntry) getID() *string {
	return pndheT.id
}

type peeringNetDynamicHandlerLosingChannel struct {
	probability int // probability to deliver a message (in percents)
}

func (lcT *peeringNetDynamicHandlerLosingChannel) handleSendMessage(
	msg *peeringMsg,
	dstPubKey *cryptolib.PublicKey,
	nextHandlers []peeringNetDynamicHandlerEntry,
	callHandlersAndSendFun func(nextHandlers []peeringNetDynamicHandlerEntry),
	log *logger.Logger,
) {
	if rand.Intn(100) > lcT.probability {
		log.Debugf("Network dropped message %v -%v-> %v", msg.from.String(), msg.msg.MsgType, dstPubKey.String())
		return
	}
	callHandlersAndSendFun(nextHandlers)
}

type peeringNetDynamicHandlerRepeatingChannel struct {
	probability int // Probability to repeat a message (in percents), 0 meaning no repeat, 100 a certain repeat, 250 - 50% that message will be sent out three times and 50% - that four times
}

func (rcT *peeringNetDynamicHandlerRepeatingChannel) handleSendMessage(
	msg *peeringMsg,
	dstPubKey *cryptolib.PublicKey,
	nextHandlers []peeringNetDynamicHandlerEntry,
	callHandlersAndSendFun func(nextHandlers []peeringNetDynamicHandlerEntry),
	log *logger.Logger,
) {
	numRepeat := 1 + rcT.probability/100
	if rand.Intn(100) < rcT.probability%100 {
		numRepeat++
	}
	log.Debugf("Network repeated message %v -%v-> %v %v times", msg.from.String(), msg.msg.MsgType, dstPubKey.String(), numRepeat)
	for i := 0; i < numRepeat; i++ {
		callHandlersAndSendFun(nextHandlers)
	}
}

type peeringNetDynamicHandlerDelayingChannel struct {
	from time.Duration
	till time.Duration
}

func (dcT *peeringNetDynamicHandlerDelayingChannel) handleSendMessage(
	msg *peeringMsg,
	dstPubKey *cryptolib.PublicKey,
	nextHandlers []peeringNetDynamicHandlerEntry,
	callHandlersAndSendFun func(nextHandlers []peeringNetDynamicHandlerEntry),
	log *logger.Logger,
) {
	go func() {
		fromMS := int(dcT.from.Milliseconds())
		tillMS := int(dcT.till.Milliseconds())
		var delay time.Duration
		if tillMS > 0 {
			if fromMS < tillMS {
				delay = time.Duration(rand.Intn(tillMS-fromMS)+fromMS) * time.Millisecond
			} else {
				delay = time.Duration(fromMS) * time.Millisecond
			}
			log.Debugf("Network delayed message %v -%v-> %v for %v", msg.from.String(), msg.msg.MsgType, dstPubKey.String(), delay)
			<-time.After(delay)
		}
		callHandlersAndSendFun(nextHandlers)
	}()
}

type peeringNetDynamicHandlerPeerDisconnected struct {
	peerPubKey *cryptolib.PublicKey
}

func (pdT *peeringNetDynamicHandlerPeerDisconnected) handleSendMessage(
	msg *peeringMsg,
	dstPubKey *cryptolib.PublicKey,
	nextHandlers []peeringNetDynamicHandlerEntry,
	callHandlersAndSendFun func(nextHandlers []peeringNetDynamicHandlerEntry),
	log *logger.Logger,
) {
	if dstPubKey.Equals(pdT.peerPubKey) {
		log.Debugf("Network dropped message %v -%v-> %v, because destination is disconnected", msg.from.String(), msg.msg.MsgType, dstPubKey.String())
		return
	}
	if msg.from.Equals(pdT.peerPubKey) {
		log.Debugf("Network dropped message %v -%v-> %v, because source is disconnected", msg.from.String(), msg.msg.MsgType, dstPubKey.String())
		return
	}
	callHandlersAndSendFun(nextHandlers)
}
