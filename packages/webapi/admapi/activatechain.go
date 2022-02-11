// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type chainWebAPI struct {
	registry   registry.Provider
	chains     chains.Provider
	network    peering.NetworkProvider
	allMetrics *metrics.Metrics
	w          *wal.WAL
}

func addChainEndpoints(adm echoswagger.ApiGroup, registryProvider registry.Provider, chainsProvider chains.Provider, network peering.NetworkProvider, allMetrics *metrics.Metrics, w *wal.WAL) {
	c := &chainWebAPI{
		registryProvider,
		chainsProvider,
		network,
		allMetrics,
		w,
	}

	adm.POST(routes.ActivateChain(":chainID"), c.handleActivateChain).
		AddParamPath("", "chainID", "ChainID (base58)").
		SetSummary("Activate a chain")

	adm.POST(routes.DeactivateChain(":chainID"), c.handleDeactivateChain).
		AddParamPath("", "chainID", "ChainID (base58)").
		SetSummary("Deactivate a chain")

	adm.GET(routes.GetChainInfo(":chainID"), c.handleGetChainInfo).
		AddParamPath("", "chainID", "ChainID (base58)").
		SetSummary("Get basic chain info.")
}

func (w *chainWebAPI) handleActivateChain(c echo.Context) error {
	panic("TODO implement")
	// aliasAddress, err := iotago.AliasAddressFromBase58EncodedString(c.Param("chainID"))
	// if err != nil {
	// 	return httperrors.BadRequest(fmt.Sprintf("Invalid alias address: %s", c.Param("chainID")))
	// }
	// chainID, err := iscp.ChainIDFromAddress(aliasAddress)
	// if err != nil {
	// 	return err
	// }
	// rec, err := w.registry().ActivateChainRecord(chainID)
	// if err != nil {
	// 	return err
	// }

	// log.Debugw("calling Chains.Activate", "chainID", rec.ChainID.String())
	// if err := w.chains().Activate(rec, w.registry, w.allMetrics, w.w); err != nil {
	// 	return err
	// }

	// return c.NoContent(http.StatusOK)
}

func (w *chainWebAPI) handleDeactivateChain(c echo.Context) error {
	panic("TODO implement")
	// scAddress, err := iotago.AddressFromBase58EncodedString(c.Param("chainID"))
	// if err != nil {
	// 	return httperrors.BadRequest(fmt.Sprintf("Invalid chain id: %s", c.Param("chainID")))
	// }
	// chainID, err := iscp.ChainIDFromAddress(scAddress)
	// if err != nil {
	// 	return err
	// }
	// bd, err := w.registry().DeactivateChainRecord(chainID)
	// if err != nil {
	// 	return err
	// }

	// err = w.chains().Deactivate(bd)
	// if err != nil {
	// 	return err
	// }

	// return c.NoContent(http.StatusOK)
}

func (w *chainWebAPI) handleGetChainInfo(c echo.Context) error {
	scAddress, err := ledgerstate.AddressFromBase58EncodedString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain id: %s", c.Param("chainID")))
	}
	chainID, err := iscp.ChainIDFromAddress(scAddress)
	if err != nil {
		return err
	}

	chain := w.chains().Get(chainID, true)
	committeeInfo := chain.GetCommitteeInfo()
	chainRecord, err := w.registry().GetChainRecordByChainID(chainID)
	if err != nil {
		return err
	}
	dkShare, err := w.registry().LoadDKShare(committeeInfo.Address)
	if err != nil {
		return err
	}

	chainNodes := chain.GetChainNodes()
	peeringStatus := make(map[ed25519.PublicKey]peering.PeerStatusProvider)
	for _, n := range w.network.PeerStatus() {
		peeringStatus[*n.PubKey()] = n
	}
	candidateNodes := make(map[ed25519.PublicKey]*governance.AccessNodeInfo)
	for _, n := range chain.GetCandidateNodes() {
		pubKey, _, err := ed25519.PublicKeyFromBytes(n.NodePubKey)
		if err != nil {
			return err
		}
		candidateNodes[pubKey] = n
	}

	inChainNodes := make(map[ed25519.PublicKey]bool)

	//
	// Committee nodes.
	cmtNodes := makeCmtNodes(dkShare, peeringStatus, candidateNodes, inChainNodes)

	//
	// Access nodes: accepted as access nodes and not included in the committee.
	acnNodes := makeAcnNodes(dkShare, chainNodes, peeringStatus, candidateNodes, inChainNodes)

	//
	// Candidate nodes have suplied applications, but are not included
	// in the committee and to the set of the access nodes.
	cndNodes, err := makeCndNodes(peeringStatus, candidateNodes, inChainNodes)
	if err != nil {
		return err
	}

	res := model.ChainInfo{
		ChainID:        model.ChainID(chainID.Base58()),
		Active:         chainRecord.Active,
		StateAddress:   model.NewAddress(committeeInfo.Address),
		CommitteeNodes: cmtNodes,
		AccessNodes:    acnNodes,
		CandidateNodes: cndNodes,
	}

	return c.JSON(http.StatusOK, res)
}

