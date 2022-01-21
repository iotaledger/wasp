// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
)

type AliasOutputWithID struct {
	output *iotago.AliasOutput
	id     *iotago.UTXOInput
}

func NewAliasOutputWithID(output *iotago.AliasOutput, id *iotago.UTXOInput) *AliasOutputWithID {
	return &AliasOutputWithID{
		output: output,
		id:     id,
	}
}

func (aowiT *AliasOutputWithID) ID() *iotago.UTXOInput {
	return aowiT.id
}

func (aowiT *AliasOutputWithID) GetStateIndex() uint32 {
	return aowiT.output.StateIndex
}

func (aowiT *AliasOutputWithID) GetStateMetadata() []byte {
	return aowiT.output.StateMetadata
}

func (aowiT *AliasOutputWithID) GetStateCommitment() (hashing.HashValue, error) {
	sd, err := StateDataFromBytes(aowiT.output.StateMetadata)
	if err != nil {
		return hashing.NilHash, err
	}
	return sd.Commitment, nil
}
