package node

import (
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	peeringService interfaces.PeeringService
}

func NewNodeController(log *loggerpkg.Logger, peeringService interfaces.PeeringService) interfaces.APIController {
	return &Controller{
		log:            log,
		peeringService: peeringService,
	}
}

func (c *Controller) Name() string {
	return "node"
}

func (c *Controller) RegisterExampleData(mock interfaces.Mocker) {
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("peering/trusted", c.GetTrustedPeers)
	adminAPI.DELETE("peering/trusted", c.DistrustPeer)
	adminAPI.POST("peering/trusted", c.TrustPeer)

	adminAPI.GET("peering/identity", c.GetIdentity)
	adminAPI.GET("peering", c.GetRegisteredPeers)
}
