// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// TODO: BDN is actually not used (only functions used, that delegate to the tbls directly). Update to use it!

package tcrypto

import (
	"bytes"
	"io"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/sign/eddsa"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/crypto/bls"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util"
)

// secretShare is an implementation for dss.DistKeyShare.
type secretShare struct {
	priShare    *share.PriShare
	commitments []kyber.Point
}

func (d *secretShare) PriShare() *share.PriShare {
	return d.priShare
}

func (d *secretShare) Commitments() []kyber.Point {
	return d.commitments
}

// dkShareImpl stands for the information stored on
// a node as a result of the DKG procedure.
type dkShareImpl struct {
	address     iotago.Address
	index       *uint16 // nil, if the current node is not a member of a group sharing the key.
	n           uint16
	t           uint16
	nodePrivKey *cryptolib.PrivateKey // Transient.
	nodePubKeys []*cryptolib.PublicKey
	//
	// Shares for the Schnorr signatures (for L1).
	edSuite         suites.Suite // Transient, only needed for un-marshaling.
	edSharedPublic  kyber.Point
	edPublicCommits []kyber.Point
	edPublicShares  []kyber.Point
	edPrivateShare  kyber.Scalar
	//
	// Shares for the randomness in the consensus et al.
	blsSuite         Suite // Transient, only needed for un-marshaling.
	blsSharedPublic  kyber.Point
	blsPublicCommits []kyber.Point
	blsPublicShares  []kyber.Point
	blsPrivateShare  kyber.Scalar
}

var _ DKShare = &dkShareImpl{}

// NewDKShare creates new share of the key.
func NewDKShare(
	index uint16,
	n uint16,
	t uint16,
	nodePrivKey *cryptolib.PrivateKey,
	nodePubKeys []*cryptolib.PublicKey,
	edSuite suites.Suite,
	edSharedPublic kyber.Point,
	edPublicCommits []kyber.Point,
	edPublicShares []kyber.Point,
	edPrivateShare kyber.Scalar,
	blsSuite Suite,
	blsSharedPublic kyber.Point,
	blsPublicCommits []kyber.Point,
	blsPublicShares []kyber.Point,
	blsPrivateShare kyber.Scalar,
) (DKShare, error) {
	//
	// Derive the ChainID.
	pubBytes, err := edSharedPublic.MarshalBinary()
	if err != nil {
		return nil, err
	}
	sharedAddress := iotago.Ed25519AddressFromPubKey(pubBytes)
	//
	// Construct the DKShare.
	dkShare := dkShareImpl{
		address:          &sharedAddress,
		index:            &index,
		n:                n,
		t:                t,
		nodePrivKey:      nodePrivKey,
		nodePubKeys:      nodePubKeys,
		edSuite:          edSuite,
		edSharedPublic:   edSharedPublic,
		edPublicCommits:  edPublicCommits,
		edPublicShares:   edPublicShares,
		edPrivateShare:   edPrivateShare,
		blsSuite:         blsSuite,
		blsSharedPublic:  blsSharedPublic,
		blsPublicCommits: blsPublicCommits,
		blsPublicShares:  blsPublicShares,
		blsPrivateShare:  blsPrivateShare,
	}
	return &dkShare, nil
}

// NewDKSharePublic creates a DKShare containing only the publicly accessible information.
func NewDKSharePublic(
	sharedAddress iotago.Address,
	n uint16,
	t uint16,
	nodePrivKey *cryptolib.PrivateKey,
	nodePubKeys []*cryptolib.PublicKey,
	edSuite suites.Suite,
	edSharedPublic kyber.Point,
	edPublicShares []kyber.Point,
	blsSuite Suite,
	blsSharedPublic kyber.Point,
	blsPublicShares []kyber.Point,
) DKShare {
	s := dkShareImpl{
		address:          sharedAddress,
		index:            nil, // Not meaningful in this case.
		n:                n,
		t:                t,
		nodePrivKey:      nodePrivKey,
		nodePubKeys:      nodePubKeys,
		edSuite:          edSuite,
		edSharedPublic:   edSharedPublic,
		edPublicCommits:  nil, // Not meaningful in this case.
		edPublicShares:   edPublicShares,
		edPrivateShare:   nil, // Not meaningful in this case.
		blsSuite:         blsSuite,
		blsSharedPublic:  blsSharedPublic,
		blsPublicCommits: nil, // Not meaningful in this case.
		blsPublicShares:  blsPublicShares,
		blsPrivateShare:  nil, // Not meaningful in this case.
	}
	return &s
}

