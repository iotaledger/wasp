package migrations

import (
	old_cryptolib "github.com/nnikolash/wasp-types-exported/packages/cryptolib"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
	"github.com/samber/lo"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

func MigrateGovernanceContract(
	oldChainState old_kv.KVStoreReader,
	newChainState kv.KVStore,
	oldChainID old_isc.ChainID,
	chainOwner *cryptolib.Address,
) (blockKeepAmount int32) {

	oldContractState := oldstate.GetContactStateReader(oldChainState, old_governance.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, governance.Contract.Hname())

	cli.DebugLog("Migrating governance contract\n")

	migrateChainOwnerID(oldChainState, newContractState, oldChainID, chainOwner) // WARNING: oldChainState is specifically used here
	migrateChainOwnerIDDelegetaed(oldContractState, newContractState, oldChainID)
	migratePayoutAgent(oldContractState, newContractState, oldChainID)
	migrateGasFeePolicy(oldContractState, newContractState)
	migrateGasLimits(oldContractState, newContractState)
	migrateAccessNodes(oldContractState, newContractState)
	migrateAccessNodeCandidates(oldContractState, newContractState)
	migrateMaintenanceStatus(oldContractState, newContractState)
	migrateMetadata(oldContractState, newContractState)
	migratePublicURL(oldContractState, newContractState)
	blockKeepAmount = migrateBlockKeepAmount(oldContractState, newContractState)
	// NOTE: VarRotateToAddress ignored
	// NOTE: VarMinBaseTokensOnCommonAccount ignored, thus deleted
	// NOTE: VarAllowedStateControllerAddresses ignored

	cli.DebugLog("Migrated governance contract\n")

	return blockKeepAmount
}

func migrateChainOwnerID(oldChainState old_kv.KVStoreReader, newContractState kv.KVStore, oldChainID old_isc.ChainID, chainOwner *cryptolib.Address) {
	cli.DebugLog("Migrating chain owner...\n")

	governance.NewStateWriter(newContractState).SetChainAdmin(isc.NewAddressAgentID(chainOwner))

	cli.DebugLog("Migrated chain owner\n")
}

func migrateChainOwnerIDDelegetaed(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore, oldChainID old_isc.ChainID) {
	cli.DebugLog("Migrating chain owner delegated...\n")

	oldChainOwnerDelegatedIDBytes := oldContractState.Get(old_governance.VarChainOwnerIDDelegated)
	if len(oldChainOwnerDelegatedIDBytes) != 0 {
		oldChainOwnerDelegatedID := lo.Must(old_codec.DecodeAgentID(oldChainOwnerDelegatedIDBytes))
		newChainIDOwnerDelegatedID := OldAgentIDtoNewAgentID(oldChainOwnerDelegatedID, oldChainID)
		governance.NewStateWriter(newContractState).SetChainAdmin(newChainIDOwnerDelegatedID)
	}

	cli.DebugLog("Migrated chain owner delegated\n")
}

func migratePayoutAgent(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
	oldChainID old_isc.ChainID,
) {
	cli.DebugLogf("Migrating Payout agent...\n")

	oldPayoudAgentID := old_governance.MustGetPayoutAgentID(oldContractState)
	newPayoutAgentID := OldAgentIDtoNewAgentID(oldPayoudAgentID, oldChainID)

	governance.NewStateWriter(newContractState).SetPayoutAgentID(newPayoutAgentID)

	cli.DebugLogf("Migrated Payout agent\n")
}

func migrateGasFeePolicy(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore) {
	cli.DebugLog("Migrating gas fee policy...\n")

	oldPolicy := old_governance.MustGetGasFeePolicy(oldContractState)
	newGasPerToken := OldGasPerTokenToNew(oldPolicy)

	newPolicy := gas.FeePolicy{
		EVMGasRatio:       lo.Must(util.Ratio32FromString(oldPolicy.EVMGasRatio.String())),
		GasPerToken:       newGasPerToken,
		ValidatorFeeShare: oldPolicy.ValidatorFeeShare,
	}

	governance.NewStateWriter(newContractState).SetGasFeePolicy(&newPolicy)

	cli.DebugLog("Migrated gas fee policy\n")
}

func migrateGasLimits(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore) {
	cli.DebugLog("Migrating gas limits...\n")

	oldLimits := old_governance.MustGetGasLimits(oldContractState)
	newLimits := gas.Limits{
		MaxGasPerBlock:         oldLimits.MaxGasPerBlock,
		MinGasPerRequest:       oldLimits.MinGasPerRequest,
		MaxGasPerRequest:       oldLimits.MaxGasPerRequest,
		MaxGasExternalViewCall: oldLimits.MaxGasExternalViewCall,
	}

	governance.NewStateWriter(newContractState).SetGasLimits(&newLimits)

	cli.DebugLog("Migrated gas limits\n")
}

func migrateAccessNodes(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
) {
	cli.DebugLog("Migrating access nodes...\n")

	oldAccessNodes := old_governance.AccessNodesMapR(oldContractState)
	newAccessNodes := governance.NewStateWriter(newContractState).AccessNodesMap()

	oldAccessNodes.Iterate(func(k []byte, v []byte) bool {
		oldNodePubKey := lo.Must(old_cryptolib.PublicKeyFromBytes(k))
		oldV := old_codec.MustDecodeBool(v)

		newNodePubKey := lo.Must(cryptolib.PublicKeyFromBytes(oldNodePubKey.AsBytes()))

		newAccessNodes.SetAt(newNodePubKey.Bytes(), codec.Encode(oldV))
		return true
	})

	cli.DebugLog("Migrated access nodes\n")
}

func migrateAccessNodeCandidates(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
) {
	cli.DebugLog("Migrating access node candidates...\n")

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

	cli.DebugLog("Migrated access node candidates\n")
}

func migrateMaintenanceStatus(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
) {
	cli.DebugLog("Migrating maintenance status...\n")

	maintenanceStatus := old_codec.MustDecodeBool(oldContractState.Get(old_governance.VarMaintenanceStatus))
	governance.NewStateWriter(newContractState).SetMaintenanceStatus(maintenanceStatus)

	cli.DebugLog("Migrated maintenance status\n")
}

func migrateMetadata(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
) {
	cli.DebugLog("Migrating public chain metadata...\n")

	oldMetadata := old_governance.MustGetMetadata(oldContractState)
	newMetadata := isc.PublicChainMetadata{
		EVMJsonRPCURL:   oldMetadata.EVMJsonRPCURL,
		EVMWebSocketURL: oldMetadata.EVMWebSocketURL,
		Name:            oldMetadata.Name,
		Description:     oldMetadata.Description,
		Website:         oldMetadata.Website,
	}

	governance.NewStateWriter(newContractState).SetMetadata(&newMetadata)

	cli.DebugLog("Migrated ublic chain metadata\n")
}

func migratePublicURL(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
) {
	cli.DebugLog("Migrating public URL...\n")

	publicURL := lo.Must(old_governance.GetPublicURL(oldContractState))
	governance.NewStateWriter(newContractState).SetPublicURL(publicURL)

	cli.DebugLog("Migrated public URL\n")
}

func migrateBlockKeepAmount(
	oldContractState old_kv.KVStoreReader,
	newContractState kv.KVStore,
) (blockKeepAmount int32) {
	cli.DebugLog("Migrating block keep amount...\n")

	blockKeepAmount = old_governance.GetBlockKeepAmount(oldContractState)
	governance.NewStateWriter(newContractState).SetBlockKeepAmount(blockKeepAmount)

	cli.DebugLog("Migrated block keep amount\n")

	return blockKeepAmount
}
