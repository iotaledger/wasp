// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"math/rand"

	"github.com/iotaledger/hive.go/crypto/bls"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

// NOTE: this is a temporar mock of DKShare
// TODO: remove this when DKShare is migrated to iotago
type mockedDKShare struct {
	Env         *MockedEnv
	Address     iotago.Address
	T           uint16
	NodePubKeys []*cryptolib.PublicKey
	Log         *logger.Logger
}

var _ tcrypto.DKShare = &mockedDKShare{}

func NewMockedDKShare(env *MockedEnv, quorum uint16, nodePubKeys []*cryptolib.PublicKey) *mockedDKShare {
	addr := make([]byte, iotago.AliasAddressBytesLength)
	rand.Read(addr)
	var aliasAddr iotago.AliasAddress
	copy(aliasAddr[:], addr)
	ret := &mockedDKShare{
		Env:         env,
		Address:     &aliasAddr,
		T:           quorum,
		NodePubKeys: nodePubKeys,
		Log:         env.Log.Named("dks"),
	}
	ret.Log.Debugf("DKShare mocked, address: %s", ret.Address)
	return ret
}

func (mdksT *mockedDKShare) GetAddress() iotago.Address {
	return mdksT.Address
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

func (*mockedDKShare) GetIndex() *uint16                { panic("TODO") }
func (*mockedDKShare) GetSharedPublic() kyber.Point     { panic("TODO") }
func (*mockedDKShare) GetPublicShares() []kyber.Point   { panic("TODO") }
func (*mockedDKShare) SetPublicShares(ps []kyber.Point) { panic("TODO") }
func (*mockedDKShare) GetPrivateShare() kyber.Scalar    { panic("TODO") }

func (*mockedDKShare) Bytes() []byte                                            { panic("TODO") }
func (*mockedDKShare) VerifySigShare(data []byte, sigshare tbls.SigShare) error { panic("TODO") }
func (*mockedDKShare) VerifyMasterSignature(data, signature []byte) error       { panic("TODO") }
func (*mockedDKShare) SignShare(data []byte) (tbls.SigShare, error)             { panic("TODO") }
func (*mockedDKShare) RecoverFullSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error) {
	panic("TODO")
}

func SetupDkg(env *MockedEnv, threshold uint16, identities []*cryptolib.KeyPair) (iotago.Address, []registry.DKShareRegistryProvider) {
	result := make([]registry.DKShareRegistryProvider, len(identities))
	pubKeys := make([]*cryptolib.PublicKey, len(identities))
	for i := range identities {
		pubKeys[i] = identities[i].GetPublicKey()
	}
	var address iotago.Address
	for i := range result {
		result[i] = testutil.NewDkgRegistryProvider(tcrypto.DefaultSuite())
		dks := NewMockedDKShare(env, threshold, pubKeys)
		if i == 0 {
			address = dks.GetAddress()
		}
		err := result[i].SaveDKShare(dks)
		if err != nil {
			env.Log.Debugf("Unable to save new DKShare: %v", err)
		}
	}
	return address, result
}
