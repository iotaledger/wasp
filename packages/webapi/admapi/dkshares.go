// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

// Endpoints for creating and getting Distributed key shares.

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

func addDKSharesEndpoints(adm echoswagger.ApiGroup, registryProvider registry.Provider, nodeProvider dkg.NodeProvider) {
	requestExample := model.DKSharesPostRequest{
		PeerPubKeys: []string{base64.StdEncoding.EncodeToString([]byte("key"))},
		Threshold:   3,
		TimeoutMS:   10000,
	}
	addr1 := iscp.RandomChainID().AsAddress()
	infoExample := model.DKSharesInfo{
		Address:      addr1.Base58(),
		SharedPubKey: base64.StdEncoding.EncodeToString([]byte("key")),
		PubKeyShares: []string{base64.StdEncoding.EncodeToString([]byte("key"))},
		PeerPubKeys:  []string{base64.StdEncoding.EncodeToString([]byte("key"))},
		Threshold:    3,
		PeerIndex:    nil,
	}

	s := &dkSharesService{registry: registryProvider, dkgNode: nodeProvider}

	adm.POST(routes.DKSharesPost(), s.handleDKSharesPost).
		AddParamBody(requestExample, "DKSharesPostRequest", "Request parameters", true).
		AddResponse(http.StatusOK, "DK shares info", infoExample, nil).
		SetSummary("Generate a new distributed key")

	adm.GET(routes.DKSharesGet(":sharedAddress"), s.handleDKSharesGet).
		AddParamPath("", "sharedAddress", "Address of the DK share (base58)").
		AddResponse(http.StatusOK, "DK shares info", infoExample, nil).
		SetSummary("Get distributed key properties")
}

type dkSharesService struct {
	registry registry.Provider
	dkgNode  dkg.NodeProvider
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

	var peerPubKeys []*ed25519.PublicKey
	if req.PeerPubKeys != nil {
		peerPubKeys = make([]*ed25519.PublicKey, len(req.PeerPubKeys))
		for i := range req.PeerPubKeys {
			peerPubKey, err := ed25519.PublicKeyFromString(req.PeerPubKeys[i])
			if err != nil {
				return httperrors.BadRequest(fmt.Sprintf("Invalid PeerPubKeys[%v]=%v", i, req.PeerPubKeys[i]))
			}
			peerPubKeys[i] = &peerPubKey
		}
	}

	var dkShare *tcrypto.DKShare
	dkShare, err = s.dkgNode().GenerateDistributedKey(
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
	var dkShare *tcrypto.DKShare
	var sharedAddress ledgerstate.Address
	if sharedAddress, err = ledgerstate.AddressFromBase58EncodedString(c.Param("sharedAddress")); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if dkShare, err = s.registry().LoadDKShare(sharedAddress); err != nil {
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

	peerPubKeys := make([]string, len(dkShare.NodePubKeys))
	for i := range dkShare.NodePubKeys {
		peerPubKeys[i] = base64.StdEncoding.EncodeToString(dkShare.NodePubKeys[i].Bytes())
	}

	return &model.DKSharesInfo{
		Address:      dkShare.Address.Base58(),
		SharedPubKey: sharedPubKey,
		PubKeyShares: pubKeyShares,
		PeerPubKeys:  peerPubKeys,
		Threshold:    dkShare.T,
		PeerIndex:    dkShare.Index,
	}, nil
}