// DKShareFromBytes reads DKShare from bytes.
func DKShareFromBytes(buf []byte, edSuite suites.Suite, blsSuite Suite, nodePrivKey *cryptolib.PrivateKey) (DKShare, error) {
	r := bytes.NewReader(buf)
	s := dkShareImpl{nodePrivKey: nodePrivKey, edSuite: edSuite, blsSuite: blsSuite}
	if err := s.Read(r); err != nil {
		return nil, err
	}
	return &s, nil
}

// Bytes returns byte representation of the share.
func (s *dkShareImpl) Bytes() []byte {
	var buf bytes.Buffer
	if err := s.Write(&buf); err != nil {
		panic(xerrors.Errorf("DKShare.Bytes: %w", err))
	}
	return buf.Bytes()
}

// Write returns byte representation of this struct.
//

//nolint:gocyclo
func (s *dkShareImpl) Write(w io.Writer) error {
	var err error
	//
	// Common attributes.
	addressType := s.address.Type()
	addressBytes, err := s.address.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return xerrors.Errorf("cannot serialize an address: %w", err)
	}
	if err := util.WriteByte(w, byte(addressType)); err != nil {
		return err
	}
	if err := util.WriteBytes16(w, addressBytes); err != nil {
		return err
	}
	if err := util.WriteUint16(w, *s.index); err != nil { // It must be not nil here.
		return err
	}
	if err := util.WriteUint16(w, s.n); err != nil {
		return err
	}
	if err := util.WriteUint16(w, s.t); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(s.nodePubKeys))); err != nil {
		return err
	}
	for _, nodePubKey := range s.nodePubKeys {
		if err := util.WriteBytes16(w, nodePubKey.AsBytes()); err != nil {
			return err
		}
	}
	//
	// Ed25519 part of the key shares.
	if err := util.WriteMarshaled(w, s.edSharedPublic); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(s.edPublicCommits))); err != nil {
		return err
	}
	for i := 0; i < len(s.edPublicCommits); i++ {
		if err := util.WriteMarshaled(w, s.edPublicCommits[i]); err != nil {
			return err
		}
	}
	if err := util.WriteUint16(w, uint16(len(s.edPublicShares))); err != nil {
		return err
	}
	for i := 0; i < len(s.edPublicShares); i++ {
		if err := util.WriteMarshaled(w, s.edPublicShares[i]); err != nil {
			return err
		}
	}
	if err := util.WriteMarshaled(w, s.edPrivateShare); err != nil {
		return err
	}
	//
	// BLS part of the key shares.
	if err := util.WriteMarshaled(w, s.blsSharedPublic); err != nil {
		return err
	}
	if err := util.WriteUint16(w, uint16(len(s.blsPublicCommits))); err != nil {
		return err
	}
	for i := 0; i < len(s.blsPublicCommits); i++ {
		if err := util.WriteMarshaled(w, s.blsPublicCommits[i]); err != nil {
			return err
		}
	}
	if err := util.WriteUint16(w, uint16(len(s.blsPublicShares))); err != nil {
		return err
	}
	for i := 0; i < len(s.blsPublicShares); i++ {
		if err := util.WriteMarshaled(w, s.blsPublicShares[i]); err != nil {
			return err
		}
	}
	if err := util.WriteMarshaled(w, s.blsPrivateShare); err != nil {
		return err
	}
	return nil
}