func makeCmtNodes(
	dkShare *tcrypto.DKShare,
	peeringStatus map[ed25519.PublicKey]peering.PeerStatusProvider,
	candidateNodes map[ed25519.PublicKey]*governance.AccessNodeInfo,
	inChainNodes map[ed25519.PublicKey]bool,
) []*model.ChainNodeStatus {
	cmtNodes := make([]*model.ChainNodeStatus, 0)
	for _, cmtNodePubKey := range dkShare.NodePubKeys {
		cmtNodes = append(cmtNodes, makeChainNodeStatus(cmtNodePubKey, peeringStatus, candidateNodes))
		inChainNodes[*cmtNodePubKey] = true
	}
	return cmtNodes
}

func makeAcnNodes(
	dkShare *tcrypto.DKShare,
	chainNodes []peering.PeerStatusProvider,
	peeringStatus map[ed25519.PublicKey]peering.PeerStatusProvider,
	candidateNodes map[ed25519.PublicKey]*governance.AccessNodeInfo,
	inChainNodes map[ed25519.PublicKey]bool,
) []*model.ChainNodeStatus {
	acnNodes := make([]*model.ChainNodeStatus, 0)
	for _, chainNode := range chainNodes {
		acnPubKey := chainNode.PubKey()
		skip := false
		for _, cmtNodePubKey := range dkShare.NodePubKeys {
			if *acnPubKey == *cmtNodePubKey {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		acnNodes = append(acnNodes, makeChainNodeStatus(acnPubKey, peeringStatus, candidateNodes))
		inChainNodes[*acnPubKey] = true
	}
	return acnNodes
}

func makeCndNodes(
	peeringStatus map[ed25519.PublicKey]peering.PeerStatusProvider,
	candidateNodes map[ed25519.PublicKey]*governance.AccessNodeInfo,
	inChainNodes map[ed25519.PublicKey]bool,
) ([]*model.ChainNodeStatus, error) {
	cndNodes := make([]*model.ChainNodeStatus, 0)
	for _, c := range candidateNodes {
		pubKey, _, err := ed25519.PublicKeyFromBytes(c.NodePubKey)
		if err != nil {
			return nil, err
		}
		if _, ok := inChainNodes[pubKey]; ok {
			continue // Only include unused candidates here.
		}
		cndNodes = append(cndNodes, makeChainNodeStatus(&pubKey, peeringStatus, candidateNodes))
	}
	return cndNodes, nil
}

func makeChainNodeStatus(
	pubKey *ed25519.PublicKey,
	peeringStatus map[ed25519.PublicKey]peering.PeerStatusProvider,
	candidateNodes map[ed25519.PublicKey]*governance.AccessNodeInfo,
) *model.ChainNodeStatus {
	cns := model.ChainNodeStatus{
		Node: model.PeeringNodeStatus{
			PubKey: pubKey.String(),
		},
	}
	if n, ok := peeringStatus[*pubKey]; ok {
		cns.Node.NetID = n.NetID()
		cns.Node.IsAlive = n.IsAlive()
		cns.Node.NumUsers = n.NumUsers()
	}
	if n, ok := candidateNodes[*pubKey]; ok {
		cns.ForCommittee = n.ForCommittee
		cns.ForAccess = true
		cns.AccessAPI = n.AccessAPI
	}
	return &cns
}
