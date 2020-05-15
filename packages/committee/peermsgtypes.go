package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/plugins/peering"
	"time"
)

const (
	MsgNotifyRequests         = 0 + peering.FirstCommitteeMsgCode
	MsgStartProcessingRequest = 1 + peering.FirstCommitteeMsgCode
	MsgSignedHash             = 2 + peering.FirstCommitteeMsgCode
	MsgGetStateUpdate         = 3 + peering.FirstCommitteeMsgCode
	MsgStateUpdate            = 4 + peering.FirstCommitteeMsgCode
	MsgTestTrace              = 5 + peering.FirstCommitteeMsgCode
)

type TimerTick int

// all peer messages have this
type PeerMsgHeader struct {
	// is set upon receive the message
	SenderIndex uint16
	// state index in the context of which the message is sent
	StateIndex uint32
}

// message is sent to the leader of the state processing
// it is sent upon state change or upon arrival of the new request
// the receiving operator will ignore repeating messages
type NotifyReqMsg struct {
	PeerMsgHeader
	// list of request ids ordered by the time of arrival
	RequestIds []*sctransaction.RequestId
}

// message is sent by the leader to other peers to initiate request processing
// other peers are expected to check is timestamp is acceptable then
// process request and sign the result hash with the timestamp proposed by the leader
type StartProcessingReqMsg struct {
	PeerMsgHeader
	// timestamp of the message. Field is set upon receive the message to sender's timestamp
	Timestamp time.Time
	// request id
	RequestId *sctransaction.RequestId
	// reward address
	RewardAddress *address.Address
	// balances/outputs
	Balances map[transaction.ID][]*balance.Balance
}

// after calculations the result peer responds to the start processing msg
// with SignedHashMsg, which contains result hash and signatures
type SignedHashMsg struct {
	PeerMsgHeader
	// timestamp of this message. Field is set upon receive the message to sender's timestamp
	Timestamp time.Time
	// request id
	RequestId *sctransaction.RequestId
	// original timestamp, the parameter for calculations, which is signed as part of the essence
	OrigTimestamp time.Time
	// hash of the signed data (essence)
	EssenceHash *hashing.HashValue
	// signature
	SigShare tbdn.SigShare
}

// request state update from peer. Used in syn process
type GetStateUpdateMsg struct {
	PeerMsgHeader
}

// state update sent to peer. Used in sync process
type StateUpdateMsg struct {
	PeerMsgHeader
	// state update
	StateUpdate state.StateUpdate
	// locally calculated by VM (needed for syncing)
	FromVM bool
}

// used for testing of the communications
type TestTraceMsg struct {
	PeerMsgHeader
	InitTime      int64
	InitPeerIndex uint16
	Sequence      []uint16
}

// state manager notifies consensus operator about changed state
// only sent internally within committee
// state transition is always from state N to state N+1
type StateTransitionMsg struct {
	// new variable state
	VariableState state.VariableState
	// corresponding state transaction
	StateTransaction *sctransaction.Transaction
	// processed requests
	RequestIds []*sctransaction.RequestId
}
