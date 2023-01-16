package smGPAUtils

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/state"
)

func FillDBFromBlockWAL(db state.Store, bw BlockWAL) error {
	storedBlockHashes := make(map[state.BlockHash]bool)
	var storeBlockFun func(blockHash state.BlockHash) error
	storeBlockFun = func(blockHash state.BlockHash) error {
		block, err := bw.Read(blockHash)
		if err != nil {
			return fmt.Errorf("Error filling the store from WAL: failed to read block %s from WAL: %w", blockHash, err)
		}
		previousBlockCommitment := block.PreviousL1Commitment()
		_, alreadyStored := storedBlockHashes[previousBlockCommitment.BlockHash()]
		if !alreadyStored && !previousBlockCommitment.Equals(state.OriginL1Commitment()) {
			err := storeBlockFun(previousBlockCommitment.BlockHash())
			if err != nil {
				return err
			}
		}
		stateDraft, err := db.NewEmptyStateDraft(previousBlockCommitment)
		if err != nil {
			return fmt.Errorf("Error filling the store from WAL: failed to create state draft to store block %s: %w", blockHash, err)
		}
		block.Mutations().ApplyTo(stateDraft)
		committedBlock := db.Commit(stateDraft)
		committedCommitment := committedBlock.L1Commitment()
		if !committedCommitment.Equals(block.L1Commitment()) {
			return fmt.Errorf("Error filling the store from WAL: committed block has different commitment than block in WAL: %s =/= %s",
				committedCommitment, block.L1Commitment())
		}
		storedBlockHashes[blockHash] = true
		return nil
	}
	for _, blockHash := range bw.Contents() {
		_, alreadyStored := storedBlockHashes[blockHash]
		if !alreadyStored {
			err := storeBlockFun(blockHash)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
