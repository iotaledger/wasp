// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) NodeOwnershipCertificate(nodePubKey ed25519.PublicKey, ownerAddress ledgerstate.Address) (governance.NodeOwnershipCertificate, error) {
	req := model.NodeOwnerCertificateRequest{
		NodePubKey:   model.NewBytes(nodePubKey.Bytes()),
		OwnerAddress: model.NewAddress(ownerAddress),
	}
	res := model.NodeOwnerCertificateResponse{}
	if err := c.do(http.MethodPost, routes.AdmNodeOwnerCertificate(), req, &res); err != nil {
		return nil, err
	}
	return governance.NewNodeOwnershipCertificateFromBytes(res.Certificate.Bytes()), nil
}
