// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"io"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

// NodeOwnershipCertificate is a proof that a specified address is an owner of the specified node.
// It is implemented as a signature over the node pub key concatenated with the owner address.
type NodeOwnershipCertificate []byte

func NewNodeOwnershipCertificate(nodeKeyPair *cryptolib.KeyPair, ownerAddress *cryptolib.Address) NodeOwnershipCertificate {
	ww := rwutil.NewBytesWriter()
	ww.Write(nodeKeyPair.GetPublicKey())
	ww.Write(ownerAddress)
	return nodeKeyPair.GetPrivateKey().Sign(ww.Bytes())
}

func NodeOwnershipCertificateFromBytes(data []byte) NodeOwnershipCertificate {
	return data
}

func (c NodeOwnershipCertificate) Verify(nodePubKey *cryptolib.PublicKey, ownerAddress *cryptolib.Address) bool {
	ww := rwutil.NewBytesWriter()
	ww.Write(nodePubKey)
	ww.Write(ownerAddress)
	return nodePubKey.Verify(ww.Bytes(), c.Bytes())
}

func (c NodeOwnershipCertificate) Bytes() []byte {
	return c
}

// AccessNodeData conveys all the information that is maintained
// on the governance SC about a specific node.
type AccessNodeData struct {
	ValidatorAddr *cryptolib.Address       // Address of the validator owning the node. Not sent via parameters.
	Certificate   NodeOwnershipCertificate // Proof that Validator owns the Node.
	ForCommittee  bool                     // true, if Node should be a candidate to a committee.
	AccessAPI     string                   // API URL, if any.
}

func (a *AccessNodeData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(a.ValidatorAddr)
	ww.WriteBytes(a.Certificate)
	ww.WriteBool(a.ForCommittee)
	ww.WriteString(a.AccessAPI)
	return ww.Err
}

func (a *AccessNodeData) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	a.ValidatorAddr = rwutil.ReadStruct(rr, new(cryptolib.Address))
	a.Certificate = rr.ReadBytes()
	a.ForCommittee = rr.ReadBool()
	a.AccessAPI = rr.ReadString()
	return rr.Err
}

// AccessNodeInfo conveys all the information that is maintained
// on the governance SC about a specific node.
type AccessNodeInfo struct {
	NodePubKey *cryptolib.PublicKey // Public Key of the node. Stored as a key in the SC State and Params.
	AccessNodeData
}

func (a *AccessNodeInfo) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(a.NodePubKey)
	ww.Write(&a.AccessNodeData)
	return ww.Err
}

func (a *AccessNodeInfo) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(a.NodePubKey)
	rr.Read(&a.AccessNodeData)
	return rr.Err
}

var errInvalidCertificate = coreerrors.Register("invalid certificate").Create()

func (a *AccessNodeInfo) AddCertificate(nodeKeyPair *cryptolib.KeyPair, ownerAddress *cryptolib.Address) *AccessNodeInfo {
	a.Certificate = NewNodeOwnershipCertificate(nodeKeyPair, ownerAddress).Bytes()
	return a
}

type ChangeAccessNodeAction byte

const (
	ChangeAccessNodeActionRemove = ChangeAccessNodeAction(iota)
	ChangeAccessNodeActionAccept
	ChangeAccessNodeActionDrop
	ChangeAccessNodeActionLast
)

type ChangeAccessNodeRequests []*ChangeAccessNodeRequest

func NewChangeAccessNodeRequests() ChangeAccessNodeRequests {
	return nil
}

func (req ChangeAccessNodeRequests) Remove(pubKey *cryptolib.PublicKey) ChangeAccessNodeRequests {
	return append(req, &ChangeAccessNodeRequest{
		PublicKey: pubKey,
		Action:    ChangeAccessNodeActionRemove,
	})
}

func (req ChangeAccessNodeRequests) Accept(pubKey *cryptolib.PublicKey) ChangeAccessNodeRequests {
	return append(req, &ChangeAccessNodeRequest{
		PublicKey: pubKey,
		Action:    ChangeAccessNodeActionAccept,
	})
}

func (req ChangeAccessNodeRequests) Drop(pubKey *cryptolib.PublicKey) ChangeAccessNodeRequests {
	return append(req, &ChangeAccessNodeRequest{
		PublicKey: pubKey,
		Action:    ChangeAccessNodeActionDrop,
	})
}

type ChangeAccessNodeRequest struct {
	PublicKey *cryptolib.PublicKey
	Action    ChangeAccessNodeAction
}

func ChangeAccessNodesRequestFromBytes(b []byte) (*ChangeAccessNodeRequest, error) {
	return rwutil.ReadFromBytes(b, new(ChangeAccessNodeRequest))
}

func (c *ChangeAccessNodeRequest) Bytes() []byte {
	return rwutil.WriteToBytes(c)
}

func (c *ChangeAccessNodeRequest) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	c.PublicKey = rwutil.ReadStruct(rr, new(cryptolib.PublicKey))
	c.Action = ChangeAccessNodeAction(rr.ReadByte())
	return rr.Err
}

func (c *ChangeAccessNodeRequest) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(c.PublicKey)
	ww.WriteByte(byte(c.Action))
	return ww.Err
}
