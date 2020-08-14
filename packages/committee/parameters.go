package committee

import "time"

const (
	// network latency assumption: time to gossip transaction to the whole network
	networkLatency = 2 * time.Second

	// confirmation time assumption. Average time from posting a transaction to finality
	confirmationTime = 10 * time.Second
)

// Params are time parameters for consensus ans state manager
type Parameters struct {
	// -- general
	// how often timer events are triggered in committee
	// each second timer tick is sent to consensus, another to state manager, so timer events
	// in objects are triggered each 200 milliseconds
	TimerTickPeriod time.Duration

	// each committee have input channel. If channel is congested, sending to it timeouts and message is lost
	ReceiveMsgChannelTimeout time.Duration

	// -- consensus operator
	// expected leader reaction after posting notifications. It it the time enough to collect quorum
	// of notifications from peers
	LeaderReactionToNotifications time.Duration

	// expected leader reaction time after sending calculated result to it
	LeaderReactionToCalculatedResult time.Duration

	// timeout for confirmation after the leader notifies the committee it has posted result transaction to the tangle
	// or after evidence about posted transaction has reached consensus operator
	ConfirmationWaitingPeriod time.Duration

	// when idle, consensus object periodically refreshes balances of its own address
	RequestBalancesPeriod time.Duration

	// -- state manager
	// if node is behind the current state (not synced) it send GetBatch messages to pseudo-randomly
	// selected peer to get the batch it needs. Node expects answer, if not the message is repeated to another
	// peer after some time
	PeriodBetweenSyncMessages time.Duration

	// State Manager is requesting transaction to confirm a pending batch from the goshimmer node.
	// Request is repeated if necessary.
	StateTransactionRequestTimeout time.Duration
}

var DefaultParameters = &Parameters{
	TimerTickPeriod:                  100 * time.Millisecond,
	ReceiveMsgChannelTimeout:         500 * time.Millisecond,
	LeaderReactionToNotifications:    (networkLatency * 3) / 2,
	LeaderReactionToCalculatedResult: 2 * time.Second,
	ConfirmationWaitingPeriod:        confirmationTime + networkLatency,
	RequestBalancesPeriod:            10 * time.Second,
	PeriodBetweenSyncMessages:        1 * time.Second,
	StateTransactionRequestTimeout:   10 * time.Second,
}
