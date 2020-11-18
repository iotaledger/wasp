package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/schnorr"
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
	threshold uint32,
	version address.Version,
	timeout time.Duration,
	suite kyber.Group,
	netProvider CoordNodeProvider,
	log *logger.Logger,
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
		Threshold: threshold,
		Version:   version,
		TimeoutMS: uint64(timeout.Milliseconds()),
	}
	timeInit := time.Now() // TODO: Nicer way to handle that.
	if err = netProvider.DkgInit(peerLocs, dkgID, &initReq); err != nil {
		return nil, nil, err
	}
	//
	// Perform the DKG steps, each step in parallel, all steps sequentially.
	// Step numbering (R) is according to <https://github.com/dedis/kyber/blob/master/share/dkg/rabin/dkg.go>.
	timeStep1 := time.Now()
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "1-R2.1-SendDeals"}); err != nil {
		return nil, nil, err
	}
	timeStep2 := time.Now()
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "2-R2.2-SendResponses"}); err != nil {
		return nil, nil, err
	}
	timeStep3 := time.Now()
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "3-R2.3-SendJustifications"}); err != nil {
		return nil, nil, err
	}
	timeStep4 := time.Now()
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "4-R4-SendSecretCommits"}); err != nil {
		return nil, nil, err
	}
	timeStep5 := time.Now()
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "5-R5-SendComplaintCommits"}); err != nil {
		return nil, nil, err
	}
	//
	// Now get the public keys.
	// This also performs the "6-R6-SendReconstructCommits" step implicitly.
	timeStep6 := time.Now()
	var pubKeyResponses []*PubKeyResp
	if pubKeyResponses, err = netProvider.DkgPubKey(peerLocs, dkgID); err != nil {
		return nil, nil, err
	}
	chainIDBytes := pubKeyResponses[0].ChainID
	sharedPublicBytes := pubKeyResponses[0].SharedPublic
	signatures := [][]byte{}
	for i := range pubKeyResponses {
		if !bytes.Equal(pubKeyResponses[i].ChainID, chainIDBytes) {
			return nil, nil, fmt.Errorf("nodes generated different addresses")
		}
		if !bytes.Equal(pubKeyResponses[i].SharedPublic, sharedPublicBytes) {
			return nil, nil, fmt.Errorf("nodes generated different shared public keys")
		}
		{
			publicShare := suite.Point()
			publicShare.UnmarshalBinary(pubKeyResponses[i].PublicShare)
			switch version {
			case address.VersionED25519:
				err = schnorr.Verify(
					suite,
					publicShare,
					pubKeyResponses[i].PublicShare,
					pubKeyResponses[i].Signature,
				)
				if err != nil {
					return nil, nil, err
				}
			case address.VersionBLS:
				err = bdn.Verify(
					suite.(pairing.Suite),
					publicShare,
					pubKeyResponses[i].PublicShare,
					pubKeyResponses[i].Signature,
				)
				if err != nil {
					return nil, nil, err
				}
			}
		}
		signatures = append(signatures, pubKeyResponses[i].Signature)
	}
	var generatedChainID coretypes.ChainID
	if generatedChainID, err = coretypes.NewChainIDFromBytes(chainIDBytes); err != nil {
		return nil, nil, err
	}
	sharedPublic := suite.Point()
	sharedPublic.UnmarshalBinary(sharedPublicBytes)
	//
	// TODO: Verify signatures.
	// switch version {
	// case address.VersionED25519:
	// 	// TODO
	// case address.VersionBLS:
	// 	var pairingSuite = suite.(pairing.Suite)
	// 	var signatureMask sign.Mask
	// 	var aggregatedSig kyber.Point
	// 	if aggregatedSig, err = bdn.AggregateSignatures(pairingSuite, signatures, &signatureMask); err != nil {
	// 		return nil, nil, err
	// 	}
	// 	var aggregatedBin []byte
	// 	if aggregatedBin, err = aggregatedSig.MarshalBinary(); err != nil {
	// 		return nil, nil, err
	// 	}
	// 	if err = bdn.Verify(pairingSuite, sharedPublic, pubKeyBytes, aggregatedBin); err != nil {
	// 		return nil, nil, err
	// 	}
	// }
	log.Debugf("COORD: Generated ChainID=%v, shared public key: %v", generatedChainID, sharedPublic)
	//
	// Commit the keys to persistent storage.
	timeStep7 := time.Now()
	if err = netProvider.DkgStep(peerLocs, dkgID, &StepReq{Step: "7-CommitAndTerminate"}); err != nil {
		return nil, nil, err
	}
	timeDone := time.Now()
	log.Debugf(
		"COORD: Timing: init=%v, Step1=%v, Step2=%v, Step3=%v, Step4=%v, Step5=%v, Step6=%v, Step7=%v",
		timeStep1.Sub(timeInit).Milliseconds(),
		timeStep2.Sub(timeStep1).Milliseconds(),
		timeStep3.Sub(timeStep2).Milliseconds(),
		timeStep4.Sub(timeStep3).Milliseconds(),
		timeStep5.Sub(timeStep4).Milliseconds(),
		timeStep6.Sub(timeStep5).Milliseconds(),
		timeStep7.Sub(timeStep6).Milliseconds(),
		timeDone.Sub(timeStep7).Milliseconds(),
	)
	return &generatedChainID, &sharedPublic, nil
}
