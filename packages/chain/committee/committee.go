// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committee

import (
	"crypto/rand"
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/consensus/commonsubset"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"
)

type committee struct {
	isReady        *atomic.Bool
	address        iotago.Address
	validatorNodes peering.GroupProvider
	acsRunner      chain.AsynchronousCommonSubsetRunner
	size           uint16
	quorum         uint16
	ownIndex       uint16
	dkshare        *tcrypto.DKShare
	log            *logger.Logger
}

var _ chain.Committee = &committee{}

const waitReady = false

func New(
	dkShare *tcrypto.DKShare,
	chainID *iscp.ChainID,
	netProvider peering.NetworkProvider,
	log *logger.Logger,
	acsRunner ...chain.AsynchronousCommonSubsetRunner, // Only for mocking.
) (chain.Committee, peering.GroupProvider, error) {
	var err error
	if dkShare.Index == nil {
		return nil, nil, xerrors.Errorf("NewCommittee: wrong DKShare record for address %s: nil index", dkShare.Address.Base58())
	}
	// peerGroupID is calculated by XORing chainID and stateAddr.
	// It allows to use same statAddr for different chains
	peerGroupID := dkShare.Address.Array()
	var chainArr [33]byte
	if chainID != nil {
		chainArr = chainID.Array()
	}
	for i := range peerGroupID {
		peerGroupID[i] ^= chainArr[i]
	}
	var peers peering.GroupProvider
	if peers, err = netProvider.PeerGroup(peerGroupID, dkShare.NodePubKeys); err != nil {
		return nil, nil, xerrors.Errorf("NewCommittee: failed to create peer group for committee: %+v: %w", dkShare.NodePubKeys, err)
	}
	log.Debugf("NewCommittee: peer group: %+v", dkShare.NodePubKeys)
	ret := &committee{
		isReady:        atomic.NewBool(false),
		address:        dkShare.Address,
		validatorNodes: peers,
		size:           dkShare.N,
		quorum:         dkShare.T,
		ownIndex:       *dkShare.Index,
		dkshare:        dkShare,
		log:            log,
	}
	if len(acsRunner) > 0 {
		ret.acsRunner = acsRunner[0]
	} else {
		// That's the default implementation of the ACS.
		// We use it, if the mocked variant was not passed.
		ret.acsRunner = commonsubset.NewCommonSubsetCoordinator(
			netProvider,
			ret.validatorNodes,
			dkShare,
			log,
		)
	}
	go ret.waitReady(waitReady)

	return ret, ret.validatorNodes, nil
}

func (c *committee) Address() iotago.Address {
	return c.address
}

func (c *committee) Size() uint16 {
	return c.size
}

func (c *committee) Quorum() uint16 {
	return c.quorum
}

func (c *committee) IsReady() bool {
	return c.isReady.Load()
}

func (c *committee) OwnPeerIndex() uint16 {
	return c.ownIndex
}

func (c *committee) DKShare() *tcrypto.DKShare {
	return c.dkshare
}

func (c *committee) IsAlivePeer(peerIndex uint16) bool {
	allNodes := c.validatorNodes.AllNodes()
	if int(peerIndex) >= len(allNodes) {
		return false
	}
	if peerIndex == c.ownIndex {
		return true
	}
	if allNodes[peerIndex] == nil {
		panic(xerrors.Errorf("c.validatorNodes[peerIndex] == nil. peerIndex: %d, ownIndex: %d", peerIndex, c.ownIndex))
	}
	return allNodes[peerIndex].IsAlive()
}

func (c *committee) QuorumIsAlive(quorum ...uint16) bool {
	q := c.quorum
	if len(quorum) > 0 {
		q = quorum[0]
	}
	count := uint16(0)
	for _, peer := range c.validatorNodes.OtherNodes() {
		if peer.IsAlive() {
			count++
		}
		if count+1 >= q {
			return true
		}
	}
	return false
}

func (c *committee) PeerStatus() []*chain.PeerStatus {
	ret := make([]*chain.PeerStatus, 0)
	for i, peer := range c.validatorNodes.AllNodes() {
		status := &chain.PeerStatus{
			Index:     int(i),
			NetID:     peer.NetID(),
			PubKey:    peer.PubKey(),
			Connected: peer.IsAlive(),
		}
		ret = append(ret, status)
	}
	return ret
}

func (c *committee) Close() {
	c.acsRunner.Close()
	c.isReady.Store(false)
	c.validatorNodes.Close()
}

func (c *committee) RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte)) {
	c.acsRunner.RunACSConsensus(value, sessionID, stateIndex, callback)
}

func (c *committee) waitReady(waitReady bool) {
	if waitReady {
		c.log.Infof("wait for at least quorum of committee validatorNodes (%d) to connect before activating the committee", c.Quorum())
		for !c.QuorumIsAlive() {
			time.Sleep(100 * time.Millisecond)
		}
	}
	c.log.Infof("committee started for address %s", c.dkshare.Address.Base58())
	c.log.Debugf("peer status: %s", c.PeerStatus())
	c.isReady.Store(true)
}

func (c *committee) GetRandomValidators(upToN int) []*cryptolib.PublicKey {
	validators := c.validatorNodes.OtherNodes()
	if upToN >= len(validators) {
		valPubKeys := make([]*cryptolib.PublicKey, 0)
		for i := range validators {
			valPubKeys = append(valPubKeys, validators[i].PubKey())
		}
		return valPubKeys
	}

	var b [8]byte
	seed := b[:]
	_, _ = rand.Read(seed)
	permutation := util.NewPermutation16(uint16(len(validators)), seed)
	permutation.Shuffle(seed)
	ret := make([]*cryptolib.PublicKey, 0)
	for len(ret) < upToN {
		i := permutation.Next()
		ret = append(ret, validators[i].PubKey())
	}

	return ret
}