func (s *dkShareImpl) Read(r io.Reader) error {
	var err error
	var arrLen uint16
	//
	// Common attributes.
	var addressTypeByte byte
	var addressBytes []byte
	if addressTypeByte, err = util.ReadByte(r); err != nil {
		return err
	}
	if addressBytes, err = util.ReadBytes16(r); err != nil {
		return err
	}
	s.address, err = iotago.AddressSelector(uint32(addressTypeByte))
	if err != nil {
		return err
	}
	if _, err = s.address.Deserialize(addressBytes, serializer.DeSeriModeNoValidation, nil); err != nil {
		return err
	}
	var index uint16
	if err := util.ReadUint16(r, &index); err != nil {
		return err
	}
	s.index = &index
	if err := util.ReadUint16(r, &s.n); err != nil {
		return err
	}
	if err := util.ReadUint16(r, &s.t); err != nil {
		return err
	}
	//
	// NodePubKeys
	if err := util.ReadUint16(r, &arrLen); err != nil {
		return err
	}
	s.nodePubKeys = make([]*cryptolib.PublicKey, arrLen)
	for i := range s.nodePubKeys {
		var nodePubKeyBin []byte
		var nodePubKey *cryptolib.PublicKey
		if nodePubKeyBin, err = util.ReadBytes16(r); err != nil {
			return err
		}
		if nodePubKey, err = cryptolib.NewPublicKeyFromBytes(nodePubKeyBin); err != nil {
			return err
		}
		s.nodePubKeys[i] = nodePubKey
	}
	//
	// DSS / Ed25519 shares.
	if err := s.readDSSAttrs(r); err != nil {
		return err
	}
	//
	// BLS Shares.
	if err := s.readBLSAttrs(r); err != nil {
		return err
	}
	return nil
}

// Read function was split just to make the linter happy.
func (s *dkShareImpl) readDSSAttrs(r io.Reader) error {
	var arrLen uint16
	s.edSharedPublic = s.edSuite.Point()
	if err := util.ReadMarshaled(r, s.edSharedPublic); err != nil {
		return err
	}
	//
	// Ed25519 shares: PublicCommits
	if err := util.ReadUint16(r, &arrLen); err != nil {
		return err
	}
	s.edPublicCommits = make([]kyber.Point, arrLen)
	for i := uint16(0); i < arrLen; i++ {
		s.edPublicCommits[i] = s.edSuite.Point()
		if err := util.ReadMarshaled(r, s.edPublicCommits[i]); err != nil {
			return err
		}
	}
	//
	// Ed25519 shares: PublicShares
	if err := util.ReadUint16(r, &arrLen); err != nil {
		return err
	}
	s.edPublicShares = make([]kyber.Point, arrLen)
	for i := uint16(0); i < arrLen; i++ {
		s.edPublicShares[i] = s.edSuite.Point()
		if err := util.ReadMarshaled(r, s.edPublicShares[i]); err != nil {
			return err
		}
	}
	//
	// Ed25519 shares: Private share.
	s.edPrivateShare = s.edSuite.Scalar()
	if err := util.ReadMarshaled(r, s.edPrivateShare); err != nil {
		return err
	}
	return nil
}

// Read function was split just to make the linter happy.
func (s *dkShareImpl) readBLSAttrs(r io.Reader) error {
	var arrLen uint16
	s.blsSharedPublic = s.blsSuite.G2().Point()
	if err := util.ReadMarshaled(r, s.blsSharedPublic); err != nil {
		return err
	}
	//
	// BLS shares: PublicCommits
	if err := util.ReadUint16(r, &arrLen); err != nil {
		return err
	}
	s.blsPublicCommits = make([]kyber.Point, arrLen)
	for i := uint16(0); i < arrLen; i++ {
		s.blsPublicCommits[i] = s.blsSuite.G2().Point()
		if err := util.ReadMarshaled(r, s.blsPublicCommits[i]); err != nil {
			return err
		}
	}
	//
	// BLS shares: PublicShares
	if err := util.ReadUint16(r, &arrLen); err != nil {
		return err
	}
	s.blsPublicShares = make([]kyber.Point, arrLen)
	for i := uint16(0); i < arrLen; i++ {
		s.blsPublicShares[i] = s.blsSuite.G2().Point()
		if err := util.ReadMarshaled(r, s.blsPublicShares[i]); err != nil {
			return err
		}
	}
	//
	// BLS shares: Private share.
	s.blsPrivateShare = s.blsSuite.G2().Scalar()
	if err := util.ReadMarshaled(r, s.blsPrivateShare); err != nil {
		return err
	}
	return nil
}

func (s *dkShareImpl) GetAddress() iotago.Address {
	return s.address
}

func (s *dkShareImpl) GetIndex() *uint16 {
	return s.index
}

func (s *dkShareImpl) GetN() uint16 {
	return s.n
}

func (s *dkShareImpl) GetT() uint16 {
	return s.t
}

func (s *dkShareImpl) GetNodePubKeys() []*cryptolib.PublicKey {
	return s.nodePubKeys
}

