package node

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/services"

	"github.com/iotaledger/hive.go/core/configuration"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	config *configuration.Configuration

	dkgService     *services.DKGService
	nodeService    interfaces.NodeService
	peeringService interfaces.PeeringService
}

func NewNodeController(log *loggerpkg.Logger, config *configuration.Configuration, dkgService *services.DKGService, nodeService interfaces.NodeService, peeringService interfaces.PeeringService) interfaces.APIController {
	return &Controller{
		log:            log,
		config:         config,
		dkgService:     dkgService,
		nodeService:    nodeService,
		peeringService: peeringService,
	}
}

func (c *Controller) Name() string {
	return "node"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	publicAPI.GET("node/info", c.getInfo).
		AddResponse(http.StatusOK, "Returns public information about this node.", nil, nil).
		SetOperationId("getInfo").
		SetSummary("Returns public information about this node.")
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("node/peers/trusted", c.getTrustedPeers).
		AddResponse(http.StatusOK, "A list of trusted peers", mocker.Get([]models.PeeringNodeIdentityResponse{}), nil).
		SetSummary("Get trusted peers").
		SetOperationId("getTrustedPeers")

	adminAPI.DELETE("node/peers/trusted", c.distrustPeer).
		AddParamBody(mocker.Get(models.PeeringTrustRequest{}), "body", "Info of the peer to distrust", true).
		AddResponse(http.StatusOK, "Peer was successfully distrusted", nil, nil).
		SetSummary("Distrust a peering node").
		SetOperationId("distrustPeer")

	adminAPI.POST("node/peers/trusted", c.trustPeer).
		AddParamBody(mocker.Get(models.PeeringTrustRequest{}), "body", "Info of the peer to trust", true).
		AddResponse(http.StatusOK, "Peer was successfully trusted", nil, nil).
		SetSummary("Trust a peering node").
		SetOperationId("trustPeer")

	adminAPI.POST("node/dks", c.generateDKS).
		AddParamBody(mocker.Get(models.DKSharesPostRequest{}), "DKSharesPostRequest", "Request parameters", true).
		AddResponse(http.StatusOK, "DK shares info", mocker.Get(models.DKSharesPostRequest{}), nil).
		SetSummary("Generate a new distributed key").
		SetOperationId("generateDKS")

	adminAPI.GET("node/dks/:sharedAddress", c.getDKSInfo).
		AddParamPath("", "sharedAddress", "SharedAddress (Bech32)").
		AddResponse(http.StatusOK, "DK shares info", mocker.Get(models.DKSharesInfo{}), nil).
		SetSummary("Get information about the shared address DKS configuration").
		SetOperationId("getDKSInfo")

	adminAPI.GET("node/peers/identity", c.getIdentity).
		AddResponse(http.StatusOK, "This node peering identity", mocker.Get(models.PeeringNodeIdentityResponse{}), nil).
		SetSummary("Get basic peer info of the current node").
		SetOperationId("getPeeringIdentity")

	adminAPI.GET("node/peers", c.getRegisteredPeers).
		AddResponse(http.StatusOK, "A list of all peers", mocker.Get([]models.PeeringNodeStatusResponse{}), nil).
		SetSummary("Get basic information about all configured peers").
		SetOperationId("getAllPeers")

	adminAPI.POST("node/shutdown", c.shutdownNode).
		AddResponse(http.StatusOK, "The node has been shut down", nil, nil).
		SetSummary("Shut down the node").
		SetOperationId("shutdownNode")

	adminAPI.GET("node/config", c.getConfiguration).
		AddResponse(http.StatusOK, "Dumped configuration", nil, nil).
		SetOperationId("getConfiguration").
		SetSummary("Return the Wasp configuration")
}
