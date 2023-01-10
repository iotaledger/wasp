// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/registry"
)

type ChainRecord struct {
	ChainID     ChainIDBech32 `json:"chainId" swagger:"desc(ChainID (bech32))"`
	Active      bool          `json:"active" swagger:"desc(Whether or not the chain is active)"`
	AccessNodes []string      `json:"accessNodes" swagger:"desc(list of access nodes public keys, hex encoded)"`
}

func NewChainRecord(rec *registry.ChainRecord) *ChainRecord {
	chainID := rec.ChainID()
	return &ChainRecord{
		ChainID: NewChainIDBech32(chainID),
		Active:  rec.Active,
		AccessNodes: lo.Map(rec.AccessNodes, func(accessNode *cryptolib.PublicKey, _ int) string {
			return accessNode.String()
		}),
	}
}

func (bd *ChainRecord) Record() (*registry.ChainRecord, error) {
	accessNodes := make([]*cryptolib.PublicKey, len(bd.AccessNodes))

	for i, pubKeyStr := range bd.AccessNodes {
		pubKey, err := cryptolib.NewPublicKeyFromString(pubKeyStr)
		if err != nil {
			return nil, err
		}
		accessNodes[i] = pubKey
	}
	rec := registry.NewChainRecord(bd.ChainID.ChainID(), bd.Active, accessNodes)
	return rec, nil
}
