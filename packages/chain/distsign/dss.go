// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package dss runs a NonceDistKeyGeneration and signs the supplied hash.
//
// This is a simplified implementation.
// Later the DistKeyGeneration part can be run in advance, while waiting for transactions.
//
// The general workflow is the following:
//
//  1. Start it upon activation of a step (last stateOutput is approved).
//  2. Exchange the underlying messages until:
//     2.1. ACSS Intermediate output is received.
//  3. Then wait for the ACS and then the VM to complete:
//     3.1. pass the ACS result to the nonce-distKeyGen (to complete the nonces).
//     3.2. pass the VM output as a message to sign (its hash).
//  4. Exchange messages until the signature is produced.
//  5. Output the signature.
//
// TODO: Make sure no two signatures are ever produced by the nonce-distKeyGen for the same
//
//	base TX. That would reveal the permanent private key of the committee.
package distsign

import (
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/adkg/nonce"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
)

type DistributedSignature interface {
	AsGPA() gpa.GPA
}

type Output struct {
	ProposedIndexes []int  // Intermediate output.
	Signature       []byte // Final output.
}

const (
	subsystemDistKeyGeneration byte = iota
)

type distributedSignatureImpl struct {
	suite                           suites.Suite
	withWrappers                    gpa.GPA // This instance, with all the wrappers.
	me                              gpa.NodeID
	mySK                            kyber.Scalar
	nodeIDs                         []gpa.NodeID
	nodePKs                         map[gpa.NodeID]kyber.Point
	f                               int
	longTermSecretShare             tcrypto.SecretShare
	distKeyGen                      gpa.GPA
	distKeyGenOutIndexes            []int                // Intermediate DistKeyGeneration output.
	distKeyGenDecidedIndexProposals map[gpa.NodeID][]int // ACS decision.
	distKeyGenOutNonce              dss.DistKeyShare     // Final DistKeyGeneration output.
	messageToSign                   []byte
	dssPartialSigBuffer             *shrinkingmap.ShrinkingMap[gpa.NodeID, *dss.PartialSig] // Accumulate early partial signatures
	dssSigner                       *dss.DSS
	signature                       []byte // The output.
	msgWrapper                      *gpa.MsgWrapper
	log                             log.Logger
}

var _ DistributedSignature = &distributedSignatureImpl{}

func New(
	suite suites.Suite,
	nodeIDs []gpa.NodeID,
	nodePKs map[gpa.NodeID]kyber.Point,
	f int,
	me gpa.NodeID,
	mySK kyber.Scalar,
	longTermSecretShare tcrypto.SecretShare,
	log log.Logger,
) DistributedSignature {
	d := &distributedSignatureImpl{
		suite:                           suite,
		withWrappers:                    nil, // Set bellow.
		me:                              me,
		mySK:                            mySK,
		nodeIDs:                         nodeIDs,
		nodePKs:                         nodePKs,
		f:                               f,
		longTermSecretShare:             longTermSecretShare,
		distKeyGen:                      nonce.New(suite, nodeIDs, nodePKs, f, me, mySK, log),
		distKeyGenOutIndexes:            nil, // To be decided.
		distKeyGenDecidedIndexProposals: nil, // To be received.
		distKeyGenOutNonce:              nil, // To be decided.
		messageToSign:                   nil, // Will be received later.
		dssPartialSigBuffer:             shrinkingmap.New[gpa.NodeID, *dss.PartialSig](),
		dssSigner:                       nil, // Will be created when indexProposals and message to sign will be created.
		log:                             log,
	}
	d.msgWrapper = gpa.NewMsgWrapper(msgTypeWrapped, d.msgWrapperFunc)
	d.withWrappers = gpa.NewOwnHandler(me, d)
	return d
}

// DSS Specific Interface: Get a GPA instance to pass messages with all the intermediate layers.
func (d *distributedSignatureImpl) AsGPA() gpa.GPA {
	return d.withWrappers
}

// Handle the input to the protocol.
func (d *distributedSignatureImpl) Input(input gpa.Input) gpa.OutMessages {
	d.log.LogDebugf("Input %+v", input)
	switch input := input.(type) {
	case *inputStart:
		msgs := d.msgWrapper.WrapMessages(subsystemDistKeyGeneration, 0, d.distKeyGen.Input(nonce.NewInputStart()))
		return d.tryHandleDkgOutput(msgs)
	case *inputDecided:
		return d.handleDecided(input)
	}
	panic(fmt.Errorf("unexpected input: %T: %+v", input, input))
}

