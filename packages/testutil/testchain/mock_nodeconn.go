package testchain

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

type MockedNodeConn struct {
	log                                    *logger.Logger
	ledgers                                *MockedLedgers
	id                                     string
	publishTransactionAllowedFun           func(chainID *isc.ChainID, tx *iotago.Transaction) bool
	publishGovernanceTransactionAllowedFun func(chainID *isc.ChainID, tx *iotago.Transaction) bool
	pullLatestOutputAllowed                bool
	pullTxInclusionStateAllowedFun         func(chainID *isc.ChainID, txID iotago.TransactionID) bool
	pullOutputByIDAllowedFun               func(chainID *isc.ChainID, outputID iotago.OutputID) bool
	stopChannel                            chan bool
	attachMilestonesClosures               map[isc.ChainID]*events.Closure
}

var _ chain.NodeConnection = &MockedNodeConn{}

func NewMockedNodeConnection(id string, ledgers *MockedLedgers, log *logger.Logger) *MockedNodeConn {
	result := &MockedNodeConn{
		log:                      log.Named("mnc"),
		id:                       id,
		ledgers:                  ledgers,
		stopChannel:              make(chan bool),
		attachMilestonesClosures: make(map[isc.ChainID]*events.Closure),
	}
	//result.SetPublishStateTransactionAllowed(true)
	//result.SetPublishGovernanceTransactionAllowed(true)
	//result.SetPullLatestOutputAllowed(true)
	//result.SetPullTxInclusionStateAllowed(true)
	//result.SetPullOutputByIDAllowed(true)
	result.log.Debugf("Nodeconn created")
	return result
}

// Publishing can be canceled via the context.
// The result must be returned via the callback, unless ctx is canceled first.
// PublishTX handles promoting and reattachments until the tx is confirmed or the context is canceled.
func (mncT *MockedNodeConn) PublishTX(
	ctx context.Context,
	chainID *isc.ChainID,
	tx *iotago.Transaction,
	callback chain.TxPostHandler,
) error {
	if mncT.publishTransactionAllowedFun(chainID, tx) {
		return mncT.ledgers.GetLedger(chainID).PublishTransaction(tx)
	}
	return fmt.Errorf("Publishing state transaction for chain %s is not allowed", chainID)
}

func (mncT *MockedNodeConn) AttachChain(
	ctx context.Context,
	chainID *isc.ChainID,
	recvRequestCB chain.RequestOutputHandler,
	recvAliasOutput chain.AliasOutputHandler,
	recvMilestone chain.MilestoneHandler,
) {
	panic("IMPLEMENT: (mncT *MockedNodeConn) AttachChain") // TODO: Implement.
}

func (mncT *MockedNodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return nodeconnmetrics.NewEmptyNodeConnectionMetrics()
}

func (mncT *MockedNodeConn) Run(ctx context.Context) {
	panic("should be unused in test")
}

/*
func (mncT *MockedNodeConn) ID() string {
	return mncT.id
}

func (mncT *MockedNodeConn) RegisterChain(
	chainID *isc.ChainID,
	stateOutputHandler,
	outputHandler func(iotago.OutputID, iotago.Output),
	milestoneHandler func(*nodebridge.Milestone),
) {
	mncT.ledgers.GetLedger(chainID).Register(mncT.id, stateOutputHandler, outputHandler)
	mncT.attachMilestonesClosures[*chainID] = mncT.AttachMilestones(milestoneHandler)
}

func (mncT *MockedNodeConn) UnregisterChain(chainID *isc.ChainID) {
	mncT.ledgers.GetLedger(chainID).Unregister(mncT.id)
	mncT.DetachMilestones(mncT.attachMilestonesClosures[*chainID])
}

func (mncT *MockedNodeConn) PullLatestOutput(chainID *isc.ChainID) {
	if mncT.pullLatestOutputAllowed {
		mncT.ledgers.GetLedger(chainID).PullLatestOutput(mncT.id)
	} else {
		mncT.log.Errorf("Pull latest output for address %s is not allowed", chainID)
	}
}

func (mncT *MockedNodeConn) PullStateOutputByID(chainID *isc.ChainID, outputID iotago.OutputID) {
	if mncT.pullOutputByIDAllowedFun(chainID, outputID) {
		mncT.ledgers.GetLedger(chainID).PullStateOutputByID(mncT.id, outputID)
	} else {
		mncT.log.Errorf("Pull output by ID %s for address %s is not allowed", outputID, chainID)
	}
}

func (mncT *MockedNodeConn) SetPublishStateTransactionAllowed(flag bool) {
	mncT.SetPublishStateTransactionAllowedFun(func(*isc.ChainID, *iotago.Transaction) bool { return flag })
}

func (mncT *MockedNodeConn) SetPublishStateTransactionAllowedFun(fun func(chainID *isc.ChainID, tx *iotago.Transaction) bool) {
	mncT.publishTransactionAllowedFun = fun
}

func (mncT *MockedNodeConn) SetPublishGovernanceTransactionAllowed(flag bool) {
	mncT.SetPublishGovernanceTransactionAllowedFun(func(*isc.ChainID, *iotago.Transaction) bool { return flag })
}

func (mncT *MockedNodeConn) SetPublishGovernanceTransactionAllowedFun(fun func(chainID *isc.ChainID, tx *iotago.Transaction) bool) {
	mncT.publishGovernanceTransactionAllowedFun = fun
}

func (mncT *MockedNodeConn) SetPullLatestOutputAllowed(flag bool) {
	mncT.pullLatestOutputAllowed = flag
}

func (mncT *MockedNodeConn) SetPullTxInclusionStateAllowed(flag bool) {
	mncT.SetPullTxInclusionStateAllowedFun(func(*isc.ChainID, iotago.TransactionID) bool { return flag })
}

func (mncT *MockedNodeConn) SetPullTxInclusionStateAllowedFun(fun func(chainID *isc.ChainID, txID iotago.TransactionID) bool) {
	mncT.pullTxInclusionStateAllowedFun = fun
}

func (mncT *MockedNodeConn) SetPullOutputByIDAllowed(flag bool) {
	mncT.SetPullOutputByIDAllowedFun(func(*isc.ChainID, iotago.OutputID) bool { return flag })
}

func (mncT *MockedNodeConn) SetPullOutputByIDAllowedFun(fun func(chainID *isc.ChainID, outputID iotago.OutputID) bool) {
	mncT.pullOutputByIDAllowedFun = fun
}

func (mncT *MockedNodeConn) AttachMilestones(handler func(*nodebridge.Milestone)) *events.Closure {
	return mncT.ledgers.AttachMilestones(handler)
}

func (mncT *MockedNodeConn) DetachMilestones(attachID *events.Closure) {
	mncT.ledgers.DetachMilestones(attachID)
}
*/
