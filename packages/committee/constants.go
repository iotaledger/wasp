package committee

import "time"

const (
	// how often timer events are triggered in commiitee
	// each second timer tick is sent to consensus, another to state manager, so timer events
	// in objects are triggered each 200 milliseconds
	TimerTickPeriod = 100 * time.Millisecond

	// if leader actions does not reset deadline, leader is rotated according to current permutation
	// after this period of time. usually this period of time should be larger than expected confirmation
	// by the transaction on the value tangle (expected 6-10 sec)
	// if this parameters is too small, conflicts are possible. If it is too large, the committee is too slow
	// to change the leader when it is needed to come out of meta stable state.
	LeaderRotationPeriod = 3 * time.Second

	// when idle, consensus object periodically refreshes balances of its own address
	RequestBalancesPeriod = 10 * time.Second

	// if node is behind the current state (not synced) it send GetBatch messages to pseudo-randomly
	// selected peer to get the batch it needs. Node expects answer, if not the message is repeated to another
	// peer after some time
	PeriodBetweenSyncMessages = 1 * time.Second

	// State Manager is requesting transaction to confirm a pending batch from the goshimmer node.
	// Request is repeated if necessary.
	StateTransactionRequestTimeout = 10 * time.Second

	// each committee have input channel. If channel is congested, sending to it timeouts and message is lost
	ReceiveMsgChannelTimeout = 500 * time.Millisecond
)
