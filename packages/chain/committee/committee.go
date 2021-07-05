package committee

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/chain/consensus/commonsubset"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry/committee_record"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"
)

type committee struct {
	isReady        *atomic.Bool
	address        ledgerstate.Address
	peerConfig     coretypes.PeerNetworkConfigProvider
	validatorNodes peering.GroupProvider
	ccProvider     commoncoin.Provider // Just to close it afterwards.
	acsRunner      chain.AsynchronousCommonSubsetRunner
	peeringID      peering.PeeringID
	size           uint16
	quorum         uint16
	ownIndex       uint16
	dkshare        *tcrypto.DKShare
	attachID       interface{}
	log            *logger.Logger
}

const waitReady = false

func New(
	cmtRec *committee_record.CommitteeRecord,
	chainID *chainid.ChainID,
	netProvider peering.NetworkProvider,
	peerConfig coretypes.PeerNetworkConfigProvider,
	dksProvider coretypes.DKShareRegistryProvider,
	log *logger.Logger,
	acsRunner ...chain.AsynchronousCommonSubsetRunner, // Only for mocking.
) (chain.Committee, error) {
	// load DKShare from the registry
	dkshare, err := dksProvider.LoadDKShare(cmtRec.Address)
	if err != nil {
		return nil, xerrors.Errorf("NewCommittee: failed loading DKShare for address %s: %w", cmtRec.Address.Base58(), err)
	}
	if dkshare.Index == nil {
		return nil, xerrors.Errorf("NewCommittee: wrong DKShare record for address %s: %w", cmtRec.Address.Base58(), err)
	}
	if err := checkValidatorNodeIDs(peerConfig, dkshare.N, *dkshare.Index, cmtRec.Nodes); err != nil {
		return nil, xerrors.Errorf("NewCommittee: %w", err)
	}
	var peers peering.GroupProvider
	if peers, err = netProvider.PeerGroup(cmtRec.Nodes); err != nil {
		return nil, xerrors.Errorf("NewCommittee: failed to create peer group for committee: %+v: %w", cmtRec.Nodes, err)
	}
	log.Debugf("NewCommittee: peer group: %+v", cmtRec.Nodes)
	// peerGroupID is calculated by XORing chainID and stateAddr.
	// It allows to use same statAddr for different chains
	peerGroupID := cmtRec.Address.Array()
	var chainArr [33]byte
	if chainID != nil {
		chainArr = chainID.Array()
	}
	for i := range peerGroupID {
		peerGroupID[i] ^= chainArr[i]
	}
	ret := &committee{
		isReady:        atomic.NewBool(false),
		address:        cmtRec.Address,
		validatorNodes: peers,
		peeringID:      peerGroupID,
		peerConfig:     peerConfig,
		size:           dkshare.N,
		quorum:         dkshare.T,
		ownIndex:       *dkshare.Index,
		dkshare:        dkshare,
		log:            log,
	}
	if len(acsRunner) > 0 {
		ret.acsRunner = acsRunner[0]
	} else {
		// That's the default implementation of the ACS.
		// We use it, of the mocked variant was not passed.
		ret.ccProvider = commoncoin.NewCommonCoinNode(
			nil,
			dkshare,
			peerGroupID,
			peers,
			log,
		)
		ret.acsRunner = commonsubset.NewCommonSubsetCoordinator(
			peerGroupID,
			netProvider,
			peers,
			dkshare.T,
			ret.ccProvider,
			log,
		)
	}
	go ret.waitReady(waitReady)

	return ret, nil
}

func (c *committee) Address() ledgerstate.Address {
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

func (c *committee) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if peer, ok := c.validatorNodes.OtherNodes()[targetPeerIndex]; ok {
		peer.SendMsg(&peering.PeerMessage{
			PeeringID:   c.peeringID,
			SenderIndex: c.ownIndex,
			MsgType:     msgType,
			MsgData:     msgData,
		})
		return nil
	}
	return fmt.Errorf("SendMsg: wrong peer index")
}

func (c *committee) SendMsgToPeers(msgType byte, msgData []byte, ts int64) {
	msg := &peering.PeerMessage{
		PeeringID:   c.peeringID,
		SenderIndex: c.ownIndex,
		Timestamp:   ts,
		MsgType:     msgType,
		MsgData:     msgData,
	}
	c.validatorNodes.Broadcast(msg, false)
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
		isSelf := peer == nil || peer.NetID() == c.peerConfig.OwnNetID()
		status := &chain.PeerStatus{
			Index:  int(i),
			IsSelf: isSelf,
		}
		if isSelf {
			status.PeeringID = c.peerConfig.OwnNetID()
			status.Connected = true
		} else {
			status.PeeringID = peer.NetID()
			status.Connected = peer.IsAlive()
		}
		ret = append(ret, status)
	}
	return ret
}

func (c *committee) Attach(ch chain.ChainCore) {
	c.attachID = c.validatorNodes.Attach(&c.peeringID, func(recv *peering.RecvEvent) {
		c.log.Infof("XXX: Attach(received=%+v)-0", recv)
		if c.ccProvider != nil && c.ccProvider.TryHandleMessage(recv) {
			return
		}
		c.log.Infof("XXX: Attach(received=%+v)-1", recv)
		if c.acsRunner != nil && c.acsRunner.TryHandleMessage(recv) {
			return
		}
		c.log.Infof("XXX: Attach(received=%+v)-2", recv)
		ch.ReceiveMessage(recv.Msg)
	})
}

func (c *committee) Close() {
	c.acsRunner.Close()
	if c.ccProvider != nil {
		c.ccProvider.Close()
	}
	c.isReady.Store(false)
	if c.attachID != nil {
		c.validatorNodes.Detach(c.attachID)
	}
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

func checkValidatorNodeIDs(cfg coretypes.PeerNetworkConfigProvider, n, ownIndex uint16, validatorNetIDs []string) error {
	if !util.AllDifferentStrings(validatorNetIDs) {
		return xerrors.Errorf("checkValidatorNodeIDs: list of validators nodes contains duplicates: %+v", validatorNetIDs)
	}
	if len(validatorNetIDs) != int(n) {
		return xerrors.Errorf("checkValidatorNodeIDs: number of validator nodes must be equal to the N parameter of the committee")
	}
	if ownIndex >= n {
		return xerrors.New("checkValidatorNodeIDs: wrong own validator index")
	}
	if validatorNetIDs[ownIndex] != cfg.OwnNetID() {
		return xerrors.New("checkValidatorNodeIDs: own netID is expected at own validator index")
	}
	// check if all validator node IDs are among known validatorNodes
	allPeers := []string{cfg.OwnNetID()}
	allPeers = append(allPeers, cfg.Neighbors()...)
	if !util.IsSubset(validatorNetIDs, allPeers) {
		return xerrors.Errorf("not all validator nodes are among known neighbors: all peers: %+v, committee: %+v",
			allPeers, validatorNetIDs)
	}
	return nil
}
