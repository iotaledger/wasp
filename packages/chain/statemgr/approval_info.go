// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type approvalInfo struct {
	outputID            *iotago.UTXOInput
	nextStateCommitment common.VCommitment
	blockHash           state.BlockHash
}

func newApprovalInfo(output *isc.AliasOutputWithID) (*approvalInfo, error) {
	l1Commitment, err := state.L1CommitmentFromAliasOutput(output.GetAliasOutput())
	if err != nil {
		return nil, err
	}
	return &approvalInfo{
		outputID:            output.ID(),
		nextStateCommitment: l1Commitment.StateCommitment,
		blockHash:           l1Commitment.BlockHash,
	}, nil
}

func (aiT *approvalInfo) getNextStateCommitment() common.VCommitment {
	return aiT.nextStateCommitment
}

func (aiT *approvalInfo) getBlockHash() state.BlockHash {
	return aiT.blockHash
}

func (aiT *approvalInfo) String() string {
	return fmt.Sprintf("output ID: %v, next state commitment %s, block hash %s",
		isc.OID(aiT.outputID), aiT.nextStateCommitment, aiT.blockHash)
}
