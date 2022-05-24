package admapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addChainMetricsEndpoints(adm echoswagger.ApiGroup, chainsProvider chains.Provider) {
	cms := &chainMetricsService{
		chainsProvider,
	}
	addChainNodeConnMetricsEndpoints(adm, cms)
	addChainConsensusMetricsEndpoints(adm, cms)
	addChainConcensusPipeMetricsEndpoints(adm, cms)
}

func addChainNodeConnMetricsEndpoints(adm echoswagger.ApiGroup, cms *chainMetricsService) {
	chainExample := &model.NodeConnectionMessagesMetrics{
		OutPublishTransaction: &model.NodeConnectionMessageMetrics{
			Total:       3,
			LastEvent:   time.Now().Add(-2 * time.Millisecond),
			LastMessage: "Last sent PublishTransaction message structure",
		},
		OutPullLatestOutput: &model.NodeConnectionMessageMetrics{
			Total:       15,
			LastEvent:   time.Now().Add(-10 * time.Second),
			LastMessage: "Last sent PullLatestOutput message structure",
		},
		OutPullTxInclusionState: &model.NodeConnectionMessageMetrics{
			Total:       28,
			LastEvent:   time.Now().Add(-5 * time.Second),
			LastMessage: "Last sent PullTxInclusionState message structure",
		},
		OutPullOutputByID: &model.NodeConnectionMessageMetrics{
			Total:       132,
			LastEvent:   time.Now().Add(100 * time.Second),
			LastMessage: "Last sent PullOutputByID message structure",
		},
		InStateOutput: &model.NodeConnectionMessageMetrics{
			Total:       101,
			LastEvent:   time.Now().Add(-8 * time.Second),
			LastMessage: "Last received State output message structure",
		},
		InAliasOutput: &model.NodeConnectionMessageMetrics{
			Total:       203,
			LastEvent:   time.Now().Add(-123 * time.Millisecond),
			LastMessage: "Last received AliasOutput message structure",
		},
		InOutput: &model.NodeConnectionMessageMetrics{
			Total:       101,
			LastEvent:   time.Now().Add(-8 * time.Second),
			LastMessage: "Last received Output message structure",
		},
		InOnLedgerRequest: &model.NodeConnectionMessageMetrics{
			Total:       85,
			LastEvent:   time.Now().Add(-2 * time.Second),
			LastMessage: "Last received OnLedgerRequest message structure",
		},
		InTxInclusionState: &model.NodeConnectionMessageMetrics{
			Total:       999,
			LastEvent:   time.Now().Add(-1 * time.Second),
			LastMessage: "Last received TxInclusionState message structure",
		},
	}

	example := &model.NodeConnectionMetrics{
		NodeConnectionMessagesMetrics: *chainExample,
		InMilestone: &model.NodeConnectionMessageMetrics{
			Total:       1234,
			LastEvent:   time.Now().Add(1 * time.Second),
			LastMessage: "Last received Milestone message structure",
		},
		Registered: []model.ChainID{
			model.NewChainID(iscp.RandomChainID()),
			model.NewChainID(iscp.RandomChainID()),
		},
	}

	adm.GET(routes.GetChainsNodeConnectionMetrics(), cms.handleGetChainsNodeConnMetrics).
		SetSummary("Get cummulative chains node connection metrics").
		AddResponse(http.StatusOK, "Cummulative chains metrics", example, nil)

	adm.GET(routes.GetChainNodeConnectionMetrics(":chainID"), cms.handleGetChainNodeConnMetrics).
		SetSummary("Get chain node connection metrics for the given chain ID").
		AddParamPath("", "chainID", "ChainID (bech32)").
		AddResponse(http.StatusOK, "Chain metrics", chainExample, nil)
}

