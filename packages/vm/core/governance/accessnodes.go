// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
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

func (a *AccessNodeInfo) Bytes() []byte {
	ww := rwutil.NewBytesWriter()
	ww.WriteBytes(a.validatorAddr)
	ww.WriteBytes(a.Certificate)
	ww.WriteBool(a.ForCommittee)
	ww.WriteString(a.AccessAPI)
	return ww.Bytes()
}

var errSenderMustHaveL1Address = coreerrors.Register("sender must have L1 address").Create()

func AccessNodeInfoFromAddCandidateNodeParams(ctx isc.Sandbox) *AccessNodeInfo {
	validatorAddr, _ := isc.AddressFromAgentID(ctx.Request().SenderAccount()) // Not from params, to have it validated.
	if validatorAddr == nil {
		panic(errSenderMustHaveL1Address)
	}
	params := ctx.Params()
	ani := AccessNodeInfo{
		NodePubKey:    params.MustGetBytes(ParamAccessNodeInfoPubKey),
		validatorAddr: isc.AddressToBytes(validatorAddr),
		Certificate:   params.MustGetBytes(ParamAccessNodeInfoCertificate),
		ForCommittee:  params.MustGetBool(ParamAccessNodeInfoForCommittee, false),
		AccessAPI:     params.MustGetString(ParamAccessNodeInfoAccessAPI, ""),
	}
	return &ani
}

func (a *AccessNodeInfo) ToAddCandidateNodeParams() dict.Dict {
	d := dict.New()
	d.Set(ParamAccessNodeInfoForCommittee, codec.EncodeBool(a.ForCommittee))
	d.Set(ParamAccessNodeInfoPubKey, a.NodePubKey)
	d.Set(ParamAccessNodeInfoCertificate, a.Certificate)
	d.Set(ParamAccessNodeInfoAccessAPI, codec.EncodeString(a.AccessAPI))
	return d
}

func AccessNodeInfoFromRevokeAccessNodeParams(ctx isc.Sandbox) *AccessNodeInfo {
	validatorAddr, _ := isc.AddressFromAgentID(ctx.Request().SenderAccount()) // Not from params, to have it validated.
	if validatorAddr == nil {
		panic(errSenderMustHaveL1Address)
	}
	params := ctx.Params()
	ani := AccessNodeInfo{
		NodePubKey:    params.MustGetBytes(ParamAccessNodeInfoPubKey),
		validatorAddr: isc.AddressToBytes(validatorAddr), // Not from params, to have it validated.
		Certificate:   params.MustGetBytes(ParamAccessNodeInfoCertificate),
	}
	return &ani
}

func (a *AccessNodeInfo) ToRevokeAccessNodeParams() dict.Dict {
	d := dict.New()
	d.Set(ParamAccessNodeInfoPubKey, a.NodePubKey)
	d.Set(ParamAccessNodeInfoCertificate, a.Certificate)
	return d
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

// GetChainNodesRequest
type GetChainNodesRequest struct{}

func (req GetChainNodesRequest) AsDict() dict.Dict {
	return dict.New()
}

// GetChainNodesResponse
type GetChainNodesResponse struct {
	AccessNodeCandidates []*AccessNodeInfo      // Application info for the AccessNodes.
	AccessNodes          []*cryptolib.PublicKey // Public Keys of Access Nodes.
}

func GetChainNodesResponseFromDict(d dict.Dict) *GetChainNodesResponse {
	res := GetChainNodesResponse{
		AccessNodeCandidates: make([]*AccessNodeInfo, 0),
		AccessNodes:          make([]*cryptolib.PublicKey, 0),
	}

	ac := collections.NewMapReadOnly(d, ParamGetChainNodesAccessNodeCandidates)
	ac.Iterate(func(pubKey, value []byte) bool {
		ani, err := AccessNodeInfoFromBytes(pubKey, value)
		if err != nil {
			panic(fmt.Errorf("unable to decode access node info: %w", err))
		}
		res.AccessNodeCandidates = append(res.AccessNodeCandidates, ani)
		return true
	})

	an := collections.NewMapReadOnly(d, ParamGetChainNodesAccessNodes)
	an.Iterate(func(pubKeyBin, value []byte) bool {
		publicKey, err := cryptolib.PublicKeyFromBytes(pubKeyBin)
		if err != nil {
			panic(fmt.Errorf("unable to decode public key: %w", err))
		}
		res.AccessNodes = append(res.AccessNodes, publicKey)
		return true
	})
	return &res
}

//
//	ChangeAccessNodesRequest
//

type ChangeAccessNodeAction byte

const (
	ChangeAccessNodeActionRemove = ChangeAccessNodeAction(iota)
	ChangeAccessNodeActionAccept
	ChangeAccessNodeActionDrop
)

type ChangeAccessNodesRequest struct {
	actions map[cryptolib.PublicKeyKey]ChangeAccessNodeAction
}

func NewChangeAccessNodesRequest() *ChangeAccessNodesRequest {
	return &ChangeAccessNodesRequest{
		actions: make(map[cryptolib.PublicKeyKey]ChangeAccessNodeAction),
	}
}

func (req *ChangeAccessNodesRequest) Remove(pubKey *cryptolib.PublicKey) *ChangeAccessNodesRequest {
	req.actions[pubKey.AsKey()] = ChangeAccessNodeActionRemove
	return req
}

func (req *ChangeAccessNodesRequest) Accept(pubKey *cryptolib.PublicKey) *ChangeAccessNodesRequest {
	req.actions[pubKey.AsKey()] = ChangeAccessNodeActionAccept
	return req
}

func (req *ChangeAccessNodesRequest) Drop(pubKey *cryptolib.PublicKey) *ChangeAccessNodesRequest {
	req.actions[pubKey.AsKey()] = ChangeAccessNodeActionDrop
	return req
}

func (req *ChangeAccessNodesRequest) AsDict() dict.Dict {
	d := dict.New()
	actionsMap := collections.NewMap(d, ParamChangeAccessNodesActions)
	for pubKey, action := range req.actions {
		actionsMap.SetAt(pubKey[:], []byte{byte(action)})
	}
	return d
}