func (s *dkShareImpl) SetPublicShares(edPublicShares, blsPublicShares []kyber.Point) {
	s.edPublicShares = edPublicShares
	s.blsPublicShares = blsPublicShares
}

func (s *dkShareImpl) GetSharedPublic() *cryptolib.PublicKey {
	pubKeyBytes, err := s.edSharedPublic.MarshalBinary()
	if err != nil {
		panic(xerrors.Errorf("cannot convert kyber.Point to cryptolib.PublicKey, failed to serialize: %w", err))
	}
	pubKeyCL, err := cryptolib.NewPublicKeyFromBytes(pubKeyBytes)
	if err != nil {
		panic(xerrors.Errorf("cannot convert kyber.Point to cryptolib.PublicKey, failed to deserialize: %w", err))
	}
	return pubKeyCL
}

//////////////////// Schnorr based signatures.

func (s *dkShareImpl) DSSSharedPublic() kyber.Point {
	return s.edSharedPublic
}

func (s *dkShareImpl) DSSPublicShares() []kyber.Point {
	return s.edPublicShares
}

// SignShare signs the data with the own key share.
// returns SigShare, which contains signature and the index
func (s *dkShareImpl) DSSSignShare(data []byte, nonce SecretShare) (*dss.PartialSig, error) {
	if s.n == 1 {
		// Do not use the DSS in the case of a single node.
		sig, err := schnorr.Sign(s.edSuite, s.edPrivateShare, data)
		if err != nil {
			return nil, err
		}
		partSig := dss.PartialSig{
			Partial: &share.PriShare{ // TODO: Do not provide it to outside.
				I: 0,
				V: s.edSuite.Scalar(),
			},
			SessionID: []byte{},
			Signature: sig,
		}
		return &partSig, nil
	}
	signer, err := s.makeSigner(data, nonce)
	if err != nil {
		return nil, err
	}
	psi, err := signer.PartialSig()
	if err != nil {
		return nil, err
	}
	// TODO: maybe we have to serialize it here, to avoid spreading the specific types everywhere?
	return psi, nil
}

// VerifySigShare verifies the signature of a particular share.
func (s *dkShareImpl) DSSVerifySigShare(data []byte, sigshare *dss.PartialSig) error {
	// TODO: Is that working?
	return dss.Verify(s.edPublicShares[sigshare.Partial.I], data, sigshare.Signature)
}

// RecoverMasterSignature generates (recovers) master signature from partial sigshares.
// returns signature as defined in the value Tangle
func (s *dkShareImpl) DSSRecoverMasterSignature(sigShares []*dss.PartialSig, data []byte, nonce SecretShare) ([]byte, error) {
	if s.n == 1 {
		// Use a regular signature in the case of single node.
		// The signature is stored in the share.
		return sigShares[0].Signature, nil
	}
	signer, err := s.makeSigner(data, nonce)
	if err != nil {
		return nil, xerrors.Errorf("cannot create DSS object: %w", err)
	}
	for i := range sigShares {
		err = signer.ProcessPartialSig(sigShares[i])
		if err != nil {
			return nil, xerrors.Errorf("cannot process partial signature: %w", err)
		}
	}
	if !signer.EnoughPartialSig() {
		return nil, xerrors.Errorf("not enough partial signatures")
	}
	aggregatedSig, err := signer.Signature()
	if err != nil {
		return nil, xerrors.Errorf("cannot aggregate signature: %w", err)
	}
	return aggregatedSig, nil
}

// VerifyMasterSignature checks signature against master public key
// NOTE: Not used.
func (s *dkShareImpl) DSSVerifyMasterSignature(data, signature []byte) error {
	return dss.Verify(s.edSharedPublic, data, signature)
}

func (s *dkShareImpl) DSSSecretShare() SecretShare {
	return &secretShare{
		priShare: &share.PriShare{
			I: int(*s.index),
			V: s.edPrivateShare,
		},
		commitments: s.edPublicCommits,
	}
}

