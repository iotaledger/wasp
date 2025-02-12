package migrations

import (
	"log"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	old_cryptolib "github.com/nnikolash/wasp-types-exported/packages/cryptolib"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
	"github.com/samber/lo"
)

func MigrateGovernanceContract(
	oldChainState old_kv.KVStoreReader,
	newChainState state.StateDraft,
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
) {

	oldContractState := oldstate.GetContactStateReader(oldChainState, old_governance.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, governance.Contract.Hname())

	log.Print("Migrating governance contract\n")

	migrateChainOwnerID(oldChainState, newContractState, oldChainID, newChainID) // WARNING: oldChainState is specifically used here
	migrateChainOwnerIDDelegetaed(oldContractState, newContractState, oldChainID, newChainID)
	migratePayoutAgent(oldContractState, newContractState, oldChainID, newChainID)
	migrateGasFeePolicy(oldContractState, newContractState)
	migrateGasLimits(oldContractState, newContractState)
	migrateAccessNodes(oldContractState, newContractState)
	migrateAccessNodeCandidates(oldContractState, newContractState)
	// NOTE: VarRotateToAddress ignored
	// NOTE: VarMinBaseTokensOnCommonAccount ignored, thus deleted

	// TODO: VarAllowedStateControllerAddresses

	log.Print("Migrated governance contract\n")
}

func migrateChainOwnerID(
	oldChainState old_kv.KVStoreReader,
	newContractState kv.KVStore,
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
) {
	log.Print("Migrating chain owner...\n")

	oldChainOwnerID := old_governance.NewStateAccess(oldChainState).ChainOwnerID()
	newChainOwnerID := OldAgentIDtoNewAgentID(oldChainOwnerID, oldChainID, newChainID)
	governance.NewStateWriter(newContractState).SetChainOwnerID(newChainOwnerID)

	log.Print("Migrated chain owner\n")
}

func migrateChainOwnerIDDelegetaed(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
) {
	log.Print("Migrating chain owner delegated...\n")

	oldChainOwnerDelegatedIDBytes := oldContractState.Get(old_governance.VarChainOwnerIDDelegated)
	if len(oldChainOwnerDelegatedIDBytes) != 0 {
		oldChainOwnerDelegatedID := lo.Must(old_codec.DecodeAgentID(oldChainOwnerDelegatedIDBytes))
		newChainIDOwnerDelegatedID := OldAgentIDtoNewAgentID(oldChainOwnerDelegatedID, oldChainID, newChainID)
		governance.NewStateWriter(newContractState).SetChainOwnerIDDelegated(newChainIDOwnerDelegatedID)
	}

	log.Print("Migrated chain owner delegated\n")
}

func migratePayoutAgent(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
) {
	log.Printf("Migrating Payout agent...\n")

	oldPayoudAgentID := old_governance.MustGetPayoutAgentID(oldContractState)
	newPayoutAgentID := OldAgentIDtoNewAgentID(oldPayoudAgentID, oldChainID, newChainID)

	governance.NewStateWriter(newContractState).SetPayoutAgentID(newPayoutAgentID)

	log.Printf("Migrated Payout agent\n")
}

func migrateGasFeePolicy(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore) {
	log.Print("Migrating gas fee policy...\n")

	oldPolicy := old_governance.MustGetGasFeePolicy(oldContractState)
	newPolicy := gas.FeePolicy{
		EVMGasRatio:       lo.Must(util.Ratio32FromString(oldPolicy.EVMGasRatio.String())),
		GasPerToken:       lo.Must(util.Ratio32FromString(oldPolicy.GasPerToken.String())),
		ValidatorFeeShare: oldPolicy.ValidatorFeeShare,
	}

	governance.NewStateWriter(newContractState).SetGasFeePolicy(&newPolicy)

	log.Print("Migrated gas fee policy\n")
}

func migrateGasLimits(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore) {
	log.Print("Migrating gas limits...\n")

	oldLimits := old_governance.MustGetGasLimits(oldContractState)
	newLimits := gas.Limits{
		MaxGasPerBlock:         oldLimits.MaxGasPerBlock,
		MinGasPerRequest:       oldLimits.MinGasPerRequest,
		MaxGasPerRequest:       oldLimits.MaxGasPerRequest,
		MaxGasExternalViewCall: oldLimits.MaxGasExternalViewCall,
	}

	governance.NewStateWriter(newContractState).SetGasLimits(&newLimits)

	log.Print("Migrated gas limits\n")
}

func migrateAccessNodes(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
) {
	log.Print("Migrating access nodes...\n")

	oldAccessNodes := old_governance.AccessNodesMapR(oldContractState)
	newAccessNodes := governance.NewStateWriter(newContractState).AccessNodesMap()

	oldAccessNodes.Iterate(func(k []byte, v []byte) bool {
		oldNodePubKey := lo.Must(old_cryptolib.PublicKeyFromBytes(k))
		oldV := old_codec.MustDecodeBool(v)

		newNodePubKey := lo.Must(cryptolib.PublicKeyFromBytes(oldNodePubKey.AsBytes()))

		newAccessNodes.SetAt(newNodePubKey.Bytes(), codec.Encode(oldV))
		return true
	})

	log.Print("Migrated access nodes\n")
}

func migrateAccessNodeCandidates(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
) {
	log.Print("Migrating access node candidates...\n")

	oldCandidates := old_governance.AccessNodeCandidatesMapR(oldContractState)
	newCandidates := governance.NewStateWriter(newContractState).AccessNodeCandidatesMap()

	oldCandidates.Iterate(func(k []byte, v []byte) bool {
		oldNodePubKey := k
		oldAccessNodeInfo := lo.Must(old_governance.AccessNodeInfoFromBytes(oldNodePubKey, v))

		oldValidatorAddr := lo.Must(old_isc.AddressFromBytes(oldAccessNodeInfo.ValidatorAddr()))
		newValidatorAddr := lo.Must(cryptolib.NewAddressFromHexString(oldValidatorAddr.String()))

		newAccessNodeInfo := governance.AccessNodeInfo{
			NodePubKey: lo.Must(cryptolib.PublicKeyFromBytes(oldNodePubKey)),
			AccessNodeData: governance.AccessNodeData{
				ValidatorAddr: newValidatorAddr,
				Certificate:   oldAccessNodeInfo.Certificate,
				ForCommittee:  oldAccessNodeInfo.ForCommittee,
				AccessAPI:     oldAccessNodeInfo.AccessAPI,
			},
		}

		newCandidates.SetAt(newAccessNodeInfo.NodePubKey.Bytes(), bcs.MustMarshal(&newAccessNodeInfo.AccessNodeData))

		return true
	})

	log.Print("Migrated access node candidates\n")
}
