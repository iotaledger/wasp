// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package commoncoin implements a common coin abstraction needed by
// the HoneyBadgerBFT for synchronization and randomness.
//
// See A. Miller et al. The honey badger of bft protocols. 2016.
// <https://eprint.iacr.org/2016/199.pdf>, Appendix C.
package commoncoin

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/anthdm/hbbft"
	"github.com/iotaledger/hive.go/crypto/bls"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"golang.org/x/xerrors"
)

// region blsCommonCoin ////////////////////////////////////////////////////////

type blsCommonCoin struct {
	dkShare   tcrypto.DKShare // Key set used to generate the coin.
	allRandom bool
	salt      []byte // Bytes to use when generating SID.
	nodeID    uint64
	epochs    map[uint32]*blsCommonCoinEpoch
}

func NewBlsCommonCoin(dkShare tcrypto.DKShare, salt []byte, allRandom bool) hbbft.CommonCoin {
	cc := blsCommonCoin{
		dkShare:   dkShare,
		allRandom: allRandom,
		salt:      salt,
		nodeID:    0,
		epochs:    make(map[uint32]*blsCommonCoinEpoch),
	}
	return &cc
}

func (cc *blsCommonCoin) ForNodeID(nodeID uint64) hbbft.CommonCoin {
	return &blsCommonCoin{
		dkShare: cc.dkShare,
		salt:    cc.salt,
		nodeID:  nodeID,
		epochs:  make(map[uint32]*blsCommonCoinEpoch),
	}
}

func (cc *blsCommonCoin) HandleRequest(epoch uint32, payload interface{}) (*bool, []interface{}, error) {
	return cc.fetchEpoch(epoch).handleRequest(payload)
}

func (cc *blsCommonCoin) StartCoinFlip(epoch uint32) (*bool, []interface{}, error) {
	return cc.fetchEpoch(epoch).startCoinFlip()
}

func (cc *blsCommonCoin) fetchEpoch(epoch uint32) *blsCommonCoinEpoch {
	if e, ok := cc.epochs[epoch]; ok {
		return e
	}
	sid := make([]byte, 12)
	binary.BigEndian.PutUint64(sid[0:8], cc.nodeID)
	binary.BigEndian.PutUint32(sid[8:12], epoch)
	if cc.salt != nil {
		sid = append(cc.salt, sid...)
	}
	e := &blsCommonCoinEpoch{
		epoch:  epoch,
		sid:    sid,
		shares: make(map[int]tbls.SigShare),
		coin:   nil,
		cc:     cc,
	}
	cc.epochs[epoch] = e
	return e
}

// endregion ///////////////////////////////////////////////////////////////////

// region blsCommonCoinEpoch ///////////////////////////////////////////////////
type blsCommonCoinEpoch struct {
	epoch  uint32                // The current epoch we are working on.
	sid    []byte                // SID used for the current coin.
	shares map[int]tbls.SigShare // Shares being aggregated.
	coin   *bool                 // The generated coin.
	cc     *blsCommonCoin
}

func (cce *blsCommonCoinEpoch) startCoinFlip() (*bool, []interface{}, error) {
	if cce.coin != nil {
		return cce.coin, []interface{}{}, nil
	}
	//
	// Some coins can ne non-random.
	if !cce.cc.allRandom {
		mod5 := cce.epoch % 5
		if mod5 < 2 {
			coin := true
			cce.coin = &coin
			return cce.coin, []interface{}{}, nil
		}
		if mod5 < 4 {
			coin := false
			cce.coin = &coin
			return cce.coin, []interface{}{}, nil
		}
	}
	//
	// Make the coin share.
	var err error
	var sigShare tbls.SigShare
	if sigShare, err = cce.cc.dkShare.BlsSignShare(cce.sid); err != nil {
		return nil, nil, xerrors.Errorf("failed to sign our share: %w", err)
	}
	if err = cce.acceptShare(sigShare); err != nil && err.Error() != "threshold not reached" {
		return nil, nil, xerrors.Errorf("failed to accept our share: %v", err)
	}
	broadcastMsg := BlsCommonCoinMsg{coinShare: sigShare}
	return cce.coin, []interface{}{&broadcastMsg}, nil
}

func (cce *blsCommonCoinEpoch) handleRequest(payload interface{}) (*bool, []interface{}, error) {
	if cce.coin != nil {
		return cce.coin, []interface{}{}, nil
	}
	sigShare := payload.(*BlsCommonCoinMsg).coinShare
	if err := cce.acceptShare(sigShare); err != nil && err.Error() != "threshold not reached" {
		return nil, nil, xerrors.Errorf("failed to accept share: %v", err)
	}
	return cce.coin, []interface{}{}, nil
}

func (cce *blsCommonCoinEpoch) acceptShare(share tbls.SigShare) error {
	dkShare := cce.cc.dkShare
	var err error
	var index int
	if index, err = share.Index(); err != nil {
		return xerrors.Errorf("unable to extract coin share index: %w", err)
	}
	if uint16(index) > dkShare.GetN() {
		return xerrors.Errorf("invalid coin share index %v > N", index)
	}
	if _, ok := cce.shares[index]; ok {
		// Ignore duplicated info.
		return nil
	}
	cce.shares[index] = share
	if uint16(len(cce.shares)) < dkShare.GetT() {
		return xerrors.New("threshold not reached")
	}
	receivedShares := make([][]byte, 0)
	for _, share := range cce.shares {
		receivedShares = append(receivedShares, share)
	}
	var sig *bls.SignatureWithPublicKey
	if sig, err = dkShare.BlsRecoverMasterSignature(receivedShares, cce.sid); err != nil {
		return xerrors.Errorf("unable to reconstruct the signature: %w", err)
	}
	value := sig.Signature.Bytes()
	if err = dkShare.VerifyMasterSignature(cce.sid, value); err != nil {
		return xerrors.Errorf("unable to verify the master signature: %w", err)
	}
	coin := value[len(value)-1]%2 == 1
	cce.coin = &coin
	return nil
}

// endregion ///////////////////////////////////////////////////////////////////

// region BlsCommonCoinMsg /////////////////////////////////////////////////////

type BlsCommonCoinMsg struct {
	coinShare tbls.SigShare
}

func (m *BlsCommonCoinMsg) Write(w io.Writer) error {
	var err error
	if err = util.WriteBytes16(w, m.coinShare); err != nil {
		return xerrors.Errorf("failed to write BlsCommonCoinMsg.coinShare: %w", err)
	}
	return nil
}

func (m *BlsCommonCoinMsg) Read(r io.Reader) error {
	var err error
	if m.coinShare, err = util.ReadBytes16(r); err != nil {
		return xerrors.Errorf("failed to read BlsCommonCoinMsg.coinShare: %w", err)
	}
	return nil
}

func (m *BlsCommonCoinMsg) FromBytes(buf []byte) error {
	r := bytes.NewReader(buf)
	return m.Read(r)
}

func (m *BlsCommonCoinMsg) Bytes() []byte {
	var buf bytes.Buffer
	_ = m.Write(&buf)
	return buf.Bytes()
}

// endregion ///////////////////////////////////////////////////////////////////
