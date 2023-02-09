// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

// Endpoints for creating and getting Distributed key shares.

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

func addDKSharesEndpoints(adm echoswagger.ApiGroup, dkShareRegistryProvider registry.DKShareRegistryProvider, nodeProvider dkg.NodeProvider) {
	requestExample := model.DKSharesPostRequest{
		PeerPubKeys: []string{base64.StdEncoding.EncodeToString([]byte("key"))},
		Threshold:   3,
		TimeoutMS:   10000,
	}
	addr1 := isc.RandomChainID().AsAddress()
	infoExample := model.DKSharesInfo{
		Address:      addr1.Bech32(parameters.L1().Protocol.Bech32HRP),
		SharedPubKey: base64.StdEncoding.EncodeToString([]byte("key")),
		PubKeyShares: []string{base64.StdEncoding.EncodeToString([]byte("key"))},
		PeerPubKeys:  []string{base64.StdEncoding.EncodeToString([]byte("key"))},
		Threshold:    3,
		PeerIndex:    nil,
	}

	s := &dkSharesService{
		dkShareRegistryProvider: dkShareRegistryProvider,
		dkgNode:                 nodeProvider,
	}

	adm.POST(routes.DKSharesPost(), s.handleDKSharesPost).
		SetDeprecated().
		AddParamBody(requestExample, "DKSharesPostRequest", "Request parameters", true).
		AddResponse(http.StatusOK, "DK shares info", infoExample, nil).
		SetSummary("Generate a new distributed key")

	adm.GET(routes.DKSharesGet(":sharedAddress"), s.handleDKSharesGet).
		SetDeprecated().
		AddParamPath("", "sharedAddress", "Address of the DK share (hex)").
		AddResponse(http.StatusOK, "DK shares info", infoExample, nil).
		SetSummary("Get distributed key properties")
}

type dkSharesService struct {
	dkShareRegistryProvider registry.DKShareRegistryProvider
	dkgNode                 dkg.NodeProvider
}

func (s *dkSharesService) handleDKSharesPost(c echo.Context) error {
	var req model.DKSharesPostRequest
	var err error

	if err = c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body.")
	}

	if req.PeerPubKeys == nil || len(req.PeerPubKeys) < 1 {
		return httperrors.BadRequest("PeerPubKeys are mandatory")
	}

	var peerPubKeys []*cryptolib.PublicKey
	if req.PeerPubKeys != nil {
		peerPubKeys = make([]*cryptolib.PublicKey, len(req.PeerPubKeys))
		for i := range req.PeerPubKeys {
			peerPubKey, err2 := cryptolib.NewPublicKeyFromString(req.PeerPubKeys[i])
			if err2 != nil {
				return httperrors.BadRequest(fmt.Sprintf("Invalid PeerPubKeys[%v]=%v", i, req.PeerPubKeys[i]))
			}
			peerPubKeys[i] = peerPubKey
		}
	}

	dkShare, err := s.dkgNode().GenerateDistributedKey(
		peerPubKeys,
		req.Threshold,
		1*time.Second,
		3*time.Second,
		time.Duration(req.TimeoutMS)*time.Millisecond,
	)
	if err != nil {
		if _, ok := err.(dkg.InvalidParamsError); ok {
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

func (s *dkSharesService) handleDKSharesGet(c echo.Context) error {
	var err error
	var dkShare tcrypto.DKShare
	var sharedAddress iotago.Address
	if _, sharedAddress, err = iotago.ParseBech32(c.Param("sharedAddress")); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if dkShare, err = s.dkShareRegistryProvider.LoadDKShare(sharedAddress); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	var response *model.DKSharesInfo
	if response, err = makeDKSharesInfo(dkShare); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, response)
}

func makeDKSharesInfo(dkShare tcrypto.DKShare) (*model.DKSharesInfo, error) {
	var err error

	b, err := dkShare.DSSSharedPublic().MarshalBinary()
	if err != nil {
		return nil, err
	}
	sharedPubKey := base64.StdEncoding.EncodeToString(b)

	pubKeyShares := make([]string, len(dkShare.DSSPublicShares()))
	for i := range dkShare.DSSPublicShares() {
		b, err := dkShare.DSSPublicShares()[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		pubKeyShares[i] = base64.StdEncoding.EncodeToString(b)
	}

	peerPubKeys := make([]string, len(dkShare.GetNodePubKeys()))
	for i := range dkShare.GetNodePubKeys() {
		peerPubKeys[i] = base64.StdEncoding.EncodeToString(dkShare.GetNodePubKeys()[i].AsBytes())
	}

	return &model.DKSharesInfo{
		Address:      dkShare.GetAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
		SharedPubKey: sharedPubKey,
		PubKeyShares: pubKeyShares,
		PeerPubKeys:  peerPubKeys,
		Threshold:    dkShare.GetT(),
		PeerIndex:    dkShare.GetIndex(),
	}, nil
}
