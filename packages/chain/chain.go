// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"sync"
)

type Chain interface {
	ID() *coretypes.ChainID
	Address() ledgerstate.Address
	Size() uint16
	Quorum() uint16
	OwnPeerIndex() uint16
	NumPeers() uint16
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToCommitteePeers(msgType byte, msgData []byte, ts int64) uint16
	IsAlivePeer(peerIndex uint16) bool
	HasQuorum() bool
	PeerStatus() []*PeerStatus
	BlobCache() coretypes.BlobCache
	// TODO distinguish external messages from internal and peer messages
	ReceiveMessage(interface{})
	ReceiveTransaction(*ledgerstate.Transaction)
	ReceiveInclusionState(ledgerstate.TransactionID, ledgerstate.InclusionState)
	ReceiveRequest(coretypes.Request)
	//
	SetReadyStateManager()
	SetReadyConsensus()
	Dismiss()
	IsDismissed() bool
	// requests
	GetRequestProcessingStatus(id *coretypes.RequestID) RequestProcessingStatus
	EventRequestProcessed() *events.Event
	// chain processors
	Processors() *processors.ProcessorCache
}

type PeerStatus struct {
	Index     int
	PeeringID string
	IsSelf    bool
	Connected bool
}

func (p *PeerStatus) String() string {
	return fmt.Sprintf("%+v", *p)
}

type RequestProcessingStatus int

const (
	RequestProcessingStatusUnknown = RequestProcessingStatus(iota)
	RequestProcessingStatusBacklog
	RequestProcessingStatusCompleted
)

type StateManager interface {
	EvidenceStateIndex(idx uint32)
	EventStateIndexPingPongMsg(msg *StateIndexPingPongMsg)
	EventGetBlockMsg(msg *GetBlockMsg)
	EventBlockHeaderMsg(msg *BlockHeaderMsg)
	EventStateUpdateMsg(msg *StateUpdateMsg)
	EventStateTransactionMsg(msg *ledgerstate.AliasOutput)
	EventPendingBlockMsg(msg PendingBlockMsg)
	EventTimerMsg(msg TimerTick)
	Close()
}

type Operator interface {
	EventStateTransitionMsg(*StateTransitionMsg)
	EventRequestMsg(coretypes.Request)
	EventNotifyReqMsg(*NotifyReqMsg)
	EventStartProcessingBatchMsg(*StartProcessingBatchMsg)
	EventResultCalculated(msg *VMResultMsg)
	EventSignedHashMsg(*SignedHashMsg)
	EventNotifyFinalResultPostedMsg(*NotifyFinalResultPostedMsg)
	EventTransactionInclusionStateMsg(msg *InclusionStateMsg)
	EventTimerMsg(TimerTick)
	Close()
	//
	IsRequestInBacklog(*coretypes.RequestID) bool
}

type chainConstructor func(
	chr *registry.ChainRecord,
	log *logger.Logger,
	netProvider peering.NetworkProvider,
	dksProvider tcrypto.RegistryProvider,
	blobProvider coretypes.BlobCache,
	onActivation func(),
) Chain

var constructorNew chainConstructor
var constructorNewMutex sync.Mutex

func RegisterChainConstructor(constr chainConstructor) {
	constructorNewMutex.Lock()
	defer constructorNewMutex.Unlock()

	if constructorNew != nil {
		panic("RegisterChainConstructor: already registered")
	}
	constructorNew = constr
}

func New(
	chr *registry.ChainRecord,
	log *logger.Logger,
	netProvider peering.NetworkProvider,
	dksProvider tcrypto.RegistryProvider,
	blobProvider coretypes.BlobCache,
	onActivation func(),
) Chain {
	return constructorNew(chr, log, netProvider, dksProvider, blobProvider, onActivation)
}
