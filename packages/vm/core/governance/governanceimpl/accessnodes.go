// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This file provides implementation for the governance SC, the ChainNode
// management functions.
//
// State of the SC (the ChainNodes part):
//
//	VarAccessNodeCandidates:  map[pubKey] => AccessNodeInfo    // A set of Access Node Info.
//	VarAccessNodes:           map[pubKey] => byte[0]           // A set of nodes.
//	VarValidatorNodes:        pubKey[]                         // An ordered list of nodes.
package governanceimpl

import (
	"encoding/base64"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

var errSenderMustHaveL1Address = coreerrors.Register("sender must have L1 address").Create()

func getValidatorAddressAndVerify(
	ctx isc.Sandbox,
	pubKey *cryptolib.PublicKey,
	cert []byte,
) *cryptolib.Address {
	validatorAddr, _ := isc.AddressFromAgentID(ctx.Request().SenderAccount()) // Not from params, to have it validated.
	if validatorAddr == nil {
		panic(errSenderMustHaveL1Address)
	}
	governance.NodeOwnershipCertificateFromBytes(cert).Verify(pubKey, validatorAddr)
	return validatorAddr
}

// Can only be invoked by the access node owner (verified via the Certificate field).
func addCandidateNode(
	ctx isc.Sandbox,
	pubKey *cryptolib.PublicKey,
	cert []byte,
	accessAPI string,
	forCommittee bool,
) {
	validatorAddress := getValidatorAddressAndVerify(ctx, pubKey, cert)
	ani := &governance.AccessNodeInfo{
		NodePubKey: pubKey,
		AccessNodeData: governance.AccessNodeData{
			ValidatorAddr: validatorAddress,
			Certificate:   cert,
			ForCommittee:  forCommittee,
			AccessAPI:     accessAPI,
		},
	}
	pubKeyStr := base64.StdEncoding.EncodeToString(ani.NodePubKey.Bytes())

	state := governance.NewStateWriterFromSandbox(ctx)
	state.AccessNodeCandidatesMap().SetAt(ani.NodePubKey.Bytes(), bcs.MustMarshal(&ani.AccessNodeData))
	ctx.Log().Infof("Governance::AddCandidateNode: accessNodeCandidate added, pubKey=%s", pubKeyStr)

	if ctx.ChainAdmin().Equals(ctx.Request().SenderAccount()) {
		state.AccessNodesMap().SetAt(ani.NodePubKey.Bytes(), codec.Encode(true))
		ctx.Log().Infof("Governance::AddCandidateNode: accessNode added, pubKey=%s", pubKeyStr)
	}
}

// Can only be invoked by the access node owner (verified via the Certificate field).
//
// It is possible that after executing `revokeAccessNode(...)` a node will stay
// in the list of validators, and will be absent in the candidate or an access node set.
// The node is removed from the list of access nodes immediately, but the validator rotation
// must be initiated by the chain admin explicitly.
func revokeAccessNode(
	ctx isc.Sandbox,
	pubKey *cryptolib.PublicKey,
	cert []byte,
) {
	_ = getValidatorAddressAndVerify(ctx, pubKey, cert)
	state := governance.NewStateWriterFromSandbox(ctx)
	state.AccessNodeCandidatesMap().DelAt(pubKey.Bytes())
	state.AccessNodesMap().DelAt(pubKey.Bytes())
}

// Can only be invoked by the chain admin.
func changeAccessNodes(ctx isc.Sandbox, reqs governance.ChangeAccessNodeActions) {
	ctx.RequireCallerIsChainAdmin()

	state := governance.NewStateWriterFromSandbox(ctx)
	accessNodeCandidates := state.AccessNodeCandidatesMap()
	accessNodes := state.AccessNodesMap()
	ctx.Log().Debugf("changeAccessNodes: actions len: %d", len(reqs))

	for _, req := range reqs {
		pubKey, action := req.Unpack()
		switch action {
		case governance.ChangeAccessNodeActionRemove:
			accessNodes.DelAt(pubKey.Bytes())
		case governance.ChangeAccessNodeActionAccept:
			// TODO should the list of candidates be checked? we are just adding any pubkey
			accessNodes.SetAt(pubKey.Bytes(), codec.Encode(true))
			// TODO should the node be removed from the list of candidates? // accessNodeCandidates.DelAt(pubKey)
		case governance.ChangeAccessNodeActionDrop:
			accessNodes.DelAt(pubKey.Bytes())
			accessNodeCandidates.DelAt(pubKey.Bytes())
		default:
			panic("invalid action")
		}
	}
}

func getChainNodes(ctx isc.SandboxView) (
	candidates []*governance.AccessNodeInfo,
	accessNodes []*cryptolib.PublicKey,
) {
	state := governance.NewStateReaderFromSandbox(ctx)
	state.AccessNodeCandidatesMap().Iterate(func(pubKeyBytes, accessNodeDataBytes []byte) bool {
		pubKey := lo.Must(cryptolib.PublicKeyFromBytes(pubKeyBytes))
		and := bcs.MustUnmarshal[governance.AccessNodeData](accessNodeDataBytes)
		candidates = append(candidates, &governance.AccessNodeInfo{
			NodePubKey:     pubKey,
			AccessNodeData: and,
		})
		return true
	})
	state.AccessNodesMap().IterateKeys(func(pubKeyBytes []byte) bool {
		pubKey := lo.Must(cryptolib.PublicKeyFromBytes(pubKeyBytes))
		accessNodes = append(accessNodes, pubKey)
		return true
	})
	return
}
