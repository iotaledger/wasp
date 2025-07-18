// Package node implements webapi node controller
package node

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/packages/webapi/params"
)

func (c *Controller) getRegisteredPeers(e echo.Context) error {
	peers := c.peeringService.GetRegisteredPeers()
	peerModels := make([]models.PeeringNodeStatusResponse, len(peers))

	for k, v := range peers {
		peerModels[k] = models.PeeringNodeStatusResponse{
			Name:       v.Name,
			IsAlive:    v.IsAlive,
			PeeringURL: v.PeeringURL,
			NumUsers:   v.NumUsers,
			PublicKey:  v.PublicKey.String(),
			IsTrusted:  v.IsTrusted,
		}
	}

	return e.JSON(http.StatusOK, peerModels)
}

func (c *Controller) getTrustedPeers(e echo.Context) error {
	peers, err := c.peeringService.GetTrustedPeers()
	if err != nil {
		panic(err)
	}

	peerModels := make([]models.PeeringNodeIdentityResponse, len(peers))
	for k, v := range peers {
		peerModels[k] = models.PeeringNodeIdentityResponse{
			Name:       v.Name,
			PeeringURL: v.PeeringURL,
			PublicKey:  v.PublicKey.String(),
			IsTrusted:  v.IsTrusted,
		}
	}

	return e.JSON(http.StatusOK, peerModels)
}

func (c *Controller) getIdentity(e echo.Context) error {
	self := c.peeringService.GetIdentity()

	peerModel := models.PeeringNodeIdentityResponse{
		PeeringURL: self.PeeringURL,
		PublicKey:  self.PublicKey.String(),
		IsTrusted:  self.IsTrusted,
	}

	return e.JSON(http.StatusOK, peerModel)
}

func (c *Controller) trustPeer(e echo.Context) error {
	var trustedPeer models.PeeringTrustRequest

	if err := e.Bind(&trustedPeer); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	publicKey, err := cryptolib.PublicKeyFromString(trustedPeer.PublicKey)
	if err != nil {
		return apierrors.InvalidPropertyError("publicKey", err)
	}

	selfPubKey := c.peeringService.GetIdentity().PublicKey
	if publicKey.Equals(selfPubKey) {
		return apierrors.SelfAsPeerError()
	}

	if !util.IsSlug(trustedPeer.Name) {
		return apierrors.InvalidPeerName()
	}
	_, err = c.peeringService.TrustPeer(trustedPeer.Name, publicKey, trustedPeer.PeeringURL)
	if err != nil {
		panic(err)
	}

	return e.NoContent(http.StatusOK)
}

func (c *Controller) distrustPeer(e echo.Context) error {
	peer := e.Param(params.ParamPeer)

	if peer == "" {
		return apierrors.InvalidPropertyError(params.ParamPeer, errors.New("invalid peer name"))
	}

	if _, err := c.peeringService.DistrustPeer(peer); err != nil {
		if errors.Is(err, interfaces.ErrPeerNotFound) {
			return apierrors.PeerNameNotFoundError(peer)
		}

		panic(err)
	}

	return e.NoContent(http.StatusOK)
}
