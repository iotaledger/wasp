import { ResponseContext, RequestContext, HttpFile, HttpInfo } from '../http/http';
import { Configuration} from '../configuration'

import { AccountNonceResponse } from '../models/AccountNonceResponse';
import { AccountObjectsResponse } from '../models/AccountObjectsResponse';
import { AddUserRequest } from '../models/AddUserRequest';
import { AnchorMetricItem } from '../models/AnchorMetricItem';
import { AssetsJSON } from '../models/AssetsJSON';
import { AssetsResponse } from '../models/AssetsResponse';
import { AuthInfoModel } from '../models/AuthInfoModel';
import { BigInt } from '../models/BigInt';
import { BlockInfoResponse } from '../models/BlockInfoResponse';
import { BurnRecord } from '../models/BurnRecord';
import { CallTargetJSON } from '../models/CallTargetJSON';
import { ChainInfoResponse } from '../models/ChainInfoResponse';
import { ChainMessageMetrics } from '../models/ChainMessageMetrics';
import { ChainRecord } from '../models/ChainRecord';
import { CoinJSON } from '../models/CoinJSON';
import { CommitteeInfoResponse } from '../models/CommitteeInfoResponse';
import { CommitteeNode } from '../models/CommitteeNode';
import { ConsensusPipeMetrics } from '../models/ConsensusPipeMetrics';
import { ConsensusWorkflowMetrics } from '../models/ConsensusWorkflowMetrics';
import { ContractCallViewRequest } from '../models/ContractCallViewRequest';
import { ControlAddressesResponse } from '../models/ControlAddressesResponse';
import { DKSharesInfo } from '../models/DKSharesInfo';
import { DKSharesPostRequest } from '../models/DKSharesPostRequest';
import { ErrorMessageFormatResponse } from '../models/ErrorMessageFormatResponse';
import { EstimateGasRequestOffledger } from '../models/EstimateGasRequestOffledger';
import { EstimateGasRequestOnledger } from '../models/EstimateGasRequestOnledger';
import { EventJSON } from '../models/EventJSON';
import { EventsResponse } from '../models/EventsResponse';
import { FeePolicy } from '../models/FeePolicy';
import { GovChainAdminResponse } from '../models/GovChainAdminResponse';
import { GovChainInfoResponse } from '../models/GovChainInfoResponse';
import { GovPublicChainMetadata } from '../models/GovPublicChainMetadata';
import { InfoResponse } from '../models/InfoResponse';
import { Int } from '../models/Int';
import { IotaCoinInfo } from '../models/IotaCoinInfo';
import { IotaObject } from '../models/IotaObject';
import { L1Params } from '../models/L1Params';
import { Limits } from '../models/Limits';
import { LoginRequest } from '../models/LoginRequest';
import { LoginResponse } from '../models/LoginResponse';
import { NodeOwnerCertificateResponse } from '../models/NodeOwnerCertificateResponse';
import { ObjectType } from '../models/ObjectType';
import { OffLedgerRequest } from '../models/OffLedgerRequest';
import { OnLedgerRequest } from '../models/OnLedgerRequest';
import { OnLedgerRequestMetricItem } from '../models/OnLedgerRequestMetricItem';
import { PeeringNodeIdentityResponse } from '../models/PeeringNodeIdentityResponse';
import { PeeringNodeStatusResponse } from '../models/PeeringNodeStatusResponse';
import { PeeringTrustRequest } from '../models/PeeringTrustRequest';
import { Protocol } from '../models/Protocol';
import { PublicChainMetadata } from '../models/PublicChainMetadata';
import { PublisherStateTransactionItem } from '../models/PublisherStateTransactionItem';
import { Ratio32 } from '../models/Ratio32';
import { ReceiptResponse } from '../models/ReceiptResponse';
import { RequestIDsResponse } from '../models/RequestIDsResponse';
import { RequestJSON } from '../models/RequestJSON';
import { RequestProcessedResponse } from '../models/RequestProcessedResponse';
import { RotateChainRequest } from '../models/RotateChainRequest';
import { StateAnchor } from '../models/StateAnchor';
import { StateResponse } from '../models/StateResponse';
import { StateTransaction } from '../models/StateTransaction';
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
    public authInfoWithHttpInfo(_options?: Configuration): Promise<HttpInfo<AuthInfoModel>> {
        const result = this.api.authInfoWithHttpInfo(_options);
        return result.toPromise();
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
    public authenticateWithHttpInfo(loginRequest: LoginRequest, _options?: Configuration): Promise<HttpInfo<LoginResponse>> {
        const result = this.api.authenticateWithHttpInfo(loginRequest, _options);
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
     * @param chainID ChainID (Hex Address)
     */
    public activateChainWithHttpInfo(chainID: string, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.activateChainWithHttpInfo(chainID, _options);
        return result.toPromise();
    }

    /**
     * Activate a chain
     * @param chainID ChainID (Hex Address)
     */
    public activateChain(chainID: string, _options?: Configuration): Promise<void> {
        const result = this.api.activateChain(chainID, _options);
        return result.toPromise();
    }

    /**
     * Configure a trusted node to be an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public addAccessNodeWithHttpInfo(peer: string, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.addAccessNodeWithHttpInfo(peer, _options);
        return result.toPromise();
    }

    /**
     * Configure a trusted node to be an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public addAccessNode(peer: string, _options?: Configuration): Promise<void> {
        const result = this.api.addAccessNode(peer, _options);
        return result.toPromise();
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param contractCallViewRequest Parameters
     */
    public callViewWithHttpInfo(contractCallViewRequest: ContractCallViewRequest, _options?: Configuration): Promise<HttpInfo<Array<string>>> {
        const result = this.api.callViewWithHttpInfo(contractCallViewRequest, _options);
        return result.toPromise();
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param contractCallViewRequest Parameters
     */
    public callView(contractCallViewRequest: ContractCallViewRequest, _options?: Configuration): Promise<Array<string>> {
        const result = this.api.callView(contractCallViewRequest, _options);
        return result.toPromise();
    }

    /**
     * Deactivate a chain
     */
    public deactivateChainWithHttpInfo(_options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.deactivateChainWithHttpInfo(_options);
        return result.toPromise();
    }

    /**
     * Deactivate a chain
     */
    public deactivateChain(_options?: Configuration): Promise<void> {
        const result = this.api.deactivateChain(_options);
        return result.toPromise();
    }

    /**
     * dump accounts information into a humanly-readable format
     */
    public dumpAccountsWithHttpInfo(_options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.dumpAccountsWithHttpInfo(_options);
        return result.toPromise();
    }

    /**
     * dump accounts information into a humanly-readable format
     */
    public dumpAccounts(_options?: Configuration): Promise<void> {
        const result = this.api.dumpAccounts(_options);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param request Request
     */
    public estimateGasOffledgerWithHttpInfo(request: EstimateGasRequestOffledger, _options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        const result = this.api.estimateGasOffledgerWithHttpInfo(request, _options);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param request Request
     */
    public estimateGasOffledger(request: EstimateGasRequestOffledger, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.estimateGasOffledger(request, _options);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param request Request
     */
    public estimateGasOnledgerWithHttpInfo(request: EstimateGasRequestOnledger, _options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        const result = this.api.estimateGasOnledgerWithHttpInfo(request, _options);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param request Request
     */
    public estimateGasOnledger(request: EstimateGasRequestOnledger, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.estimateGasOnledger(request, _options);
        return result.toPromise();
    }

    /**
     * Get information about the chain
     * @param [block] Block index or trie root
     */
    public getChainInfoWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<ChainInfoResponse>> {
        const result = this.api.getChainInfoWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get information about the chain
     * @param [block] Block index or trie root
     */
    public getChainInfo(block?: string, _options?: Configuration): Promise<ChainInfoResponse> {
        const result = this.api.getChainInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get information about the deployed committee
     * @param [block] Block index or trie root
     */
    public getCommitteeInfoWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<CommitteeInfoResponse>> {
        const result = this.api.getCommitteeInfoWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get information about the deployed committee
     * @param [block] Block index or trie root
     */
    public getCommitteeInfo(block?: string, _options?: Configuration): Promise<CommitteeInfoResponse> {
        const result = this.api.getCommitteeInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get the contents of the mempool.
     */
    public getMempoolContentsWithHttpInfo(_options?: Configuration): Promise<HttpInfo<Array<number>>> {
        const result = this.api.getMempoolContentsWithHttpInfo(_options);
        return result.toPromise();
    }

    /**
     * Get the contents of the mempool.
     */
    public getMempoolContents(_options?: Configuration): Promise<Array<number>> {
        const result = this.api.getMempoolContents(_options);
        return result.toPromise();
    }

    /**
     * Get a receipt from a request ID
     * @param requestID RequestID (Hex)
     */
    public getReceiptWithHttpInfo(requestID: string, _options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        const result = this.api.getReceiptWithHttpInfo(requestID, _options);
        return result.toPromise();
    }

    /**
     * Get a receipt from a request ID
     * @param requestID RequestID (Hex)
     */
    public getReceipt(requestID: string, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.getReceipt(requestID, _options);
        return result.toPromise();
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param stateKey State Key (Hex)
     */
    public getStateValueWithHttpInfo(stateKey: string, _options?: Configuration): Promise<HttpInfo<StateResponse>> {
        const result = this.api.getStateValueWithHttpInfo(stateKey, _options);
        return result.toPromise();
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param stateKey State Key (Hex)
     */
    public getStateValue(stateKey: string, _options?: Configuration): Promise<StateResponse> {
        const result = this.api.getStateValue(stateKey, _options);
        return result.toPromise();
    }

    /**
     * Remove an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public removeAccessNodeWithHttpInfo(peer: string, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.removeAccessNodeWithHttpInfo(peer, _options);
        return result.toPromise();
    }

    /**
     * Remove an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public removeAccessNode(peer: string, _options?: Configuration): Promise<void> {
        const result = this.api.removeAccessNode(peer, _options);
        return result.toPromise();
    }

    /**
     * Rotate a chain
     * @param [rotateRequest] RotateRequest
     */
    public rotateChainWithHttpInfo(rotateRequest?: RotateChainRequest, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.rotateChainWithHttpInfo(rotateRequest, _options);
        return result.toPromise();
    }

    /**
     * Rotate a chain
     * @param [rotateRequest] RotateRequest
     */
    public rotateChain(rotateRequest?: RotateChainRequest, _options?: Configuration): Promise<void> {
        const result = this.api.rotateChain(rotateRequest, _options);
        return result.toPromise();
    }

    /**
     * Sets the chain record.
     * @param chainID ChainID (Hex Address)
     * @param chainRecord Chain Record
     */
    public setChainRecordWithHttpInfo(chainID: string, chainRecord: ChainRecord, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.setChainRecordWithHttpInfo(chainID, chainRecord, _options);
        return result.toPromise();
    }

    /**
     * Sets the chain record.
     * @param chainID ChainID (Hex Address)
     * @param chainRecord Chain Record
     */
    public setChainRecord(chainID: string, chainRecord: ChainRecord, _options?: Configuration): Promise<void> {
        const result = this.api.setChainRecord(chainID, chainRecord, _options);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC
     */
    public v1ChainEvmPostWithHttpInfo(_options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.v1ChainEvmPostWithHttpInfo(_options);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC
     */
    public v1ChainEvmPost(_options?: Configuration): Promise<void> {
        const result = this.api.v1ChainEvmPost(_options);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     */
    public v1ChainEvmWsGetWithHttpInfo(_options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.v1ChainEvmWsGetWithHttpInfo(_options);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     */
    public v1ChainEvmWsGet(_options?: Configuration): Promise<void> {
        const result = this.api.v1ChainEvmWsGet(_options);
        return result.toPromise();
    }

    /**
     * Wait until the given request has been processed by the node
     * @param requestID RequestID (Hex)
     * @param [timeoutSeconds] The timeout in seconds, maximum 60s
     * @param [waitForL1Confirmation] Wait for the block to be confirmed on L1
     */
    public waitForRequestWithHttpInfo(requestID: string, timeoutSeconds?: number, waitForL1Confirmation?: boolean, _options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        const result = this.api.waitForRequestWithHttpInfo(requestID, timeoutSeconds, waitForL1Confirmation, _options);
        return result.toPromise();
    }

    /**
     * Wait until the given request has been processed by the node
     * @param requestID RequestID (Hex)
     * @param [timeoutSeconds] The timeout in seconds, maximum 60s
     * @param [waitForL1Confirmation] Wait for the block to be confirmed on L1
     */
    public waitForRequest(requestID: string, timeoutSeconds?: number, waitForL1Confirmation?: boolean, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.waitForRequest(requestID, timeoutSeconds, waitForL1Confirmation, _options);
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
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountBalanceWithHttpInfo(agentID: string, block?: string, _options?: Configuration): Promise<HttpInfo<AssetsResponse>> {
        const result = this.api.accountsGetAccountBalanceWithHttpInfo(agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all assets belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountBalance(agentID: string, block?: string, _options?: Configuration): Promise<AssetsResponse> {
        const result = this.api.accountsGetAccountBalance(agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the current nonce of an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountNonceWithHttpInfo(agentID: string, block?: string, _options?: Configuration): Promise<HttpInfo<AccountNonceResponse>> {
        const result = this.api.accountsGetAccountNonceWithHttpInfo(agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the current nonce of an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountNonce(agentID: string, block?: string, _options?: Configuration): Promise<AccountNonceResponse> {
        const result = this.api.accountsGetAccountNonce(agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all object ids belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountObjectIDsWithHttpInfo(agentID: string, block?: string, _options?: Configuration): Promise<HttpInfo<AccountObjectsResponse>> {
        const result = this.api.accountsGetAccountObjectIDsWithHttpInfo(agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all object ids belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountObjectIDs(agentID: string, block?: string, _options?: Configuration): Promise<AccountObjectsResponse> {
        const result = this.api.accountsGetAccountObjectIDs(agentID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all stored assets
     * @param [block] Block index or trie root
     */
    public accountsGetTotalAssetsWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<AssetsResponse>> {
        const result = this.api.accountsGetTotalAssetsWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get all stored assets
     * @param [block] Block index or trie root
     */
    public accountsGetTotalAssets(block?: string, _options?: Configuration): Promise<AssetsResponse> {
        const result = this.api.accountsGetTotalAssets(block, _options);
        return result.toPromise();
    }

    /**
     * Get the block info of a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetBlockInfoWithHttpInfo(blockIndex: number, block?: string, _options?: Configuration): Promise<HttpInfo<BlockInfoResponse>> {
        const result = this.api.blocklogGetBlockInfoWithHttpInfo(blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get the block info of a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetBlockInfo(blockIndex: number, block?: string, _options?: Configuration): Promise<BlockInfoResponse> {
        const result = this.api.blocklogGetBlockInfo(blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get the control addresses
     * @param [block] Block index or trie root
     */
    public blocklogGetControlAddressesWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<ControlAddressesResponse>> {
        const result = this.api.blocklogGetControlAddressesWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get the control addresses
     * @param [block] Block index or trie root
     */
    public blocklogGetControlAddresses(block?: string, _options?: Configuration): Promise<ControlAddressesResponse> {
        const result = this.api.blocklogGetControlAddresses(block, _options);
        return result.toPromise();
    }

    /**
     * Get events of a block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfBlockWithHttpInfo(blockIndex: number, block?: string, _options?: Configuration): Promise<HttpInfo<EventsResponse>> {
        const result = this.api.blocklogGetEventsOfBlockWithHttpInfo(blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get events of a block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfBlock(blockIndex: number, block?: string, _options?: Configuration): Promise<EventsResponse> {
        const result = this.api.blocklogGetEventsOfBlock(blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get events of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfLatestBlockWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<EventsResponse>> {
        const result = this.api.blocklogGetEventsOfLatestBlockWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get events of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfLatestBlock(block?: string, _options?: Configuration): Promise<EventsResponse> {
        const result = this.api.blocklogGetEventsOfLatestBlock(block, _options);
        return result.toPromise();
    }

    /**
     * Get events of a request
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfRequestWithHttpInfo(requestID: string, block?: string, _options?: Configuration): Promise<HttpInfo<EventsResponse>> {
        const result = this.api.blocklogGetEventsOfRequestWithHttpInfo(requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get events of a request
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfRequest(requestID: string, block?: string, _options?: Configuration): Promise<EventsResponse> {
        const result = this.api.blocklogGetEventsOfRequest(requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the block info of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetLatestBlockInfoWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<BlockInfoResponse>> {
        const result = this.api.blocklogGetLatestBlockInfoWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get the block info of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetLatestBlockInfo(block?: string, _options?: Configuration): Promise<BlockInfoResponse> {
        const result = this.api.blocklogGetLatestBlockInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get the request ids for a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForBlockWithHttpInfo(blockIndex: number, block?: string, _options?: Configuration): Promise<HttpInfo<RequestIDsResponse>> {
        const result = this.api.blocklogGetRequestIDsForBlockWithHttpInfo(blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get the request ids for a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForBlock(blockIndex: number, block?: string, _options?: Configuration): Promise<RequestIDsResponse> {
        const result = this.api.blocklogGetRequestIDsForBlock(blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get the request ids for the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForLatestBlockWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<RequestIDsResponse>> {
        const result = this.api.blocklogGetRequestIDsForLatestBlockWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get the request ids for the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForLatestBlock(block?: string, _options?: Configuration): Promise<RequestIDsResponse> {
        const result = this.api.blocklogGetRequestIDsForLatestBlock(block, _options);
        return result.toPromise();
    }

    /**
     * Get the request processing status
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIsProcessedWithHttpInfo(requestID: string, block?: string, _options?: Configuration): Promise<HttpInfo<RequestProcessedResponse>> {
        const result = this.api.blocklogGetRequestIsProcessedWithHttpInfo(requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the request processing status
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIsProcessed(requestID: string, block?: string, _options?: Configuration): Promise<RequestProcessedResponse> {
        const result = this.api.blocklogGetRequestIsProcessed(requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the receipt of a certain request id
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptWithHttpInfo(requestID: string, block?: string, _options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        const result = this.api.blocklogGetRequestReceiptWithHttpInfo(requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the receipt of a certain request id
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceipt(requestID: string, block?: string, _options?: Configuration): Promise<ReceiptResponse> {
        const result = this.api.blocklogGetRequestReceipt(requestID, block, _options);
        return result.toPromise();
    }

    /**
     * Get all receipts of a certain block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfBlockWithHttpInfo(blockIndex: number, block?: string, _options?: Configuration): Promise<HttpInfo<Array<ReceiptResponse>>> {
        const result = this.api.blocklogGetRequestReceiptsOfBlockWithHttpInfo(blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get all receipts of a certain block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfBlock(blockIndex: number, block?: string, _options?: Configuration): Promise<Array<ReceiptResponse>> {
        const result = this.api.blocklogGetRequestReceiptsOfBlock(blockIndex, block, _options);
        return result.toPromise();
    }

    /**
     * Get all receipts of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<Array<ReceiptResponse>>> {
        const result = this.api.blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Get all receipts of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfLatestBlock(block?: string, _options?: Configuration): Promise<Array<ReceiptResponse>> {
        const result = this.api.blocklogGetRequestReceiptsOfLatestBlock(block, _options);
        return result.toPromise();
    }

    /**
     * Get the error message format of a specific error id
     * @param chainID ChainID (Hex Address)
     * @param contractHname Contract (Hname as Hex)
     * @param errorID Error Id (uint16)
     * @param [block] Block index or trie root
     */
    public errorsGetErrorMessageFormatWithHttpInfo(chainID: string, contractHname: string, errorID: number, block?: string, _options?: Configuration): Promise<HttpInfo<ErrorMessageFormatResponse>> {
        const result = this.api.errorsGetErrorMessageFormatWithHttpInfo(chainID, contractHname, errorID, block, _options);
        return result.toPromise();
    }

    /**
     * Get the error message format of a specific error id
     * @param chainID ChainID (Hex Address)
     * @param contractHname Contract (Hname as Hex)
     * @param errorID Error Id (uint16)
     * @param [block] Block index or trie root
     */
    public errorsGetErrorMessageFormat(chainID: string, contractHname: string, errorID: number, block?: string, _options?: Configuration): Promise<ErrorMessageFormatResponse> {
        const result = this.api.errorsGetErrorMessageFormat(chainID, contractHname, errorID, block, _options);
        return result.toPromise();
    }

    /**
     * Returns the chain admin
     * Get the chain admin
     * @param [block] Block index or trie root
     */
    public governanceGetChainAdminWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<GovChainAdminResponse>> {
        const result = this.api.governanceGetChainAdminWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * Returns the chain admin
     * Get the chain admin
     * @param [block] Block index or trie root
     */
    public governanceGetChainAdmin(block?: string, _options?: Configuration): Promise<GovChainAdminResponse> {
        const result = this.api.governanceGetChainAdmin(block, _options);
        return result.toPromise();
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chain\' to get information about the chain.
     * Get the chain info
     * @param [block] Block index or trie root
     */
    public governanceGetChainInfoWithHttpInfo(block?: string, _options?: Configuration): Promise<HttpInfo<GovChainInfoResponse>> {
        const result = this.api.governanceGetChainInfoWithHttpInfo(block, _options);
        return result.toPromise();
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chain\' to get information about the chain.
     * Get the chain info
     * @param [block] Block index or trie root
     */
    public governanceGetChainInfo(block?: string, _options?: Configuration): Promise<GovChainInfoResponse> {
        const result = this.api.governanceGetChainInfo(block, _options);
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
    public getHealthWithHttpInfo(_options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.getHealthWithHttpInfo(_options);
        return result.toPromise();
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
    public v1WsGetWithHttpInfo(_options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.v1WsGetWithHttpInfo(_options);
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
     */
    public getChainMessageMetricsWithHttpInfo(_options?: Configuration): Promise<HttpInfo<ChainMessageMetrics>> {
        const result = this.api.getChainMessageMetricsWithHttpInfo(_options);
        return result.toPromise();
    }

    /**
     * Get chain specific message metrics.
     */
    public getChainMessageMetrics(_options?: Configuration): Promise<ChainMessageMetrics> {
        const result = this.api.getChainMessageMetrics(_options);
        return result.toPromise();
    }

    /**
     * Get chain pipe event metrics.
     */
    public getChainPipeMetricsWithHttpInfo(_options?: Configuration): Promise<HttpInfo<ConsensusPipeMetrics>> {
        const result = this.api.getChainPipeMetricsWithHttpInfo(_options);
        return result.toPromise();
    }

    /**
     * Get chain pipe event metrics.
     */
    public getChainPipeMetrics(_options?: Configuration): Promise<ConsensusPipeMetrics> {
        const result = this.api.getChainPipeMetrics(_options);
        return result.toPromise();
    }

    /**
     * Get chain workflow metrics.
     */
    public getChainWorkflowMetricsWithHttpInfo(_options?: Configuration): Promise<HttpInfo<ConsensusWorkflowMetrics>> {
        const result = this.api.getChainWorkflowMetricsWithHttpInfo(_options);
        return result.toPromise();
    }

    /**
     * Get chain workflow metrics.
     */
    public getChainWorkflowMetrics(_options?: Configuration): Promise<ConsensusWorkflowMetrics> {
        const result = this.api.getChainWorkflowMetrics(_options);
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
    public distrustPeerWithHttpInfo(peer: string, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.distrustPeerWithHttpInfo(peer, _options);
        return result.toPromise();
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
    public generateDKSWithHttpInfo(dKSharesPostRequest: DKSharesPostRequest, _options?: Configuration): Promise<HttpInfo<DKSharesInfo>> {
        const result = this.api.generateDKSWithHttpInfo(dKSharesPostRequest, _options);
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
    public getAllPeersWithHttpInfo(_options?: Configuration): Promise<HttpInfo<Array<PeeringNodeStatusResponse>>> {
        const result = this.api.getAllPeersWithHttpInfo(_options);
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
    public getConfigurationWithHttpInfo(_options?: Configuration): Promise<HttpInfo<{ [key: string]: string; }>> {
        const result = this.api.getConfigurationWithHttpInfo(_options);
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
     * @param sharedAddress SharedAddress (Hex Address)
     */
    public getDKSInfoWithHttpInfo(sharedAddress: string, _options?: Configuration): Promise<HttpInfo<DKSharesInfo>> {
        const result = this.api.getDKSInfoWithHttpInfo(sharedAddress, _options);
        return result.toPromise();
    }

    /**
     * Get information about the shared address DKS configuration
     * @param sharedAddress SharedAddress (Hex Address)
     */
    public getDKSInfo(sharedAddress: string, _options?: Configuration): Promise<DKSharesInfo> {
        const result = this.api.getDKSInfo(sharedAddress, _options);
        return result.toPromise();
    }

    /**
     * Returns private information about this node.
     */
    public getInfoWithHttpInfo(_options?: Configuration): Promise<HttpInfo<InfoResponse>> {
        const result = this.api.getInfoWithHttpInfo(_options);
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
    public getPeeringIdentityWithHttpInfo(_options?: Configuration): Promise<HttpInfo<PeeringNodeIdentityResponse>> {
        const result = this.api.getPeeringIdentityWithHttpInfo(_options);
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
    public getTrustedPeersWithHttpInfo(_options?: Configuration): Promise<HttpInfo<Array<PeeringNodeIdentityResponse>>> {
        const result = this.api.getTrustedPeersWithHttpInfo(_options);
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
    public getVersionWithHttpInfo(_options?: Configuration): Promise<HttpInfo<VersionResponse>> {
        const result = this.api.getVersionWithHttpInfo(_options);
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
    public ownerCertificateWithHttpInfo(_options?: Configuration): Promise<HttpInfo<NodeOwnerCertificateResponse>> {
        const result = this.api.ownerCertificateWithHttpInfo(_options);
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
    public shutdownNodeWithHttpInfo(_options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.shutdownNodeWithHttpInfo(_options);
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
    public trustPeerWithHttpInfo(peeringTrustRequest: PeeringTrustRequest, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.trustPeerWithHttpInfo(peeringTrustRequest, _options);
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
    public offLedgerWithHttpInfo(offLedgerRequest: OffLedgerRequest, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.offLedgerWithHttpInfo(offLedgerRequest, _options);
        return result.toPromise();
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
    public addUserWithHttpInfo(addUserRequest: AddUserRequest, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.addUserWithHttpInfo(addUserRequest, _options);
        return result.toPromise();
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
    public changeUserPasswordWithHttpInfo(username: string, updateUserPasswordRequest: UpdateUserPasswordRequest, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.changeUserPasswordWithHttpInfo(username, updateUserPasswordRequest, _options);
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
    public changeUserPermissionsWithHttpInfo(username: string, updateUserPermissionsRequest: UpdateUserPermissionsRequest, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.changeUserPermissionsWithHttpInfo(username, updateUserPermissionsRequest, _options);
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
    public deleteUserWithHttpInfo(username: string, _options?: Configuration): Promise<HttpInfo<void>> {
        const result = this.api.deleteUserWithHttpInfo(username, _options);
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
    public getUserWithHttpInfo(username: string, _options?: Configuration): Promise<HttpInfo<User>> {
        const result = this.api.getUserWithHttpInfo(username, _options);
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
    public getUsersWithHttpInfo(_options?: Configuration): Promise<HttpInfo<Array<User>>> {
        const result = this.api.getUsersWithHttpInfo(_options);
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



