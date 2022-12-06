package node

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"

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
	publicAPI.GET("node/version", c.getPublicInfo).
		AddResponse(http.StatusOK, "Returns the version of the node.", "", nil).
		SetOperationId("getVersion").
		SetSummary("Returns the node version.")
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("node/info", c.getInfo).
		AddResponse(http.StatusOK, "Returns information about this node.", mocker.Get(models.InfoResponse{}), nil).
		SetOperationId("getInfo").
		SetSummary("Returns private information about this node.")

	adminAPI.GET("node/peers/trusted", c.getTrustedPeers, authentication.ValidatePermissions([]string{permissions.PeeringRead})).
		AddResponse(http.StatusOK, "A list of trusted peers", mocker.Get([]models.PeeringNodeIdentityResponse{}), nil).
		SetSummary("Get trusted peers").
		SetOperationId("getTrustedPeers")

	adminAPI.DELETE("node/peers/trusted", c.distrustPeer, authentication.ValidatePermissions([]string{permissions.PeeringWrite})).
		AddParamBody(mocker.Get(models.PeeringTrustRequest{}), "", "Info of the peer to distrust", true).
		AddResponse(http.StatusOK, "Peer was successfully distrusted", nil, nil).
		SetSummary("Distrust a peering node").
		SetOperationId("distrustPeer")

	adminAPI.POST("node/owner/certificate", c.setNodeOwner, authentication.ValidatePermissions([]string{permissions.NodeWrite})).
		AddParamBody(mocker.Get(models.NodeOwnerCertificateRequest{}), "", "The node owner certificate", true).
		AddResponse(http.StatusOK, "Node owner was successfully changed", nil, nil).
		SetSummary("Sets the node owner").
		SetOperationId("setNodeOwner")

	adminAPI.POST("node/peers/trusted", c.trustPeer, authentication.ValidatePermissions([]string{permissions.PeeringWrite})).
		AddParamBody(mocker.Get(models.PeeringTrustRequest{}), "", "Info of the peer to trust", true).
		AddResponse(http.StatusOK, "Peer was successfully trusted", nil, nil).
		SetSummary("Trust a peering node").
		SetOperationId("trustPeer")

	adminAPI.POST("node/dks", c.generateDKS, authentication.ValidatePermissions([]string{permissions.PeeringWrite})).
		AddParamBody(mocker.Get(models.DKSharesPostRequest{}), "DKSharesPostRequest", "Request parameters", true).
		AddResponse(http.StatusOK, "DK shares info", mocker.Get(models.DKSharesPostRequest{}), nil).
		SetSummary("Generate a new distributed key").
		SetOperationId("generateDKS")

	adminAPI.GET("node/dks/:sharedAddress", c.getDKSInfo, authentication.ValidatePermissions([]string{permissions.NodeRead})).
		AddParamPath("", "sharedAddress", "SharedAddress (Bech32)").
		AddResponse(http.StatusOK, "DK shares info", mocker.Get(models.DKSharesInfo{}), nil).
		SetSummary("Get information about the shared address DKS configuration").
		SetOperationId("getDKSInfo")

	adminAPI.GET("node/peers/identity", c.getIdentity, authentication.ValidatePermissions([]string{permissions.NodeRead})).
		AddResponse(http.StatusOK, "This node peering identity", mocker.Get(models.PeeringNodeIdentityResponse{}), nil).
		SetSummary("Get basic peer info of the current node").
		SetOperationId("getPeeringIdentity")

	adminAPI.GET("node/peers", c.getRegisteredPeers, authentication.ValidatePermissions([]string{permissions.PeeringRead})).
		AddResponse(http.StatusOK, "A list of all peers", mocker.Get([]models.PeeringNodeStatusResponse{}), nil).
		SetSummary("Get basic information about all configured peers").
		SetOperationId("getAllPeers")

	adminAPI.POST("node/shutdown", c.shutdownNode, authentication.ValidatePermissions([]string{permissions.NodeWrite})).
		AddResponse(http.StatusOK, "The node has been shut down", nil, nil).
		SetSummary("Shut down the node").
		SetOperationId("shutdownNode")

	adminAPI.GET("node/config", c.getConfiguration, authentication.ValidatePermissions([]string{permissions.NodeRead})).
		AddResponse(http.StatusOK, "Dumped configuration", map[string]interface{}{}, nil).
		SetOperationId("getConfiguration").
		SetSummary("Return the Wasp configuration")
}
