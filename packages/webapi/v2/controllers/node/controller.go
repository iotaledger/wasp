package node

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	nodeService    interfaces.NodeService
	peeringService interfaces.PeeringService
}

func NewNodeController(log *loggerpkg.Logger, peeringService interfaces.PeeringService, nodeService interfaces.NodeService) interfaces.APIController {
	return &Controller{
		log:            log,
		nodeService:    nodeService,
		peeringService: peeringService,
	}
}

func (c *Controller) Name() string {
	return "node"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("node/peers/trusted", c.GetTrustedPeers).
		AddResponse(http.StatusOK, "A list of trusted peers.", mocker.Get([]models.PeeringNodeIdentity{}), nil).
		SetOperationId("getTrustedPeers")

	adminAPI.DELETE("node/peers/trusted", c.DistrustPeer).
		AddParamBody(mocker.Get(models.PeeringTrustRequest{}), "body", "Info of the peer to distrust.", true).
		AddResponse(http.StatusOK, "Peer was successfully distrusted", nil, nil).
		SetSummary("Distrusts a peering node.").
		SetOperationId("distrustPeer")

	adminAPI.POST("node/peers/trusted", c.TrustPeer).
		AddParamBody(mocker.Get(models.PeeringTrustRequest{}), "body", "Info of the peer to trust.", true).
		AddResponse(http.StatusOK, "Peer was successfully trusted", nil, nil).
		SetSummary("Trusts a peering node.").
		SetOperationId("trustPeer")

	adminAPI.GET("node/peers/identity", c.GetIdentity).
		AddResponse(http.StatusOK, "This node as a peer.", mocker.Get(models.PeeringNodeIdentity{}), nil).
		SetSummary("Basic peer info of the current node.").
		SetOperationId("getPeeringIdentity")

	adminAPI.GET("node/peers", c.GetRegisteredPeers).
		AddResponse(http.StatusOK, "A list of all peers.", mocker.Get([]models.PeeringNodeStatus{}), nil).
		SetSummary("Basic information about all configured peers.").
		SetOperationId("getAllPeers")

	adminAPI.POST("node/shutdown", c.ShutdownNode).
		AddResponse(http.StatusOK, "The node has been shut down", nil, nil).
		SetSummary("Shuts down the node").
		SetOperationId("shutdownNode")

}
