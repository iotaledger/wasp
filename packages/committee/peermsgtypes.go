package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/plugins/peering"
)

const (
	MsgNotifyRequests          = 0 + peering.FirstCommitteeMsgCode
	MsgNotifyFinalResultPosted = 1 + peering.FirstCommitteeMsgCode
	MsgStartProcessingRequest  = 2 + peering.FirstCommitteeMsgCode
	MsgSignedHash              = 3 + peering.FirstCommitteeMsgCode
	MsgGetBatch                = 4 + peering.FirstCommitteeMsgCode
	MsgStateUpdate             = 5 + peering.FirstCommitteeMsgCode
	MsgBatchHeader             = 6 + peering.FirstCommitteeMsgCode
	MsgTestTrace               = 7 + peering.FirstCommitteeMsgCode
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
	RequestIds []sctransaction.RequestId
}

// message is sent by the leader to all peers immediately after the final transaction is posted
// to the tangle. Main purpose of the message is to prevent unnecessary leader rotation
// in long confirmation times
// Final signature is sent to prevent possibility for a leader node to lie (is it necessary)
type NotifyFinalResultPostedMsg struct {
	PeerMsgHeader
	Signature signaturescheme.Signature
}

// message is sent by the leader to other peers to initiate request processing
// other peers are expected to check is timestamp is acceptable then
// process request batch and sign the result hash with the timestamp proposed by the leader
type StartProcessingBatchMsg struct {
	PeerMsgHeader
	// timestamp of the message. Field is set upon receive the message to sender's timestamp
	Timestamp int64
	// batch of request ids
	RequestIds []sctransaction.RequestId
	// reward address
	RewardAddress address.Address
	// balances/outputs
	Balances map[valuetransaction.ID][]*balance.Balance
}

// after calculations the result peer responds to the start processing msg
// with SignedHashMsg, which contains result hash and signatures
type SignedHashMsg struct {
	PeerMsgHeader
	// timestamp of this message. Field is set upon receive the message to sender's timestamp
	Timestamp int64
	// returns hash of all req ids
	BatchHash hashing.HashValue
	// original timestamp, the parameter for calculations, which is signed as part of the essence
	OrigTimestamp int64
	// hash of the signed data (essence)
	EssenceHash hashing.HashValue
	// signature
	SigShare tbdn.SigShare
}

// request batch of updates from peer. Used in syn process
type GetBatchMsg struct {
	PeerMsgHeader
}

// the header of the batch message sent by peers in the process of syncing
// it is sent as a first message while syncing a batch
type BatchHeaderMsg struct {
	PeerMsgHeader
	// state index of the batch
	Size uint16
	// approving transaction id
	StateTransactionId valuetransaction.ID
}

// state update sent to peer. Used in sync process, as part of batch
type StateUpdateMsg struct {
	PeerMsgHeader
	// state update
	StateUpdate state.StateUpdate
	// position in a batch
	BatchIndex uint16
}

// used for testing of the communications
type TestTraceMsg struct {
	PeerMsgHeader
	InitTime      int64
	InitPeerIndex uint16
	Sequence      []uint16
	NumHops       uint16
}

// state manager notifies consensus operator about changed state
// only sent internally within committee
// state transition is always from state N to state N+1
type StateTransitionMsg struct {
	// new variable state
	VariableState state.VirtualState
	// corresponding state transaction
	StateTransaction *sctransaction.Transaction
	// processed requests
	RequestIds []*sctransaction.RequestId
	// is the state index last seen
	Synchronized bool
}

// message of complete batch. Is sent by consensus operator to the state manager as a VM result
// - state manager to itself when batch is completed after syncing
type PendingBatchMsg struct {
	Batch state.Batch
}

// message sent to notify VM processor is ready. It is a successful finish of asynchronous loading of the processor
type ProcessorIsReady struct {
	ProgramHash string // base58
}

// message is sent to the consensus manager after it receives state transaction
// which is valid but not confirmed yet.
type StateTransactionEvidenced struct {
	TxId      valuetransaction.ID
	StateHash hashing.HashValue
}
