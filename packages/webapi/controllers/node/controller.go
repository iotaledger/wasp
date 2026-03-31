package node

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/wasp/v2/packages/authentication"
	"github.com/iotaledger/wasp/v2/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/packages/webapi/params"
	"github.com/iotaledger/wasp/v2/packages/webapi/services"
)

type Controller struct {
	waspVersion    string
	config         *configuration.Configuration
	dkgService     *services.DKGService
	nodeService    interfaces.NodeService
	peeringService interfaces.PeeringService
}

func NewNodeController(waspVersion string, config *configuration.Configuration, dkgService *services.DKGService, nodeService interfaces.NodeService, peeringService interfaces.PeeringService) interfaces.APIController {
	return &Controller{
		waspVersion:    waspVersion,
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
		AddResponse(http.StatusOK, "Returns the version of the node.", mocker.Get(models.VersionResponse{}), nil).
		SetOperationId("getVersion").
		SetSummary("Returns the node version.")
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
	adminAPI.GET("node/info", c.getInfo).
		AddResponse(http.StatusOK, "Returns information about this node.", mocker.Get(models.InfoResponse{}), nil).
		SetOperationId("getInfo").
		SetSummary("Returns private information about this node.")

	adminAPI.GET("node/peers/trusted", c.getTrustedPeers, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusOK, "A list of trusted peers", mocker.Get([]models.PeeringNodeIdentityResponse{}), nil).
		SetSummary("Get trusted peers").
		SetOperationId("getTrustedPeers")

	adminAPI.DELETE("node/peers/trusted/:peer", c.distrustPeer, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamPath("", params.ParamPeer, params.DescriptionPeer).
		AddResponse(http.StatusNotFound, "Peer not found", nil, nil).
		AddResponse(http.StatusOK, "Peer was successfully distrusted", nil, nil).
		SetSummary("Distrust a peering node").
		SetOperationId("distrustPeer")

	adminAPI.GET("node/owner/certificate", c.nodeOwnerCertificate, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusOK, "Node Certificate", mocker.Get(models.NodeOwnerCertificateResponse{}), nil).
		SetSummary("Gets the node owner").
		SetOperationId("ownerCertificate")

	adminAPI.POST("node/peers/trusted", c.trustPeer, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamBody(mocker.Get(models.PeeringTrustRequest{}), "", "Info of the peer to trust", true).
		AddResponse(http.StatusOK, "Peer was successfully trusted", nil, nil).
		SetSummary("Trust a peering node").
		SetOperationId("trustPeer")

	adminAPI.POST("node/dks", c.generateDKS, authentication.ValidatePermissions([]string{permissions.Write})).
		AddParamBody(mocker.Get(models.DKSharesPostRequest{}), "DKSharesPostRequest", "Request parameters", true).
		AddResponse(http.StatusOK, "DK shares info", mocker.Get(models.DKSharesInfo{}), nil).
		SetSummary("Generate a new distributed key").
		SetOperationId("generateDKS")

	adminAPI.GET("node/dks/:sharedAddress", c.getDKSInfo, authentication.ValidatePermissions([]string{permissions.Read})).
		AddParamPath("", params.ParamSharedAddress, params.DescriptionSharedAddress).
		AddResponse(http.StatusNotFound, "Shared address not found", nil, nil).
		AddResponse(http.StatusOK, "DK shares info", mocker.Get(models.DKSharesInfo{}), nil).
		SetSummary("Get information about the shared address DKS configuration").
		SetOperationId("getDKSInfo")

	adminAPI.GET("node/peers/identity", c.getIdentity, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusOK, "This node peering identity", mocker.Get(models.PeeringNodeIdentityResponse{}), nil).
		SetSummary("Get basic peer info of the current node").
		SetOperationId("getPeeringIdentity")

	adminAPI.GET("node/peers", c.getRegisteredPeers, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusOK, "A list of all peers", mocker.Get([]models.PeeringNodeStatusResponse{}), nil).
		SetSummary("Get basic information about all configured peers").
		SetOperationId("getAllPeers")

	adminAPI.POST("node/shutdown", c.shutdownNode, authentication.ValidatePermissions([]string{permissions.Write})).
		AddResponse(http.StatusOK, "The node has been shut down", nil, nil).
		SetSummary("Shut down the node").
		SetOperationId("shutdownNode")

	fakeConfigMap := make(map[string]interface{})
	fakeConfigMap["app.checkForUpdates"] = true
	fakeConfigMap["logger.level"] = "info"

	adminAPI.GET("node/config", c.getConfiguration, authentication.ValidatePermissions([]string{permissions.Read})).
		AddResponse(http.StatusOK, "Dumped configuration", fakeConfigMap, nil).
		SetOperationId("getConfiguration").
		SetSummary("Return the Wasp configuration")
}
