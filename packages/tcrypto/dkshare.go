// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// TODO: BDN is actually not used (only functions used, that delegate to the tbls directly). Update to use it!

package tcrypto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/dss"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/onchangemap"
	"github.com/iotaledger/wasp/v2/packages/tcrypto/bls"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

// secretShareImpl is an implementation for SecretShare.
type secretShareImpl struct {
	priShare    *share.PriShare
	commitments []kyber.Point
	nodeCount   int
	threshold   int
}

var _ SecretShare = &secretShareImpl{}

func NewDistKeyShare(priShare *share.PriShare, commitments []kyber.Point, nodeCount, threshold int) SecretShare {
	return newDistKeyShare(priShare, commitments, nodeCount, threshold)
}

func newDistKeyShare(priShare *share.PriShare, commitments []kyber.Point, nodeCount, threshold int) *secretShareImpl {
	return &secretShareImpl{
		priShare:    priShare,
		commitments: commitments,
		nodeCount:   nodeCount,
		threshold:   threshold,
	}
}

func (d *secretShareImpl) NodeCount() int {
	return d.nodeCount
}

// F = N - T.
func (d *secretShareImpl) MaxFaulty() int {
	return d.NodeCount() - d.Threshold()
}

func (d *secretShareImpl) Threshold() int {
	return d.threshold
}

func (d *secretShareImpl) Clone() *secretShareImpl {
	return newDistKeyShare(
		&share.PriShare{
			I: d.priShare.I,
			V: d.priShare.V.Clone(),
		},
		util.CloneSlice(d.commitments),
		d.nodeCount,
		d.threshold,
	)
}

func (d *secretShareImpl) PriShare() *share.PriShare {
	return d.priShare
}

func (d *secretShareImpl) Commitments() []kyber.Point {
	return d.commitments
}

// dkShareImpl stands for the information stored on
// a node as a result of the DKG procedure.
type dkShareImpl struct {
	address     *util.ComparableAddress
	index       *uint16 // nil, if the current node is not a member of a group sharing the key.
	n           uint16
	t           uint16
	nodePrivKey *cryptolib.PrivateKey // Transient.
	nodePubKeys []*cryptolib.PublicKey
	//
	// Shares for the Schnorr signatures (for L1).
	edSuite         suites.Suite // Used for unmarshalling and signing
	edSharedPublic  kyber.Point
	edPublicCommits []kyber.Point
	edPublicShares  []kyber.Point
	edPrivateShare  kyber.Scalar
	//
	// Shares for the randomness in the consensus et al.
	blsSuite         Suite  // Used for unmarshalling, signing and verification
	blsThreshold     uint16 // BLS Threshold has to be low (F+1)
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
	blsThreshold uint16,
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
	// sharedAddress := iotago.Ed25519AddressFromPubKey(pubBytes)
	publicKey, _ := cryptolib.PublicKeyFromBytes(pubBytes)
	sharedAddress := publicKey.AsAddress()
	//
	// Construct the DKShare.
	dkShare := dkShareImpl{
		address:          util.NewComparableAddress(sharedAddress),
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
		blsThreshold:     blsThreshold,
		blsSharedPublic:  blsSharedPublic,
		blsPublicCommits: blsPublicCommits,
		blsPublicShares:  blsPublicShares,
		blsPrivateShare:  blsPrivateShare,
	}
	return &dkShare, nil
}

func NewEmptyDKShare(nodePrivKey *cryptolib.PrivateKey, edSuite suites.Suite, blsSuite Suite) DKShare {
	return &dkShareImpl{
		nodePrivKey: nodePrivKey,
		edSuite:     edSuite,
		blsSuite:    blsSuite,
	}
}

