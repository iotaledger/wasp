// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package node is a front-end to the Distributed Schnorr Signature subsystem.
// Each node runs an instance of the DSSNode. It is responsible for exchanging
// the messages over the network, and to maintain DSS protocol instances.
//
// DSS Instances are organized in series. Each series is represented by a base
// hash. DSS instances in a single series is indexed by state index. This is
// done to have a way to pre-generate the nonces for future signatures.
//
// The main interaction with this subsystem is the following:
//
//  1. Start the DSS instance (the most expensive part).
//  2. Wait for DSS input to the ACS, provide the ACS output to the DSS.
//  3. Provide the data to sign.
//  4. Wait for a signature.
//
// TODO: Instance cleanup.
package node

import (
	"context"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

type DSSNode interface {
	Start(key string, index int, dkShare tcrypto.DKShare, partCB func([]int), sigCB func([]byte)) error
	DecidedIndexProposals(key string, index int, decidedIndexProposals [][]int, messageToSign []byte) error
	StatusString(key string, index int) string
	Close()
}

type recvMsg struct {
	sender  gpa.NodeID
	msgData []byte
}

type dssNodeImpl struct {
	lock      *sync.RWMutex
	suite     suites.Suite
	net       peering.NetworkProvider
	netAttach interface{}
	peeringID *peering.PeeringID
	nid       *cryptolib.KeyPair
	series    map[string]*dssSeriesImpl
	seriesBuf map[string]map[int][]*recvMsg
	ctx       context.Context
	ctxCancel context.CancelFunc
	log       *logger.Logger
}

var _ DSSNode = &dssNodeImpl{}

func New(peeringID *peering.PeeringID, net peering.NetworkProvider, nid *cryptolib.KeyPair, log *logger.Logger) DSSNode {
	ctx, ctcCancel := context.WithCancel(context.Background())
	n := &dssNodeImpl{
		lock:      &sync.RWMutex{},
		suite:     tcrypto.DefaultEd25519Suite(),
		net:       net,
		netAttach: nil, // Set bellow.
		peeringID: peeringID,
		nid:       nid,
		series:    map[string]*dssSeriesImpl{},
		seriesBuf: map[string]map[int][]*recvMsg{},
		ctx:       ctx,
		ctxCancel: ctcCancel,
		log:       log,
	}
	n.netAttach = net.Attach(peeringID, peering.PeerMessageReceiverChainDSS, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeDSS {
			n.log.Warnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		key, index, msgData, err := fromMsgData(recv.MsgData)
		if err != nil {
			n.log.Warnf("Cannot parse a message: %v", err)
			return
		}
		n.recv(key, index, msgData, pubKeyAsNodeID(recv.SenderPubKey))
	})
	go func() { // Send ticks periodically for message redelivery, etc.
		for {
			select {
			case <-time.After(1 * time.Second):
				n.lock.Lock()
				series := n.series
				now := time.Now()
				for i := range series {
					series[i].tick(now)
				}
				n.lock.Unlock()
			case <-n.ctx.Done():
				n.net.Detach(n.netAttach)
				return
			}
		}
	}()
	return n
}

func (n *dssNodeImpl) Start(key string, index int, dkShare tcrypto.DKShare, partCB func([]int), sigCB func([]byte)) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	if _, ok := n.series[key]; !ok {
		n.series[key] = newSeries(n, key, dkShare)
		n.recvFromBuf(key)
	}
	return n.series[key].start(index, partCB, sigCB)
}

func (n *dssNodeImpl) DecidedIndexProposals(key string, index int, decidedIndexProposals [][]int, messageToSign []byte) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	if series, ok := n.series[key]; ok {
		return series.decidedIndexProposals(index, decidedIndexProposals, messageToSign)
	}
	return xerrors.Errorf("DSS series for key=%v not found", key)
}

func (n *dssNodeImpl) StatusString(key string, index int) string {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.series[key].statusString(index)
}

func (n *dssNodeImpl) Close() {
	n.ctxCancel()
}

func (n *dssNodeImpl) recv(key string, index int, msgData []byte, sender gpa.NodeID) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if _, ok := n.series[key]; !ok {
		n.saveToBuf(key, index, msgData, sender)
		return
	}
	n.series[key].recvMessage(index, msgData, sender)
}

func (n *dssNodeImpl) saveToBuf(key string, index int, msgData []byte, sender gpa.NodeID) {
	if _, ok := n.seriesBuf[key]; !ok {
		n.seriesBuf[key] = map[int][]*recvMsg{}
	}
	if _, ok := n.seriesBuf[key][index]; !ok {
		n.seriesBuf[key][index] = []*recvMsg{}
	}
	n.seriesBuf[key][index] = append(n.seriesBuf[key][index], &recvMsg{sender, msgData})
}

func (n *dssNodeImpl) recvFromBuf(key string) {
	indexes, ok := n.seriesBuf[key]
	if !ok {
		return
	}
	for index, msgs := range indexes {
		for i := range msgs {
			n.series[key].recvMessage(index, msgs[i].msgData, msgs[i].sender)
		}
	}
	delete(n.seriesBuf, key)
}
