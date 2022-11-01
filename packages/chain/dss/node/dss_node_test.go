// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package node_test

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/sign/eddsa"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/crypto/bls"
	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	dss_node "github.com/iotaledger/wasp/packages/chain/dss/node"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/adkg"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/util"
)

const ( // HT = High Threshold, LT = Low Threshold.
	dkgTypeRobustLT byte = iota
	dkgTypeTrivialHT
)

func TestBasic(t *testing.T) {
	t.Run("n=4,f=1,reliable,dkgTypeRobustLT", func(tt *testing.T) { testGeneric(tt, 4, 1, true, dkgTypeRobustLT) })
	t.Run("n=4,f=1,unreliable,dkgTypeTrivialHT", func(tt *testing.T) { testGeneric(tt, 4, 1, false, dkgTypeTrivialHT) })
	t.Run("n=10,f=3,unreliable,dkgTypeTrivialHT", func(tt *testing.T) { testGeneric(tt, 10, 3, false, dkgTypeTrivialHT) })
}

func testGeneric(t *testing.T, n, f int, reliable bool, dkgType byte) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Create a fake network and keys for the tests.
	peerNetIDs, peerIdentities := testpeers.SetupKeys(uint16(n))
	peerPubKeys := make([]*cryptolib.PublicKey, len(peerIdentities))
	for i := range peerPubKeys {
		peerPubKeys[i] = peerIdentities[i].GetPublicKey()
	}
	var networkBehaviour testutil.PeeringNetBehavior
	if reliable {
		networkBehaviour = testutil.NewPeeringNetReliable(log)
	} else {
		networkBehaviour = testutil.NewPeeringNetUnreliable(80, 20, 10*time.Millisecond, 200*time.Millisecond, log)
	}
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		peerNetIDs, peerIdentities, 10000,
		networkBehaviour,
		testlogger.WithLevel(log, logger.LevelWarn, false),
	)
	defer peeringNetwork.Close()
	var networkProviders []peering.NetworkProvider = peeringNetwork.NetworkProviders()
	peeringID := peering.RandomPeeringID()
	//
	// Initialize the DSS subsystem in each node / chain.
	dssNodes := make([]dss_node.DSSNode, len(peerIdentities))
	for i := range peerIdentities {
		dssNodes[i] = dss_node.New(&peeringID, networkProviders[i], peerIdentities[i], log.Named(fmt.Sprintf("dssNode#%v", i)))
	}
	defer func() {
		for _, n := range dssNodes {
			n.Close()
		}
	}()
	dkShares := longTermDKG(t, dkgType, peerIdentities, f, log)
	//
	//	Start the DSS instances.
	key := hashing.HashData([]byte{1, 2, 3}).String()
	outPartChs := make([]chan []int, len(dssNodes))
	outPartVals := make([][]int, len(dssNodes))
	outSigChs := make([]chan []byte, len(dssNodes))
	outSigVals := make([][]byte, len(dssNodes))
	index := 0
	for i := range dssNodes {
		outPartChs[i] = make(chan []int, 1)
		outSigChs[i] = make(chan []byte, 1)
		ii := i
		require.NoError(t, dssNodes[i].Start(key, index, dkShares[i],
			func(part []int) { outPartChs[ii] <- part },
			func(sig []byte) { outSigChs[ii] <- sig },
		))
	}
	//
	// Wait for partial outputs and submit the decisions.
	for i := range dssNodes {
		outPartVals[i] = <-outPartChs[i]
		t.Logf("DSS: PartialOutput[%v]=%v", i, outPartVals[i])
	}
	messageToSign := []byte{112, 117, 116, 105, 110, 32, 99, 104, 117, 105, 108, 111}
	for i := range dssNodes {
		require.NoError(t, dssNodes[i].DecidedIndexProposals(key, index, outPartVals, messageToSign))
	}
	//
	// Wait for partial outputs.
	for i := range dssNodes {
		outSigVals[i] = <-outSigChs[i]
	}
	for i := range outSigVals {
		require.NotNil(t, outSigVals[i])
		t.Logf("DSS: FinalSignature[%v]=%v", i, outSigVals[i])
		assert.True(t, bytes.Equal(outSigVals[0], outSigVals[i]))
		assert.NoError(t, eddsa.Verify(dkShares[i].DSSSharedPublic(), messageToSign, outSigVals[i]))
	}
}

func longTermDKG(t *testing.T, dkgType byte, peerIdentities []*cryptolib.KeyPair, f int, log *logger.Logger) []tcrypto.DKShare {
	switch dkgType {
	case dkgTypeRobustLT:
		return longTermDKGRobustLT(t, peerIdentities, f, log)
	case dkgTypeTrivialHT:
		return longTermDKGTrivialHT(t, peerIdentities, f)
	}
	panic("unknown dkg type")
}

