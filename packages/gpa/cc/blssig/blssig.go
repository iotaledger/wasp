// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// blssig package implements a Common Coin (CC) based on a BLS Threshold signatures as
// described in the Appendix C of
//
//	Andrew Miller, Yu Xia, Kyle Croman, Elaine Shi, and Dawn Song. 2016.
//	The Honey Badger of BFT Protocols. In Proceedings of the 2016 ACM SIGSAC
//	Conference on Computer and Communications Security (CCS '16).
//	Association for Computing Machinery, New York, NY, USA, 31–42.
//	DOI:https://doi.org/10.1145/2976749.2978399
//
// We con't use the DKShare here, because in some cases this CC will be used while
// creating the DKShare.
package blssig

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"golang.org/x/xerrors"
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
}

var _ gpa.GPA = &ccImpl{}

func New(
	suite pairing.Suite,
	nodeIDs []gpa.NodeID,
	pubPoly *share.PubPoly,
	priShare *share.PriShare,
	n int,
	t int,
	me gpa.NodeID,
	sid []byte,
) gpa.GPA {
	cc := &ccImpl{
		suite:    suite,
		nodeIDs:  nodeIDs,
		pubPoly:  pubPoly,
		priShare: priShare,
		n:        n,
		t:        t,
		me:       me,
		sid:      sid,
		output:   nil,
	}
	return cc
}

func (cc *ccImpl) Input(input gpa.Input) gpa.OutMessages {
	if input != nil {
		panic(xerrors.Errorf("input must be nil"))
	}
	if _, ok := cc.sigShares[cc.me]; ok {
		// Only consider the first input.
		return gpa.NoMessages()
	}
	sigShare, err := tbls.Sign(cc.suite, cc.priShare, cc.sid)
	if err != nil {
		panic(xerrors.Errorf("cannot sign a sid: %v", err))
	}
	cc.sigShares[cc.me] = sigShare
	msgs := gpa.NoMessages()
	for _, nid := range cc.nodeIDs {
		if nid != cc.me {
			msgs.Add(&msgSigShare{recipient: nid, sender: cc.me, sigShare: sigShare})
		}
	}
	return msgs
}

func (cc *ccImpl) Message(msg gpa.Message) gpa.OutMessages {
	shareMsg, ok := msg.(*msgSigShare)
	if !ok {
		panic(xerrors.Errorf("unexpected message: %+v", msg))
	}
	if _, ok := cc.sigShares[shareMsg.sender]; ok {
		// Drop a duplicate.
		return gpa.NoMessages()
	}
	cc.sigShares[shareMsg.sender] = shareMsg.sigShare
	if len(cc.sigShares) >= cc.t && cc.output == nil {
		sigs := make([][]byte, len(cc.nodeIDs))
		for i := range sigs {
			if sig, ok := cc.sigShares[cc.nodeIDs[i]]; ok {
				sigs[i] = sig
			}
		}
		mainSig, err := tbls.Recover(cc.suite, cc.pubPoly, cc.sid, sigs, cc.t, cc.n)
		if err != nil {
			// TODO: Log the error.
			return gpa.NoMessages()
		}
		if err := tbls.Verify(cc.suite, cc.pubPoly, cc.sid, mainSig); err != nil {
			// TODO: Log the error.
			return gpa.NoMessages()
		}
		coin := mainSig[len(mainSig)-1]%2 == 1
		cc.output = &coin
		return gpa.NoMessages()
	}
	return gpa.NoMessages()
}

func (cc *ccImpl) Output() gpa.Output {
	return cc.output
}

func (cc *ccImpl) StatusString() string {
	return fmt.Sprintf("{CC:blssig, sigShares=%v, output=%v}", cc.sigShares, cc.output)
}

func (cc *ccImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return nil, xerrors.Errorf("not implemented") // TODO: XXX: Impl.
}