// NewDKSharePublic creates a DKShare containing only the publicly accessible information.
func NewDKSharePublic(
	sharedAddress *cryptolib.Address,
	n uint16,
	t uint16,
	nodePrivKey *cryptolib.PrivateKey,
	nodePubKeys []*cryptolib.PublicKey,
	edSuite suites.Suite,
	edSharedPublic kyber.Point,
	edPublicShares []kyber.Point,
	blsSuite Suite,
	blsThreshold uint16,
	blsSharedPublic kyber.Point,
	blsPublicShares []kyber.Point,
) DKShare {
	s := dkShareImpl{
		address:          util.NewComparableAddress(sharedAddress),
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
		blsThreshold:     blsThreshold,
		blsSharedPublic:  blsSharedPublic,
		blsPublicCommits: nil, // Not meaningful in this case.
		blsPublicShares:  blsPublicShares,
		blsPrivateShare:  nil, // Not meaningful in this case.
	}
	return &s
}

func (s *dkShareImpl) ID() *util.ComparableAddress {
	return s.address
}

func (s *dkShareImpl) Clone() onchangemap.Item[cryptolib.AddressKey, *util.ComparableAddress] {
	index := *s.index

	return &dkShareImpl{
		address:          util.NewComparableAddress(s.GetAddress().Clone()),
		index:            &index,
		n:                s.n,
		t:                s.t,
		nodePrivKey:      s.nodePrivKey.Clone(),
		nodePubKeys:      util.CloneSlice(s.nodePubKeys),
		edSuite:          s.edSuite,
		edSharedPublic:   s.edSharedPublic.Clone(),
		edPublicCommits:  util.CloneSlice(s.edPublicCommits),
		edPublicShares:   util.CloneSlice(s.edPublicShares),
		edPrivateShare:   s.edPrivateShare.Clone(),
		blsSuite:         s.blsSuite,
		blsThreshold:     s.blsThreshold,
		blsSharedPublic:  s.blsSharedPublic.Clone(),
		blsPublicCommits: util.CloneSlice(s.blsPublicCommits),
		blsPublicShares:  util.CloneSlice(s.blsPublicShares),
		blsPrivateShare:  s.blsPrivateShare.Clone(),
	}
}

// DKShareFromBytes reads DKShare from bytes.
func DKShareFromBytes(buf []byte, edSuite suites.Suite, blsSuite Suite, nodePrivKey *cryptolib.PrivateKey) (DKShare, error) {
	s := &dkShareImpl{nodePrivKey: nodePrivKey, edSuite: edSuite, blsSuite: blsSuite}
	return rwutil.ReadFromBytes(buf, s)
}

// Bytes returns byte representation of the share.
func (s *dkShareImpl) Bytes() []byte {
	return rwutil.WriteToBytes(s)
}

func (s *dkShareImpl) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	address := cryptolib.NewEmptyAddress()
	rr.Read(address)
	if rr.Err == nil {
		s.address = util.NewComparableAddress(address)
	}

	index := rr.ReadUint16()
	s.index = &index
	s.n = rr.ReadUint16()
	s.t = rr.ReadUint16()

	size := rr.ReadSize16()
	s.nodePubKeys = make([]*cryptolib.PublicKey, size)
	for i := range s.nodePubKeys {
		s.nodePubKeys[i] = cryptolib.NewEmptyPublicKey()
		rr.Read(s.nodePubKeys[i])
	}

	// DSS / Ed25519 part of the key shares.
	edSuite := s.edSuite
	s.edSharedPublic = cryptolib.PointFromReader(rr, edSuite)
	size = rr.ReadSize16()
	s.edPublicCommits = make([]kyber.Point, size)
	for i := range s.edPublicCommits {
		s.edPublicCommits[i] = cryptolib.PointFromReader(rr, edSuite)
	}
	size = rr.ReadSize16()
	s.edPublicShares = make([]kyber.Point, size)
	for i := range s.edPublicShares {
		s.edPublicShares[i] = cryptolib.PointFromReader(rr, edSuite)
	}
	s.edPrivateShare = cryptolib.ScalarFromReader(rr, edSuite)

	// BLS part of the key shares.
	blsGroup := s.blsSuite.G2()
	s.blsThreshold = rr.ReadUint16()
	s.blsSharedPublic = cryptolib.PointFromReader(rr, blsGroup)
	size = rr.ReadSize16()
	s.blsPublicCommits = make([]kyber.Point, size)
	for i := range s.blsPublicCommits {
		s.blsPublicCommits[i] = cryptolib.PointFromReader(rr, blsGroup)
	}
	size = rr.ReadSize16()
	s.blsPublicShares = make([]kyber.Point, size)
	for i := range s.blsPublicShares {
		s.blsPublicShares[i] = cryptolib.PointFromReader(rr, blsGroup)
	}
	s.blsPrivateShare = cryptolib.ScalarFromReader(rr, blsGroup)
	return rr.Err
}

