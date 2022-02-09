// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"bytes"
	"crypto"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/xerrors"
)

// NodeOwnershipCertificate is a proof that a specified address is an owner of the specified node.
// It is implemented as a signature over the node pub key concatenated with the owner address.
type NodeOwnershipCertificate []byte

func NewNodeOwnershipCertificate(nodeKeyPair *cryptolib.KeyPair, ownerAddress iotago.Address) NodeOwnershipCertificate {
	certData := bytes.Buffer{}
	certData.Write(nodeKeyPair.PublicKey)
	certData.Write(iscp.BytesFromAddress(ownerAddress))
	result, err := nodeKeyPair.PrivateKey.Sign(nil, certData.Bytes(), crypto.Hash(0))
	if err != nil {
		panic(err)
	}
	return result
}

func NewNodeOwnershipCertificateFromBytes(data []byte) NodeOwnershipCertificate {
	return NodeOwnershipCertificate(data)
}

func (c NodeOwnershipCertificate) Verify(nodePubKey ed25519.PublicKey, ownerAddress iotago.Address) bool {
	certData := bytes.Buffer{}
	certData.Write(nodePubKey)
	certData.Write(iscp.BytesFromAddress(ownerAddress))
	return cryptolib.Verify(nodePubKey, certData.Bytes(), c.Bytes())
}

func (c NodeOwnershipCertificate) Bytes() []byte {
	return c
}

// AccessNodeInfo conveys all the information that is maintained
// on the governance SC about a specific node.
type AccessNodeInfo struct {
	NodePubKey    []byte // Public Key of the node. Stored as a key in the SC State and Params.
	ValidatorAddr []byte // Address of the validator owning the node. Not sent via parameters.
	Certificate   []byte // Proof that Validator owns the Node.
	ForCommittee  bool   // true, if Node should be a candidate to a committee.
	AccessAPI     string // API URL, if any.
}

func NewAccessNodeInfoFromBytes(pubKey, value []byte) (*AccessNodeInfo, error) {
	var a AccessNodeInfo
	var err error
	r := bytes.NewReader(value)
	a.NodePubKey = pubKey // NodePubKey stored as a map key.
	if a.ValidatorAddr, err = util.ReadBytes16(r); err != nil {
		return nil, xerrors.Errorf("failed to read AccessNodeInfo.ValidatorAddr: %v", err)
	}
	if a.Certificate, err = util.ReadBytes16(r); err != nil {
		return nil, xerrors.Errorf("failed to read AccessNodeInfo.Certificate: %v", err)
	}
	if err := util.ReadBoolByte(r, &a.ForCommittee); err != nil {
		return nil, xerrors.Errorf("failed to read AccessNodeInfo.ForCommittee: %v", err)
	}
	if a.AccessAPI, err = util.ReadString16(r); err != nil {
		return nil, xerrors.Errorf("failed to read AccessNodeInfo.AccessAPI: %v", err)
	}
	return &a, nil
}

func NewAccessNodeInfoListFromMap(infoMap *collections.ImmutableMap) ([]*AccessNodeInfo, error) {
	res := make([]*AccessNodeInfo, 0)
	var accErr error
	err := infoMap.Iterate(func(elemKey, value []byte) bool {
		var a *AccessNodeInfo
		if a, accErr = NewAccessNodeInfoFromBytes(elemKey, value); accErr != nil {
			return false
		}
		res = append(res, a)
		return true
	})
	if accErr != nil {
		return nil, xerrors.Errorf("failed to iterate over AccessNodeInfo list: %v", accErr)
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to iterate over AccessNodeInfo list: %v", err)
	}
	return res, nil
}

func (a *AccessNodeInfo) Bytes() []byte {
	w := bytes.Buffer{}
	// NodePubKey stored as a map key.
	if err := util.WriteBytes16(&w, a.ValidatorAddr); err != nil {
		panic(xerrors.Errorf("failed to write AccessNodeInfo.ValidatorAddr: %v", err))
	}
	if err := util.WriteBytes16(&w, a.Certificate); err != nil {
		panic(xerrors.Errorf("failed to write AccessNodeInfo.Certificate: %v", err))
	}
	if err := util.WriteBoolByte(&w, a.ForCommittee); err != nil {
		panic(xerrors.Errorf("failed to write AccessNodeInfo.ForCommittee: %v", err))
	}
	if err := util.WriteString16(&w, a.AccessAPI); err != nil {
		panic(xerrors.Errorf("failed to write AccessNodeInfo.AccessAPI: %v", err))
	}
	return w.Bytes()
}

