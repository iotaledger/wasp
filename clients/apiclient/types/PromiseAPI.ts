import { ResponseContext, RequestContext, HttpFile, HttpInfo } from '../http/http';
import { Configuration, PromiseConfigurationOptions, wrapOptions } from '../configuration'
import { PromiseMiddleware, Middleware, PromiseMiddlewareWrapper } from '../middleware';

import { AccountNonceResponse } from '../models/AccountNonceResponse';
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
    public authInfoWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<AuthInfoModel>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.authInfoWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get information about the current authentication mode
     */
    public authInfo(_options?: PromiseConfigurationOptions): Promise<AuthInfoModel> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.authInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Authenticate towards the node
     * @param loginRequest The login request
     */
    public authenticateWithHttpInfo(loginRequest: LoginRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<LoginResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.authenticateWithHttpInfo(loginRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Authenticate towards the node
     * @param loginRequest The login request
     */
    public authenticate(loginRequest: LoginRequest, _options?: PromiseConfigurationOptions): Promise<LoginResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.authenticate(loginRequest, observableOptions);
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
    public activateChainWithHttpInfo(chainID: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.activateChainWithHttpInfo(chainID, observableOptions);
        return result.toPromise();
    }

    /**
     * Activate a chain
     * @param chainID ChainID (Hex Address)
     */
    public activateChain(chainID: string, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.activateChain(chainID, observableOptions);
        return result.toPromise();
    }

    /**
     * Configure a trusted node to be an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public addAccessNodeWithHttpInfo(peer: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.addAccessNodeWithHttpInfo(peer, observableOptions);
        return result.toPromise();
    }

    /**
     * Configure a trusted node to be an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public addAccessNode(peer: string, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.addAccessNode(peer, observableOptions);
        return result.toPromise();
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param contractCallViewRequest Parameters
     */
    public callViewWithHttpInfo(contractCallViewRequest: ContractCallViewRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<Array<string>>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.callViewWithHttpInfo(contractCallViewRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param contractCallViewRequest Parameters
     */
    public callView(contractCallViewRequest: ContractCallViewRequest, _options?: PromiseConfigurationOptions): Promise<Array<string>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.callView(contractCallViewRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Deactivate a chain
     */
    public deactivateChainWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.deactivateChainWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Deactivate a chain
     */
    public deactivateChain(_options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.deactivateChain(observableOptions);
        return result.toPromise();
    }

    /**
     * dump accounts information into a humanly-readable format
     */
    public dumpAccountsWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.dumpAccountsWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * dump accounts information into a humanly-readable format
     */
    public dumpAccounts(_options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.dumpAccounts(observableOptions);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param request Request
     */
    public estimateGasOffledgerWithHttpInfo(request: EstimateGasRequestOffledger, _options?: PromiseConfigurationOptions): Promise<HttpInfo<ReceiptResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.estimateGasOffledgerWithHttpInfo(request, observableOptions);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param request Request
     */
    public estimateGasOffledger(request: EstimateGasRequestOffledger, _options?: PromiseConfigurationOptions): Promise<ReceiptResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.estimateGasOffledger(request, observableOptions);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param request Request
     */
    public estimateGasOnledgerWithHttpInfo(request: EstimateGasRequestOnledger, _options?: PromiseConfigurationOptions): Promise<HttpInfo<ReceiptResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.estimateGasOnledgerWithHttpInfo(request, observableOptions);
        return result.toPromise();
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param request Request
     */
    public estimateGasOnledger(request: EstimateGasRequestOnledger, _options?: PromiseConfigurationOptions): Promise<ReceiptResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.estimateGasOnledger(request, observableOptions);
        return result.toPromise();
    }

    /**
     * Get information about a specific chain
     * @param [block] Block index or trie root
     */
    public getChainInfoWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<ChainInfoResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getChainInfoWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get information about a specific chain
     * @param [block] Block index or trie root
     */
    public getChainInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<ChainInfoResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getChainInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get information about the deployed committee
     * @param [block] Block index or trie root
     */
    public getCommitteeInfoWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<CommitteeInfoResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getCommitteeInfoWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get information about the deployed committee
     * @param [block] Block index or trie root
     */
    public getCommitteeInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<CommitteeInfoResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getCommitteeInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all available chain contracts
     * @param [block] Block index or trie root
     */
    public getContractsWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<Array<ContractInfoResponse>>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getContractsWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all available chain contracts
     * @param [block] Block index or trie root
     */
    public getContracts(block?: string, _options?: PromiseConfigurationOptions): Promise<Array<ContractInfoResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getContracts(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the contents of the mempool.
     */
    public getMempoolContentsWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<Array<number>>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getMempoolContentsWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get the contents of the mempool.
     */
    public getMempoolContents(_options?: PromiseConfigurationOptions): Promise<Array<number>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getMempoolContents(observableOptions);
        return result.toPromise();
    }

    /**
     * Get a receipt from a request ID
     * @param requestID RequestID (Hex)
     */
    public getReceiptWithHttpInfo(requestID: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<ReceiptResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getReceiptWithHttpInfo(requestID, observableOptions);
        return result.toPromise();
    }

    /**
     * Get a receipt from a request ID
     * @param requestID RequestID (Hex)
     */
    public getReceipt(requestID: string, _options?: PromiseConfigurationOptions): Promise<ReceiptResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getReceipt(requestID, observableOptions);
        return result.toPromise();
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param stateKey State Key (Hex)
     */
    public getStateValueWithHttpInfo(stateKey: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<StateResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getStateValueWithHttpInfo(stateKey, observableOptions);
        return result.toPromise();
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param stateKey State Key (Hex)
     */
    public getStateValue(stateKey: string, _options?: PromiseConfigurationOptions): Promise<StateResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getStateValue(stateKey, observableOptions);
        return result.toPromise();
    }

    /**
     * Remove an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public removeAccessNodeWithHttpInfo(peer: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.removeAccessNodeWithHttpInfo(peer, observableOptions);
        return result.toPromise();
    }

    /**
     * Remove an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public removeAccessNode(peer: string, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.removeAccessNode(peer, observableOptions);
        return result.toPromise();
    }

    /**
     * Rotate a chain
     * @param [rotateRequest] RotateRequest
     */
    public rotateChainWithHttpInfo(rotateRequest?: RotateChainRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.rotateChainWithHttpInfo(rotateRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Rotate a chain
     * @param [rotateRequest] RotateRequest
     */
    public rotateChain(rotateRequest?: RotateChainRequest, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.rotateChain(rotateRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Sets the chain record.
     * @param chainID ChainID (Hex Address)
     * @param chainRecord Chain Record
     */
    public setChainRecordWithHttpInfo(chainID: string, chainRecord: ChainRecord, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.setChainRecordWithHttpInfo(chainID, chainRecord, observableOptions);
        return result.toPromise();
    }

    /**
     * Sets the chain record.
     * @param chainID ChainID (Hex Address)
     * @param chainRecord Chain Record
     */
    public setChainRecord(chainID: string, chainRecord: ChainRecord, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.setChainRecord(chainID, chainRecord, observableOptions);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC
     */
    public v1ChainEvmPostWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.v1ChainEvmPostWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC
     */
    public v1ChainEvmPost(_options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.v1ChainEvmPost(observableOptions);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     */
    public v1ChainEvmWsGetWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.v1ChainEvmWsGetWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     */
    public v1ChainEvmWsGet(_options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.v1ChainEvmWsGet(observableOptions);
        return result.toPromise();
    }

    /**
     * Wait until the given request has been processed by the node
     * @param requestID RequestID (Hex)
     * @param [timeoutSeconds] The timeout in seconds, maximum 60s
     * @param [waitForL1Confirmation] Wait for the block to be confirmed on L1
     */
    public waitForRequestWithHttpInfo(requestID: string, timeoutSeconds?: number, waitForL1Confirmation?: boolean, _options?: PromiseConfigurationOptions): Promise<HttpInfo<ReceiptResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.waitForRequestWithHttpInfo(requestID, timeoutSeconds, waitForL1Confirmation, observableOptions);
        return result.toPromise();
    }

    /**
     * Wait until the given request has been processed by the node
     * @param requestID RequestID (Hex)
     * @param [timeoutSeconds] The timeout in seconds, maximum 60s
     * @param [waitForL1Confirmation] Wait for the block to be confirmed on L1
     */
    public waitForRequest(requestID: string, timeoutSeconds?: number, waitForL1Confirmation?: boolean, _options?: PromiseConfigurationOptions): Promise<ReceiptResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.waitForRequest(requestID, timeoutSeconds, waitForL1Confirmation, observableOptions);
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
    public accountsGetAccountBalanceWithHttpInfo(agentID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<AssetsResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.accountsGetAccountBalanceWithHttpInfo(agentID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all assets belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountBalance(agentID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<AssetsResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.accountsGetAccountBalance(agentID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the current nonce of an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountNonceWithHttpInfo(agentID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<AccountNonceResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.accountsGetAccountNonceWithHttpInfo(agentID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the current nonce of an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountNonce(agentID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<AccountNonceResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.accountsGetAccountNonce(agentID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all stored assets
     * @param [block] Block index or trie root
     */
    public accountsGetTotalAssetsWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<AssetsResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.accountsGetTotalAssetsWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all stored assets
     * @param [block] Block index or trie root
     */
    public accountsGetTotalAssets(block?: string, _options?: PromiseConfigurationOptions): Promise<AssetsResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.accountsGetTotalAssets(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the block info of a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetBlockInfoWithHttpInfo(blockIndex: number, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<BlockInfoResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetBlockInfoWithHttpInfo(blockIndex, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the block info of a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetBlockInfo(blockIndex: number, block?: string, _options?: PromiseConfigurationOptions): Promise<BlockInfoResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetBlockInfo(blockIndex, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the control addresses
     * @param [block] Block index or trie root
     */
    public blocklogGetControlAddressesWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<ControlAddressesResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetControlAddressesWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the control addresses
     * @param [block] Block index or trie root
     */
    public blocklogGetControlAddresses(block?: string, _options?: PromiseConfigurationOptions): Promise<ControlAddressesResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetControlAddresses(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get events of a block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfBlockWithHttpInfo(blockIndex: number, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<EventsResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetEventsOfBlockWithHttpInfo(blockIndex, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get events of a block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfBlock(blockIndex: number, block?: string, _options?: PromiseConfigurationOptions): Promise<EventsResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetEventsOfBlock(blockIndex, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get events of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfLatestBlockWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<EventsResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetEventsOfLatestBlockWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get events of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfLatestBlock(block?: string, _options?: PromiseConfigurationOptions): Promise<EventsResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetEventsOfLatestBlock(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get events of a request
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfRequestWithHttpInfo(requestID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<EventsResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetEventsOfRequestWithHttpInfo(requestID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get events of a request
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfRequest(requestID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<EventsResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetEventsOfRequest(requestID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the block info of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetLatestBlockInfoWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<BlockInfoResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetLatestBlockInfoWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the block info of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetLatestBlockInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<BlockInfoResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetLatestBlockInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the request ids for a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForBlockWithHttpInfo(blockIndex: number, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<RequestIDsResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestIDsForBlockWithHttpInfo(blockIndex, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the request ids for a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForBlock(blockIndex: number, block?: string, _options?: PromiseConfigurationOptions): Promise<RequestIDsResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestIDsForBlock(blockIndex, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the request ids for the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForLatestBlockWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<RequestIDsResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestIDsForLatestBlockWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the request ids for the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForLatestBlock(block?: string, _options?: PromiseConfigurationOptions): Promise<RequestIDsResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestIDsForLatestBlock(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the request processing status
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIsProcessedWithHttpInfo(requestID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<RequestProcessedResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestIsProcessedWithHttpInfo(requestID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the request processing status
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIsProcessed(requestID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<RequestProcessedResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestIsProcessed(requestID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the receipt of a certain request id
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptWithHttpInfo(requestID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<ReceiptResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestReceiptWithHttpInfo(requestID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the receipt of a certain request id
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceipt(requestID: string, block?: string, _options?: PromiseConfigurationOptions): Promise<ReceiptResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestReceipt(requestID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all receipts of a certain block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfBlockWithHttpInfo(blockIndex: number, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<Array<ReceiptResponse>>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestReceiptsOfBlockWithHttpInfo(blockIndex, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all receipts of a certain block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfBlock(blockIndex: number, block?: string, _options?: PromiseConfigurationOptions): Promise<Array<ReceiptResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestReceiptsOfBlock(blockIndex, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all receipts of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<Array<ReceiptResponse>>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get all receipts of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfLatestBlock(block?: string, _options?: PromiseConfigurationOptions): Promise<Array<ReceiptResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.blocklogGetRequestReceiptsOfLatestBlock(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the error message format of a specific error id
     * @param chainID ChainID (Hex Address)
     * @param contractHname Contract (Hname as Hex)
     * @param errorID Error Id (uint16)
     * @param [block] Block index or trie root
     */
    public errorsGetErrorMessageFormatWithHttpInfo(chainID: string, contractHname: string, errorID: number, block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<ErrorMessageFormatResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.errorsGetErrorMessageFormatWithHttpInfo(chainID, contractHname, errorID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Get the error message format of a specific error id
     * @param chainID ChainID (Hex Address)
     * @param contractHname Contract (Hname as Hex)
     * @param errorID Error Id (uint16)
     * @param [block] Block index or trie root
     */
    public errorsGetErrorMessageFormat(chainID: string, contractHname: string, errorID: number, block?: string, _options?: PromiseConfigurationOptions): Promise<ErrorMessageFormatResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.errorsGetErrorMessageFormat(chainID, contractHname, errorID, block, observableOptions);
        return result.toPromise();
    }

    /**
     * Returns the chain admin
     * Get the chain admin
     * @param [block] Block index or trie root
     */
    public governanceGetChainAdminWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<GovChainAdminResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.governanceGetChainAdminWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * Returns the chain admin
     * Get the chain admin
     * @param [block] Block index or trie root
     */
    public governanceGetChainAdmin(block?: string, _options?: PromiseConfigurationOptions): Promise<GovChainAdminResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.governanceGetChainAdmin(block, observableOptions);
        return result.toPromise();
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chains/:chainID\' to get information about a chain.
     * Get the chain info
     * @param [block] Block index or trie root
     */
    public governanceGetChainInfoWithHttpInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<GovChainInfoResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.governanceGetChainInfoWithHttpInfo(block, observableOptions);
        return result.toPromise();
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chains/:chainID\' to get information about a chain.
     * Get the chain info
     * @param [block] Block index or trie root
     */
    public governanceGetChainInfo(block?: string, _options?: PromiseConfigurationOptions): Promise<GovChainInfoResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.governanceGetChainInfo(block, observableOptions);
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
    public getHealthWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getHealthWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Returns 200 if the node is healthy.
     */
    public getHealth(_options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getHealth(observableOptions);
        return result.toPromise();
    }

    /**
     * The websocket connection service
     */
    public v1WsGetWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.v1WsGetWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * The websocket connection service
     */
    public v1WsGet(_options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.v1WsGet(observableOptions);
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
    public getChainMessageMetricsWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<ChainMessageMetrics>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getChainMessageMetricsWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get chain specific message metrics.
     */
    public getChainMessageMetrics(_options?: PromiseConfigurationOptions): Promise<ChainMessageMetrics> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getChainMessageMetrics(observableOptions);
        return result.toPromise();
    }

    /**
     * Get chain pipe event metrics.
     */
    public getChainPipeMetricsWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<ConsensusPipeMetrics>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getChainPipeMetricsWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get chain pipe event metrics.
     */
    public getChainPipeMetrics(_options?: PromiseConfigurationOptions): Promise<ConsensusPipeMetrics> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getChainPipeMetrics(observableOptions);
        return result.toPromise();
    }

    /**
     * Get chain workflow metrics.
     */
    public getChainWorkflowMetricsWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<ConsensusWorkflowMetrics>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getChainWorkflowMetricsWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get chain workflow metrics.
     */
    public getChainWorkflowMetrics(_options?: PromiseConfigurationOptions): Promise<ConsensusWorkflowMetrics> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getChainWorkflowMetrics(observableOptions);
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
    public distrustPeerWithHttpInfo(peer: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.distrustPeerWithHttpInfo(peer, observableOptions);
        return result.toPromise();
    }

    /**
     * Distrust a peering node
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public distrustPeer(peer: string, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.distrustPeer(peer, observableOptions);
        return result.toPromise();
    }

    /**
     * Generate a new distributed key
     * @param dKSharesPostRequest Request parameters
     */
    public generateDKSWithHttpInfo(dKSharesPostRequest: DKSharesPostRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<DKSharesInfo>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.generateDKSWithHttpInfo(dKSharesPostRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Generate a new distributed key
     * @param dKSharesPostRequest Request parameters
     */
    public generateDKS(dKSharesPostRequest: DKSharesPostRequest, _options?: PromiseConfigurationOptions): Promise<DKSharesInfo> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.generateDKS(dKSharesPostRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Get basic information about all configured peers
     */
    public getAllPeersWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<Array<PeeringNodeStatusResponse>>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getAllPeersWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get basic information about all configured peers
     */
    public getAllPeers(_options?: PromiseConfigurationOptions): Promise<Array<PeeringNodeStatusResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getAllPeers(observableOptions);
        return result.toPromise();
    }

    /**
     * Return the Wasp configuration
     */
    public getConfigurationWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<{ [key: string]: string; }>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getConfigurationWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Return the Wasp configuration
     */
    public getConfiguration(_options?: PromiseConfigurationOptions): Promise<{ [key: string]: string; }> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getConfiguration(observableOptions);
        return result.toPromise();
    }

    /**
     * Get information about the shared address DKS configuration
     * @param sharedAddress SharedAddress (Hex Address)
     */
    public getDKSInfoWithHttpInfo(sharedAddress: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<DKSharesInfo>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getDKSInfoWithHttpInfo(sharedAddress, observableOptions);
        return result.toPromise();
    }

    /**
     * Get information about the shared address DKS configuration
     * @param sharedAddress SharedAddress (Hex Address)
     */
    public getDKSInfo(sharedAddress: string, _options?: PromiseConfigurationOptions): Promise<DKSharesInfo> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getDKSInfo(sharedAddress, observableOptions);
        return result.toPromise();
    }

    /**
     * Returns private information about this node.
     */
    public getInfoWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<InfoResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getInfoWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Returns private information about this node.
     */
    public getInfo(_options?: PromiseConfigurationOptions): Promise<InfoResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get basic peer info of the current node
     */
    public getPeeringIdentityWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<PeeringNodeIdentityResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getPeeringIdentityWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get basic peer info of the current node
     */
    public getPeeringIdentity(_options?: PromiseConfigurationOptions): Promise<PeeringNodeIdentityResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getPeeringIdentity(observableOptions);
        return result.toPromise();
    }

    /**
     * Get trusted peers
     */
    public getTrustedPeersWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<Array<PeeringNodeIdentityResponse>>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getTrustedPeersWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get trusted peers
     */
    public getTrustedPeers(_options?: PromiseConfigurationOptions): Promise<Array<PeeringNodeIdentityResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getTrustedPeers(observableOptions);
        return result.toPromise();
    }

    /**
     * Returns the node version.
     */
    public getVersionWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<VersionResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getVersionWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Returns the node version.
     */
    public getVersion(_options?: PromiseConfigurationOptions): Promise<VersionResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getVersion(observableOptions);
        return result.toPromise();
    }

    /**
     * Gets the node owner
     */
    public ownerCertificateWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<NodeOwnerCertificateResponse>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.ownerCertificateWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Gets the node owner
     */
    public ownerCertificate(_options?: PromiseConfigurationOptions): Promise<NodeOwnerCertificateResponse> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.ownerCertificate(observableOptions);
        return result.toPromise();
    }

    /**
     * Shut down the node
     */
    public shutdownNodeWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.shutdownNodeWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Shut down the node
     */
    public shutdownNode(_options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.shutdownNode(observableOptions);
        return result.toPromise();
    }

    /**
     * Trust a peering node
     * @param peeringTrustRequest Info of the peer to trust
     */
    public trustPeerWithHttpInfo(peeringTrustRequest: PeeringTrustRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.trustPeerWithHttpInfo(peeringTrustRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Trust a peering node
     * @param peeringTrustRequest Info of the peer to trust
     */
    public trustPeer(peeringTrustRequest: PeeringTrustRequest, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.trustPeer(peeringTrustRequest, observableOptions);
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
    public offLedgerWithHttpInfo(offLedgerRequest: OffLedgerRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.offLedgerWithHttpInfo(offLedgerRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Post an off-ledger request
     * @param offLedgerRequest Offledger request as JSON. Request encoded in Hex
     */
    public offLedger(offLedgerRequest: OffLedgerRequest, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.offLedger(offLedgerRequest, observableOptions);
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
    public addUserWithHttpInfo(addUserRequest: AddUserRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.addUserWithHttpInfo(addUserRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Add a user
     * @param addUserRequest The user data
     */
    public addUser(addUserRequest: AddUserRequest, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.addUser(addUserRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Change user password
     * @param username The username
     * @param updateUserPasswordRequest The users new password
     */
    public changeUserPasswordWithHttpInfo(username: string, updateUserPasswordRequest: UpdateUserPasswordRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.changeUserPasswordWithHttpInfo(username, updateUserPasswordRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Change user password
     * @param username The username
     * @param updateUserPasswordRequest The users new password
     */
    public changeUserPassword(username: string, updateUserPasswordRequest: UpdateUserPasswordRequest, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.changeUserPassword(username, updateUserPasswordRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Change user permissions
     * @param username The username
     * @param updateUserPermissionsRequest The users new permissions
     */
    public changeUserPermissionsWithHttpInfo(username: string, updateUserPermissionsRequest: UpdateUserPermissionsRequest, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.changeUserPermissionsWithHttpInfo(username, updateUserPermissionsRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Change user permissions
     * @param username The username
     * @param updateUserPermissionsRequest The users new permissions
     */
    public changeUserPermissions(username: string, updateUserPermissionsRequest: UpdateUserPermissionsRequest, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.changeUserPermissions(username, updateUserPermissionsRequest, observableOptions);
        return result.toPromise();
    }

    /**
     * Deletes a user
     * @param username The username
     */
    public deleteUserWithHttpInfo(username: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<void>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.deleteUserWithHttpInfo(username, observableOptions);
        return result.toPromise();
    }

    /**
     * Deletes a user
     * @param username The username
     */
    public deleteUser(username: string, _options?: PromiseConfigurationOptions): Promise<void> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.deleteUser(username, observableOptions);
        return result.toPromise();
    }

    /**
     * Get a user
     * @param username The username
     */
    public getUserWithHttpInfo(username: string, _options?: PromiseConfigurationOptions): Promise<HttpInfo<User>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getUserWithHttpInfo(username, observableOptions);
        return result.toPromise();
    }

    /**
     * Get a user
     * @param username The username
     */
    public getUser(username: string, _options?: PromiseConfigurationOptions): Promise<User> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getUser(username, observableOptions);
        return result.toPromise();
    }

    /**
     * Get a list of all users
     */
    public getUsersWithHttpInfo(_options?: PromiseConfigurationOptions): Promise<HttpInfo<Array<User>>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getUsersWithHttpInfo(observableOptions);
        return result.toPromise();
    }

    /**
     * Get a list of all users
     */
    public getUsers(_options?: PromiseConfigurationOptions): Promise<Array<User>> {
        const observableOptions = wrapOptions(_options);
        const result = this.api.getUsers(observableOptions);
        return result.toPromise();
    }


}



