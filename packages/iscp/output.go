// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"bytes"

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

func (aowiT *AliasOutputWithID) GetAliasOutput() *iotago.AliasOutput {
	return aowiT.output
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

func (aowiT *AliasOutputWithID) GetStateAddress() iotago.Address {
	return aowiT.output.StateController
}

func AliasOutputsEqual(ao1, ao2 *iotago.AliasOutput) bool {
	if ao1 == nil {
		return ao2 == nil
	}
	if ao2 == nil {
		return false
	}
	if ao1.Amount != ao2.Amount {
		return false
	}
	if !ao1.NativeTokens.Equal(ao2.NativeTokens) {
		return false
	}
	if ao1.AliasID != ao2.AliasID {
		return false
	}
	if ao1.StateController == nil {
		if ao2.StateController != nil {
			return false
		}
	} else if !ao1.StateController.Equal(ao2.StateController) {
		return false
	}
	if ao1.GovernanceController == nil {
		if ao2.StateController != nil {
			return false
		}
	} else if !ao1.GovernanceController.Equal(ao2.GovernanceController) {
		return false
	}
	if ao1.StateIndex != ao2.StateIndex {
		return false
	}
	if !bytes.Equal(ao1.StateMetadata, ao2.StateMetadata) {
		return false
	}
	if ao1.FoundryCounter != ao2.FoundryCounter {
		return false
	}
	return ao1.Blocks.Equal(ao2.Blocks)
}
