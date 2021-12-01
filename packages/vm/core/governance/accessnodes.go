// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"bytes"
	"crypto/ed25519"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type AccessNodeInfo struct {
	PubKey    []byte
	Validator bool
	API       string
}

func NewAccessNodeInfoFromBytes(pubKey, value []byte) (*AccessNodeInfo, error) {
	var a AccessNodeInfo
	var err error
	r := bytes.NewReader(value)
	if err := util.ReadBoolByte(r, &a.Validator); err != nil {
		return nil, xerrors.Errorf("failed to read AccessNodeInfo.Validator: %v", err)
	}
	if a.API, err = util.ReadString16(r); err != nil {
		return nil, xerrors.Errorf("failed to read AccessNodeInfo.API: %v", err)
	}
	a.PubKey = pubKey
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
	if err := util.WriteBoolByte(&w, a.Validator); err != nil {
		panic(xerrors.Errorf("failed to write AccessNodeInfo.Validator: %v", err))
	}
	if err := util.WriteString16(&w, a.API); err != nil {
		panic(xerrors.Errorf("failed to write AccessNodeInfo.Validator: %v", err))
	}
	return w.Bytes()
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

	ac := collections.NewMapReadOnly(d, ParamGetChainNodesAccessNodeCandidates)
	ac.MustIterate(func(pubKey, value []byte) bool {
		ani, err := NewAccessNodeInfoFromBytes(pubKey, value)
		if err != nil {
			panic(xerrors.Errorf("unable to decode access node info: %v", err))
		}
		res.AccessNodeCandidates = append(res.AccessNodeCandidates, ani)
		return true
	})

	an := collections.NewMapReadOnly(d, ParamGetChainNodesAccessNodes)
	an.MustIterate(func(pubKey, value []byte) bool {
		res.AccessNodes = append(res.AccessNodes, ed25519.PublicKey(pubKey))
		return true
	})
	return &res
}

//
// CandidateNodeRequest
//
type CandidateNodeRequest struct {
	Candidate bool
	Validator bool
	PubKey    []byte
	Cert      []byte
	API       string
}

func (req CandidateNodeRequest) AsDict() dict.Dict {
	d := dict.New()
	d.Set(ParamCandidateNodeCandidate, codec.EncodeBool(req.Candidate))
	d.Set(ParamCandidateNodeValidator, codec.EncodeBool(req.Validator))
	d.Set(ParamCandidateNodePubKey, req.PubKey)
	d.Set(ParamCandidateNodeCert, req.Cert)
	d.Set(ParamCandidateNodeAPI, codec.EncodeString(req.API))
	return d
}

//
//	ChangeAccessNodesRequest
//
type ChangeAccessNodesRequest struct {
	actions map[string]byte
}

func NewChangeAccessNodesRequest() *ChangeAccessNodesRequest {
	return &ChangeAccessNodesRequest{
		actions: make(map[string]byte),
	}
}

func (req *ChangeAccessNodesRequest) Remove(pubKey ed25519.PublicKey) *ChangeAccessNodesRequest {
	req.actions[string(pubKey)] = 0
	return req
}

func (req *ChangeAccessNodesRequest) Accept(pubKey ed25519.PublicKey) *ChangeAccessNodesRequest {
	req.actions[string(pubKey)] = 1
	return req
}

func (req *ChangeAccessNodesRequest) Drop(pubKey ed25519.PublicKey) *ChangeAccessNodesRequest {
	req.actions[string(pubKey)] = 2
	return req
}

func (req *ChangeAccessNodesRequest) AsDict() dict.Dict {
	d := dict.New()
	actionsMap := collections.NewMap(d, ParamChangeAccessNodesActions)
	for pubKey, action := range req.actions {
		actionsMap.MustSetAt([]byte(pubKey), []byte{action})
	}
	return d
}
