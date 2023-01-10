// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

func (c *WaspClient) NodeOwnershipCertificate(nodePubKey *cryptolib.PublicKey, ownerAddress iotago.Address) (governance.NodeOwnershipCertificate, error) {
	req := model.NodeOwnerCertificateRequest{
		NodePubKey:   model.NewBytes(nodePubKey.AsBytes()),
		OwnerAddress: model.NewAddress(ownerAddress),
	}
	res := model.NodeOwnerCertificateResponse{}
	if err := c.do(http.MethodPost, routes.AdmNodeOwnerCertificate(), req, &res); err != nil {
		return nil, err
	}
	return governance.NewNodeOwnershipCertificateFromBytes(res.Certificate.Bytes()), nil
}
