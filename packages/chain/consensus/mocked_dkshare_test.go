// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/hive.go/crypto/bls"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"go.dedis.ch/kyber/v3/util/key"
)

var publicKey = bls.PrivateKeyFromRandomness().PublicKey()

// NOTE: this is a temporar mock of DKShare
// TODO: remove this when DKShare is migrated to iotago
type mockedDKShare struct {
	Env          *MockedEnv
	Address      iotago.Address
	Index        uint16
	T            uint16
	NodePubKeys  []*cryptolib.PublicKey
	PrivateShare kyber.Scalar
	Log          *logger.Logger
}

var _ tcrypto.DKShare = &mockedDKShare{}

func NewMockedDKShare(env *MockedEnv, address iotago.Address, index uint16, quorum uint16, nodePubKeys []*cryptolib.PublicKey) *mockedDKShare {
	ret := &mockedDKShare{
		Env:          env,
		Address:      address,
		Index:        index,
		T:            quorum,
		NodePubKeys:  nodePubKeys,
		PrivateShare: key.NewKeyPair(tcrypto.DefaultSuite()).Private,
		Log:          env.Log.Named("dks"),
	}
	ret.Log.Debugf("DKShare mocked, address: %s", ret.Address)
	return ret
}

func (mdksT *mockedDKShare) GetAddress() iotago.Address {
	return mdksT.Address
}

func (mdksT *mockedDKShare) GetIndex() *uint16 {
	return &mdksT.Index
}

func (mdksT *mockedDKShare) GetN() uint16 {
	return uint16(len(mdksT.NodePubKeys))
}

func (mdksT *mockedDKShare) GetT() uint16 {
	return mdksT.T
}

func (mdksT *mockedDKShare) GetNodePubKeys() []*cryptolib.PublicKey {
	return mdksT.NodePubKeys
}

func (mdksT *mockedDKShare) SignShare(data []byte) (tbls.SigShare, error) {
	mdksT.Log.Debugf("DKShare mock: signing data")
	priShare := &share.PriShare{
		I: int(*mdksT.GetIndex()),
		V: mdksT.PrivateShare,
	}
	return tbls.Sign(tcrypto.DefaultSuite(), priShare, data)
}

func (mdksT *mockedDKShare) VerifySigShare(data []byte, sigshare tbls.SigShare) error {
	mdksT.Log.Debugf("DKShare mock: verifying data - always success")
	return nil
}

func (*mockedDKShare) RecoverFullSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error) {
	var signature bls.Signature
	for i := range signature {
		base := len(sigShares) + 1
		iMod := i % base
		iDiv := i / base
		switch iMod {
		case 0:
			signature[i] = data[iDiv]
		default:
			signature[i] = sigShares[iMod-1][iDiv]
		}
	}
	signatureWithPK := bls.NewSignatureWithPublicKey(publicKey, signature)
	return &signatureWithPK, nil
}

func (*mockedDKShare) GetSharedPublic() kyber.Point     { panic("TODO") }
func (*mockedDKShare) GetPublicShares() []kyber.Point   { panic("TODO") }
func (*mockedDKShare) SetPublicShares(ps []kyber.Point) { panic("TODO") }
func (*mockedDKShare) GetPrivateShare() kyber.Scalar    { panic("TODO") }

func (*mockedDKShare) Bytes() []byte                                      { panic("TODO") }
func (*mockedDKShare) VerifyMasterSignature(data, signature []byte) error { panic("TODO") }