func NewAccessNodeInfoFromAddCandidateNodeParams(ctx iscp.Sandbox) *AccessNodeInfo {
	ani := AccessNodeInfo{
		NodePubKey:    ctx.ParamDecoder().MustGetBytes(ParamAccessNodeInfoPubKey),
		ValidatorAddr: iscp.BytesFromAddress(ctx.Request().SenderAddress()), // Not from params, to have it validated.
		Certificate:   ctx.ParamDecoder().MustGetBytes(ParamAccessNodeInfoCertificate),
		ForCommittee:  ctx.ParamDecoder().MustGetBool(ParamAccessNodeInfoForCommittee, false),
		AccessAPI:     ctx.ParamDecoder().MustGetString(ParamAccessNodeInfoAccessAPI, ""),
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

func NewAccessNodeInfoFromRevokeAccessNodeParams(ctx iscp.Sandbox) *AccessNodeInfo {
	ani := AccessNodeInfo{
		NodePubKey:    ctx.ParamDecoder().MustGetBytes(ParamAccessNodeInfoPubKey),
		ValidatorAddr: iscp.BytesFromAddress(ctx.Request().SenderAddress()), // Not from params, to have it validated.
		Certificate:   ctx.ParamDecoder().MustGetBytes(ParamAccessNodeInfoCertificate),
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

func (a *AccessNodeInfo) ValidateCertificate(ctx iscp.Sandbox) bool {
	nodePubKey, err := cryptolib.PublicKeyFromBytes(a.NodePubKey)
	if err != nil {
		return false
	}
	validatorAddr, _, err := iscp.AddressFromBytes(a.ValidatorAddr)
	if err != nil {
		return false
	}
	cert := NewNodeOwnershipCertificateFromBytes(a.Certificate)
	return cert.Verify(nodePubKey, validatorAddr)
}

//
// GetChainNodesRequest
//
type GetChainNodesRequest struct{}

func (req GetChainNodesRequest) AsDict() dict.Dict {
	return dict.New()
}

//
// GetChainNodesResponse
//
type GetChainNodesResponse struct {
	AccessNodeCandidates []*AccessNodeInfo   // Application info for the AccessNodes.
	AccessNodes          []ed25519.PublicKey // Public Keys of Access Nodes.
}

func NewGetChainNodesResponseFromDict(d dict.Dict) *GetChainNodesResponse {
	res := GetChainNodesResponse{
		AccessNodeCandidates: make([]*AccessNodeInfo, 0),
		AccessNodes:          make([]ed25519.PublicKey, 0),
	}

	ac := collections.NewMapReadOnly(d, string(ParamGetChainNodesAccessNodeCandidates))
	ac.MustIterate(func(pubKey, value []byte) bool {
		ani, err := NewAccessNodeInfoFromBytes(pubKey, value)
		if err != nil {
			panic(xerrors.Errorf("unable to decode access node info: %v", err))
		}
		res.AccessNodeCandidates = append(res.AccessNodeCandidates, ani)
		return true
	})

	an := collections.NewMapReadOnly(d, string(ParamGetChainNodesAccessNodes))
	an.MustIterate(func(pubKeyBin, value []byte) bool {
		res.AccessNodes = append(res.AccessNodes, pubKeyBin)
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

type fixedSizePubKey [ed25519.PublicKeySize]byte // needed because iotago pub key is a byte slice ([]byte) and cannot be used as a key to an array

func getFixedSizePubKey(pubKey []byte) fixedSizePubKey {
	var ret fixedSizePubKey
	copy(ret[:], pubKey)
	return ret
}

type ChangeAccessNodesRequest struct {
	actions map[fixedSizePubKey]ChangeAccessNodeAction
}

func NewChangeAccessNodesRequest() *ChangeAccessNodesRequest {
	return &ChangeAccessNodesRequest{
		actions: make(map[fixedSizePubKey]ChangeAccessNodeAction),
	}
}

func (req *ChangeAccessNodesRequest) Remove(pubKey ed25519.PublicKey) *ChangeAccessNodesRequest {
	req.actions[getFixedSizePubKey(pubKey)] = ChangeAccessNodeActionRemove
	return req
}

func (req *ChangeAccessNodesRequest) Accept(pubKey ed25519.PublicKey) *ChangeAccessNodesRequest {
	req.actions[getFixedSizePubKey(pubKey)] = ChangeAccessNodeActionAccept
	return req
}

func (req *ChangeAccessNodesRequest) Drop(pubKey ed25519.PublicKey) *ChangeAccessNodesRequest {
	req.actions[getFixedSizePubKey(pubKey)] = ChangeAccessNodeActionDrop
	return req
}

func (req *ChangeAccessNodesRequest) AsDict() dict.Dict {
	d := dict.New()
	actionsMap := collections.NewMap(d, string(ParamChangeAccessNodesActions))
	for pubKey, action := range req.actions {
		actionsMap.MustSetAt(pubKey[:], []byte{byte(action)})
	}
	return d
}
