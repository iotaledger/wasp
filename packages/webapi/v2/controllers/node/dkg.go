package node

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
)

func parsePeerPubKeys(dkgRequestModel models.DKSharesPostRequest) ([]*cryptolib.PublicKey, error) {
	if dkgRequestModel.PeerPubKeys == nil || len(dkgRequestModel.PeerPubKeys) == 0 {
		return nil, apierrors.InvalidPropertyError("PeerPubKeys", errors.New("PeerPubKeys are mandatory"))
	}

	peerPubKeys := make([]*cryptolib.PublicKey, len(dkgRequestModel.PeerPubKeys))
	invalidPeerPubKeys := make([]string, 0)

	for i, publicKey := range dkgRequestModel.PeerPubKeys {
		peerPubKey, err := cryptolib.NewPublicKeyFromString(publicKey)

		if err != nil {
			invalidPeerPubKeys = append(invalidPeerPubKeys, publicKey)
		}

		peerPubKeys[i] = peerPubKey
	}

	if len(invalidPeerPubKeys) > 0 {
		return nil, apierrors.InvalidPeerPublicKeys(invalidPeerPubKeys)
	}

	return peerPubKeys, nil
}

func (c *Controller) generateDKS(e echo.Context) error {
	generateDKSRequest := models.DKSharesPostRequest{}

	if err := e.Bind(&generateDKSRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	peerPublicKeys, err := parsePeerPubKeys(generateDKSRequest)
	if err != nil {
		return err
	}

	sharesInfo, err := c.dkgService.GenerateDistributedKey(peerPublicKeys, generateDKSRequest.Threshold, generateDKSRequest.TimeoutMS)
	if err != nil {
		return apierrors.InternalServerError(err)
	}

	return e.JSON(http.StatusOK, sharesInfo)
}

func (c *Controller) getDKSInfo(e echo.Context) error {
	_, sharedAddress, err := iotago.ParseBech32(e.Param("sharedAddress"))
	if err != nil {
		return apierrors.InvalidPropertyError("sharedAddress", err)
	}

	sharesInfo, err := c.dkgService.GetShares(sharedAddress)
	if err != nil {
		return apierrors.InternalServerError(err)
	}

	return e.JSON(http.StatusOK, sharesInfo)
}
