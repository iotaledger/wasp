package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/routes"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/samber/lo"
)

const pubKeyParam = "PublicKey"

func addAccessNodesEndpoints(
	adm echoswagger.ApiGroup,
	registryProvider registry.ChainRecordRegistryProvider,
	tnm peering.TrustedNetworkManager,
) {
	a := &accessNodesService{
		registry:   registryProvider,
		networkMgr: tnm,
	}
	adm.POST(routes.AdmAddAccessNode(":chainID"), a.handleAddAccessNode).
		AddParamPath("", "chainID", "ChainID (bech32))").
		AddParamPath("", pubKeyParam, "PublicKey (hex string)").
		SetSummary("Add an access node to a chain")

	adm.POST(routes.AdmRemoveAccessNode(":chainID"), a.handleRemoveAccessNode).
		AddParamPath("", "chainID", "ChainID (bech32))").
		AddParamPath("", pubKeyParam, "PublicKey (hex string)").
		SetSummary("Remove an access node from a chain")
}

type accessNodesService struct {
	registry   registry.ChainRecordRegistryProvider
	networkMgr peering.TrustedNetworkManager
}

func paramsPubKey(c echo.Context) (*cryptolib.PublicKey, error) {
	return cryptolib.NewPublicKeyFromHexString(c.Param(pubKeyParam))
}

func (a *accessNodesService) handleAddAccessNode(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest("invalid chainID")
	}
	pubKey, err := paramsPubKey(c)
	if err != nil {
		return httperrors.BadRequest("invalid pub key")
	}
	peers, err := a.networkMgr.TrustedPeers()
	if err != nil {
		return httperrors.ServerError("error getting trusted peers")
	}
	_, ok := lo.Find(peers, func(p *peering.TrustedPeer) bool {
		return p.PubKey().Equals(pubKey)
	})
	if !ok {
		return httperrors.NotFound(fmt.Sprintf("couldn't find peer with public key %s", pubKey))
	}
	_, err = a.registry.UpdateChainRecord(*chainID, func(rec *registry.ChainRecord) bool {
		rec.AddAccessNode(pubKey)
		// TODO what should this return?
		return false
	})
	if err != nil {
		return httperrors.ServerError("error saving chain record.")
	}
	return c.NoContent(http.StatusOK)
}

func (a *accessNodesService) handleRemoveAccessNode(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest("invalid chainID")
	}
	pubKey, err := paramsPubKey(c)
	if err != nil {
		return httperrors.BadRequest("invalid pub key")
	}
	_, err = a.registry.UpdateChainRecord(*chainID, func(rec *registry.ChainRecord) bool {
		return rec.RemoveAccessNode(pubKey)
	})
	if err != nil {
		return httperrors.ServerError("error saving chain record.")
	}
	return c.NoContent(http.StatusOK)
}
