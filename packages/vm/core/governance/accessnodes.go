// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"io"

	"github.com/samber/lo"

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

// AccessNodeInfo conveys all the information that is maintained
// on the governance SC about a specific node.
type AccessNodeInfo struct {
	NodePubKey *cryptolib.PublicKey // Public Key of the node. Stored as a key in the SC State and Params.
	AccessNodeData
}

var errInvalidCertificate = coreerrors.Register("invalid certificate").Create()

func (a *AccessNodeInfo) AddCertificate(nodeKeyPair *cryptolib.KeyPair, ownerAddress *cryptolib.Address) *AccessNodeInfo {
	a.Certificate = NewNodeOwnershipCertificate(nodeKeyPair, ownerAddress).Bytes()
	return a
}

type ChangeAccessNodeAction byte

func (a *ChangeAccessNodeAction) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(byte(*a))
	return ww.Err
}

func (a *ChangeAccessNodeAction) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	*a = ChangeAccessNodeAction(rr.ReadByte())
	return rr.Err
}

const (
	ChangeAccessNodeActionRemove = ChangeAccessNodeAction(iota)
	ChangeAccessNodeActionAccept
	ChangeAccessNodeActionDrop
	ChangeAccessNodeActionLast
)

func RemoveAccessNodeAction(pubKey *cryptolib.PublicKey) lo.Tuple2[*cryptolib.PublicKey, *ChangeAccessNodeAction] {
	action := ChangeAccessNodeActionRemove
	return lo.T2(pubKey, &action)
}

func AcceptAccessNodeAction(pubKey *cryptolib.PublicKey) lo.Tuple2[*cryptolib.PublicKey, *ChangeAccessNodeAction] {
	action := ChangeAccessNodeActionAccept
	return lo.T2(pubKey, &action)
}

func DropAccessNodeAction(pubKey *cryptolib.PublicKey) lo.Tuple2[*cryptolib.PublicKey, *ChangeAccessNodeAction] {
	action := ChangeAccessNodeActionDrop
	return lo.T2(pubKey, &action)
}
