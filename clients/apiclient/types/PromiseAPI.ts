import { ResponseContext, RequestContext, HttpFile } from '../http/http';
import { Configuration} from '../configuration'

import { AccountFoundriesResponse } from '../models/AccountFoundriesResponse';
import { AccountNFTsResponse } from '../models/AccountNFTsResponse';
import { AccountNonceResponse } from '../models/AccountNonceResponse';
import { AddUserRequest } from '../models/AddUserRequest';
import { AliasOutputMetricItem } from '../models/AliasOutputMetricItem';
import { AssetsJSON } from '../models/AssetsJSON';
import { AssetsResponse } from '../models/AssetsResponse';
import { AuthInfoModel } from '../models/AuthInfoModel';
import { BaseToken } from '../models/BaseToken';
import { BlobInfoResponse } from '../models/BlobInfoResponse';
import { BlobValueResponse } from '../models/BlobValueResponse';
import { BlockInfoResponse } from '../models/BlockInfoResponse';
import { BurnRecord } from '../models/BurnRecord';
import { CallTargetJSON } from '../models/CallTargetJSON';
import { ChainInfoResponse } from '../models/ChainInfoResponse';
import { ChainMessageMetrics } from '../models/ChainMessageMetrics';
import { ChainRecord } from '../models/ChainRecord';
import { CommitteeInfoResponse } from '../models/CommitteeInfoResponse';
import { CommitteeNode } from '../models/CommitteeNode';
import { ConsensusPipeMetrics } from '../models/ConsensusPipeMetrics';
import { ConsensusWorkflowMetrics } from '../models/ConsensusWorkflowMetrics';
import { ContractCallViewRequest } from '../models/ContractCallViewRequest';
import { ContractInfoResponse } from '../models/ContractInfoResponse';
import { ControlAddressesResponse } from '../models/ControlAddressesResponse';
import { DKSharesInfo } from '../models/DKSharesInfo';
import { DKSharesPostRequest } from '../models/DKSharesPostRequest';
import { ErrorMessageFormatResponse } from '../models/ErrorMessageFormatResponse';
import { EstimateGasRequestOffledger } from '../models/EstimateGasRequestOffledger';
import { EstimateGasRequestOnledger } from '../models/EstimateGasRequestOnledger';
import { EventJSON } from '../models/EventJSON';
import { EventsResponse } from '../models/EventsResponse';
import { FeePolicy } from '../models/FeePolicy';
import { FoundryOutputResponse } from '../models/FoundryOutputResponse';
import { GovAllowedStateControllerAddressesResponse } from '../models/GovAllowedStateControllerAddressesResponse';
import { GovChainInfoResponse } from '../models/GovChainInfoResponse';
import { GovChainOwnerResponse } from '../models/GovChainOwnerResponse';
import { GovPublicChainMetadata } from '../models/GovPublicChainMetadata';
import { InOutput } from '../models/InOutput';
import { InOutputMetricItem } from '../models/InOutputMetricItem';
import { InStateOutput } from '../models/InStateOutput';
import { InStateOutputMetricItem } from '../models/InStateOutputMetricItem';
import { InfoResponse } from '../models/InfoResponse';
import { InterfaceMetricItem } from '../models/InterfaceMetricItem';
import { Item } from '../models/Item';
import { JSONDict } from '../models/JSONDict';
import { L1Params } from '../models/L1Params';
import { Limits } from '../models/Limits';
import { LoginRequest } from '../models/LoginRequest';
import { LoginResponse } from '../models/LoginResponse';
import { MilestoneInfo } from '../models/MilestoneInfo';
import { MilestoneMetricItem } from '../models/MilestoneMetricItem';
import { NFTJSON } from '../models/NFTJSON';
import { NativeTokenIDRegistryResponse } from '../models/NativeTokenIDRegistryResponse';
import { NativeTokenJSON } from '../models/NativeTokenJSON';
import { NodeMessageMetrics } from '../models/NodeMessageMetrics';
import { NodeOwnerCertificateResponse } from '../models/NodeOwnerCertificateResponse';
import { OffLedgerRequest } from '../models/OffLedgerRequest';
import { OnLedgerRequest } from '../models/OnLedgerRequest';
import { OnLedgerRequestMetricItem } from '../models/OnLedgerRequestMetricItem';
import { Output } from '../models/Output';
import { OutputID } from '../models/OutputID';
import { PeeringNodeIdentityResponse } from '../models/PeeringNodeIdentityResponse';
import { PeeringNodeStatusResponse } from '../models/PeeringNodeStatusResponse';
import { PeeringTrustRequest } from '../models/PeeringTrustRequest';
import { ProtocolParameters } from '../models/ProtocolParameters';
import { PublicChainMetadata } from '../models/PublicChainMetadata';
import { PublisherStateTransactionItem } from '../models/PublisherStateTransactionItem';
import { Ratio32 } from '../models/Ratio32';
import { ReceiptResponse } from '../models/ReceiptResponse';
import { RentStructure } from '../models/RentStructure';
import { RequestIDsResponse } from '../models/RequestIDsResponse';
import { RequestJSON } from '../models/RequestJSON';
import { RequestProcessedResponse } from '../models/RequestProcessedResponse';
import { StateResponse } from '../models/StateResponse';
import { StateTransaction } from '../models/StateTransaction';
import { Transaction } from '../models/Transaction';
import { TransactionIDMetricItem } from '../models/TransactionIDMetricItem';
import { TransactionMetricItem } from '../models/TransactionMetricItem';
import { TxInclusionStateMsg } from '../models/TxInclusionStateMsg';
import { TxInclusionStateMsgMetricItem } from '../models/TxInclusionStateMsgMetricItem';
import { UTXOInputMetricItem } from '../models/UTXOInputMetricItem';
import { UnresolvedVMErrorJSON } from '../models/UnresolvedVMErrorJSON';
import { UpdateUserPasswordRequest } from '../models/UpdateUserPasswordRequest';
import { UpdateUserPermissionsRequest } from '../models/UpdateUserPermissionsRequest';
import { User } from '../models/User';
import { ValidationError } from '../models/ValidationError';
import { VersionResponse } from '../models/VersionResponse';
import { ObservableAuthApi } from './ObservableAPI';

