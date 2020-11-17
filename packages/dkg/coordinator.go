package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"go.dedis.ch/kyber/v3"
)

// GenerateDistributedKey is called from the client node to initiate the DKG
// procedure on a set of nodes. The client is not required to have an instance
// of the DkgNode, but may have it (be one of the peers sharing the secret).
// This function works synchronously, so the user should run it async if needed.
func GenerateDistributedKey(
	coordKey kyber.Scalar,
	coordPub kyber.Point,
	peerLocs []string,
	peerPubs []kyber.Point,
	treshold uint32,
	version byte, // address.VersionED25519 = 1 | address.VersionBLS = 2
	timeout time.Duration,
	suite kyber.Group,
	netProvider CoordNodeProvider,
) (*coretypes.ChainID, *kyber.Point, error) {
	var err error
	var dkgID string = address.Random().String()
	//
	// Initialize the peers.
	var peerPubsBytes [][]byte
	if peerPubsBytes, err = pubsToBytes(peerPubs); err != nil {
		return nil, nil, err
	}
	var coordPubBytes []byte
	if coordPubBytes, err = pubToBytes(coordPub); err != nil {
		return nil, nil, err
	}
	initReq := InitReq{
		PeerLocs:  peerLocs,
		PeerPubs:  peerPubsBytes,
		CoordPub:  coordPubBytes,
		Treshold:  treshold,
		Version:   version,
		TimeoutMS: uint64(timeout.Milliseconds()),
	}
	if err = netProvider.DkgInit(peerLocs, dkgID, &initReq); err != nil {
		return nil, nil, err
	}
	//
	// Perform the DKG steps, each step in parallel, all steps sequentially.
	// Step numbering (R) is according to <https://github.com/dedis/kyber/blob/master/share/dkg/rabin/dkg.go>.
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "1-R2.1-SendDeals"}); err != nil {
		return nil, nil, err
	}
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "2-R2.2-SendResponses"}); err != nil {
		return nil, nil, err
	}
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "3-R2.3-SendJustifications"}); err != nil {
		return nil, nil, err
	}
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "4-R4-SendSecretCommits"}); err != nil {
		return nil, nil, err
	}
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "5-R5-SendComplaintCommits"}); err != nil {
		return nil, nil, err
	}
	//
	// Now get the public keys.
	// This also performs the "6-R6-SendReconstructCommits" step implicitly.
	var pubKeyResponses []*PubKeyResp
	if pubKeyResponses, err = netProvider.DkgPubKey(peerLocs, dkgID); err != nil {
		return nil, nil, err
	}
	chainIDBytes := pubKeyResponses[0].ChainID
	pubKeyBytes := pubKeyResponses[0].PubKey
	signatures := [][]byte{}
	for i := range pubKeyResponses {
		if !bytes.Equal(pubKeyResponses[i].ChainID, chainIDBytes) {
			return nil, nil, fmt.Errorf("nodes generated different addresses")
		}
		if !bytes.Equal(pubKeyResponses[i].PubKey, pubKeyBytes) {
			return nil, nil, fmt.Errorf("nodes generated different public keys")
		}
		signatures = append(signatures, pubKeyResponses[i].Signature)
	}
	var generatedChainID coretypes.ChainID
	if generatedChainID, err = coretypes.NewChainIDFromBytes(chainIDBytes); err != nil {
		return nil, nil, err
	}
	sharedPublic := suite.Point()
	sharedPublic.UnmarshalBinary(pubKeyBytes)
	//
	// Verify signatures.
	switch version {
	case address.VersionED25519:
		// TODO
	case address.VersionBLS:
		// var pairingSuite = suite.(pairing.Suite) // TODO
		// var signatureMask sign.Mask
		// var aggregatedSig kyber.Point
		// if aggregatedSig, err = bdn.AggregateSignatures(pairingSuite, signatures, &signatureMask); err != nil {
		// 	return nil, nil, err
		// }
		// var aggregatedBin []byte
		// if aggregatedBin, err = aggregatedSig.MarshalBinary(); err != nil {
		// 	return nil, nil, err
		// }
		// if err = bdn.Verify(pairingSuite, sharedPublic, pubKeyBytes, aggregatedBin); err != nil {
		// 	return nil, nil, err
		// }
	}
	fmt.Printf("COORD: Generated ChainID=%v, shared public key: %v\n", generatedChainID, sharedPublic)
	//
	// Commit the keys to persistent storage.
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "7-CommitAndTerminate"}); err != nil {
		return nil, nil, err
	}
	return &generatedChainID, &sharedPublic, nil
}
