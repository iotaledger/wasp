// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// NodeOwnershipCertificate is a proof that a specified address is an owner of the specified node.
// It is implemented as a signature over the node pub key concatenated with the owner address.
type NodeOwnershipCertificate []byte

type NodeOwnershipCertificateFields struct {
	NodePubKey   *cryptolib.PublicKey
	OwnerAddress *cryptolib.Address
}

func NewNodeOwnershipCertificate(nodeKeyPair *cryptolib.KeyPair, ownerAddress *cryptolib.Address) NodeOwnershipCertificate {
	cert := bcs.MustMarshal(&NodeOwnershipCertificateFields{
		NodePubKey:   nodeKeyPair.GetPublicKey(),
		OwnerAddress: ownerAddress,
	})

	return nodeKeyPair.GetPrivateKey().Sign(cert)
}

func NodeOwnershipCertificateFromBytes(data []byte) NodeOwnershipCertificate {
	return data
}

func (c NodeOwnershipCertificate) Verify(nodePubKey *cryptolib.PublicKey, ownerAddress *cryptolib.Address) bool {
	cert := bcs.MustMarshal(&NodeOwnershipCertificateFields{
		NodePubKey:   nodePubKey,
		OwnerAddress: ownerAddress,
	})

	return nodePubKey.Verify(cert, c.Bytes())
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

type ChangeAccessNodeActions = []lo.Tuple2[*cryptolib.PublicKey, ChangeAccessNodeAction]

func RemoveAccessNodeAction(pubKey *cryptolib.PublicKey) lo.Tuple2[*cryptolib.PublicKey, ChangeAccessNodeAction] {
	return lo.T2(pubKey, ChangeAccessNodeActionRemove)
}

func AcceptAccessNodeAction(pubKey *cryptolib.PublicKey) lo.Tuple2[*cryptolib.PublicKey, ChangeAccessNodeAction] {
	return lo.T2(pubKey, ChangeAccessNodeActionAccept)
}

func DropAccessNodeAction(pubKey *cryptolib.PublicKey) lo.Tuple2[*cryptolib.PublicKey, ChangeAccessNodeAction] {
	return lo.T2(pubKey, ChangeAccessNodeActionDrop)
}