func addChainConsensusMetricsEndpoints(adm echoswagger.ApiGroup, cms *chainMetricsService) {
	example := &model.ConsensusWorkflowStatus{
		FlagStateReceived:        true,
		FlagBatchProposalSent:    true,
		FlagConsensusBatchKnown:  true,
		FlagVMStarted:            false,
		FlagVMResultSigned:       false,
		FlagTransactionFinalized: false,
		FlagTransactionPosted:    false,
		FlagTransactionSeen:      false,
		FlagInProgress:           true,

		TimeBatchProposalSent:    time.Now().Add(-10 * time.Second),
		TimeConsensusBatchKnown:  time.Now().Add(-5 * time.Second),
		TimeVMStarted:            time.Time{},
		TimeVMResultSigned:       time.Time{},
		TimeTransactionFinalized: time.Time{},
		TimeTransactionPosted:    time.Time{},
		TimeTransactionSeen:      time.Time{},
		TimeCompleted:            time.Time{},

		CurrentStateIndex: 0,
	}

	adm.GET(routes.GetChainConsensusWorkflowStatus(":chainID"), cms.handleGetChainConsensusWorkflowStatus).
		SetSummary("Get chain state statistics for the given chain ID").
		AddParamPath("", "chainID", "ChainID (bech32)").
		AddResponse(http.StatusOK, "Chain consensus stats", example, nil).
		AddResponse(http.StatusNotFound, "Chain consensus hasn't been created", nil, nil)
}

func addChainConcensusPipeMetricsEndpoints(adm echoswagger.ApiGroup, cms *chainMetricsService) {
	example := &model.ConsensusPipeMetrics{
		EventStateTransitionMsgPipeSize: 0,
		EventSignedResultMsgPipeSize:    0,
		EventSignedResultAckMsgPipeSize: 0,
		EventInclusionStateMsgPipeSize:  0,
		EventACSMsgPipeSize:             0,
		EventVMResultMsgPipeSize:        0,
		EventTimerMsgPipeSize:           0,
	}

	adm.GET(routes.GetChainConsensusPipeMetrics(":chainID"), cms.handleGetChainConsensusPipeMetrics).
		SetSummary("Get consensus pipe metrics").
		AddParamPath("", "chainID", "CHAINid (base58)").
		AddResponse(http.StatusOK, "Chain consensus pipe metrics", example, nil).
		AddResponse(http.StatusNotFound, "Chain consensus hasn't been created", nil, nil)
}

type chainMetricsService struct {
	chains chains.Provider
}

func (cssT *chainMetricsService) handleGetChainsNodeConnMetrics(c echo.Context) error {
	metrics := cssT.chains().GetNodeConnectionMetrics()
	metricsModel := model.NewNodeConnectionMetrics(metrics)

	return c.JSON(http.StatusOK, metricsModel)
}

func (cssT *chainMetricsService) handleGetChainNodeConnMetrics(c echo.Context) error {
	chainID, err := iscp.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(err.Error())
	}
	theChain := cssT.chains().Get(chainID)
	if theChain == nil {
		return httperrors.NotFound(fmt.Sprintf("Active chain %s not found", chainID))
	}
	metrics := theChain.GetNodeConnectionMetrics()
	metricsModel := model.NewNodeConnectionMessagesMetrics(metrics)

	return c.JSON(http.StatusOK, metricsModel)
}

func (cssT *chainMetricsService) handleGetChainConsensusWorkflowStatus(c echo.Context) error {
	theChain, err := cssT.getChain(c)
	if err != nil {
		return err
	}
	status := theChain.GetConsensusWorkflowStatus()
	if status == nil {
		return c.NoContent(http.StatusNotFound)
	}
	statusModel := model.NewConsensusWorkflowStatus(status)

	return c.JSON(http.StatusOK, statusModel)
}

func (cssT *chainMetricsService) handleGetChainConsensusPipeMetrics(c echo.Context) error {
	theChain, err := cssT.getChain(c)
	if err != nil {
		return err
	}
	pipeMetrics := theChain.GetConsensusPipeMetrics()
	if pipeMetrics == nil {
		return c.NoContent(http.StatusNotFound)
	}
	pipeMetricsModel := model.NewConsensusPipeMetrics(pipeMetrics)
	return c.JSON(http.StatusOK, pipeMetricsModel)
}

func (cssT *chainMetricsService) getChain(c echo.Context) (chain.Chain, error) {
	chainID, err := iscp.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return nil, httperrors.BadRequest(err.Error())
	}
	theChain := cssT.chains().Get(chainID)
	if theChain == nil {
		return nil, httperrors.NotFound(fmt.Sprintf("Active chain %s not found", chainID))
	}
	return theChain, nil
}
