// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"bytes"
	"net/http"

	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

func addNodeOwnerEndpoints(adm echoswagger.ApiGroup, nodeIdentityProvider registry.NodeIdentityProvider, nodeOwnerAddresses []string) {
	nos := &nodeOwnerService{
		nodeOwnerAddresses: nodeOwnerAddresses,
	}
	addCtx := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("reg", nodeIdentityProvider)
			return next(c)
		}
	}
	reqExample := model.NodeOwnerCertificateRequest{
		NodePubKey:   model.NewBytes([]byte{0, 1, 17}),
		OwnerAddress: model.Address("any_address"),
	}
	resExample := model.NodeOwnerCertificateResponse{
		Certificate: model.NewBytes([]byte{0, 1, 17, 19}),
	}
	adm.POST(routes.AdmNodeOwnerCertificate(), nos.handleAdmNodeOwnerCertificate, addCtx).
		AddParamBody(reqExample, "Request", "Certificate request", true).
		AddResponse(http.StatusOK, "Generated certificate.", resExample, nil).
		SetSummary("Provides a certificate, if the node recognizes the owner.")
}

type nodeOwnerService struct {
	nodeOwnerAddresses []string
}

func (n *nodeOwnerService) handleAdmNodeOwnerCertificate(c echo.Context) error {
	nodeIdentityProvider := c.Get("reg").(registry.NodeIdentityProvider)

	var req model.NodeOwnerCertificateRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}
	reqOwnerAddress := req.OwnerAddress.Address()
	reqNodePubKeyBytes := req.NodePubKey.Bytes()

	nodeIdentity := nodeIdentityProvider.NodeIdentity()

	//
	// Check, if supplied node PubKey matches.
	if !bytes.Equal(nodeIdentity.GetPublicKey().AsBytes(), reqNodePubKeyBytes) {
		return &httperrors.HTTPError{Code: 400, Message: "Wrong NodePubKey"}
	}

	//
	// Check, if owner is presented in the configuration.
	ownerAuthorized := false
	for _, nodeOwnerAddressStr := range n.nodeOwnerAddresses {
		_, nodeOwnerAddress, err := iotago.ParseBech32(nodeOwnerAddressStr)
		if err != nil {
			continue
		}
		if bytes.Equal(isc.BytesFromAddress(reqOwnerAddress), isc.BytesFromAddress(nodeOwnerAddress)) {
			ownerAuthorized = true
			break
		}
	}
	if !ownerAuthorized {
		return &httperrors.HTTPError{Code: 403, Message: "unauthorized"}
	}

	//
	// Create the certificate. It consists of signature only. The data is not included.
	cert := governance.NewNodeOwnershipCertificate(nodeIdentity, reqOwnerAddress)
	resp := model.NodeOwnerCertificateResponse{
		Certificate: model.NewBytes(cert.Bytes()),
	}

	return c.JSON(http.StatusOK, resp)
}