func (s *dkShareImpl) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(s.address.Address())

	ww.WriteUint16(*s.index)
	ww.WriteUint16(s.n)
	ww.WriteUint16(s.t)

	ww.WriteSize16(len(s.nodePubKeys))
	for _, nodePubKey := range s.nodePubKeys {
		ww.Write(nodePubKey)
	}

	// DSS / Ed25519 part of the key shares.
	cryptolib.PointToWriter(ww, s.edSharedPublic)
	ww.WriteSize16(len(s.edPublicCommits))
	for i := 0; i < len(s.edPublicCommits); i++ {
		cryptolib.PointToWriter(ww, s.edPublicCommits[i])
	}
	ww.WriteSize16(len(s.edPublicShares))
	for i := 0; i < len(s.edPublicShares); i++ {
		cryptolib.PointToWriter(ww, s.edPublicShares[i])
	}
	cryptolib.ScalarToWriter(ww, s.edPrivateShare)

	// BLS part of the key shares.
	ww.WriteUint16(s.blsThreshold)
	cryptolib.PointToWriter(ww, s.blsSharedPublic)
	ww.WriteSize16(len(s.blsPublicCommits))
	for i := 0; i < len(s.blsPublicCommits); i++ {
		cryptolib.PointToWriter(ww, s.blsPublicCommits[i])
	}
	ww.WriteSize16(len(s.blsPublicShares))
	for i := 0; i < len(s.blsPublicShares); i++ {
		cryptolib.PointToWriter(ww, s.blsPublicShares[i])
	}
	cryptolib.ScalarToWriter(ww, s.blsPrivateShare)
	return ww.Err
}

func (s *dkShareImpl) GetAddress() *cryptolib.Address {
	return s.address.Address()
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
		panic(fmt.Errorf("cannot convert kyber.Point to cryptolib.PublicKey, failed to serialize: %w", err))
	}
	pubKeyCL, err := cryptolib.PublicKeyFromBytes(pubKeyBytes)
	if err != nil {
		panic(fmt.Errorf("cannot convert kyber.Point to cryptolib.PublicKey, failed to deserialize: %w", err))
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
		return nil, fmt.Errorf("cannot create DSS object: %w", err)
	}
	for i := range sigShares {
		err = signer.ProcessPartialSig(sigShares[i])
		if err != nil {
			return nil, fmt.Errorf("cannot process partial signature: %w", err)
		}
	}
	if !signer.EnoughPartialSig() {
		return nil, errors.New("not enough partial signatures")
	}
	aggregatedSig, err := signer.Signature()
	if err != nil {
		return nil, fmt.Errorf("cannot aggregate signature: %w", err)
	}
	return aggregatedSig, nil
}

// VerifyMasterSignature checks signature against master public key
// NOTE: Not used.
func (s *dkShareImpl) DSSVerifyMasterSignature(data, signature []byte) error {
	return dss.Verify(s.edSharedPublic, data, signature)
}

func (s *dkShareImpl) DSS() SecretShare {
	return newDistKeyShare( // TODO: Use a single instance.
		&share.PriShare{
			I: int(*s.index),
			V: s.edPrivateShare.Clone(),
		},
		util.CloneSlice(s.edPublicCommits),
		int(s.n),
		int(s.t),
	)
}