import { AuthApiRequestFactory, AuthApiResponseProcessor} from "../apis/AuthApi";
export class PromiseAuthApi {
    private api: ObservableAuthApi

    public constructor(
        configuration: Configuration,
        requestFactory?: AuthApiRequestFactory,
        responseProcessor?: AuthApiResponseProcessor
    ) {
        this.api = new ObservableAuthApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Get information about the current authentication mode
     */
    public authInfo(_options?: Configuration): Promise<AuthInfoModel> {
        const result = this.api.authInfo(_options);
        return result.toPromise();
    }

    /**
     * Authenticate towards the node
     * @param loginRequest The login request
     */
    public authenticate(loginRequest: LoginRequest, _options?: Configuration): Promise<LoginResponse> {
        const result = this.api.authenticate(loginRequest, _options);
        return result.toPromise();
    }


}



import { ObservableChainsApi } from './ObservableAPI';

import { ChainsApiRequestFactory, ChainsApiResponseProcessor} from "../apis/ChainsApi";
export class PromiseChainsApi {
    private api: ObservableChainsApi

    public constructor(
        configuration: Configuration,
        requestFactory?: ChainsApiRequestFactory,
        responseProcessor?: ChainsApiResponseProcessor
    ) {
        this.api = new ObservableChainsApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Activate a chain
     * @param chainID ChainID (Bech32)
     */
    public activateChain(chainID: string, _options?: Configuration): Promise<void> {
        const result = this.api.activateChain(chainID, _options);
        return result.toPromise();
    }

    /**
     * Configure a trusted node to be an access node.
     * @param chainID ChainID (Bech32)
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public addAccessNode(chainID: string, peer: string, _options?: Configuration): Promise<void> {
        const result = this.api.addAccessNode(chainID, peer, _options);
        return result.toPromise();
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param chainID ChainID (Bech32)
     * @param contractCallViewRequest Parameters
     */
    public callView(chainID: string, contractCallViewRequest: ContractCallViewRequest, _options?: Configuration): Promise<JSONDict> {
        const result = this.api.callView(chainID, contractCallViewRequest, _options);
        return result.toPromise();
    }

    /**
     * Deactivate a chain
     * @param chainID ChainID (Bech32)
     */
    public deactivateChain(chainID: string, _options?: Configuration): Promise<void> {
        const result = this.api.deactivateChain(chainID, _options);
        return result.toPromise();
    }

    /**
     * dump accounts information into a humanly-readable format
     * @param chainID ChainID (Bech32)
     */
    public dumpAccounts(chainID: string, _options?: Configuration): Promise<void> {
        const result = this.api.dumpAccounts(chainID, _options);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param chainID ChainID (Bech32)
     * @param request Request
     */
    public estimateGasOffledger(chainID: string, request: EstimateGasRequestOffledger, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.estimateGasOffledger(chainID, request, _options);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param chainID ChainID (Bech32)
     * @param request Request
     */
    public estimateGasOnledger(chainID: string, request: EstimateGasRequestOnledger, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.estimateGasOnledger(chainID, request, _options);
        return result.toPromise();
    }

    /**
     * Get information about a specific chain
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public getChainInfo(chainID: string, block?: string, _options?: Configuration): Promise<ChainInfoResponse> {
        const result = this.api.getChainInfo(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get a list of all chains
     */
    public getChains(_options?: Configuration): Promise<Array<ChainInfoResponse>> {
        const result = this.api.getChains(_options);
        return result.toPromise();
    }

    /**
     * Get information about the deployed committee
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public getCommitteeInfo(chainID: string, block?: string, _options?: Configuration): Promise<CommitteeInfoResponse> {
        const result = this.api.getCommitteeInfo(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all available chain contracts
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public getContracts(chainID: string, block?: string, _options?: Configuration): Promise<Array<ContractInfoResponse>> {
        const result = this.api.getContracts(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the contents of the mempool.
     * @param chainID ChainID (Bech32)
     */
    public getMempoolContents(chainID: string, _options?: Configuration): Promise<Array<number>> {
        const result = this.api.getMempoolContents(chainID, _options);
        return result.toPromise();
    }

    /**
     * Get a receipt from a request ID
     * @param chainID ChainID (Bech32)
     * @param requestID RequestID (Hex)
     */
    public getReceipt(chainID: string, requestID: string, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.getReceipt(chainID, requestID, _options);
        return result.toPromise();
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param chainID ChainID (Bech32)
     * @param stateKey State Key (Hex)
     */
    public getStateValue(chainID: string, stateKey: string, _options?: Configuration): Promise<StateResponse> {
        const result = this.api.getStateValue(chainID, stateKey, _options);
        return result.toPromise();
    }

    /**
     * Remove an access node.
     * @param chainID ChainID (Bech32)
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public removeAccessNode(chainID: string, peer: string, _options?: Configuration): Promise<void> {
        const result = this.api.removeAccessNode(chainID, peer, _options);
        return result.toPromise();
    }

    /**
     * Sets the chain record.
     * @param chainID ChainID (Bech32)
     * @param chainRecord Chain Record
     */
    public setChainRecord(chainID: string, chainRecord: ChainRecord, _options?: Configuration): Promise<void> {
        const result = this.api.setChainRecord(chainID, chainRecord, _options);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC
     * @param chainID ChainID (Bech32)
     */
    public v1ChainsChainIDEvmPost(chainID: string, _options?: Configuration): Promise<void> {
        const result = this.api.v1ChainsChainIDEvmPost(chainID, _options);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     * @param chainID ChainID (Bech32)
     */
    public v1ChainsChainIDEvmWsGet(chainID: string, _options?: Configuration): Promise<void> {
        const result = this.api.v1ChainsChainIDEvmWsGet(chainID, _options);
        return result.toPromise();
    }

    /**
     * Wait until the given request has been processed by the node
     * @param chainID ChainID (Bech32)
     * @param requestID RequestID (Hex)
     * @param timeoutSeconds The timeout in seconds, maximum 60s
     * @param waitForL1Confirmation Wait for the block to be confirmed on L1
     */
    public waitForRequest(chainID: string, requestID: string, timeoutSeconds?: number, waitForL1Confirmation?: boolean, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.waitForRequest(chainID, requestID, timeoutSeconds, waitForL1Confirmation, _options);
        return result.toPromise();
    }


}



import { ObservableCorecontractsApi } from './ObservableAPI';

import { CorecontractsApiRequestFactory, CorecontractsApiResponseProcessor} from "../apis/CorecontractsApi";
export class PromiseCorecontractsApi {
    private api: ObservableCorecontractsApi

    public constructor(
        configuration: Configuration,
        requestFactory?: CorecontractsApiRequestFactory,
        responseProcessor?: CorecontractsApiResponseProcessor
    ) {
        this.api = new ObservableCorecontractsApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Get all assets belonging to an account
     * @param chainID ChainID (Bech32)
     * @param agentID AgentID (Bech32 for WasmVM | Hex for EVM)
     * @param block Block index or trie root
     */
    public accountsGetAccountBalance(chainID: string, agentID: string, block?: string, _options?: Configuration): Promise<AssetsResponse> {
        const result = this.api.accountsGetAccountBalance(chainID, agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all foundries owned by an account
     * @param chainID ChainID (Bech32)
     * @param agentID AgentID (Bech32 for WasmVM | Hex for EVM)
     * @param block Block index or trie root
     */
    public accountsGetAccountFoundries(chainID: string, agentID: string, block?: string, _options?: Configuration): Promise<AccountFoundriesResponse> {
        const result = this.api.accountsGetAccountFoundries(chainID, agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all NFT ids belonging to an account
     * @param chainID ChainID (Bech32)
     * @param agentID AgentID (Bech32 for WasmVM | Hex for EVM)
     * @param block Block index or trie root
     */
    public accountsGetAccountNFTIDs(chainID: string, agentID: string, block?: string, _options?: Configuration): Promise<AccountNFTsResponse> {
        const result = this.api.accountsGetAccountNFTIDs(chainID, agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the current nonce of an account
     * @param chainID ChainID (Bech32)
     * @param agentID AgentID (Bech32 for WasmVM | Hex for EVM)
     * @param block Block index or trie root
     */
    public accountsGetAccountNonce(chainID: string, agentID: string, block?: string, _options?: Configuration): Promise<AccountNonceResponse> {
        const result = this.api.accountsGetAccountNonce(chainID, agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the foundry output
     * @param chainID ChainID (Bech32)
     * @param serialNumber Serial Number (uint32)
     * @param block Block index or trie root
     */
    public accountsGetFoundryOutput(chainID: string, serialNumber: number, block?: string, _options?: Configuration): Promise<FoundryOutputResponse> {
        const result = this.api.accountsGetFoundryOutput(chainID, serialNumber, block, _options);
        return result.toPromise();
    }

    /**
     * Get the NFT data by an ID
     * @param chainID ChainID (Bech32)
     * @param nftID NFT ID (Hex)
     * @param block Block index or trie root
     */
    public accountsGetNFTData(chainID: string, nftID: string, block?: string, _options?: Configuration): Promise<NFTJSON> {
        const result = this.api.accountsGetNFTData(chainID, nftID, block, _options);
        return result.toPromise();
    }

    /**
     * Get a list of all registries
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public accountsGetNativeTokenIDRegistry(chainID: string, block?: string, _options?: Configuration): Promise<NativeTokenIDRegistryResponse> {
        const result = this.api.accountsGetNativeTokenIDRegistry(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all stored assets
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public accountsGetTotalAssets(chainID: string, block?: string, _options?: Configuration): Promise<AssetsResponse> {
        const result = this.api.accountsGetTotalAssets(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all fields of a blob
     * @param chainID ChainID (Bech32)
     * @param blobHash BlobHash (Hex)
     * @param block Block index or trie root
     */
    public blobsGetBlobInfo(chainID: string, blobHash: string, block?: string, _options?: Configuration): Promise<BlobInfoResponse> {
        const result = this.api.blobsGetBlobInfo(chainID, blobHash, block, _options);
        return result.toPromise();
    }

    /**
     * Get the value of the supplied field (key)
     * @param chainID ChainID (Bech32)
     * @param blobHash BlobHash (Hex)
     * @param fieldKey FieldKey (String)
     * @param block Block index or trie root
     */
    public blobsGetBlobValue(chainID: string, blobHash: string, fieldKey: string, block?: string, _options?: Configuration): Promise<BlobValueResponse> {
        const result = this.api.blobsGetBlobValue(chainID, blobHash, fieldKey, block, _options);
        return result.toPromise();
    }

    /**
     * Get the block info of a certain block index
     * @param chainID ChainID (Bech32)
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public blocklogGetBlockInfo(chainID: string, blockIndex: number, block?: string, _options?: Configuration): Promise<BlockInfoResponse> {
        const result = this.api.blocklogGetBlockInfo(chainID, blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get the control addresses
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public blocklogGetControlAddresses(chainID: string, block?: string, _options?: Configuration): Promise<ControlAddressesResponse> {
        const result = this.api.blocklogGetControlAddresses(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get events of a block
     * @param chainID ChainID (Bech32)
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public blocklogGetEventsOfBlock(chainID: string, blockIndex: number, block?: string, _options?: Configuration): Promise<EventsResponse> {
        const result = this.api.blocklogGetEventsOfBlock(chainID, blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get events of the latest block
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public blocklogGetEventsOfLatestBlock(chainID: string, block?: string, _options?: Configuration): Promise<EventsResponse> {
        const result = this.api.blocklogGetEventsOfLatestBlock(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get events of a request
     * @param chainID ChainID (Bech32)
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public blocklogGetEventsOfRequest(chainID: string, requestID: string, block?: string, _options?: Configuration): Promise<EventsResponse> {
        const result = this.api.blocklogGetEventsOfRequest(chainID, requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the block info of the latest block
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public blocklogGetLatestBlockInfo(chainID: string, block?: string, _options?: Configuration): Promise<BlockInfoResponse> {
        const result = this.api.blocklogGetLatestBlockInfo(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the request ids for a certain block index
     * @param chainID ChainID (Bech32)
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public blocklogGetRequestIDsForBlock(chainID: string, blockIndex: number, block?: string, _options?: Configuration): Promise<RequestIDsResponse> {
        const result = this.api.blocklogGetRequestIDsForBlock(chainID, blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get the request ids for the latest block
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public blocklogGetRequestIDsForLatestBlock(chainID: string, block?: string, _options?: Configuration): Promise<RequestIDsResponse> {
        const result = this.api.blocklogGetRequestIDsForLatestBlock(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the request processing status
     * @param chainID ChainID (Bech32)
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public blocklogGetRequestIsProcessed(chainID: string, requestID: string, block?: string, _options?: Configuration): Promise<RequestProcessedResponse> {
        const result = this.api.blocklogGetRequestIsProcessed(chainID, requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the receipt of a certain request id
     * @param chainID ChainID (Bech32)
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public blocklogGetRequestReceipt(chainID: string, requestID: string, block?: string, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.blocklogGetRequestReceipt(chainID, requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all receipts of a certain block
     * @param chainID ChainID (Bech32)
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public blocklogGetRequestReceiptsOfBlock(chainID: string, blockIndex: number, block?: string, _options?: Configuration): Promise<Array<ReceiptResponse>> {
        const result = this.api.blocklogGetRequestReceiptsOfBlock(chainID, blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get all receipts of the latest block
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public blocklogGetRequestReceiptsOfLatestBlock(chainID: string, block?: string, _options?: Configuration): Promise<Array<ReceiptResponse>> {
        const result = this.api.blocklogGetRequestReceiptsOfLatestBlock(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the error message format of a specific error id
     * @param chainID ChainID (Bech32)
     * @param contractHname Contract (Hname as Hex)
     * @param errorID Error Id (uint16)
     * @param block Block index or trie root
     */
    public errorsGetErrorMessageFormat(chainID: string, contractHname: string, errorID: number, block?: string, _options?: Configuration): Promise<ErrorMessageFormatResponse> {
        const result = this.api.errorsGetErrorMessageFormat(chainID, contractHname, errorID, block, _options);
        return result.toPromise();
    }

    /**
     * Returns the allowed state controller addresses
     * Get the allowed state controller addresses
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public governanceGetAllowedStateControllerAddresses(chainID: string, block?: string, _options?: Configuration): Promise<GovAllowedStateControllerAddressesResponse> {
        const result = this.api.governanceGetAllowedStateControllerAddresses(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * If you are using the common API functions, you most likely rather want to use '/v1/chains/:chainID' to get information about a chain.
     * Get the chain info
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public governanceGetChainInfo(chainID: string, block?: string, _options?: Configuration): Promise<GovChainInfoResponse> {
        const result = this.api.governanceGetChainInfo(chainID, block, _options);
        return result.toPromise();
    }

    /**
     * Returns the chain owner
     * Get the chain owner
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public governanceGetChainOwner(chainID: string, block?: string, _options?: Configuration): Promise<GovChainOwnerResponse> {
        const result = this.api.governanceGetChainOwner(chainID, block, _options);
        return result.toPromise();
    }


}



import { ObservableDefaultApi } from './ObservableAPI';

import { DefaultApiRequestFactory, DefaultApiResponseProcessor} from "../apis/DefaultApi";
export class PromiseDefaultApi {
    private api: ObservableDefaultApi

    public constructor(
        configuration: Configuration,
        requestFactory?: DefaultApiRequestFactory,
        responseProcessor?: DefaultApiResponseProcessor
    ) {
        this.api = new ObservableDefaultApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Returns 200 if the node is healthy.
     */
    public getHealth(_options?: Configuration): Promise<void> {
        const result = this.api.getHealth(_options);
        return result.toPromise();
    }

    /**
     * The websocket connection service
     */
    public v1WsGet(_options?: Configuration): Promise<void> {
        const result = this.api.v1WsGet(_options);
        return result.toPromise();
    }


}



import { ObservableMetricsApi } from './ObservableAPI';

import { MetricsApiRequestFactory, MetricsApiResponseProcessor} from "../apis/MetricsApi";
export class PromiseMetricsApi {
    private api: ObservableMetricsApi

    public constructor(
        configuration: Configuration,
        requestFactory?: MetricsApiRequestFactory,
        responseProcessor?: MetricsApiResponseProcessor
    ) {
        this.api = new ObservableMetricsApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Get chain specific message metrics.
     * @param chainID ChainID (Bech32)
     */
    public getChainMessageMetrics(chainID: string, _options?: Configuration): Promise<ChainMessageMetrics> {
        const result = this.api.getChainMessageMetrics(chainID, _options);
        return result.toPromise();
    }

    /**
     * Get chain pipe event metrics.
     * @param chainID ChainID (Bech32)
     */
    public getChainPipeMetrics(chainID: string, _options?: Configuration): Promise<ConsensusPipeMetrics> {
        const result = this.api.getChainPipeMetrics(chainID, _options);
        return result.toPromise();
    }

    /**
     * Get chain workflow metrics.
     * @param chainID ChainID (Bech32)
     */
    public getChainWorkflowMetrics(chainID: string, _options?: Configuration): Promise<ConsensusWorkflowMetrics> {
        const result = this.api.getChainWorkflowMetrics(chainID, _options);
        return result.toPromise();
    }

    /**
     * Get accumulated message metrics.
     */
    public getNodeMessageMetrics(_options?: Configuration): Promise<NodeMessageMetrics> {
        const result = this.api.getNodeMessageMetrics(_options);
        return result.toPromise();
    }


}



import { ObservableNodeApi } from './ObservableAPI';

import { NodeApiRequestFactory, NodeApiResponseProcessor} from "../apis/NodeApi";
export class PromiseNodeApi {
    private api: ObservableNodeApi

    public constructor(
        configuration: Configuration,
        requestFactory?: NodeApiRequestFactory,
        responseProcessor?: NodeApiResponseProcessor
    ) {
        this.api = new ObservableNodeApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Distrust a peering node
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public distrustPeer(peer: string, _options?: Configuration): Promise<void> {
        const result = this.api.distrustPeer(peer, _options);
        return result.toPromise();
    }

    /**
     * Generate a new distributed key
     * @param dKSharesPostRequest Request parameters
     */
    public generateDKS(dKSharesPostRequest: DKSharesPostRequest, _options?: Configuration): Promise<DKSharesInfo> {
        const result = this.api.generateDKS(dKSharesPostRequest, _options);
        return result.toPromise();
    }

    /**
     * Get basic information about all configured peers
     */
    public getAllPeers(_options?: Configuration): Promise<Array<PeeringNodeStatusResponse>> {
        const result = this.api.getAllPeers(_options);
        return result.toPromise();
    }

    /**
     * Return the Wasp configuration
     */
    public getConfiguration(_options?: Configuration): Promise<{ [key: string]: string; }> {
        const result = this.api.getConfiguration(_options);
        return result.toPromise();
    }

    /**
     * Get information about the shared address DKS configuration
     * @param sharedAddress SharedAddress (Bech32)
     */
    public getDKSInfo(sharedAddress: string, _options?: Configuration): Promise<DKSharesInfo> {
        const result = this.api.getDKSInfo(sharedAddress, _options);
        return result.toPromise();
    }

    /**
     * Returns private information about this node.
     */
    public getInfo(_options?: Configuration): Promise<InfoResponse> {
        const result = this.api.getInfo(_options);
        return result.toPromise();
    }

    /**
     * Get basic peer info of the current node
     */
    public getPeeringIdentity(_options?: Configuration): Promise<PeeringNodeIdentityResponse> {
        const result = this.api.getPeeringIdentity(_options);
        return result.toPromise();
    }

    /**
     * Get trusted peers
     */
    public getTrustedPeers(_options?: Configuration): Promise<Array<PeeringNodeIdentityResponse>> {
        const result = this.api.getTrustedPeers(_options);
        return result.toPromise();
    }

    /**
     * Returns the node version.
     */
    public getVersion(_options?: Configuration): Promise<VersionResponse> {
        const result = this.api.getVersion(_options);
        return result.toPromise();
    }

    /**
     * Gets the node owner
     */
    public ownerCertificate(_options?: Configuration): Promise<NodeOwnerCertificateResponse> {
        const result = this.api.ownerCertificate(_options);
        return result.toPromise();
    }

    /**
     * Shut down the node
     */
    public shutdownNode(_options?: Configuration): Promise<void> {
        const result = this.api.shutdownNode(_options);
        return result.toPromise();
    }

    /**
     * Trust a peering node
     * @param peeringTrustRequest Info of the peer to trust
     */
    public trustPeer(peeringTrustRequest: PeeringTrustRequest, _options?: Configuration): Promise<void> {
        const result = this.api.trustPeer(peeringTrustRequest, _options);
        return result.toPromise();
    }


}



import { ObservableRequestsApi } from './ObservableAPI';

import { RequestsApiRequestFactory, RequestsApiResponseProcessor} from "../apis/RequestsApi";
export class PromiseRequestsApi {
    private api: ObservableRequestsApi

    public constructor(
        configuration: Configuration,
        requestFactory?: RequestsApiRequestFactory,
        responseProcessor?: RequestsApiResponseProcessor
    ) {
        this.api = new ObservableRequestsApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Post an off-ledger request
     * @param offLedgerRequest Offledger request as JSON. Request encoded in Hex
     */
    public offLedger(offLedgerRequest: OffLedgerRequest, _options?: Configuration): Promise<void> {
        const result = this.api.offLedger(offLedgerRequest, _options);
        return result.toPromise();
    }


}



import { ObservableUsersApi } from './ObservableAPI';

import { UsersApiRequestFactory, UsersApiResponseProcessor} from "../apis/UsersApi";
export class PromiseUsersApi {
    private api: ObservableUsersApi

    public constructor(
        configuration: Configuration,
        requestFactory?: UsersApiRequestFactory,
        responseProcessor?: UsersApiResponseProcessor
    ) {
        this.api = new ObservableUsersApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Add a user
     * @param addUserRequest The user data
     */
    public addUser(addUserRequest: AddUserRequest, _options?: Configuration): Promise<void> {
        const result = this.api.addUser(addUserRequest, _options);
        return result.toPromise();
    }

    /**
     * Change user password
     * @param username The username
     * @param updateUserPasswordRequest The users new password
     */
    public changeUserPassword(username: string, updateUserPasswordRequest: UpdateUserPasswordRequest, _options?: Configuration): Promise<void> {
        const result = this.api.changeUserPassword(username, updateUserPasswordRequest, _options);
        return result.toPromise();
    }

    /**
     * Change user permissions
     * @param username The username
     * @param updateUserPermissionsRequest The users new permissions
     */
    public changeUserPermissions(username: string, updateUserPermissionsRequest: UpdateUserPermissionsRequest, _options?: Configuration): Promise<void> {
        const result = this.api.changeUserPermissions(username, updateUserPermissionsRequest, _options);
        return result.toPromise();
    }

    /**
     * Deletes a user
     * @param username The username
     */
    public deleteUser(username: string, _options?: Configuration): Promise<void> {
        const result = this.api.deleteUser(username, _options);
        return result.toPromise();
    }

    /**
     * Get a user
     * @param username The username
     */
    public getUser(username: string, _options?: Configuration): Promise<User> {
        const result = this.api.getUser(username, _options);
        return result.toPromise();
    }

    /**
     * Get a list of all users
     */
    public getUsers(_options?: Configuration): Promise<Array<User>> {
        const result = this.api.getUsers(_options);
        return result.toPromise();
    }


}



