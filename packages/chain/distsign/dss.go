// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package distsign runs a NonceDKG and signs the supplied hash.
//
// This is a simplified implementation.
// Later the DKG part can be run in advance, while waiting for transactions.
//
// The general workflow is the following:
//
//  1. Start it upon activation of a step (last stateOutput is approved).
//  2. Exchange the underlying messages until:
//     2.1. ACSS Intermediate output is received.
//  3. Then wait for the ACS and then the VM to complete:
//     3.1. pass the ACS result to the nonce-dkg (to complete the nonces).
//     3.2. pass the VM output as a message to sign (its hash).
//  4. Exchange messages until the signature is produced.
//  5. Output the signature.
//
// TODO: Make sure no two signatures are ever produced by the nonce-dkg for the same
//
//	base TX. That would reveal the permanent private key of the committee.
package distsign

import (
	"fmt"
	"slices"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/acss"
	"github.com/iotaledger/wasp/v2/packages/gpa/asyncdistkeygen/nonce"
	rbc "github.com/iotaledger/wasp/v2/packages/gpa/rbc/bracha"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
)

type Output struct {
	ProposedIndexes []int  // Intermediate output.
	Signature       []byte // Final output.
}

const (
	subsystemDistributedKeyGeneration byte = iota
)

type DistributedSignature struct {
	suite                           suites.Suite
	me                              gpa.NodeID
	mySK                            kyber.Scalar
	nodeIDs                         []gpa.NodeID
	nodePKs                         map[gpa.NodeID]kyber.Point
	f                               int
	longTermSecretShare             tcrypto.SecretShare
	distributedKeyGen               *nonce.NonceDistributedKeyGeneration
	distKeyGenOutIndexes            []int                // Intermediate DKG output.
	distKeyGenDecidedIndexProposals map[gpa.NodeID][]int // ACS decision.
	distKeyGenOutNonce              dss.DistKeyShare     // Final DKG output.
	messageToSign                   []byte
	distSignPartialSigBuffer        *shrinkingmap.ShrinkingMap[gpa.NodeID, *dss.PartialSig] // Accumulate early partial signatures
	distributedSignatureSigner      *dss.DSS
	signature                       []byte // The output.
	log                             log.Logger
}

func New(
	suite suites.Suite,
	nodeIDs []gpa.NodeID,
	nodePKs map[gpa.NodeID]kyber.Point,
	f int,
	me gpa.NodeID,
	mySK kyber.Scalar,
	longTermSecretShare tcrypto.SecretShare,
	log log.Logger,
) *DistributedSignature {
	d := &DistributedSignature{
		suite:                           suite,
		me:                              me,
		mySK:                            mySK,
		nodeIDs:                         nodeIDs,
		nodePKs:                         nodePKs,
		f:                               f,
		longTermSecretShare:             longTermSecretShare,
		distributedKeyGen:               nonce.New(suite, nodeIDs, nodePKs, f, me, mySK, log),
		distKeyGenOutIndexes:            nil, // To be decided.
		distKeyGenDecidedIndexProposals: nil, // To be received.
		distKeyGenOutNonce:              nil, // To be decided.
		messageToSign:                   nil, // Will be received later.
		distSignPartialSigBuffer:        shrinkingmap.New[gpa.NodeID, *dss.PartialSig](),
		distributedSignatureSigner:      nil, // Will be created when indexProposals and message to sign will be created.
		log:                             log,
	}
	return d
}

// Input handles the input to the protocol.
func (d *DistributedSignature) Input(input gpa.Input) []gpa.MessageOut {
	d.log.LogDebugf("Input %+v", input)
	switch input := input.(type) {
	case *inputStart:
		msgs := d.distributedKeyGen.Input(nonce.NewInputStart())
		return slices.Concat(msgs, d.tryHandleDistributedKeyGenerationOutput())
	case *inputDecided:
		return d.handleDecided(input)
	}
	panic(fmt.Errorf("unexpected input: %T: %+v", input, input))
}

func (d *DistributedSignature) HandleACSSMsgBracha(acssIndex int, m gpa.MessageIn[rbc.MsgBracha]) []gpa.MessageOut {
	outMsgs := d.distributedKeyGen.HandleACSSMsgBracha(acssIndex, m)
	return slices.Concat(outMsgs, d.tryHandleDistributedKeyGenerationOutput())
}

func (d *DistributedSignature) HandleACSSMsgVote(acssIndex int, msg gpa.MessageIn[acss.MsgVote]) []gpa.MessageOut {
	outMsgs := d.distributedKeyGen.HandleACSSMsgVote(acssIndex, msg)
	return slices.Concat(outMsgs, d.tryHandleDistributedKeyGenerationOutput())
}

func (d *DistributedSignature) HandleACSSMsgImplicateRecover(acssIndex int, msg gpa.MessageIn[acss.MsgImplicateRecover]) []gpa.MessageOut {
	outMsgs := d.distributedKeyGen.HandleACSSMsgImplicateRecover(acssIndex, msg)
	return slices.Concat(outMsgs, d.tryHandleDistributedKeyGenerationOutput())
}

// Output provides the output, if any.
func (d *DistributedSignature) Output() gpa.Output {
	if d.distKeyGenOutIndexes == nil && d.signature == nil {
		return nil
	}
	return &Output{
		ProposedIndexes: d.distKeyGenOutIndexes,
		Signature:       d.signature,
	}
}

