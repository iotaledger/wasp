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
	cms := &chainMetricsService{chainsProvider}
	addChainNodeConnMetricsEndpoints(adm, cms)
	addChainConsensusMetricsEndpoints(adm, cms)
}

func addChainNodeConnMetricsEndpoints(adm echoswagger.ApiGroup, cms *chainMetricsService) {
	chainExample := &model.NodeConnectionMessagesMetrics{
		OutPullState: &model.NodeConnectionMessageMetrics{
			Total:       15,
			LastEvent:   time.Now().Add(-10 * time.Second),
			LastMessage: "Last sent PullState message structure",
		},
		OutPullTransactionInclusionState: &model.NodeConnectionMessageMetrics{
			Total:       28,
			LastEvent:   time.Now().Add(-5 * time.Second),
			LastMessage: "Last sent PullTransactionInclusionState message structure",
		},
		OutPullConfirmedOutput: &model.NodeConnectionMessageMetrics{
			Total:       132,
			LastEvent:   time.Now().Add(100 * time.Second),
			LastMessage: "Last sent PullConfirmedOutput message structure",
		},
		OutPostTransaction: &model.NodeConnectionMessageMetrics{
			Total:       3,
			LastEvent:   time.Now().Add(-2 * time.Millisecond),
			LastMessage: "Last sent PostTransaction message structure",
		},
		InTransaction: &model.NodeConnectionMessageMetrics{
			Total:       101,
			LastEvent:   time.Now().Add(-8 * time.Second),
			LastMessage: "Last received Transaction message structure",
		},
		InInclusionState: &model.NodeConnectionMessageMetrics{
			Total:       203,
			LastEvent:   time.Now().Add(-123 * time.Millisecond),
			LastMessage: "Last received InclusionState message structure",
		},
		InOutput: &model.NodeConnectionMessageMetrics{
			Total:       85,
			LastEvent:   time.Now().Add(-2 * time.Second),
			LastMessage: "Last received Output message structure",
		},
		InUnspentAliasOutput: &model.NodeConnectionMessageMetrics{
			Total:       999,
			LastEvent:   time.Now().Add(-1 * time.Second),
			LastMessage: "Last received UnspentAliasOutput message structure",
		},
	}

	example := &model.NodeConnectionMetrics{
		NodeConnectionMessagesMetrics: *chainExample,
		Subscribed: []model.Address{
			model.NewAddress(iscp.RandomChainID().AsAddress()),
			model.NewAddress(iscp.RandomChainID().AsAddress()),
		},
	}

	adm.GET(routes.GetChainsNodeConnectionMetrics(), cms.handleGetChainsNodeConnMetrics).
		SetSummary("Get cummulative chains node connection metrics").
		AddResponse(http.StatusOK, "Chains Metrics", example, nil)

	adm.GET(routes.GetChainNodeConnectionMetrics(":chainID"), cms.handleGetChainNodeConnMetrics).
		SetSummary("Get chain node connection metrics for the given chain ID").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddResponse(http.StatusOK, "Chain Metrics", chainExample, nil)
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
		AddParamPath("", "chainID", "ChainID (base58)").
		AddResponse(http.StatusOK, "Chain Stats", example, nil)
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
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainID"))
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
	statusModel := model.NewConsensusWorkflowStatus(status)

	return c.JSON(http.StatusOK, statusModel)
}

func (cssT *chainMetricsService) getChain(c echo.Context) (chain.Chain, error) {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return nil, httperrors.BadRequest(err.Error())
	}
	theChain := cssT.chains().Get(chainID)
	if theChain == nil {
		return nil, httperrors.NotFound(fmt.Sprintf("Active chain %s not found", chainID))
	}
	return theChain, nil
}