// Handle the messages.
func (d *distributedSignatureImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msgT := msg.(type) {
	case *msgPartialSig:
		d.log.LogDebugf("Message %+v", msg)
		return d.handlePartialSig(msgT)
	case *gpa.WrappingMsg:
		if msgT.Subsystem() == subsystemDistKeyGeneration && msgT.Index() == 0 {
			msgs := d.msgWrapper.WrapMessages(subsystemDistKeyGeneration, 0, d.distKeyGen.Message(msgT.Wrapped()))
			return d.tryHandleDkgOutput(msgs)
		}
		d.log.LogWarnf("unknown wrapped message %+v, wrapped %T: %v", msgT, msgT.Wrapped(), msgT.Wrapped())
		return nil
	default:
		panic(fmt.Errorf("unknown message %T: %v", msg, msg))
	}
}

// Provide the output, if any.
func (d *distributedSignatureImpl) Output() gpa.Output {
	if d.distKeyGenOutIndexes == nil && d.signature == nil {
		return nil
	}
	return &Output{
		ProposedIndexes: d.distKeyGenOutIndexes,
		Signature:       d.signature,
	}
}

func (d *distributedSignatureImpl) tryHandleDkgOutput(msgs gpa.OutMessages) gpa.OutMessages {
	distKeyGenOut := d.distKeyGen.Output()
	if d.distKeyGenOutIndexes == nil && distKeyGenOut != nil && distKeyGenOut.(*nonce.Output).Indexes != nil {
		d.distKeyGenOutIndexes = distKeyGenOut.(*nonce.Output).Indexes
	}
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
		d.dssSigner = dssSigner
		partialSig, err := d.dssSigner.PartialSig()
		if err != nil {
			d.log.LogErrorf("cannot create a partial signature: %v", err)
			return msgs
		}
		//
		// Process early sent partial signatures, if any.
		if d.dssPartialSigBuffer.Size() > 0 {
			d.dssPartialSigBuffer.ForEach(func(nid gpa.NodeID, ps *dss.PartialSig) bool {
				err := d.dssSigner.ProcessPartialSig(ps)
				if err != nil {
					d.log.LogErrorf("Failed to process a buffered partial signature: %v", err)
				}

				d.dssPartialSigBuffer.Delete(nid)
				return true
			})
		}
		//
		// Broadcast it (except the current node).
		for i := range d.nodeIDs {
			if d.nodeIDs[i] == d.me {
				continue
			}
			msg := &msgPartialSig{
				BasicMessage: gpa.NewBasicMessage(d.nodeIDs[i]),
				suite:        d.suite,
				partialSig:   partialSig,
			}
			msg.SetSender(d.me)
			msgs.Add(msg)
		}
		//
		// Maybe we have everything for the signature already?
		if d.dssSigner.EnoughPartialSig() {
			sig, err := d.dssSigner.Signature()
			if err != nil {
				d.log.LogErrorf("unable to aggregate the signature: %v", err)
				return msgs
			}
			d.signature = sig
		}
	}
	return msgs
}

func (d *distributedSignatureImpl) handlePartialSig(msg *msgPartialSig) gpa.OutMessages {
	if d.signature != nil {
		// Signature already aggregated, ignore the remaining shares.
		return nil
	}
	if d.dssSigner == nil {
		if d.dssPartialSigBuffer.Has(msg.Sender()) {
			d.log.LogWarn("duplicate partial signature from %v", msg.Sender())
			return nil
		}

		d.dssPartialSigBuffer.Set(msg.Sender(), msg.partialSig)
		return nil
	}
	//
	// Then process the one received with the current message.
	err := d.dssSigner.ProcessPartialSig(msg.partialSig)
	if err != nil {
		d.log.LogWarnf("Failed to process a partial signature: %v", err)
		return nil
	}
	if !d.dssSigner.EnoughPartialSig() {
		return nil
	}

	sig, err := d.dssSigner.Signature()
	if err != nil {
		d.log.LogErrorf("unable to aggregate the signature: %v", err)
		return nil
	}
	d.signature = sig
	return nil
}

func (d *distributedSignatureImpl) handleDecided(input *inputDecided) gpa.OutMessages {
	if d.distKeyGenDecidedIndexProposals != nil {
		d.log.LogWarn("Duplicate will be dropped: DecidedIndexes=%+v", input.decidedIndexProposals)
		return nil
	}
	d.distKeyGenDecidedIndexProposals = input.decidedIndexProposals
	d.messageToSign = input.messageToSign

	decisionInput := nonce.NewInputAgreementResult(input.decidedIndexProposals)
	msgs := d.msgWrapper.WrapMessages(subsystemDistKeyGeneration, 0, d.distKeyGen.Input(decisionInput))
	return d.tryHandleDkgOutput(msgs)
}

func (d *distributedSignatureImpl) nodePKArray() []kyber.Point {
	res := make([]kyber.Point, len(d.nodeIDs))
	for i := range res {
		res[i] = d.nodePKs[d.nodeIDs[i]]
	}
	return res
}

func (d *distributedSignatureImpl) StatusString() string {
	return fmt.Sprintf("{DSS, distKeyGen=%v}", d.distKeyGen.StatusString())
}