func (d *DistributedSignature) tryHandleDistributedKeyGenerationOutput() []gpa.MessageOut {
	distKeyGenOut := d.distributedKeyGen.Output()
	if d.distKeyGenOutIndexes == nil && distKeyGenOut != nil && distKeyGenOut.(*nonce.Output).Indexes != nil {
		d.distKeyGenOutIndexes = distKeyGenOut.(*nonce.Output).Indexes
	}
	var msgs []gpa.MessageOut
	if d.distKeyGenOutNonce == nil && distKeyGenOut != nil && distKeyGenOut.(*nonce.Output).PriShare != nil {
		d.distKeyGenOutNonce = tcrypto.NewDistKeyShare(
			distKeyGenOut.(*nonce.Output).PriShare,
			distKeyGenOut.(*nonce.Output).Commits,
			len(d.nodeIDs),
			distKeyGenOut.(*nonce.Output).Threshold,
		)
		//
		// Create a partial signature.
		dssSigner, err := dss.NewDSS(d.suite, d.mySK, d.nodePKArray(), d.longTermSecretShare, d.distKeyGenOutNonce, d.messageToSign, d.longTermSecretShare.Threshold())
		if err != nil {
			d.log.LogError("Failed to create DSS Signer: %v", err)
			return msgs
		}
		d.distributedSignatureSigner = dssSigner
		partialSig, err := d.distributedSignatureSigner.PartialSig()
		if err != nil {
			d.log.LogErrorf("cannot create a partial signature: %v", err)
			return msgs
		}
		//
		// Process early sent partial signatures, if any.
		if d.distSignPartialSigBuffer.Size() > 0 {
			d.distSignPartialSigBuffer.ForEach(func(nid gpa.NodeID, ps *dss.PartialSig) bool {
				err := d.distributedSignatureSigner.ProcessPartialSig(ps)
				if err != nil {
					d.log.LogErrorf("Failed to process a buffered partial signature: %v", err)
				}

				d.distSignPartialSigBuffer.Delete(nid)
				return true
			})
		}
		//
		// Broadcast it (except the current node).
		for i := range d.nodeIDs {
			if d.nodeIDs[i] == d.me {
				continue
			}

			msg, err := NewMsgPartialSig(partialSig)
			if err != nil {
				d.log.LogErrorf("cannot create MsgPartialSig: %v", err)
				continue
			}

			msgs = append(msgs, gpa.NewMessageOut(d.nodeIDs[i], msg))
		}
		//
		// Maybe we have everything for the signature already?
		if d.distributedSignatureSigner.EnoughPartialSig() {
			sig, err := d.distributedSignatureSigner.Signature()
			if err != nil {
				d.log.LogErrorf("unable to aggregate the signature: %v", err)
				return msgs
			}
			d.signature = sig
		}
	}
	return msgs
}

func (d *DistributedSignature) HandleMsgPartialSig(msg gpa.MessageIn[MsgPartialSig]) []gpa.MessageOut {
	if d.signature != nil {
		// Signature already aggregated, ignore the remaining shares.
		return nil
	}

	partialSig, err := msg.Payload.PartialSig(d.suite)
	if err != nil {
		d.log.LogErrorf("Failed to extract partial signature from the message: %v", err)
		return nil
	}

	if d.distributedSignatureSigner == nil {
		if d.distSignPartialSigBuffer.Has(msg.Sender) {
			d.log.LogWarn("duplicate partial signature from %v", msg.Sender)
			return nil
		}

		d.distSignPartialSigBuffer.Set(msg.Sender, partialSig)
		return nil
	}
	//
	// Then process the one received with the current message.
	if err := d.distributedSignatureSigner.ProcessPartialSig(partialSig); err != nil {
		d.log.LogWarnf("Failed to process a partial signature: %v", err)
		return nil
	}
	if !d.distributedSignatureSigner.EnoughPartialSig() {
		return nil
	}

	sig, err := d.distributedSignatureSigner.Signature()
	if err != nil {
		d.log.LogErrorf("unable to aggregate the signature: %v", err)
		return nil
	}
	d.signature = sig
	return nil
}

func (d *DistributedSignature) handleDecided(input *inputDecided) []gpa.MessageOut {
	if d.distKeyGenDecidedIndexProposals != nil {
		d.log.LogWarn("Duplicate will be dropped: DecidedIndexes=%+v", input.decidedIndexProposals)
		return nil
	}
	d.distKeyGenDecidedIndexProposals = input.decidedIndexProposals
	d.messageToSign = input.messageToSign

	decisionInput := nonce.NewInputAgreementResult(input.decidedIndexProposals)
	msgs := d.distributedKeyGen.Input(decisionInput)
	return slices.Concat(msgs, d.tryHandleDistributedKeyGenerationOutput())
}

func (d *DistributedSignature) nodePKArray() []kyber.Point {
	res := make([]kyber.Point, len(d.nodeIDs))
	for i := range res {
		res[i] = d.nodePKs[d.nodeIDs[i]]
	}
	return res
}

func (d *DistributedSignature) StatusString() string {
	return fmt.Sprintf("{DSS, dkg=%v}", d.distributedKeyGen.StatusString())
}
