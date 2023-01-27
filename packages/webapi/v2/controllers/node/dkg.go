package node

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
)

func parsePeerPublicKeys(dkgRequestModel models.DKSharesPostRequest) ([]*cryptolib.PublicKey, error) {
	if dkgRequestModel.PeerIdentities == nil || len(dkgRequestModel.PeerIdentities) == 0 {
		return nil, apierrors.InvalidPropertyError("PeerPublicKeys", errors.New("PeerPublicKeys are mandatory"))
	}

	peerPublicKeys := make([]*cryptolib.PublicKey, len(dkgRequestModel.PeerIdentities))
	invalidPeerPublicKeys := make([]string, 0)

	for i, publicKey := range dkgRequestModel.PeerIdentities {
		peerPubKey, err := cryptolib.NewPublicKeyFromString(publicKey)
		if err != nil {
			invalidPeerPublicKeys = append(invalidPeerPublicKeys, publicKey)
		}

		peerPublicKeys[i] = peerPubKey
	}

	if len(invalidPeerPublicKeys) > 0 {
		return nil, apierrors.InvalidPeerPublicKeys(invalidPeerPublicKeys)
	}

	return peerPublicKeys, nil
}

func (c *Controller) generateDKS(e echo.Context) error {
	generateDKSRequest := models.DKSharesPostRequest{}

	if err := e.Bind(&generateDKSRequest); err != nil {
		return apierrors.InvalidPropertyError("body", err)
	}

	peerPublicKeys, err := parsePeerPublicKeys(generateDKSRequest)
	if err != nil {
		return err
	}

	sharesInfo, err := c.dkgService.GenerateDistributedKey(peerPublicKeys, generateDKSRequest.Threshold, time.Duration(generateDKSRequest.TimeoutMS)*time.Millisecond)
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
