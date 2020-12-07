package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/dks"
	"github.com/iotaledger/wasp/packages/peering"
	"go.dedis.ch/kyber/v3"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
)

// Node represents a node, that can participate in a DKG procedure.
// It receives commands from the initiator as a dkg.NodeProvider,
// and communicates with other DKG nodes via the peering network.
type Node struct {
	secKey      kyber.Scalar
	pubKey      kyber.Point
	suite       rabin_dkg.Suite         // Cryptography to use.
	netProvider peering.NetworkProvider // Network to communicate through.
	registry    dks.RegistryProvider    // Where to store the generated keys.
	processes   map[string]*proc        // Only for introspection.
	attachID    interface{}             // Peering attach ID
	log         *logger.Logger
}

// Init creates new node, that can participate in the DKG procedure.
// The node then can run several DKG procedures.
func Init(
	secKey kyber.Scalar,
	pubKey kyber.Point,
	suite rabin_dkg.Suite,
	netProvider peering.NetworkProvider,
	registry dks.RegistryProvider,
	log *logger.Logger,
) *Node {
	n := Node{
		secKey:      secKey,
		pubKey:      pubKey,
		suite:       suite,
		netProvider: netProvider,
		registry:    registry,
		processes:   make(map[string]*proc),
		log:         log,
	}
	n.attachID = netProvider.Attach(nil, n.onInitMsg)
	return &n
}

// onInitMsg is a callback to handle the DKG initialization messages.
func (n *Node) onInitMsg(recv *peering.RecvEvent) {
	if recv.Msg.MsgType != initiatorInitMsgType {
		return
	}
	var err error
	var p *proc
	req := initiatorInitMsg{}
	if err = req.fromBytes(recv.Msg.MsgData, n.suite); err != nil {
		n.log.Warnf("Dropping unknown message: %v", recv)
	}
	if p, err = onInitiatorInit(&recv.Msg.ChainID, &req, n); err == nil {
		n.processes[p.stringID()] = p
	}
	recv.From.SendMsg(makePeerMessage(&recv.Msg.ChainID, &initiatorStatusMsg{
		step:  req.step,
		error: err,
	}))
}

// Called by the DKG process on termination.
func (n *Node) dropProcess(p *proc) bool {
	procID := p.stringID()
	if found := n.processes[procID]; found != nil {
		delete(n.processes, procID)
		return true
	}
	return false
}