func longTermDKGRobustLT(t *testing.T, peerIdentities []*cryptolib.KeyPair, f int, log *logger.Logger) []tcrypto.DKShare {
	dkShares := make([]tcrypto.DKShare, len(peerIdentities))

	peerPubKeys := make([]*cryptolib.PublicKey, len(peerIdentities))
	for i := range peerPubKeys {
		peerPubKeys[i] = peerIdentities[i].GetPublicKey()
	}

	nodeIDs := make([]gpa.NodeID, len(peerIdentities))
	nodePKs := map[gpa.NodeID]kyber.Point{}
	nodeSKs := map[gpa.NodeID]kyber.Scalar{}
	for i := range nodeIDs {
		nodeIDs[i] = gpa.NodeID(peerIdentities[i].GetPublicKey().String())
		nodeKeyPair, err := peerIdentities[i].GetPrivateKey().AsKyberKeyPair()
		require.NoError(t, err)
		nodePKs[nodeIDs[i]] = nodeKeyPair.Public
		nodeSKs[nodeIDs[i]] = nodeKeyPair.Private
	}

	longTermPK, longTermSecretShares := adkg.MakeTestDistributedKey(t, tcrypto.DefaultEd25519Suite(), nodeIDs, nodeSKs, nodePKs, f, log)

	publicKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	sharedAddress := iotago.Ed25519AddressFromPubKey(publicKey)

	for i := range dkShares {
		dkShares[i] = &fakeDKShare{
			id:              util.NewComparableAddress(&sharedAddress),
			nodePubKeys:     peerPubKeys,
			index:           uint16(i),
			dssSecretShare:  longTermSecretShares[nodeIDs[i]],
			dssSharedPublic: longTermPK,
		}
	}
	return dkShares
}

func longTermDKGTrivialHT(t *testing.T, peerIdentities []*cryptolib.KeyPair, f int) []tcrypto.DKShare {
	sharedAddress, dkShareRegistryProviders := testpeers.SetupDkgTrivial(t, len(peerIdentities), f, peerIdentities, nil)
	dkShares := make([]tcrypto.DKShare, len(peerIdentities))
	for i := range dkShareRegistryProviders {
		dkShare, err := dkShareRegistryProviders[i].LoadDKShare(sharedAddress)
		require.NoError(t, err)
		dkShares[i] = dkShare
	}
	return dkShares
}

type fakeDKShare struct {
	id              *util.ComparableAddress
	nodePubKeys     []*cryptolib.PublicKey
	index           uint16
	dssSecretShare  tcrypto.SecretShare
	dssSharedPublic kyber.Point
}

var _ tcrypto.DKShare = &fakeDKShare{}

func (f *fakeDKShare) ID() *util.ComparableAddress {
	return f.id
}

func (f *fakeDKShare) Clone() onchangemap.Item[string, *util.ComparableAddress] {
	panic("not important")
}

func (f *fakeDKShare) Bytes() []byte {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) GetAddress() iotago.Address {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) GetIndex() *uint16 {
	return &f.index
}

func (f *fakeDKShare) GetN() uint16 {
	return uint16(len(f.nodePubKeys))
}

func (f *fakeDKShare) GetT() uint16 {
	cmtN := f.GetN()
	cmtF := (cmtN - 1) / 3
	cmtT := cmtN - cmtF
	return cmtT
}

func (f *fakeDKShare) GetNodePubKeys() []*cryptolib.PublicKey {
	return f.nodePubKeys
}

func (f *fakeDKShare) GetSharedPublic() *cryptolib.PublicKey {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) SetPublicShares(edPublicShares, blsPublicShares []kyber.Point) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) DSSPublicShares() []kyber.Point {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) DSSSharedPublic() kyber.Point {
	return f.dssSharedPublic
}

func (f *fakeDKShare) DSSSignShare(data []byte, nonce tcrypto.SecretShare) (*dss.PartialSig, error) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) DSSVerifySigShare(data []byte, sigShare *dss.PartialSig) error {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) DSSRecoverMasterSignature(sigShares []*dss.PartialSig, data []byte, nonce tcrypto.SecretShare) ([]byte, error) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) DSSVerifyMasterSignature(data, signature []byte) error {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) DSSSecretShare() tcrypto.SecretShare {
	return f.dssSecretShare
}

func (f *fakeDKShare) BLSThreshold() uint16 {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSSharedPublic() kyber.Point {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSPublicShares() []kyber.Point {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSSignShare(data []byte) (tbls.SigShare, error) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSVerifySigShare(data []byte, sigShare tbls.SigShare) error {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSRecoverMasterSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSVerifyMasterSignature(data, signature []byte) error {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSSign(data []byte) ([]byte, error) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSVerify(signer kyber.Point, data, signature []byte) error {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSCommits() *share.PubPoly {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) BLSPriShare() *share.PriShare {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) AssignNodePubKeys(nodePubKeys []*cryptolib.PublicKey) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) AssignCommonData(dks tcrypto.DKShare) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) ClearCommonData() {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) MarshalJSON() ([]byte, error) {
	panic(xerrors.New("not important"))
}

func (f *fakeDKShare) UnmarshalJSON(bytes []byte) error {
	panic(xerrors.New("not important"))
}

type fakeSecretShare struct {
	priShare    *share.PriShare
	commitments []kyber.Point
}

var _ tcrypto.SecretShare = &fakeSecretShare{}

func (f *fakeSecretShare) PriShare() *share.PriShare {
	return f.priShare
}

func (f *fakeSecretShare) Commitments() []kyber.Point {
	return f.commitments
}
