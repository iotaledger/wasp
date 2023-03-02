// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// blssig package implements a Common Coin (CC) based on a BLS Threshold signatures as
// described in the Appendix C of
//
// > Andrew Miller, Yu Xia, Kyle Croman, Elaine Shi, and Dawn Song. 2016.
// > The Honey Badger of BFT Protocols. In Proceedings of the 2016 ACM SIGSAC
// > Conference on Computer and Communications Security (CCS '16).
// > Association for Computing Machinery, New York, NY, USA, 31–42.
// > DOI:https://doi.org/10.1145/2976749.2978399
//
// We con't use the DKShare here, because in some cases this CC will be used while
// creating the DKShare.
package blssig

import (
	"errors"
	"fmt"

	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/tbls"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
)

type ccImpl struct {
	suite     pairing.Suite
	nodeIDs   []gpa.NodeID
	pubPoly   *share.PubPoly
	priShare  *share.PriShare
	n         int
	t         int
	me        gpa.NodeID
	sid       []byte
	sigShares map[gpa.NodeID][]byte
	output    *bool
	log       *logger.Logger
}

var _ gpa.GPA = &ccImpl{}

func New(
	suite pairing.Suite,
	nodeIDs []gpa.NodeID,
	pubPoly *share.PubPoly,
	priShare *share.PriShare,
	t int,
	me gpa.NodeID,
	sid []byte,
	log *logger.Logger,
) gpa.GPA {
	cc := &ccImpl{
		suite:     suite,
		nodeIDs:   nodeIDs,
		pubPoly:   pubPoly,
		priShare:  priShare,
		n:         len(nodeIDs),
		t:         t,
		me:        me,
		sid:       sid,
		sigShares: map[gpa.NodeID][]byte{},
		output:    nil,
		log:       log,
	}
	return cc
}

func (cc *ccImpl) Input(input gpa.Input) gpa.OutMessages {
	if input != nil {
		panic(errors.New("input must be nil"))
	}
	if _, ok := cc.sigShares[cc.me]; ok {
		// Only consider the first input.
		return nil
	}
	sigShare, err := tbls.Sign(cc.suite, cc.priShare, cc.sid)
	if err != nil {
		panic(fmt.Errorf("cannot sign a sid: %w", err))
	}
	cc.sigShares[cc.me] = sigShare
	if cc.n == 1 {
		coin := sigShare[len(sigShare)-1]%2 == 1
		cc.output = &coin
		return nil
	}
	cc.tryOutput()
	msgs := gpa.NoMessages()
	for _, nid := range cc.nodeIDs {
		if nid != cc.me {
			msgs.Add(&msgSigShare{recipient: nid, sigShare: sigShare})
		}
	}
	return msgs
}

func (cc *ccImpl) Message(msg gpa.Message) gpa.OutMessages {
	if cc.output != nil {
		// Decided, don't need to process messages anymore.
		return nil
	}
	shareMsg, ok := msg.(*msgSigShare)
	if !ok {
		panic(fmt.Errorf("unexpected message: %+v", msg))
	}
	if _, ok := cc.sigShares[shareMsg.sender]; ok {
		// Drop a duplicate.
		return nil
	}
	cc.sigShares[shareMsg.sender] = shareMsg.sigShare
	cc.tryOutput()
	return nil
}

func (cc *ccImpl) tryOutput() {
	if len(cc.sigShares) < cc.t || cc.output != nil {
		return
	}
	sigs := make([][]byte, 0, len(cc.nodeIDs))
	for _, sig := range cc.sigShares {
		sigs = append(sigs, sig)
	}
	mainSig, err := tbls.Recover(cc.suite, cc.pubPoly, cc.sid, sigs, cc.t, cc.n)
	if err != nil {
		cc.log.Warnf("cannot recover the signature with %v/%v shares: %v", len(cc.sigShares), cc.n, err)
		return
	}
	if err := bdn.Verify(cc.suite, cc.pubPoly.Commit(), cc.sid, mainSig); err != nil {
		cc.log.Warnf("cannot verify the signature: %v", err)
		return
	}
	coin := mainSig[len(mainSig)-1]%2 == 1
	cc.output = &coin
}

func (cc *ccImpl) Output() gpa.Output {
	if cc.output == nil {
		return nil // Untyped nil.
	}
	return cc.output
}

func (cc *ccImpl) StatusString() string {
	return fmt.Sprintf("{CC:blssig, threshold=%v, sigShares=%v/%v, output=%v}", cc.t, len(cc.sigShares), cc.n, cc.output)
}

func (cc *ccImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	msg := &msgSigShare{}
	if err := msg.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal msgSigShare: %w", err)
	}
	return msg, nil
}
