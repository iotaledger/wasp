package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/peering"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/schnorr"
)

// GenerateDistributedKeyParams is only here to make long
// list of paremeters in GenerateDistributedKey easier to specify.
type GenerateDistributedKeyParams struct {
	InitiatorPub kyber.Point
	PeerLocs     []string
	PeerPubs     []kyber.Point
	Threshold    uint16
	Version      address.Version
	Timeout      time.Duration
	Suite        kyber.Group
	NetProvider  peering.NetworkProvider
	Logger       *logger.Logger
}

// GenerateDistributedKey is called from the client node to initiate the DKG
// procedure on a set of nodes. The client is not required to have an instance
// of the DkgNode, but may have it (be one of the peers sharing the secret).
// This function works synchronously, so the user should run it async if needed.
func GenerateDistributedKey(params *GenerateDistributedKeyParams) (*coretypes.ChainID, *kyber.Point, error) {
	var log = params.Logger
	var err error
	dkgID := coretypes.NewRandomChainID()
	var netGroup peering.GroupProvider
	if netGroup, err = params.NetProvider.Group(params.PeerLocs); err != nil {
		return nil, nil, err
	}
	defer netGroup.Close()
	recvCh := make(chan *peering.RecvEvent, len(params.PeerLocs)*2)
	attachID := params.NetProvider.Attach(&dkgID, func(msgFrom *peering.RecvEvent) {
		recvCh <- msgFrom
	})
	defer params.NetProvider.Detach(attachID)
	//
	// Initialize the peers.
	timeInit := time.Now() // TODO: Nicer way to handle that.
	broadcastToGroup(netGroup, &dkgID, &initiatorInitMsg{
		step:         rabinStep0Initialize,
		peerLocs:     params.PeerLocs,
		peerPubs:     params.PeerPubs,
		initiatorPub: params.InitiatorPub,
		threshold:    params.Threshold,
		version:      params.Version,
		timeoutMS:    uint32(params.Timeout.Milliseconds()),
	})
	if err = waitForInitiatorAcks(netGroup, recvCh, rabinStep0Initialize, params); err != nil {
		return nil, nil, err
	}
	//
	// Perform the DKG steps, each step in parallel, all steps sequentially.
	// Step numbering (R) is according to <https://github.com/dedis/kyber/blob/master/share/dkg/rabin/dkg.go>.
	timeStep1 := time.Now()
	broadcastToGroup(netGroup, &dkgID, &initiatorStepMsg{step: rabinStep1R21SendDeals})
	if err = waitForInitiatorAcks(netGroup, recvCh, rabinStep1R21SendDeals, params); err != nil {
		return nil, nil, err
	}
	timeStep2 := time.Now()
	broadcastToGroup(netGroup, &dkgID, &initiatorStepMsg{step: rabinStep2R22SendResponses})
	if err = waitForInitiatorAcks(netGroup, recvCh, rabinStep2R22SendResponses, params); err != nil {
		return nil, nil, err
	}
	timeStep3 := time.Now()
	broadcastToGroup(netGroup, &dkgID, &initiatorStepMsg{step: rabinStep3R23SendJustifications})
	if err = waitForInitiatorAcks(netGroup, recvCh, rabinStep3R23SendJustifications, params); err != nil {
		return nil, nil, err
	}
	timeStep4 := time.Now()
	broadcastToGroup(netGroup, &dkgID, &initiatorStepMsg{step: rabinStep4R4SendSecretCommits})
	if err = waitForInitiatorAcks(netGroup, recvCh, rabinStep4R4SendSecretCommits, params); err != nil {
		return nil, nil, err
	}
	timeStep5 := time.Now()
	broadcastToGroup(netGroup, &dkgID, &initiatorStepMsg{step: rabinStep5R5SendComplaintCommits})
	if err = waitForInitiatorAcks(netGroup, recvCh, rabinStep5R5SendComplaintCommits, params); err != nil {
		return nil, nil, err
	}
	//
	// Now get the public keys.
	// This also performs the "6-R6-SendReconstructCommits" step implicitly.
	timeStep6 := time.Now()
	pubShareResponses := map[int]*initiatorPubShareMsg{}
	broadcastToGroup(netGroup, &dkgID, &initiatorStepMsg{step: rabinStep6R6SendReconstructCommits})
	err = waitForInitiatorMsgs(netGroup, recvCh, rabinStep6R6SendReconstructCommits, params,
		func(recv *peering.RecvEvent, initMsg initiatorMsg) error {
			switch msg := initMsg.(type) {
			case *initiatorPubShareMsg:
				pubShareResponses[int(recv.Msg.SenderIndex)] = msg
				return nil
			default:
				log.Errorf("unexpected message type instead of initiatorPubShareMsg: %V", msg)
				return errors.New("unexpected message type instead of initiatorPubShareMsg")
			}
		},
	)
	if err != nil {
		return nil, nil, err
	}
	chainID := pubShareResponses[0].chainID
	chainIDBytes := chainID.Bytes()
	sharedPublic := pubShareResponses[0].sharedPublic
	publicShares := make([]kyber.Point, len(params.PeerLocs))
	for i := range pubShareResponses {
		if !bytes.Equal(chainIDBytes, pubShareResponses[i].chainID.Bytes()) {
			return nil, nil, fmt.Errorf("nodes generated different addresses")
		}
		if !sharedPublic.Equal(pubShareResponses[i].sharedPublic) {
			return nil, nil, fmt.Errorf("nodes generated different shared public keys")
		}
		publicShares[i] = pubShareResponses[i].publicShare
		{
			var pubShareBytes []byte
			if pubShareBytes, err = pubShareResponses[i].publicShare.MarshalBinary(); err != nil {
				return nil, nil, err
			}
			switch params.Version {
			case address.VersionED25519:
				err = schnorr.Verify(
					params.Suite,
					pubShareResponses[i].publicShare,
					pubShareBytes,
					pubShareResponses[i].signature,
				)
				if err != nil {
					return nil, nil, err
				}
			case address.VersionBLS:
				err = bdn.Verify(
					params.Suite.(pairing.Suite),
					pubShareResponses[i].publicShare,
					pubShareBytes,
					pubShareResponses[i].signature,
				)
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}
	var generatedChainID coretypes.ChainID
	if generatedChainID, err = coretypes.NewChainIDFromBytes(chainIDBytes); err != nil {
		return nil, nil, err
	}
	log.Debugf("Generated ChainID=%v, shared public key: %v", generatedChainID, sharedPublic)
	//
	// Commit the keys to persistent storage.
	timeStep7 := time.Now()
	broadcastToGroup(netGroup, &dkgID, &initiatorDoneMsg{
		step:      rabinStep7CommitAndTerminate,
		pubShares: publicShares,
	})
	if err = waitForInitiatorAcks(netGroup, recvCh, rabinStep7CommitAndTerminate, params); err != nil {
		return nil, nil, err
	}
	timeDone := time.Now()
	log.Debugf(
		"Timing: init=%v, Step1=%v, Step2=%v, Step3=%v, Step4=%v, Step5=%v, Step6=%v, Step7=%v",
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

func broadcastToGroup(netGroup peering.GroupProvider, chainID *coretypes.ChainID, msg msgByteCoder) {
	netGroup.Broadcast(makePeerMessage(chainID, msg), true)
}

func waitForInitiatorAcks(
	netGroup peering.GroupProvider,
	recvCh chan *peering.RecvEvent,
	step byte,
	params *GenerateDistributedKeyParams,
) error {
	return waitForInitiatorMsgs(netGroup, recvCh, step, params,
		func(_ *peering.RecvEvent, _ initiatorMsg) error {
			return nil
		},
	)
}

func waitForInitiatorMsgs(
	netGroup peering.GroupProvider,
	recvCh chan *peering.RecvEvent,
	step byte,
	params *GenerateDistributedKeyParams,
	callback func(recv *peering.RecvEvent, initMsg initiatorMsg) error,
) error {
	var err error
	flags := make([]bool, len(netGroup.AllNodes()))
	errs := make(map[uint16]error)
	timeoutCh := time.After(params.Timeout)
	for !haveAll(flags) {
		select {
		case peerMsg := <-recvCh:
			var fromIdx uint16
			if fromIdx, err = netGroup.PeerIndex(peerMsg.From); err != nil {
				params.Logger.Warnf("Dropping message from unexpected peer %v: %v", peerMsg.From.Location(), peerMsg.Msg)
				continue
			}
			peerMsg.Msg.SenderIndex = uint16(fromIdx)
			var initMsg initiatorMsg
			var isInitMsg bool
			isInitMsg, initMsg, err = readInitiatorMsg(peerMsg.Msg, params.Suite)
			if !isInitMsg {
				continue
			}
			if err != nil {
				params.Logger.Warnf("Failed to read message from %v: %v", peerMsg.From.Location(), peerMsg.Msg)
				continue

			}
			if !initMsg.IsResponse() {
				continue
			}
			if initMsg.Step() != step {
				continue
			}
			if initMsg.Error() != nil {
				errs[fromIdx] = initMsg.Error()
				continue
			}
			if err = callback(peerMsg, initMsg); err != nil {
				errs[fromIdx] = err
			}
			flags[fromIdx] = true
		case <-timeoutCh:
			return errors.New("step_timeout")
		}
	}
	if len(errs) == 0 {
		return nil
	}
	var errMsg string
	for i := range errs {
		errMsg = errMsg + fmt.Sprintf("[%v:%v]", i, errs[i].Error())
	}
	return errors.New(errMsg)
}