func (s *dkShareImpl) makeSigner(data []byte, nonce SecretShare) (*dss.DSS, error) {
	priKeyDKS := s.DSSSecretShare()
	nodePrivKey := eddsa.EdDSA{}
	if err := nodePrivKey.UnmarshalBinary(s.nodePrivKey.AsBytes()); err != nil {
		return nil, xerrors.Errorf("cannot convert node priv key to kyber scalar: %w", err)
	}
	participants := make([]kyber.Point, len(s.nodePubKeys))
	for i := range s.nodePubKeys {
		participants[i] = s.edSuite.Point()
		if err := participants[i].UnmarshalBinary(s.nodePubKeys[i].AsBytes()); err != nil {
			return nil, xerrors.Errorf("cannot convert node public key to kyber point: %w", err)
		}
	}
	return dss.NewDSS(s.edSuite, nodePrivKey.Secret, participants, priKeyDKS, nonce, data, int(s.t))
}

///////////////////////// BLS based signatures.

func (s *dkShareImpl) BLSSharedPublic() kyber.Point {
	return s.blsSharedPublic
}

func (s *dkShareImpl) BLSPublicShares() []kyber.Point {
	return s.blsPublicShares
}

// BLSSignShare signs the data with the own key share.
// returns SigShare, which contains signature and the index
func (s *dkShareImpl) BLSSignShare(data []byte) (tbls.SigShare, error) {
	priShare := share.PriShare{
		I: int(*s.index),
		V: s.blsPrivateShare,
	}
	return tbls.Sign(s.blsSuite, &priShare, data)
}

// BLSVerifySigShare verifies the signature of a particular share.
func (s *dkShareImpl) BLSVerifySigShare(data []byte, sigshare tbls.SigShare) error {
	idx, err := sigshare.Index()
	if err != nil || idx >= int(s.n) || idx < 0 {
		return err
	}
	return bdn.Verify(s.blsSuite, s.blsPublicShares[idx], data, sigshare.Value())
}

// BLSRecoverFullSignature generates (recovers) master signature from partial sigshares.
// returns signature as defined in the value Tangle
func (s *dkShareImpl) BLSRecoverMasterSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error) {
	var err error
	var recoveredSignatureBin []byte
	if s.n > 1 {
		pubPoly := share.NewPubPoly(s.blsSuite, nil, s.blsPublicCommits)
		recoveredSignatureBin, err = tbls.Recover(s.blsSuite, pubPoly, data, sigShares, int(s.t), int(s.n))
		if err != nil {
			return nil, err
		}
	} else {
		singleSigShare := tbls.SigShare(sigShares[0])
		recoveredSignatureBin = singleSigShare.Value()
	}
	sig, _, err := bls.SignatureFromBytes(recoveredSignatureBin)
	if err != nil {
		return nil, err
	}
	ret := bls.NewSignatureWithPublicKey(bls.PublicKey{Point: s.blsSharedPublic}, sig)
	return &ret, nil
}

// BLSVerifyMasterSignature checks signature against master public key
// NOTE: Not used. // TODO: Has to be used.
func (s *dkShareImpl) BLSVerifyMasterSignature(data, signature []byte) error {
	return bdn.Verify(s.blsSuite, s.blsSharedPublic, data, signature)
}

// BLSSign considers partial key as a key and signs the specified message.
func (s *dkShareImpl) BLSSign(data []byte) ([]byte, error) {
	return bdn.Sign(s.blsSuite, s.blsPrivateShare, data)
}

// BLSVerify checks a signature made with BLSSign. It ignores the threshold sig aspects.
func (s *dkShareImpl) BLSVerify(signer kyber.Point, data, signature []byte) error {
	return bdn.Verify(s.blsSuite, signer, data, signature)
}

///////////////////////// Test support functions.

func (s *dkShareImpl) AssignNodePubKeys(nodePubKeys []*cryptolib.PublicKey) {
	s.nodePubKeys = nodePubKeys
}

func (s *dkShareImpl) AssignCommonData(dks DKShare) {
	src := dks.(*dkShareImpl)
	s.edPublicCommits = src.edPublicCommits
	s.edPublicShares = src.edPublicShares
	s.blsPublicCommits = src.blsPublicCommits
	s.blsPublicShares = src.blsPublicShares
	s.nodePubKeys = src.nodePubKeys
}

func (s *dkShareImpl) ClearCommonData() {
	s.edPublicCommits = make([]kyber.Point, 0)
	s.edPublicShares = make([]kyber.Point, 0)
	s.blsPublicCommits = make([]kyber.Point, 0)
	s.blsPublicShares = make([]kyber.Point, 0)
	s.nodePubKeys = make([]*cryptolib.PublicKey, 0)
}