func (s *dkShareImpl) makeSigner(data []byte, nonce SecretShare) (*dss.DSS, error) {
	priKeyDKS := s.DSS()
	nodeKyberKeyPair, err := s.nodePrivKey.AsKyberKeyPair()
	if err != nil {
		return nil, fmt.Errorf("cannot convert node priv key to kyber scalar: %w", err)
	}
	participants := make([]kyber.Point, len(s.nodePubKeys))
	for i := range s.nodePubKeys {
		participants[i], err = s.nodePubKeys[i].AsKyberPoint()
		if err != nil {
			return nil, fmt.Errorf("cannot convert node public key to kyber point: %w", err)
		}
	}
	return dss.NewDSS(s.edSuite, nodeKyberKeyPair.Private, participants, priKeyDKS, nonce, data, int(s.t))
}

///////////////////////// BLS based signatures.

func (s *dkShareImpl) BLSThreshold() uint16 {
	return s.blsThreshold
}

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

// BLSRecoverMasterSignature generates (recovers) master signature from partial sigshares.
// returns signature as defined in the value Tangle
func (s *dkShareImpl) BLSRecoverMasterSignature(sigShares [][]byte, data []byte) (*bls.SignatureWithPublicKey, error) {
	var err error
	var recoveredSignatureBin []byte
	if s.n > 1 {
		pubPoly := share.NewPubPoly(s.blsSuite, nil, s.blsPublicCommits)
		recoveredSignatureBin, err = tbls.Recover(s.blsSuite, pubPoly, data, sigShares, int(s.blsThreshold), int(s.n))
		if err != nil {
			return nil, err
		}
	} else {
		singleSigShare := tbls.SigShare(sigShares[0])
		recoveredSignatureBin = singleSigShare.Value()
	}
	sig, err := bls.SignatureFromBytes(recoveredSignatureBin)
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

// Needed for signatures outside of this object.
func (s *dkShareImpl) BLSCommits() *share.PubPoly {
	return share.NewPubPoly(s.blsSuite, nil, s.blsPublicCommits)
}

// Needed for signatures outside of this object.
func (s *dkShareImpl) BLSPriShare() *share.PriShare {
	return &share.PriShare{I: int(*s.index), V: s.blsPrivateShare}
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

type jsonKeyShares struct {
	SharedPublic  string   `json:"sharedPublic"`
	PublicCommits []string `json:"publicCommits"`
	PublicShares  []string `json:"publicShares"`
	PrivateShare  string   `json:"privateShare"`
}

type jsonDKShares struct {
	Address      string         `json:"address"`
	Index        uint16         `json:"index"`
	N            uint16         `json:"n"`
	T            uint16         `json:"t"`
	NodePubKeys  []string       `json:"nodePubKeys"`
	Ed25519      *jsonKeyShares `json:"ed25519"`
	BlsThreshold uint16         `json:"blsThreshold"`
	BLS          *jsonKeyShares `json:"bls"`
}

func DecodeHexKyberPoint(group kyber.Group, dataHex string) (kyber.Point, error) {
	point := group.Point()
	if err := util.DecodeHexBinaryMarshaled(dataHex, point); err != nil {
		return nil, err
	}

	return point, nil
}

func DecodeHexKyberScalar(group kyber.Group, dataHex string) (kyber.Scalar, error) {
	scalar := group.Scalar()
	if err := util.DecodeHexBinaryMarshaled(dataHex, scalar); err != nil {
		return nil, err
	}

	return scalar, nil
}

func DecodeHexKyberPoints(group kyber.Group, dataHex []string) ([]kyber.Point, error) {
	results := make([]kyber.Point, len(dataHex))

	for i, hex := range dataHex {
		point, err := DecodeHexKyberPoint(group, hex)
		if err != nil {
			return nil, err
		}
		results[i] = point
	}

	return results, nil
}

func (s *dkShareImpl) MarshalJSON() ([]byte, error) {
	jAddress := s.address.Address().String()

	nodePubKeys := make([]string, 0)
	for _, nodePubKey := range s.nodePubKeys {
		nodePubKeys = append(nodePubKeys, nodePubKey.String())
	}

	ed25519SharedPublicHex, err := util.EncodeHexBinaryMarshaled(s.edSharedPublic)
	if err != nil {
		return nil, err
	}

	ed25519PublicCommitsHex, err := util.EncodeSliceHexBinaryMarshaled(s.edPublicCommits)
	if err != nil {
		return nil, err
	}

	ed25519PublicSharesHex, err := util.EncodeSliceHexBinaryMarshaled(s.edPublicShares)
	if err != nil {
		return nil, err
	}

	ed25519PrivateShareHex, err := util.EncodeHexBinaryMarshaled(s.edPrivateShare)
	if err != nil {
		return nil, err
	}

	blsSharedPublicHex, err := util.EncodeHexBinaryMarshaled(s.blsSharedPublic)
	if err != nil {
		return nil, err
	}

	blsPublicCommitsHex, err := util.EncodeSliceHexBinaryMarshaled(s.blsPublicCommits)
	if err != nil {
		return nil, err
	}

	blsPublicSharesHex, err := util.EncodeSliceHexBinaryMarshaled(s.blsPublicShares)
	if err != nil {
		return nil, err
	}

	blsPrivateShareHex, err := util.EncodeHexBinaryMarshaled(s.blsPrivateShare)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&jsonDKShares{
		Address:     jAddress,
		Index:       *s.index,
		N:           s.n,
		T:           s.t,
		NodePubKeys: nodePubKeys,
		Ed25519: &jsonKeyShares{
			SharedPublic:  ed25519SharedPublicHex,
			PublicCommits: ed25519PublicCommitsHex,
			PublicShares:  ed25519PublicSharesHex,
			PrivateShare:  ed25519PrivateShareHex,
		},
		BlsThreshold: s.blsThreshold,
		BLS: &jsonKeyShares{
			SharedPublic:  blsSharedPublicHex,
			PublicCommits: blsPublicCommitsHex,
			PublicShares:  blsPublicSharesHex,
			PrivateShare:  blsPrivateShareHex,
		},
	})
}

// ATTENTION: edSuite and blsSuite need to be initialized already.
// Use NewEmptyDKShare for init.
func (s *dkShareImpl) UnmarshalJSON(bytes []byte) error {
	j := &jsonDKShares{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}

	address, err := cryptolib.NewAddressFromHexString(j.Address)
	if err != nil {
		return err
	}
	s.address = util.NewComparableAddress(address)

	s.index = &j.Index
	s.n = j.N
	s.t = j.T
	s.blsThreshold = j.BlsThreshold

	s.nodePubKeys = make([]*cryptolib.PublicKey, len(j.NodePubKeys))
	for i, nodePubKeyHex := range j.NodePubKeys {
		nodePubKey, err2 := cryptolib.PublicKeyFromString(nodePubKeyHex)
		if err2 != nil {
			return err2
		}

		s.nodePubKeys[i] = nodePubKey
	}

	s.edSharedPublic, err = DecodeHexKyberPoint(s.edSuite, j.Ed25519.SharedPublic)
	if err != nil {
		return err
	}

	s.edPublicCommits, err = DecodeHexKyberPoints(s.edSuite, j.Ed25519.PublicCommits)
	if err != nil {
		return err
	}

	s.edPublicShares, err = DecodeHexKyberPoints(s.edSuite, j.Ed25519.PublicShares)
	if err != nil {
		return err
	}

	s.edPrivateShare, err = DecodeHexKyberScalar(s.edSuite, j.Ed25519.PrivateShare)
	if err != nil {
		return err
	}

	s.blsSharedPublic, err = DecodeHexKyberPoint(s.blsSuite.G2(), j.BLS.SharedPublic)
	if err != nil {
		return err
	}

	s.blsPublicCommits, err = DecodeHexKyberPoints(s.blsSuite.G2(), j.BLS.PublicCommits)
	if err != nil {
		return err
	}

	s.blsPublicShares, err = DecodeHexKyberPoints(s.blsSuite.G2(), j.BLS.PublicShares)
	if err != nil {
		return err
	}

	s.blsPrivateShare, err = DecodeHexKyberScalar(s.blsSuite.G2(), j.BLS.PrivateShare)
	if err != nil {
		return err
	}

	return nil
}
