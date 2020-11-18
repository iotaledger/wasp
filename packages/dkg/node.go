package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"errors"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/plugins/peering"
	"go.dedis.ch/kyber/v3"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
)

// In a normal execution mode, there will be exactly one node registered
// in this registry. Multiple nodes are used in the unit tests.
var nodes []*node = []*node{}

// Node represents a node, that can participate in a DKG procedure.
// It receives commands from the coordinator as a CoordNodeProvider,
// and communicates with other DKG nodes via the peering network.
type node struct {
	secKey      kyber.Scalar
	pubKey      kyber.Point
	suite       rabin_dkg.Suite
	netProvider peering.NetworkProvider
	registry    RegistryProvider
	processes   map[string]*proc
	log         *logger.Logger
}

// InitNode creates new node, that can participate in the DKG procedure.
// The node then can run several DKG procedures.
func InitNode(
	secKey kyber.Scalar,
	pubKey kyber.Point,
	suite rabin_dkg.Suite,
	netProvider peering.NetworkProvider,
	registry RegistryProvider,
	log *logger.Logger,
) CoordNodeProvider {
	n := node{
		secKey:      secKey,
		pubKey:      pubKey,
		suite:       suite,
		netProvider: netProvider,
		registry:    registry,
		processes:   make(map[string]*proc),
		log:         log,
	}
	nodes = append(nodes, &n)
	return &n
}

func (n *node) dropProcess(p *proc) bool {
	if found := n.processes[p.dkgID]; found != nil {
		delete(n.processes, p.dkgID)
		return true
	}
	return false
}

// DkgInit implements CoordNodeProvider.
// peerAddrs here is always a slice of a single element equal to our node.
func (n *node) DkgInit(peerAddrs []string, dkgID string, msg *InitReq) error {
	n.log.Debugf("DkgInit, dkgID=%v, msg.PeerLocs=%v", dkgID, msg.PeerLocs)
	var err error
	var p *proc
	if p, err = onCoordInit(dkgID, msg, n); err != nil {
		return err
	}
	n.processes[dkgID] = p
	return nil

}

// DkgStep implements CoordNodeProvider.
func (n *node) DkgStep(peerAddrs []string, dkgID string, msg *StepReq) error {
	if p := n.processes[dkgID]; p != nil {
		p.log.Debugf("DkgStep, msg.Step=%v", msg.Step)
		return p.onCoordStep(msg)
	}
	return errors.New("dkgID_not_found")
}

// DkgPubKey implements CoordNodeProvider.
func (n *node) DkgPubKey(peerAddrs []string, dkgID string) ([]*PubKeyResp, error) {
	var err error
	if p := n.processes[dkgID]; p != nil {
		var resp *PubKeyResp
		if resp, err = p.onCoordPubKey(); err != nil {
			return nil, err
		}
		return []*PubKeyResp{resp}, nil
	}
	// TODO: Handle the case, when the key is taken from the registry.
	return nil, errors.New("dkgID_not_found")
}
