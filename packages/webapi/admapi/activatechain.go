// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package admapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

type chainWebAPI struct {
	registry   registry.Provider
	chains     chains.Provider
	network    peering.NetworkProvider
	allMetrics *metrics.Metrics
	w          *wal.WAL
}

func addChainEndpoints(adm echoswagger.ApiGroup, c *chainWebAPI) {
	adm.POST(routes.ActivateChain(":chainID"), c.handleActivateChain).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Activate a chain")

	adm.POST(routes.DeactivateChain(":chainID"), c.handleDeactivateChain).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Deactivate a chain")

	adm.GET(routes.GetChainInfo(":chainID"), c.handleGetChainInfo).
		AddParamPath("", "chainID", "ChainID (string)").
		SetSummary("Get basic chain info.")
}

func (w *chainWebAPI) handleActivateChain(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return err
	}
	rec, err := w.registry().ActivateChainRecord(chainID)
	if err != nil {
		return err
	}

	log.Debugw("calling Chains.Activate", "chainID", rec.ChainID.String())
	if err := w.chains().Activate(rec, w.registry, w.allMetrics, w.w); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (w *chainWebAPI) handleDeactivateChain(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return err
	}
	bd, err := w.registry().DeactivateChainRecord(chainID)
	if err != nil {
		return err
	}

	err = w.chains().Deactivate(bd)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (w *chainWebAPI) handleGetChainInfo(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid chain id: %s", c.Param("chainID")))
	}

	chainRecord, err := w.registry().GetChainRecordByChainID(chainID)
	if err != nil {
		return err
	}
	if chainRecord == nil {
		return httperrors.NotFound("")
	}
	chain := w.chains().Get(chainID, true)
	committeeInfo := chain.GetCommitteeInfo()
	dkShare, err := w.registry().LoadDKShare(committeeInfo.Address)
	if err != nil {
		return err
	}

	chainNodes := chain.GetChainNodes()
	peeringStatus := peeringStatusIncludeSelf(w.network)
	candidateNodes := make(map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo)
	for _, n := range chain.GetCandidateNodes() {
		pubKey, err := cryptolib.NewPublicKeyFromBytes(n.NodePubKey)
		if err != nil {
			return err
		}
		candidateNodes[pubKey.AsKey()] = n
	}

	inChainNodes := make(map[cryptolib.PublicKeyKey]bool)

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
		ChainID:        model.ChainID(chainID.String()),
		Active:         chainRecord.Active,
		StateAddress:   model.NewAddress(committeeInfo.Address),
		CommitteeNodes: cmtNodes,
		AccessNodes:    acnNodes,
		CandidateNodes: cndNodes,
	}

	return c.JSON(http.StatusOK, res)
}

func peeringStatusIncludeSelf(networkProvider peering.NetworkProvider) map[cryptolib.PublicKeyKey]peering.PeerStatusProvider {
	peeringStatus := make(map[cryptolib.PublicKeyKey]peering.PeerStatusProvider)
	for _, n := range networkProvider.PeerStatus() {
		peeringStatus[n.PubKey().AsKey()] = n
	}
	peeringStatus[networkProvider.Self().PubKey().AsKey()] = networkProvider.Self().Status()
	return peeringStatus
}

func makeCmtNodes(
	dkShare tcrypto.DKShare,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
	inChainNodes map[cryptolib.PublicKeyKey]bool,
) []*model.ChainNodeStatus {
	cmtNodes := make([]*model.ChainNodeStatus, 0)
	for _, cmtNodePubKey := range dkShare.GetNodePubKeys() {
		cmtNodes = append(cmtNodes, makeChainNodeStatus(cmtNodePubKey, peeringStatus, candidateNodes))
		inChainNodes[cmtNodePubKey.AsKey()] = true
	}
	return cmtNodes
}

func makeAcnNodes(
	dkShare tcrypto.DKShare,
	chainNodes []peering.PeerStatusProvider,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
	inChainNodes map[cryptolib.PublicKeyKey]bool,
) []*model.ChainNodeStatus {
	acnNodes := make([]*model.ChainNodeStatus, 0)
	for _, chainNode := range chainNodes {
		acnPubKey := chainNode.PubKey()
		skip := false
		for _, cmtNodePubKey := range dkShare.GetNodePubKeys() {
			if acnPubKey.AsKey() == cmtNodePubKey.AsKey() {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		acnNodes = append(acnNodes, makeChainNodeStatus(acnPubKey, peeringStatus, candidateNodes))
		inChainNodes[acnPubKey.AsKey()] = true
	}
	return acnNodes
}

func makeCndNodes(
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
	inChainNodes map[cryptolib.PublicKeyKey]bool,
) ([]*model.ChainNodeStatus, error) {
	cndNodes := make([]*model.ChainNodeStatus, 0)
	for _, c := range candidateNodes {
		pubKey, err := cryptolib.NewPublicKeyFromBytes(c.NodePubKey)
		if err != nil {
			return nil, err
		}
		if _, ok := inChainNodes[pubKey.AsKey()]; ok {
			continue // Only include unused candidates here.
		}
		cndNodes = append(cndNodes, makeChainNodeStatus(pubKey, peeringStatus, candidateNodes))
	}
	return cndNodes, nil
}

func makeChainNodeStatus(
	pubKey *cryptolib.PublicKey,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
) *model.ChainNodeStatus {
	cns := model.ChainNodeStatus{
		Node: model.PeeringNodeStatus{
			PubKey: pubKey.String(),
		},
	}
	if n, ok := peeringStatus[pubKey.AsKey()]; ok {
		cns.Node.NetID = n.NetID()
		cns.Node.IsAlive = n.IsAlive()
		cns.Node.NumUsers = n.NumUsers()
	}
	if n, ok := candidateNodes[pubKey.AsKey()]; ok {
		cns.ForCommittee = n.ForCommittee
		cns.ForAccess = true
		cns.AccessAPI = n.AccessAPI
	}
	return &cns
}
