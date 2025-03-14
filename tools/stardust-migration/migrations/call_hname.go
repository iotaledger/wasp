package migrations

import (
	"github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/blob"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/errors"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/root"
)

var ContractNameFuncs = make(map[isc.Hname]string)

func BuildContractNameFuncs() {

	// accounts
	ContractNameFuncs[accounts.FuncDeposit.Hname()] = accounts.FuncDeposit.Name
	ContractNameFuncs[accounts.FuncFoundryCreateNew.Hname()] = accounts.FuncFoundryCreateNew.Name
	ContractNameFuncs[accounts.FuncNativeTokenCreate.Hname()] = accounts.FuncNativeTokenCreate.Name
	ContractNameFuncs[accounts.FuncNativeTokenModifySupply.Hname()] = accounts.FuncNativeTokenModifySupply.Name
	ContractNameFuncs[accounts.FuncNativeTokenDestroy.Hname()] = accounts.FuncNativeTokenDestroy.Name
	ContractNameFuncs[accounts.FuncMintNFT.Hname()] = accounts.FuncMintNFT.Name
	ContractNameFuncs[accounts.FuncTransferAccountToChain.Hname()] = accounts.FuncTransferAccountToChain.Name
	ContractNameFuncs[accounts.FuncTransferAllowanceTo.Hname()] = accounts.FuncTransferAllowanceTo.Name
	ContractNameFuncs[accounts.FuncWithdraw.Hname()] = accounts.FuncWithdraw.Name
	ContractNameFuncs[accounts.ViewAccountFoundries.Hname()] = accounts.ViewAccountFoundries.Name
	ContractNameFuncs[accounts.ViewAccountNFTAmount.Hname()] = accounts.ViewAccountNFTAmount.Name
	ContractNameFuncs[accounts.ViewAccountNFTAmountInCollection.Hname()] = accounts.ViewAccountNFTAmountInCollection.Name
	ContractNameFuncs[accounts.ViewAccountNFTs.Hname()] = accounts.ViewAccountNFTs.Name
	ContractNameFuncs[accounts.ViewAccountNFTsInCollection.Hname()] = accounts.ViewAccountNFTsInCollection.Name
	ContractNameFuncs[accounts.ViewNFTIDbyMintID.Hname()] = accounts.ViewNFTIDbyMintID.Name
	ContractNameFuncs[accounts.ViewBalance.Hname()] = accounts.ViewBalance.Name
	ContractNameFuncs[accounts.ViewBalanceBaseToken.Hname()] = accounts.ViewBalanceBaseToken.Name
	ContractNameFuncs[accounts.ViewBalanceBaseTokenEVM.Hname()] = accounts.ViewBalanceBaseTokenEVM.Name
	ContractNameFuncs[accounts.ViewBalanceNativeToken.Hname()] = accounts.ViewBalanceNativeToken.Name
	ContractNameFuncs[accounts.ViewNativeToken.Hname()] = accounts.ViewNativeToken.Name
	ContractNameFuncs[accounts.ViewGetAccountNonce.Hname()] = accounts.ViewGetAccountNonce.Name
	ContractNameFuncs[accounts.ViewGetNativeTokenIDRegistry.Hname()] = accounts.ViewGetNativeTokenIDRegistry.Name
	ContractNameFuncs[accounts.ViewNFTData.Hname()] = accounts.ViewNFTData.Name
	ContractNameFuncs[accounts.ViewTotalAssets.Hname()] = accounts.ViewTotalAssets.Name

	// blob
	ContractNameFuncs[blob.FuncStoreBlob.Hname()] = blob.FuncStoreBlob.Name
	ContractNameFuncs[blob.ViewGetBlobInfo.Hname()] = blob.ViewGetBlobInfo.Name
	ContractNameFuncs[blob.ViewGetBlobField.Hname()] = blob.ViewGetBlobField.Name

	// blocklog
	ContractNameFuncs[blocklog.FuncRetryUnprocessable.Hname()] = blocklog.FuncRetryUnprocessable.Name
	ContractNameFuncs[blocklog.ViewGetBlockInfo.Hname()] = blocklog.ViewGetBlockInfo.Name
	ContractNameFuncs[blocklog.ViewGetRequestIDsForBlock.Hname()] = blocklog.ViewGetRequestIDsForBlock.Name
	ContractNameFuncs[blocklog.ViewGetRequestReceipt.Hname()] = blocklog.ViewGetRequestReceipt.Name
	ContractNameFuncs[blocklog.ViewGetRequestReceiptsForBlock.Hname()] = blocklog.ViewGetRequestReceiptsForBlock.Name
	ContractNameFuncs[blocklog.ViewIsRequestProcessed.Hname()] = blocklog.ViewIsRequestProcessed.Name
	ContractNameFuncs[blocklog.ViewGetEventsForRequest.Hname()] = blocklog.ViewGetEventsForRequest.Name
	ContractNameFuncs[blocklog.ViewGetEventsForBlock.Hname()] = blocklog.ViewGetEventsForBlock.Name
	ContractNameFuncs[blocklog.ViewHasUnprocessable.Hname()] = blocklog.ViewHasUnprocessable.Name

	// errors
	ContractNameFuncs[errors.FuncRegisterError.Hname()] = errors.FuncRegisterError.Name
	ContractNameFuncs[errors.ViewGetErrorMessageFormat.Hname()] = errors.ViewGetErrorMessageFormat.Name

	// evm
	ContractNameFuncs[evm.FuncSendTransaction.Hname()] = evm.FuncSendTransaction.Name
	ContractNameFuncs[evm.FuncCallContract.Hname()] = evm.FuncCallContract.Name
	ContractNameFuncs[evm.FuncGetChainID.Hname()] = evm.FuncGetChainID.Name
	ContractNameFuncs[evm.FuncRegisterERC20NativeToken.Hname()] = evm.FuncRegisterERC20NativeToken.Name
	ContractNameFuncs[evm.FuncRegisterERC20NativeTokenOnRemoteChain.Hname()] = evm.FuncRegisterERC20NativeTokenOnRemoteChain.Name
	ContractNameFuncs[evm.FuncRegisterERC20ExternalNativeToken.Hname()] = evm.FuncRegisterERC20ExternalNativeToken.Name
	ContractNameFuncs[evm.FuncGetERC20ExternalNativeTokenAddress.Hname()] = evm.FuncGetERC20ExternalNativeTokenAddress.Name
	ContractNameFuncs[evm.FuncGetERC721CollectionAddress.Hname()] = evm.FuncGetERC721CollectionAddress.Name
	ContractNameFuncs[evm.FuncRegisterERC721NFTCollection.Hname()] = evm.FuncRegisterERC721NFTCollection.Name
	ContractNameFuncs[evm.FuncNewL1Deposit.Hname()] = evm.FuncNewL1Deposit.Name

	// governance
	ContractNameFuncs[governance.FuncRotateStateController.Hname()] = governance.FuncRotateStateController.Name
	ContractNameFuncs[governance.FuncAddAllowedStateControllerAddress.Hname()] = governance.FuncAddAllowedStateControllerAddress.Name
	ContractNameFuncs[governance.FuncRemoveAllowedStateControllerAddress.Hname()] = governance.FuncRemoveAllowedStateControllerAddress.Name
	ContractNameFuncs[governance.ViewGetAllowedStateControllerAddresses.Hname()] = governance.ViewGetAllowedStateControllerAddresses.Name
	ContractNameFuncs[governance.FuncClaimChainOwnership.Hname()] = governance.FuncClaimChainOwnership.Name
	ContractNameFuncs[governance.FuncDelegateChainOwnership.Hname()] = governance.FuncDelegateChainOwnership.Name
	ContractNameFuncs[governance.FuncSetPayoutAgentID.Hname()] = governance.FuncSetPayoutAgentID.Name
	ContractNameFuncs[governance.FuncSetMinCommonAccountBalance.Hname()] = governance.FuncSetMinCommonAccountBalance.Name
	ContractNameFuncs[governance.ViewGetPayoutAgentID.Hname()] = governance.ViewGetPayoutAgentID.Name
	ContractNameFuncs[governance.ViewGetMinCommonAccountBalance.Hname()] = governance.ViewGetMinCommonAccountBalance.Name
	ContractNameFuncs[governance.ViewGetChainOwner.Hname()] = governance.ViewGetChainOwner.Name
	ContractNameFuncs[governance.FuncSetFeePolicy.Hname()] = governance.FuncSetFeePolicy.Name
	ContractNameFuncs[governance.FuncSetGasLimits.Hname()] = governance.FuncSetGasLimits.Name
	ContractNameFuncs[governance.ViewGetFeePolicy.Hname()] = governance.ViewGetFeePolicy.Name
	ContractNameFuncs[governance.ViewGetGasLimits.Hname()] = governance.ViewGetGasLimits.Name
	ContractNameFuncs[governance.FuncSetEVMGasRatio.Hname()] = governance.FuncSetEVMGasRatio.Name
	ContractNameFuncs[governance.ViewGetEVMGasRatio.Hname()] = governance.ViewGetEVMGasRatio.Name
	ContractNameFuncs[governance.ViewGetChainInfo.Hname()] = governance.ViewGetChainInfo.Name
	ContractNameFuncs[governance.FuncAddCandidateNode.Hname()] = governance.FuncAddCandidateNode.Name
	ContractNameFuncs[governance.FuncRevokeAccessNode.Hname()] = governance.FuncRevokeAccessNode.Name
	ContractNameFuncs[governance.FuncChangeAccessNodes.Hname()] = governance.FuncChangeAccessNodes.Name
	ContractNameFuncs[governance.ViewGetChainNodes.Hname()] = governance.ViewGetChainNodes.Name
	ContractNameFuncs[governance.FuncStartMaintenance.Hname()] = governance.FuncStartMaintenance.Name
	ContractNameFuncs[governance.FuncStopMaintenance.Hname()] = governance.FuncStopMaintenance.Name
	ContractNameFuncs[governance.ViewGetMaintenanceStatus.Hname()] = governance.ViewGetMaintenanceStatus.Name
	ContractNameFuncs[governance.FuncSetMetadata.Hname()] = governance.FuncSetMetadata.Name
	ContractNameFuncs[governance.ViewGetMetadata.Hname()] = governance.ViewGetMetadata.Name

	// root
	ContractNameFuncs[root.FuncDeployContract.Hname()] = root.FuncDeployContract.Name
	ContractNameFuncs[root.FuncGrantDeployPermission.Hname()] = root.FuncGrantDeployPermission.Name
	ContractNameFuncs[root.FuncRevokeDeployPermission.Hname()] = root.FuncRevokeDeployPermission.Name
	ContractNameFuncs[root.FuncRequireDeployPermissions.Hname()] = root.FuncRequireDeployPermissions.Name
	ContractNameFuncs[root.ViewFindContract.Hname()] = root.ViewFindContract.Name
	ContractNameFuncs[root.ViewGetContractRecords.Hname()] = root.ViewGetContractRecords.Name
}
