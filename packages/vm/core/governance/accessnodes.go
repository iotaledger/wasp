// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

// NodeOwnershipCertificate is a proof that a specified address is an owner of the specified node.
// It is implemented as a signature over the node pub key concatenated with the owner address.
type NodeOwnershipCertificate []byte

func NewNodeOwnershipCertificate(nodeKeyPair *cryptolib.KeyPair, ownerAddress iotago.Address) NodeOwnershipCertificate {
	ww := rwutil.NewBytesWriter()
	ww.Write(nodeKeyPair.GetPublicKey())
	isc.AddressToWriter(ww, ownerAddress)
	return nodeKeyPair.GetPrivateKey().Sign(ww.Bytes())
}

func NodeOwnershipCertificateFromBytes(data []byte) NodeOwnershipCertificate {
	return data
}

func (c NodeOwnershipCertificate) Verify(nodePubKey *cryptolib.PublicKey, ownerAddress iotago.Address) bool {
	ww := rwutil.NewBytesWriter()
	ww.Write(nodePubKey)
	isc.AddressToWriter(ww, ownerAddress)
	return nodePubKey.Verify(ww.Bytes(), c.Bytes())
}

func (c NodeOwnershipCertificate) Bytes() []byte {
	return c
}

// AccessNodeInfo conveys all the information that is maintained
// on the governance SC about a specific node.
type AccessNodeInfo struct {
	NodePubKey    []byte // Public Key of the node. Stored as a key in the SC State and Params.
	validatorAddr []byte // Address of the validator owning the node. Not sent via parameters.
	Certificate   []byte // Proof that Validator owns the Node.
	ForCommittee  bool   // true, if Node should be a candidate to a committee.
	AccessAPI     string // API URL, if any.
}

func AccessNodeInfoFromBytes(pubKey, data []byte) (*AccessNodeInfo, error) {
	rr := rwutil.NewBytesReader(data)
	return &AccessNodeInfo{
		NodePubKey:    pubKey,
		validatorAddr: rr.ReadBytes(),
		Certificate:   rr.ReadBytes(),
		ForCommittee:  rr.ReadBool(),
		AccessAPI:     rr.ReadString(),
	}, rr.Err
}

var (
	errInvalidCertificate      = coreerrors.Register("invalid certificate").Create()
	errSenderMustHaveL1Address = coreerrors.Register("sender must have L1 address").Create()
)

func AccessNodeInfoWithValidatorAddress(ctx isc.Sandbox, ani *AccessNodeInfo) *AccessNodeInfo {
	validatorAddr, _ := isc.AddressFromAgentID(ctx.Request().SenderAccount()) // Not from params, to have it validated.
	if validatorAddr == nil {
		panic(errSenderMustHaveL1Address)
	}
	ani.validatorAddr = codec.Address.Encode(validatorAddr)
	if !ani.ValidateCertificate() {
		panic(errInvalidCertificate)
	}
	return ani
}

func (a *AccessNodeInfo) Bytes() []byte {
	ww := rwutil.NewBytesWriter()
	ww.WriteBytes(a.validatorAddr)
	ww.WriteBytes(a.Certificate)
	ww.WriteBool(a.ForCommittee)
	ww.WriteString(a.AccessAPI)
	return ww.Bytes()
}

func (a *AccessNodeInfo) AddCertificate(nodeKeyPair *cryptolib.KeyPair, ownerAddress iotago.Address) *AccessNodeInfo {
	a.Certificate = NewNodeOwnershipCertificate(nodeKeyPair, ownerAddress).Bytes()
	return a
}

func (a *AccessNodeInfo) ValidateCertificate() bool {
	nodePubKey, err := cryptolib.PublicKeyFromBytes(a.NodePubKey)
	if err != nil {
		return false
	}
	validatorAddr, err := isc.AddressFromBytes(a.validatorAddr)
	if err != nil {
		return false
	}
	cert := NodeOwnershipCertificateFromBytes(a.Certificate)
	return cert.Verify(nodePubKey, validatorAddr)
}

// GetChainNodesResponse
type GetChainNodesResponse struct {
	AccessNodeCandidates map[cryptolib.PublicKeyKey]*AccessNodeInfo // Application info for the AccessNodes.
	AccessNodes          map[cryptolib.PublicKeyKey]struct{}        // Public Keys of Access Nodes.
}

//
//	ChangeAccessNodesRequest
//

type ChangeAccessNodeAction byte

const (
	ChangeAccessNodeActionRemove = ChangeAccessNodeAction(iota)
	ChangeAccessNodeActionAccept
	ChangeAccessNodeActionDrop
	ChangeAccessNodeActionLast
)

type ChangeAccessNodesRequest map[cryptolib.PublicKeyKey]ChangeAccessNodeAction

func NewChangeAccessNodesRequest() ChangeAccessNodesRequest {
	return ChangeAccessNodesRequest(make(map[cryptolib.PublicKeyKey]ChangeAccessNodeAction))
}

func (req ChangeAccessNodesRequest) Remove(pubKey *cryptolib.PublicKey) ChangeAccessNodesRequest {
	req[pubKey.AsKey()] = ChangeAccessNodeActionRemove
	return req
}

func (req ChangeAccessNodesRequest) Accept(pubKey *cryptolib.PublicKey) ChangeAccessNodesRequest {
	req[pubKey.AsKey()] = ChangeAccessNodeActionAccept
	return req
}

func (req ChangeAccessNodesRequest) Drop(pubKey *cryptolib.PublicKey) ChangeAccessNodesRequest {
	req[pubKey.AsKey()] = ChangeAccessNodeActionDrop
	return req
}
