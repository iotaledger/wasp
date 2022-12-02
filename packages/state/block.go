// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"bytes"
	"io"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

type block struct {
	mutations         *buffered.Mutations
	trieRoot          common.VCommitment
	previousTrieRoot  common.VCommitment
	approvingOutputID *iotago.UTXOInput
}

var _ Block = &block{}

func BlockFromBytes(blockBytes []byte) (*block, error) {
	buf := bytes.NewBuffer(blockBytes)

	muts := buffered.NewMutations()
	err := muts.Read(buf)
	if err != nil {
		return nil, err
	}

	trieRoot, err := common.VectorCommitmentFromBytes(commitmentModel, buf.Next(int(commitmentModel.HashSize())))
	if err != nil {
		return nil, err
	}

	var hasPrevTrieRoot bool
	if hasPrevTrieRoot, err = codec.DecodeBool(buf.Next(1)); err != nil {
		return nil, err
	}
	var prevTrieRoot common.VCommitment
	if hasPrevTrieRoot {
		prevTrieRoot, err = common.VectorCommitmentFromBytes(commitmentModel, buf.Next(int(commitmentModel.HashSize())))
		if err != nil {
			return nil, err
		}
	}

	var hasApprovingOutputID bool
	var approvingOutputID *iotago.UTXOInput
	if hasApprovingOutputID, err = codec.DecodeBool(buf.Next(1)); err != nil {
		return nil, err
	}
	if hasApprovingOutputID {
		approvingOutputID = &iotago.UTXOInput{}
		_, err := approvingOutputID.Deserialize(buf.Next(approvingOutputID.Size()), serializer.DeSeriModeNoValidation, nil)
		if err != nil {
			return nil, err
		}
	}

	return &block{
		mutations:         muts,
		trieRoot:          trieRoot,
		previousTrieRoot:  prevTrieRoot,
		approvingOutputID: approvingOutputID,
	}, nil
}

func (b *block) Mutations() *buffered.Mutations {
	return b.mutations
}

func (b *block) TrieRoot() common.VCommitment {
	return b.trieRoot
}

func (b *block) PreviousTrieRoot() common.VCommitment {
	return b.previousTrieRoot
}

func (b *block) ApprovingOutputID() *iotago.UTXOInput {
	return b.approvingOutputID
}

func (b *block) setApprovingOutputID(oid *iotago.UTXOInput) {
	b.approvingOutputID = oid
}

func (b *block) essenceBytes() []byte {
	var w bytes.Buffer
	b.writeEssence(&w)
	return w.Bytes()
}

func (b *block) writeEssence(w io.Writer) {
	w.Write(b.Mutations().Bytes())
}

func (b *block) Bytes() []byte {
	var w bytes.Buffer
	b.writeEssence(&w)
	w.Write(b.TrieRoot().Bytes())
	w.Write(codec.EncodeBool(b.PreviousTrieRoot() != nil))
	if b.PreviousTrieRoot() != nil {
		w.Write(b.PreviousTrieRoot().Bytes())
	}
	w.Write(codec.EncodeBool(b.approvingOutputID != nil))
	if b.approvingOutputID != nil {
		bytes, err := b.approvingOutputID.Serialize(serializer.DeSeriModeNoValidation, nil)
		if err != nil {
			panic(err)
		}
		_, _ = w.Write(bytes)
	}
	return w.Bytes()
}

func (b *block) Hash() BlockHash {
	return BlockHashFromData(b.essenceBytes())
}

func (b *block) L1Commitment() *L1Commitment {
	return &L1Commitment{
		StateCommitment: b.TrieRoot(),
		BlockHash:       b.Hash(),
	}
}

func (b *block) GetHash() (ret BlockHash) {
	r := blake2b.Sum256(b.essenceBytes())
	copy(ret[:BlockHashSize], r[:BlockHashSize])
	return
}

func (b *block) Equals(other Block) bool {
	return b.GetHash().Equals(other.Hash())
}
