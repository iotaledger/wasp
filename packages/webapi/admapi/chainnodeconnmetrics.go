package admapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addChainStatsEndpoints(adm echoswagger.ApiGroup, chainsProvider chains.Provider) {
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

	s := &chainStatsService{chainsProvider}

	adm.GET(routes.GetChainsNodeConnectionMetrics(), s.handleGetChainsStats).
		SetSummary("Get cummulative chains state statistics").
		AddResponse(http.StatusOK, "Chains Stats", example, nil)

	adm.GET(routes.GetChainNodeConnectionMetrics(":chainID"), s.handleGetChainStats).
		SetSummary("Get chain state statistics for the given chain ID").
		AddParamPath("", "chainID", "ChainID (base58)").
		AddResponse(http.StatusOK, "Chain Stats", chainExample, nil)
}

type chainStatsService struct {
	chains chains.Provider
}

func (cssT *chainStatsService) handleGetChainsStats(c echo.Context) error {
	metrics := cssT.chains().GetNodeConnectionMetrics()
	metricsModel := model.NewNodeConnectionMetrics(metrics)

	return c.JSON(http.StatusOK, metricsModel)
}

func (cssT *chainStatsService) handleGetChainStats(c echo.Context) error {
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
