// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

// Endpoints for creating and getting Distributed key shares.

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	dkg_pkg "github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/plugins/dkg"
	"github.com/iotaledger/wasp/plugins/registry"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"go.dedis.ch/kyber/v3"
)

func addDKSharesEndpoints(adm echoswagger.ApiGroup) {
	requestExample := model.DKSharesPostRequest{
		PeerNetIDs:  []string{"wasp1:4000", "wasp2:4000", "wasp3:4000", "wasp4:4000"},
		PeerPubKeys: []string{base64.StdEncoding.EncodeToString([]byte("key"))},
		Threshold:   3,
		TimeoutMS:   10000,
	}
	addr1 := chainid.RandomChainID().AsAddress()
	infoExample := model.DKSharesInfo{
		Address:      addr1.Base58(),
		SharedPubKey: base64.StdEncoding.EncodeToString([]byte("key")),
		PubKeyShares: []string{base64.StdEncoding.EncodeToString([]byte("key"))},
		Threshold:    3,
		PeerIndex:    nil,
	}

	adm.POST(routes.DKSharesPost(), handleDKSharesPost).
		AddParamBody(requestExample, "DKSharesPostRequest", "Request parameters", true).
		AddResponse(http.StatusOK, "DK shares info", infoExample, nil).
		SetSummary("Generate a new distributed key")

	adm.GET(routes.DKSharesGet(":sharedAddress"), handleDKSharesGet).
		AddParamPath("", "sharedAddress", "Address of the DK share (base58)").
		AddResponse(http.StatusOK, "DK shares info", infoExample, nil).
		SetSummary("Get distributed key properties")
}

func handleDKSharesPost(c echo.Context) error {
	var req model.DKSharesPostRequest
	var err error

	suite := dkg.DefaultNode().GroupSuite()

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
			b, err := base64.StdEncoding.DecodeString(req.PeerPubKeys[i])
			if err != nil {
				return httperrors.BadRequest(fmt.Sprintf("Invalid PeerPubKeys[%v]=%v", i, req.PeerPubKeys[i]))
			}
			if err = peerPubKeys[i].UnmarshalBinary(b); err != nil {
				return httperrors.BadRequest(fmt.Sprintf("Invalid PeerPubKeys[%v]=%v", i, req.PeerPubKeys[i]))
			}
		}
	}

	var dkShare *tcrypto.DKShare
	dkShare, err = dkg.DefaultNode().GenerateDistributedKey(
		req.PeerNetIDs,
		peerPubKeys,
		req.Threshold,
		1*time.Second,
		3*time.Second,
		time.Duration(req.TimeoutMS)*time.Millisecond,
	)
	if err != nil {
		if _, ok := err.(dkg_pkg.InvalidParamsError); ok {
			return httperrors.BadRequest(err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	var response *model.DKSharesInfo
	if response, err = makeDKSharesInfo(dkShare); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, response)
}

func handleDKSharesGet(c echo.Context) error {
	var err error
	var dkShare *tcrypto.DKShare
	var sharedAddress ledgerstate.Address
	if sharedAddress, err = ledgerstate.AddressFromBase58EncodedString(c.Param("sharedAddress")); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if dkShare, err = registry.DefaultRegistry().LoadDKShare(sharedAddress); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	var response *model.DKSharesInfo
	if response, err = makeDKSharesInfo(dkShare); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, response)
}

func makeDKSharesInfo(dkShare *tcrypto.DKShare) (*model.DKSharesInfo, error) {
	var err error

	b, err := dkShare.SharedPublic.MarshalBinary()
	if err != nil {
		return nil, err
	}
	sharedPubKey := base64.StdEncoding.EncodeToString(b)

	pubKeyShares := make([]string, len(dkShare.PublicShares))
	for i := range dkShare.PublicShares {
		b, err := dkShare.PublicShares[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		pubKeyShares[i] = base64.StdEncoding.EncodeToString(b)
	}

	return &model.DKSharesInfo{
		Address:      dkShare.Address.Base58(),
		SharedPubKey: sharedPubKey,
		PubKeyShares: pubKeyShares,
		Threshold:    dkShare.T,
		PeerIndex:    dkShare.Index,
	}, nil
}
