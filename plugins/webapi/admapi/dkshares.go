// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

// Endpoints for creating and getting Distributed key shares.

import (
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/dkg"
	"github.com/iotaledger/wasp/plugins/registry"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
	"go.dedis.ch/kyber/v3"
)

func addDKSharesEndpoints(adm *echo.Group) {
	adm.POST("/"+client.DKSharesPostRoute(), handleDKSharesPost)
	adm.GET("/"+client.DKSharesGetRoute(":sharedAddress"), handleDKSharesGet)
}

func handleDKSharesPost(c echo.Context) error {
	var req client.DKSharesPostRequest
	var err error

	var suite = dkg.DefaultNode().GroupSuite()

	if err = c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body.")
	}

	if req.PeerPubKeys != nil && len(req.PeerNetIDs) != len(req.PeerPubKeys) {
		return httperrors.BadRequest("Inconsistent PeerNetIDs and PeerPubKeys.")
	}

	var peerPubKeys []kyber.Point = nil
	if req.PeerPubKeys != nil {
		peerPubKeys = make([]kyber.Point, len(req.PeerPubKeys))
		for i := range req.PeerPubKeys {
			peerPubKeys[i] = suite.Point()
			if err = peerPubKeys[i].UnmarshalBinary(req.PeerPubKeys[i]); err != nil {
				return httperrors.BadRequest(fmt.Sprintf("Invalid PeerPubKeys[%v]=%v", i, req.PeerPubKeys[i]))
			}
		}
	}

	var dkShare *tcrypto.DKShare
	dkShare, err = dkg.DefaultNode().GenerateDistributedKey(
		req.PeerNetIDs,
		peerPubKeys,
		req.Threshold,
		time.Duration(req.TimeoutMS)*time.Millisecond,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	var response *client.DKSharesInfo
	if response, err = makeDKSharesInfo(dkShare); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, response)
}

func handleDKSharesGet(c echo.Context) error {
	var err error
	var dkShare *tcrypto.DKShare
	var sharedAddress address.Address
	if sharedAddress, err = address.FromBase58(c.Param("sharedAddress")); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if dkShare, err = registry.DefaultRegistry().LoadDKShare(&sharedAddress); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	var response *client.DKSharesInfo
	if response, err = makeDKSharesInfo(dkShare); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, response)
}

func makeDKSharesInfo(dkShare *tcrypto.DKShare) (*client.DKSharesInfo, error) {
	var err error

	var sharedPubKeyBin []byte
	if sharedPubKeyBin, err = dkShare.SharedPublic.MarshalBinary(); err != nil {
		return nil, err
	}

	pubKeySharesBin := make([][]byte, len(dkShare.PublicShares))
	for i := range dkShare.PublicShares {
		if pubKeySharesBin[i], err = dkShare.PublicShares[i].MarshalBinary(); err != nil {
			return nil, err
		}
	}

	return &client.DKSharesInfo{
		Address:      dkShare.Address.String(),
		SharedPubKey: sharedPubKeyBin,
		PubKeyShares: pubKeySharesBin,
		Threshold:    dkShare.T,
		PeerIndex:    dkShare.Index,
	}, nil
}
