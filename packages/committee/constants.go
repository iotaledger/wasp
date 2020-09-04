package committee

import "time"

const (
	// confirmation time assumption. Average time from posting a transaction to finality
	ConfirmationTime = 10 * time.Second

	// relaxation period after committee activation
	InitConnectPeriod = 5 * time.Second

	// time tick for consensus and state manager objects
	TimerTickPeriod = 100 * time.Millisecond

	// timeout for congested input channel for the consensus and state manager objects.channel.
	// After the timeout message is lost
	ReceiveMsgChannelTimeout = 500 * time.Millisecond

	RequestBalancesPeriod = 10 * time.Second

	// if node is behind the current state (not synced) it send GetBatch messages to pseudo-randomly
	// selected peer to get the batch it needs. Node expects answer, if not the message is repeated to another
	// peer after some time
	PeriodBetweenSyncMessages = 1 * time.Second

	// State Manager is requesting transaction to confirm a pending batch from the goshimmer node.
	// Request is repeated if necessary.
	StateTransactionRequestTimeout = 10 * time.Second
)
