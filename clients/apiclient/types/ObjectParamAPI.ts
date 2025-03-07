import { ResponseContext, RequestContext, HttpFile, HttpInfo } from '../http/http';
import { Configuration} from '../configuration'

import { AccountFoundriesResponse } from '../models/AccountFoundriesResponse';
import { AccountNFTsResponse } from '../models/AccountNFTsResponse';
import { AccountNonceResponse } from '../models/AccountNonceResponse';
import { AddUserRequest } from '../models/AddUserRequest';
import { AnchorMetricItem } from '../models/AnchorMetricItem';
import { AssetsJSON } from '../models/AssetsJSON';
import { AssetsResponse } from '../models/AssetsResponse';
import { AuthInfoModel } from '../models/AuthInfoModel';
import { BaseToken } from '../models/BaseToken';
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
import { FoundryOutputResponse } from '../models/FoundryOutputResponse';
import { GovAllowedStateControllerAddressesResponse } from '../models/GovAllowedStateControllerAddressesResponse';
import { GovChainInfoResponse } from '../models/GovChainInfoResponse';
import { GovChainOwnerResponse } from '../models/GovChainOwnerResponse';
import { GovPublicChainMetadata } from '../models/GovPublicChainMetadata';
import { InfoResponse } from '../models/InfoResponse';
import { L1Params } from '../models/L1Params';
import { Limits } from '../models/Limits';
import { LoginRequest } from '../models/LoginRequest';
import { LoginResponse } from '../models/LoginResponse';
import { NativeTokenIDRegistryResponse } from '../models/NativeTokenIDRegistryResponse';
import { NodeMessageMetrics } from '../models/NodeMessageMetrics';
import { NodeOwnerCertificateResponse } from '../models/NodeOwnerCertificateResponse';
import { OffLedgerRequest } from '../models/OffLedgerRequest';
import { OnLedgerRequest } from '../models/OnLedgerRequest';
import { OnLedgerRequestMetricItem } from '../models/OnLedgerRequestMetricItem';
import { PeeringNodeIdentityResponse } from '../models/PeeringNodeIdentityResponse';
import { PeeringNodeStatusResponse } from '../models/PeeringNodeStatusResponse';
import { PeeringTrustRequest } from '../models/PeeringTrustRequest';
import { ProtocolParameters } from '../models/ProtocolParameters';
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
import { Type } from '../models/Type';
import { UnresolvedVMErrorJSON } from '../models/UnresolvedVMErrorJSON';
import { UpdateUserPasswordRequest } from '../models/UpdateUserPasswordRequest';
import { UpdateUserPermissionsRequest } from '../models/UpdateUserPermissionsRequest';
import { User } from '../models/User';
import { ValidationError } from '../models/ValidationError';
import { VersionResponse } from '../models/VersionResponse';

import { ObservableAuthApi } from "./ObservableAPI";
import { AuthApiRequestFactory, AuthApiResponseProcessor} from "../apis/AuthApi";

export interface AuthApiAuthInfoRequest {
}

export interface AuthApiAuthenticateRequest {
    /**
     * The login request
     * @type LoginRequest
     * @memberof AuthApiauthenticate
     */
    loginRequest: LoginRequest
}

export class ObjectAuthApi {
    private api: ObservableAuthApi

    public constructor(configuration: Configuration, requestFactory?: AuthApiRequestFactory, responseProcessor?: AuthApiResponseProcessor) {
        this.api = new ObservableAuthApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Get information about the current authentication mode
     * @param param the request object
     */
    public authInfoWithHttpInfo(param: AuthApiAuthInfoRequest = {}, options?: Configuration): Promise<HttpInfo<AuthInfoModel>> {
        return this.api.authInfoWithHttpInfo( options).toPromise();
    }

    /**
     * Get information about the current authentication mode
     * @param param the request object
     */
    public authInfo(param: AuthApiAuthInfoRequest = {}, options?: Configuration): Promise<AuthInfoModel> {
        return this.api.authInfo( options).toPromise();
    }

    /**
     * Authenticate towards the node
     * @param param the request object
     */
    public authenticateWithHttpInfo(param: AuthApiAuthenticateRequest, options?: Configuration): Promise<HttpInfo<LoginResponse>> {
        return this.api.authenticateWithHttpInfo(param.loginRequest,  options).toPromise();
    }

    /**
     * Authenticate towards the node
     * @param param the request object
     */
    public authenticate(param: AuthApiAuthenticateRequest, options?: Configuration): Promise<LoginResponse> {
        return this.api.authenticate(param.loginRequest,  options).toPromise();
    }

}

import { ObservableChainsApi } from "./ObservableAPI";
import { ChainsApiRequestFactory, ChainsApiResponseProcessor} from "../apis/ChainsApi";

export interface ChainsApiActivateChainRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiactivateChain
     */
    chainID: string
}

export interface ChainsApiAddAccessNodeRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiaddAccessNode
     */
    chainID: string
    /**
     * Name or PubKey (hex) of the trusted peer
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiaddAccessNode
     */
    peer: string
}

export interface ChainsApiCallViewRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApicallView
     */
    chainID: string
    /**
     * Parameters
     * @type ContractCallViewRequest
     * @memberof ChainsApicallView
     */
    contractCallViewRequest: ContractCallViewRequest
}

export interface ChainsApiDeactivateChainRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApideactivateChain
     */
    chainID: string
}

export interface ChainsApiDumpAccountsRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApidumpAccounts
     */
    chainID: string
}

export interface ChainsApiEstimateGasOffledgerRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiestimateGasOffledger
     */
    chainID: string
    /**
     * Request
     * @type EstimateGasRequestOffledger
     * @memberof ChainsApiestimateGasOffledger
     */
    request: EstimateGasRequestOffledger
}

export interface ChainsApiEstimateGasOnledgerRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiestimateGasOnledger
     */
    chainID: string
    /**
     * Request
     * @type EstimateGasRequestOnledger
     * @memberof ChainsApiestimateGasOnledger
     */
    request: EstimateGasRequestOnledger
}

export interface ChainsApiGetChainInfoRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetChainInfo
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetChainInfo
     */
    block?: string
}

export interface ChainsApiGetChainsRequest {
}

export interface ChainsApiGetCommitteeInfoRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetCommitteeInfo
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetCommitteeInfo
     */
    block?: string
}

export interface ChainsApiGetContractsRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetContracts
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetContracts
     */
    block?: string
}

export interface ChainsApiGetMempoolContentsRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetMempoolContents
     */
    chainID: string
}

export interface ChainsApiGetReceiptRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetReceipt
     */
    chainID: string
    /**
     * RequestID (Hex)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetReceipt
     */
    requestID: string
}

export interface ChainsApiGetStateValueRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetStateValue
     */
    chainID: string
    /**
     * State Key (Hex)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApigetStateValue
     */
    stateKey: string
}

export interface ChainsApiRemoveAccessNodeRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiremoveAccessNode
     */
    chainID: string
    /**
     * Name or PubKey (hex) of the trusted peer
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiremoveAccessNode
     */
    peer: string
}

export interface ChainsApiRotateChainRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApirotateChain
     */
    chainID: string
    /**
     * RotateRequest
     * @type RotateChainRequest
     * @memberof ChainsApirotateChain
     */
    rotateRequest?: RotateChainRequest
}

export interface ChainsApiSetChainRecordRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApisetChainRecord
     */
    chainID: string
    /**
     * Chain Record
     * @type ChainRecord
     * @memberof ChainsApisetChainRecord
     */
    chainRecord: ChainRecord
}

export interface ChainsApiV1ChainsChainIDEvmPostRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiv1ChainsChainIDEvmPost
     */
    chainID: string
}

export interface ChainsApiV1ChainsChainIDEvmWsGetRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiv1ChainsChainIDEvmWsGet
     */
    chainID: string
}

export interface ChainsApiWaitForRequestRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiwaitForRequest
     */
    chainID: string
    /**
     * RequestID (Hex)
     * Defaults to: undefined
     * @type string
     * @memberof ChainsApiwaitForRequest
     */
    requestID: string
    /**
     * The timeout in seconds, maximum 60s
     * Defaults to: undefined
     * @type number
     * @memberof ChainsApiwaitForRequest
     */
    timeoutSeconds?: number
    /**
     * Wait for the block to be confirmed on L1
     * Defaults to: undefined
     * @type boolean
     * @memberof ChainsApiwaitForRequest
     */
    waitForL1Confirmation?: boolean
}

export class ObjectChainsApi {
    private api: ObservableChainsApi

    public constructor(configuration: Configuration, requestFactory?: ChainsApiRequestFactory, responseProcessor?: ChainsApiResponseProcessor) {
        this.api = new ObservableChainsApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Activate a chain
     * @param param the request object
     */
    public activateChainWithHttpInfo(param: ChainsApiActivateChainRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.activateChainWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * Activate a chain
     * @param param the request object
     */
    public activateChain(param: ChainsApiActivateChainRequest, options?: Configuration): Promise<void> {
        return this.api.activateChain(param.chainID,  options).toPromise();
    }

    /**
     * Configure a trusted node to be an access node.
     * @param param the request object
     */
    public addAccessNodeWithHttpInfo(param: ChainsApiAddAccessNodeRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.addAccessNodeWithHttpInfo(param.chainID, param.peer,  options).toPromise();
    }

    /**
     * Configure a trusted node to be an access node.
     * @param param the request object
     */
    public addAccessNode(param: ChainsApiAddAccessNodeRequest, options?: Configuration): Promise<void> {
        return this.api.addAccessNode(param.chainID, param.peer,  options).toPromise();
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param param the request object
     */
    public callViewWithHttpInfo(param: ChainsApiCallViewRequest, options?: Configuration): Promise<HttpInfo<Array<string>>> {
        return this.api.callViewWithHttpInfo(param.chainID, param.contractCallViewRequest,  options).toPromise();
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param param the request object
     */
    public callView(param: ChainsApiCallViewRequest, options?: Configuration): Promise<Array<string>> {
        return this.api.callView(param.chainID, param.contractCallViewRequest,  options).toPromise();
    }

    /**
     * Deactivate a chain
     * @param param the request object
     */
    public deactivateChainWithHttpInfo(param: ChainsApiDeactivateChainRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.deactivateChainWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * Deactivate a chain
     * @param param the request object
     */
    public deactivateChain(param: ChainsApiDeactivateChainRequest, options?: Configuration): Promise<void> {
        return this.api.deactivateChain(param.chainID,  options).toPromise();
    }

    /**
     * dump accounts information into a humanly-readable format
     * @param param the request object
     */
    public dumpAccountsWithHttpInfo(param: ChainsApiDumpAccountsRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.dumpAccountsWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * dump accounts information into a humanly-readable format
     * @param param the request object
     */
    public dumpAccounts(param: ChainsApiDumpAccountsRequest, options?: Configuration): Promise<void> {
        return this.api.dumpAccounts(param.chainID,  options).toPromise();
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param param the request object
     */
    public estimateGasOffledgerWithHttpInfo(param: ChainsApiEstimateGasOffledgerRequest, options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        return this.api.estimateGasOffledgerWithHttpInfo(param.chainID, param.request,  options).toPromise();
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param param the request object
     */
    public estimateGasOffledger(param: ChainsApiEstimateGasOffledgerRequest, options?: Configuration): Promise<ReceiptResponse> {
        return this.api.estimateGasOffledger(param.chainID, param.request,  options).toPromise();
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param param the request object
     */
    public estimateGasOnledgerWithHttpInfo(param: ChainsApiEstimateGasOnledgerRequest, options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        return this.api.estimateGasOnledgerWithHttpInfo(param.chainID, param.request,  options).toPromise();
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param param the request object
     */
    public estimateGasOnledger(param: ChainsApiEstimateGasOnledgerRequest, options?: Configuration): Promise<ReceiptResponse> {
        return this.api.estimateGasOnledger(param.chainID, param.request,  options).toPromise();
    }

    /**
     * Get information about a specific chain
     * @param param the request object
     */
    public getChainInfoWithHttpInfo(param: ChainsApiGetChainInfoRequest, options?: Configuration): Promise<HttpInfo<ChainInfoResponse>> {
        return this.api.getChainInfoWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get information about a specific chain
     * @param param the request object
     */
    public getChainInfo(param: ChainsApiGetChainInfoRequest, options?: Configuration): Promise<ChainInfoResponse> {
        return this.api.getChainInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get a list of all chains
     * @param param the request object
     */
    public getChainsWithHttpInfo(param: ChainsApiGetChainsRequest = {}, options?: Configuration): Promise<HttpInfo<Array<ChainInfoResponse>>> {
        return this.api.getChainsWithHttpInfo( options).toPromise();
    }

    /**
     * Get a list of all chains
     * @param param the request object
     */
    public getChains(param: ChainsApiGetChainsRequest = {}, options?: Configuration): Promise<Array<ChainInfoResponse>> {
        return this.api.getChains( options).toPromise();
    }

    /**
     * Get information about the deployed committee
     * @param param the request object
     */
    public getCommitteeInfoWithHttpInfo(param: ChainsApiGetCommitteeInfoRequest, options?: Configuration): Promise<HttpInfo<CommitteeInfoResponse>> {
        return this.api.getCommitteeInfoWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get information about the deployed committee
     * @param param the request object
     */
    public getCommitteeInfo(param: ChainsApiGetCommitteeInfoRequest, options?: Configuration): Promise<CommitteeInfoResponse> {
        return this.api.getCommitteeInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get all available chain contracts
     * @param param the request object
     */
    public getContractsWithHttpInfo(param: ChainsApiGetContractsRequest, options?: Configuration): Promise<HttpInfo<Array<ContractInfoResponse>>> {
        return this.api.getContractsWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get all available chain contracts
     * @param param the request object
     */
    public getContracts(param: ChainsApiGetContractsRequest, options?: Configuration): Promise<Array<ContractInfoResponse>> {
        return this.api.getContracts(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get the contents of the mempool.
     * @param param the request object
     */
    public getMempoolContentsWithHttpInfo(param: ChainsApiGetMempoolContentsRequest, options?: Configuration): Promise<HttpInfo<Array<number>>> {
        return this.api.getMempoolContentsWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * Get the contents of the mempool.
     * @param param the request object
     */
    public getMempoolContents(param: ChainsApiGetMempoolContentsRequest, options?: Configuration): Promise<Array<number>> {
        return this.api.getMempoolContents(param.chainID,  options).toPromise();
    }

    /**
     * Get a receipt from a request ID
     * @param param the request object
     */
    public getReceiptWithHttpInfo(param: ChainsApiGetReceiptRequest, options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        return this.api.getReceiptWithHttpInfo(param.chainID, param.requestID,  options).toPromise();
    }

    /**
     * Get a receipt from a request ID
     * @param param the request object
     */
    public getReceipt(param: ChainsApiGetReceiptRequest, options?: Configuration): Promise<ReceiptResponse> {
        return this.api.getReceipt(param.chainID, param.requestID,  options).toPromise();
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param param the request object
     */
    public getStateValueWithHttpInfo(param: ChainsApiGetStateValueRequest, options?: Configuration): Promise<HttpInfo<StateResponse>> {
        return this.api.getStateValueWithHttpInfo(param.chainID, param.stateKey,  options).toPromise();
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param param the request object
     */
    public getStateValue(param: ChainsApiGetStateValueRequest, options?: Configuration): Promise<StateResponse> {
        return this.api.getStateValue(param.chainID, param.stateKey,  options).toPromise();
    }

    /**
     * Remove an access node.
     * @param param the request object
     */
    public removeAccessNodeWithHttpInfo(param: ChainsApiRemoveAccessNodeRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.removeAccessNodeWithHttpInfo(param.chainID, param.peer,  options).toPromise();
    }

    /**
     * Remove an access node.
     * @param param the request object
     */
    public removeAccessNode(param: ChainsApiRemoveAccessNodeRequest, options?: Configuration): Promise<void> {
        return this.api.removeAccessNode(param.chainID, param.peer,  options).toPromise();
    }

    /**
     * Rotate a chain
     * @param param the request object
     */
    public rotateChainWithHttpInfo(param: ChainsApiRotateChainRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.rotateChainWithHttpInfo(param.chainID, param.rotateRequest,  options).toPromise();
    }

    /**
     * Rotate a chain
     * @param param the request object
     */
    public rotateChain(param: ChainsApiRotateChainRequest, options?: Configuration): Promise<void> {
        return this.api.rotateChain(param.chainID, param.rotateRequest,  options).toPromise();
    }

    /**
     * Sets the chain record.
     * @param param the request object
     */
    public setChainRecordWithHttpInfo(param: ChainsApiSetChainRecordRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.setChainRecordWithHttpInfo(param.chainID, param.chainRecord,  options).toPromise();
    }

    /**
     * Sets the chain record.
     * @param param the request object
     */
    public setChainRecord(param: ChainsApiSetChainRecordRequest, options?: Configuration): Promise<void> {
        return this.api.setChainRecord(param.chainID, param.chainRecord,  options).toPromise();
    }

    /**
     * Ethereum JSON-RPC
     * @param param the request object
     */
    public v1ChainsChainIDEvmPostWithHttpInfo(param: ChainsApiV1ChainsChainIDEvmPostRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.v1ChainsChainIDEvmPostWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * Ethereum JSON-RPC
     * @param param the request object
     */
    public v1ChainsChainIDEvmPost(param: ChainsApiV1ChainsChainIDEvmPostRequest, options?: Configuration): Promise<void> {
        return this.api.v1ChainsChainIDEvmPost(param.chainID,  options).toPromise();
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     * @param param the request object
     */
    public v1ChainsChainIDEvmWsGetWithHttpInfo(param: ChainsApiV1ChainsChainIDEvmWsGetRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.v1ChainsChainIDEvmWsGetWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     * @param param the request object
     */
    public v1ChainsChainIDEvmWsGet(param: ChainsApiV1ChainsChainIDEvmWsGetRequest, options?: Configuration): Promise<void> {
        return this.api.v1ChainsChainIDEvmWsGet(param.chainID,  options).toPromise();
    }

    /**
     * Wait until the given request has been processed by the node
     * @param param the request object
     */
    public waitForRequestWithHttpInfo(param: ChainsApiWaitForRequestRequest, options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        return this.api.waitForRequestWithHttpInfo(param.chainID, param.requestID, param.timeoutSeconds, param.waitForL1Confirmation,  options).toPromise();
    }

    /**
     * Wait until the given request has been processed by the node
     * @param param the request object
     */
    public waitForRequest(param: ChainsApiWaitForRequestRequest, options?: Configuration): Promise<ReceiptResponse> {
        return this.api.waitForRequest(param.chainID, param.requestID, param.timeoutSeconds, param.waitForL1Confirmation,  options).toPromise();
    }

}

import { ObservableCorecontractsApi } from "./ObservableAPI";
import { CorecontractsApiRequestFactory, CorecontractsApiResponseProcessor} from "../apis/CorecontractsApi";

export interface CorecontractsApiAccountsGetAccountBalanceRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountBalance
     */
    chainID: string
    /**
     * AgentID (Hex Address for L1 accounts | Hex for EVM)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountBalance
     */
    agentID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountBalance
     */
    block?: string
}

export interface CorecontractsApiAccountsGetAccountFoundriesRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountFoundries
     */
    chainID: string
    /**
     * AgentID (Hex Address for L1 accounts, Hex for EVM)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountFoundries
     */
    agentID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountFoundries
     */
    block?: string
}

export interface CorecontractsApiAccountsGetAccountNFTIDsRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountNFTIDs
     */
    chainID: string
    /**
     * AgentID (Hex Address for L1 accounts | Hex for EVM)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountNFTIDs
     */
    agentID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountNFTIDs
     */
    block?: string
}

export interface CorecontractsApiAccountsGetAccountNonceRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountNonce
     */
    chainID: string
    /**
     * AgentID (Hex Address for L1 accounts | Hex for EVM)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountNonce
     */
    agentID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetAccountNonce
     */
    block?: string
}

export interface CorecontractsApiAccountsGetFoundryOutputRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetFoundryOutput
     */
    chainID: string
    /**
     * Serial Number (uint32)
     * Minimum: 1
     * Defaults to: undefined
     * @type number
     * @memberof CorecontractsApiaccountsGetFoundryOutput
     */
    serialNumber: number
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetFoundryOutput
     */
    block?: string
}

export interface CorecontractsApiAccountsGetNFTDataRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetNFTData
     */
    chainID: string
    /**
     * NFT ID (Hex)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetNFTData
     */
    nftID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetNFTData
     */
    block?: string
}

export interface CorecontractsApiAccountsGetNativeTokenIDRegistryRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetNativeTokenIDRegistry
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetNativeTokenIDRegistry
     */
    block?: string
}

export interface CorecontractsApiAccountsGetTotalAssetsRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetTotalAssets
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiaccountsGetTotalAssets
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetBlockInfoRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetBlockInfo
     */
    chainID: string
    /**
     * BlockIndex (uint32)
     * Minimum: 1
     * Defaults to: undefined
     * @type number
     * @memberof CorecontractsApiblocklogGetBlockInfo
     */
    blockIndex: number
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetBlockInfo
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetControlAddressesRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetControlAddresses
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetControlAddresses
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetEventsOfBlockRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetEventsOfBlock
     */
    chainID: string
    /**
     * BlockIndex (uint32)
     * Minimum: 1
     * Defaults to: undefined
     * @type number
     * @memberof CorecontractsApiblocklogGetEventsOfBlock
     */
    blockIndex: number
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetEventsOfBlock
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetEventsOfLatestBlockRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetEventsOfLatestBlock
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetEventsOfLatestBlock
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetEventsOfRequestRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetEventsOfRequest
     */
    chainID: string
    /**
     * RequestID (Hex)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetEventsOfRequest
     */
    requestID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetEventsOfRequest
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetLatestBlockInfoRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetLatestBlockInfo
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetLatestBlockInfo
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetRequestIDsForBlockRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestIDsForBlock
     */
    chainID: string
    /**
     * BlockIndex (uint32)
     * Minimum: 1
     * Defaults to: undefined
     * @type number
     * @memberof CorecontractsApiblocklogGetRequestIDsForBlock
     */
    blockIndex: number
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestIDsForBlock
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetRequestIDsForLatestBlockRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestIDsForLatestBlock
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestIDsForLatestBlock
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetRequestIsProcessedRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestIsProcessed
     */
    chainID: string
    /**
     * RequestID (Hex)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestIsProcessed
     */
    requestID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestIsProcessed
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetRequestReceiptRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestReceipt
     */
    chainID: string
    /**
     * RequestID (Hex)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestReceipt
     */
    requestID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestReceipt
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetRequestReceiptsOfBlockRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestReceiptsOfBlock
     */
    chainID: string
    /**
     * BlockIndex (uint32)
     * Minimum: 1
     * Defaults to: undefined
     * @type number
     * @memberof CorecontractsApiblocklogGetRequestReceiptsOfBlock
     */
    blockIndex: number
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestReceiptsOfBlock
     */
    block?: string
}

export interface CorecontractsApiBlocklogGetRequestReceiptsOfLatestBlockRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestReceiptsOfLatestBlock
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApiblocklogGetRequestReceiptsOfLatestBlock
     */
    block?: string
}

export interface CorecontractsApiErrorsGetErrorMessageFormatRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApierrorsGetErrorMessageFormat
     */
    chainID: string
    /**
     * Contract (Hname as Hex)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApierrorsGetErrorMessageFormat
     */
    contractHname: string
    /**
     * Error Id (uint16)
     * Minimum: 1
     * Defaults to: undefined
     * @type number
     * @memberof CorecontractsApierrorsGetErrorMessageFormat
     */
    errorID: number
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApierrorsGetErrorMessageFormat
     */
    block?: string
}

export interface CorecontractsApiGovernanceGetAllowedStateControllerAddressesRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApigovernanceGetAllowedStateControllerAddresses
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApigovernanceGetAllowedStateControllerAddresses
     */
    block?: string
}

export interface CorecontractsApiGovernanceGetChainInfoRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApigovernanceGetChainInfo
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApigovernanceGetChainInfo
     */
    block?: string
}

export interface CorecontractsApiGovernanceGetChainOwnerRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApigovernanceGetChainOwner
     */
    chainID: string
    /**
     * Block index or trie root
     * Defaults to: undefined
     * @type string
     * @memberof CorecontractsApigovernanceGetChainOwner
     */
    block?: string
}

export class ObjectCorecontractsApi {
    private api: ObservableCorecontractsApi

    public constructor(configuration: Configuration, requestFactory?: CorecontractsApiRequestFactory, responseProcessor?: CorecontractsApiResponseProcessor) {
        this.api = new ObservableCorecontractsApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Get all assets belonging to an account
     * @param param the request object
     */
    public accountsGetAccountBalanceWithHttpInfo(param: CorecontractsApiAccountsGetAccountBalanceRequest, options?: Configuration): Promise<HttpInfo<AssetsResponse>> {
        return this.api.accountsGetAccountBalanceWithHttpInfo(param.chainID, param.agentID, param.block,  options).toPromise();
    }

    /**
     * Get all assets belonging to an account
     * @param param the request object
     */
    public accountsGetAccountBalance(param: CorecontractsApiAccountsGetAccountBalanceRequest, options?: Configuration): Promise<AssetsResponse> {
        return this.api.accountsGetAccountBalance(param.chainID, param.agentID, param.block,  options).toPromise();
    }

    /**
     * Get all foundries owned by an account
     * @param param the request object
     */
    public accountsGetAccountFoundriesWithHttpInfo(param: CorecontractsApiAccountsGetAccountFoundriesRequest, options?: Configuration): Promise<HttpInfo<AccountFoundriesResponse>> {
        return this.api.accountsGetAccountFoundriesWithHttpInfo(param.chainID, param.agentID, param.block,  options).toPromise();
    }

    /**
     * Get all foundries owned by an account
     * @param param the request object
     */
    public accountsGetAccountFoundries(param: CorecontractsApiAccountsGetAccountFoundriesRequest, options?: Configuration): Promise<AccountFoundriesResponse> {
        return this.api.accountsGetAccountFoundries(param.chainID, param.agentID, param.block,  options).toPromise();
    }

    /**
     * Get all NFT ids belonging to an account
     * @param param the request object
     */
    public accountsGetAccountNFTIDsWithHttpInfo(param: CorecontractsApiAccountsGetAccountNFTIDsRequest, options?: Configuration): Promise<HttpInfo<AccountNFTsResponse>> {
        return this.api.accountsGetAccountNFTIDsWithHttpInfo(param.chainID, param.agentID, param.block,  options).toPromise();
    }

    /**
     * Get all NFT ids belonging to an account
     * @param param the request object
     */
    public accountsGetAccountNFTIDs(param: CorecontractsApiAccountsGetAccountNFTIDsRequest, options?: Configuration): Promise<AccountNFTsResponse> {
        return this.api.accountsGetAccountNFTIDs(param.chainID, param.agentID, param.block,  options).toPromise();
    }

    /**
     * Get the current nonce of an account
     * @param param the request object
     */
    public accountsGetAccountNonceWithHttpInfo(param: CorecontractsApiAccountsGetAccountNonceRequest, options?: Configuration): Promise<HttpInfo<AccountNonceResponse>> {
        return this.api.accountsGetAccountNonceWithHttpInfo(param.chainID, param.agentID, param.block,  options).toPromise();
    }

    /**
     * Get the current nonce of an account
     * @param param the request object
     */
    public accountsGetAccountNonce(param: CorecontractsApiAccountsGetAccountNonceRequest, options?: Configuration): Promise<AccountNonceResponse> {
        return this.api.accountsGetAccountNonce(param.chainID, param.agentID, param.block,  options).toPromise();
    }

    /**
     * Get the foundry output
     * @param param the request object
     */
    public accountsGetFoundryOutputWithHttpInfo(param: CorecontractsApiAccountsGetFoundryOutputRequest, options?: Configuration): Promise<HttpInfo<FoundryOutputResponse>> {
        return this.api.accountsGetFoundryOutputWithHttpInfo(param.chainID, param.serialNumber, param.block,  options).toPromise();
    }

    /**
     * Get the foundry output
     * @param param the request object
     */
    public accountsGetFoundryOutput(param: CorecontractsApiAccountsGetFoundryOutputRequest, options?: Configuration): Promise<FoundryOutputResponse> {
        return this.api.accountsGetFoundryOutput(param.chainID, param.serialNumber, param.block,  options).toPromise();
    }

    /**
     * Get the NFT data by an ID
     * @param param the request object
     */
    public accountsGetNFTDataWithHttpInfo(param: CorecontractsApiAccountsGetNFTDataRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.accountsGetNFTDataWithHttpInfo(param.chainID, param.nftID, param.block,  options).toPromise();
    }

    /**
     * Get the NFT data by an ID
     * @param param the request object
     */
    public accountsGetNFTData(param: CorecontractsApiAccountsGetNFTDataRequest, options?: Configuration): Promise<void> {
        return this.api.accountsGetNFTData(param.chainID, param.nftID, param.block,  options).toPromise();
    }

    /**
     * Get a list of all registries
     * @param param the request object
     */
    public accountsGetNativeTokenIDRegistryWithHttpInfo(param: CorecontractsApiAccountsGetNativeTokenIDRegistryRequest, options?: Configuration): Promise<HttpInfo<NativeTokenIDRegistryResponse>> {
        return this.api.accountsGetNativeTokenIDRegistryWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get a list of all registries
     * @param param the request object
     */
    public accountsGetNativeTokenIDRegistry(param: CorecontractsApiAccountsGetNativeTokenIDRegistryRequest, options?: Configuration): Promise<NativeTokenIDRegistryResponse> {
        return this.api.accountsGetNativeTokenIDRegistry(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get all stored assets
     * @param param the request object
     */
    public accountsGetTotalAssetsWithHttpInfo(param: CorecontractsApiAccountsGetTotalAssetsRequest, options?: Configuration): Promise<HttpInfo<AssetsResponse>> {
        return this.api.accountsGetTotalAssetsWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get all stored assets
     * @param param the request object
     */
    public accountsGetTotalAssets(param: CorecontractsApiAccountsGetTotalAssetsRequest, options?: Configuration): Promise<AssetsResponse> {
        return this.api.accountsGetTotalAssets(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get the block info of a certain block index
     * @param param the request object
     */
    public blocklogGetBlockInfoWithHttpInfo(param: CorecontractsApiBlocklogGetBlockInfoRequest, options?: Configuration): Promise<HttpInfo<BlockInfoResponse>> {
        return this.api.blocklogGetBlockInfoWithHttpInfo(param.chainID, param.blockIndex, param.block,  options).toPromise();
    }

    /**
     * Get the block info of a certain block index
     * @param param the request object
     */
    public blocklogGetBlockInfo(param: CorecontractsApiBlocklogGetBlockInfoRequest, options?: Configuration): Promise<BlockInfoResponse> {
        return this.api.blocklogGetBlockInfo(param.chainID, param.blockIndex, param.block,  options).toPromise();
    }

    /**
     * Get the control addresses
     * @param param the request object
     */
    public blocklogGetControlAddressesWithHttpInfo(param: CorecontractsApiBlocklogGetControlAddressesRequest, options?: Configuration): Promise<HttpInfo<ControlAddressesResponse>> {
        return this.api.blocklogGetControlAddressesWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get the control addresses
     * @param param the request object
     */
    public blocklogGetControlAddresses(param: CorecontractsApiBlocklogGetControlAddressesRequest, options?: Configuration): Promise<ControlAddressesResponse> {
        return this.api.blocklogGetControlAddresses(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get events of a block
     * @param param the request object
     */
    public blocklogGetEventsOfBlockWithHttpInfo(param: CorecontractsApiBlocklogGetEventsOfBlockRequest, options?: Configuration): Promise<HttpInfo<EventsResponse>> {
        return this.api.blocklogGetEventsOfBlockWithHttpInfo(param.chainID, param.blockIndex, param.block,  options).toPromise();
    }

    /**
     * Get events of a block
     * @param param the request object
     */
    public blocklogGetEventsOfBlock(param: CorecontractsApiBlocklogGetEventsOfBlockRequest, options?: Configuration): Promise<EventsResponse> {
        return this.api.blocklogGetEventsOfBlock(param.chainID, param.blockIndex, param.block,  options).toPromise();
    }

    /**
     * Get events of the latest block
     * @param param the request object
     */
    public blocklogGetEventsOfLatestBlockWithHttpInfo(param: CorecontractsApiBlocklogGetEventsOfLatestBlockRequest, options?: Configuration): Promise<HttpInfo<EventsResponse>> {
        return this.api.blocklogGetEventsOfLatestBlockWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get events of the latest block
     * @param param the request object
     */
    public blocklogGetEventsOfLatestBlock(param: CorecontractsApiBlocklogGetEventsOfLatestBlockRequest, options?: Configuration): Promise<EventsResponse> {
        return this.api.blocklogGetEventsOfLatestBlock(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get events of a request
     * @param param the request object
     */
    public blocklogGetEventsOfRequestWithHttpInfo(param: CorecontractsApiBlocklogGetEventsOfRequestRequest, options?: Configuration): Promise<HttpInfo<EventsResponse>> {
        return this.api.blocklogGetEventsOfRequestWithHttpInfo(param.chainID, param.requestID, param.block,  options).toPromise();
    }

    /**
     * Get events of a request
     * @param param the request object
     */
    public blocklogGetEventsOfRequest(param: CorecontractsApiBlocklogGetEventsOfRequestRequest, options?: Configuration): Promise<EventsResponse> {
        return this.api.blocklogGetEventsOfRequest(param.chainID, param.requestID, param.block,  options).toPromise();
    }

    /**
     * Get the block info of the latest block
     * @param param the request object
     */
    public blocklogGetLatestBlockInfoWithHttpInfo(param: CorecontractsApiBlocklogGetLatestBlockInfoRequest, options?: Configuration): Promise<HttpInfo<BlockInfoResponse>> {
        return this.api.blocklogGetLatestBlockInfoWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get the block info of the latest block
     * @param param the request object
     */
    public blocklogGetLatestBlockInfo(param: CorecontractsApiBlocklogGetLatestBlockInfoRequest, options?: Configuration): Promise<BlockInfoResponse> {
        return this.api.blocklogGetLatestBlockInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get the request ids for a certain block index
     * @param param the request object
     */
    public blocklogGetRequestIDsForBlockWithHttpInfo(param: CorecontractsApiBlocklogGetRequestIDsForBlockRequest, options?: Configuration): Promise<HttpInfo<RequestIDsResponse>> {
        return this.api.blocklogGetRequestIDsForBlockWithHttpInfo(param.chainID, param.blockIndex, param.block,  options).toPromise();
    }

    /**
     * Get the request ids for a certain block index
     * @param param the request object
     */
    public blocklogGetRequestIDsForBlock(param: CorecontractsApiBlocklogGetRequestIDsForBlockRequest, options?: Configuration): Promise<RequestIDsResponse> {
        return this.api.blocklogGetRequestIDsForBlock(param.chainID, param.blockIndex, param.block,  options).toPromise();
    }

    /**
     * Get the request ids for the latest block
     * @param param the request object
     */
    public blocklogGetRequestIDsForLatestBlockWithHttpInfo(param: CorecontractsApiBlocklogGetRequestIDsForLatestBlockRequest, options?: Configuration): Promise<HttpInfo<RequestIDsResponse>> {
        return this.api.blocklogGetRequestIDsForLatestBlockWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get the request ids for the latest block
     * @param param the request object
     */
    public blocklogGetRequestIDsForLatestBlock(param: CorecontractsApiBlocklogGetRequestIDsForLatestBlockRequest, options?: Configuration): Promise<RequestIDsResponse> {
        return this.api.blocklogGetRequestIDsForLatestBlock(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get the request processing status
     * @param param the request object
     */
    public blocklogGetRequestIsProcessedWithHttpInfo(param: CorecontractsApiBlocklogGetRequestIsProcessedRequest, options?: Configuration): Promise<HttpInfo<RequestProcessedResponse>> {
        return this.api.blocklogGetRequestIsProcessedWithHttpInfo(param.chainID, param.requestID, param.block,  options).toPromise();
    }

    /**
     * Get the request processing status
     * @param param the request object
     */
    public blocklogGetRequestIsProcessed(param: CorecontractsApiBlocklogGetRequestIsProcessedRequest, options?: Configuration): Promise<RequestProcessedResponse> {
        return this.api.blocklogGetRequestIsProcessed(param.chainID, param.requestID, param.block,  options).toPromise();
    }

    /**
     * Get the receipt of a certain request id
     * @param param the request object
     */
    public blocklogGetRequestReceiptWithHttpInfo(param: CorecontractsApiBlocklogGetRequestReceiptRequest, options?: Configuration): Promise<HttpInfo<ReceiptResponse>> {
        return this.api.blocklogGetRequestReceiptWithHttpInfo(param.chainID, param.requestID, param.block,  options).toPromise();
    }

    /**
     * Get the receipt of a certain request id
     * @param param the request object
     */
    public blocklogGetRequestReceipt(param: CorecontractsApiBlocklogGetRequestReceiptRequest, options?: Configuration): Promise<ReceiptResponse> {
        return this.api.blocklogGetRequestReceipt(param.chainID, param.requestID, param.block,  options).toPromise();
    }

    /**
     * Get all receipts of a certain block
     * @param param the request object
     */
    public blocklogGetRequestReceiptsOfBlockWithHttpInfo(param: CorecontractsApiBlocklogGetRequestReceiptsOfBlockRequest, options?: Configuration): Promise<HttpInfo<Array<ReceiptResponse>>> {
        return this.api.blocklogGetRequestReceiptsOfBlockWithHttpInfo(param.chainID, param.blockIndex, param.block,  options).toPromise();
    }

    /**
     * Get all receipts of a certain block
     * @param param the request object
     */
    public blocklogGetRequestReceiptsOfBlock(param: CorecontractsApiBlocklogGetRequestReceiptsOfBlockRequest, options?: Configuration): Promise<Array<ReceiptResponse>> {
        return this.api.blocklogGetRequestReceiptsOfBlock(param.chainID, param.blockIndex, param.block,  options).toPromise();
    }

    /**
     * Get all receipts of the latest block
     * @param param the request object
     */
    public blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(param: CorecontractsApiBlocklogGetRequestReceiptsOfLatestBlockRequest, options?: Configuration): Promise<HttpInfo<Array<ReceiptResponse>>> {
        return this.api.blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get all receipts of the latest block
     * @param param the request object
     */
    public blocklogGetRequestReceiptsOfLatestBlock(param: CorecontractsApiBlocklogGetRequestReceiptsOfLatestBlockRequest, options?: Configuration): Promise<Array<ReceiptResponse>> {
        return this.api.blocklogGetRequestReceiptsOfLatestBlock(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Get the error message format of a specific error id
     * @param param the request object
     */
    public errorsGetErrorMessageFormatWithHttpInfo(param: CorecontractsApiErrorsGetErrorMessageFormatRequest, options?: Configuration): Promise<HttpInfo<ErrorMessageFormatResponse>> {
        return this.api.errorsGetErrorMessageFormatWithHttpInfo(param.chainID, param.contractHname, param.errorID, param.block,  options).toPromise();
    }

    /**
     * Get the error message format of a specific error id
     * @param param the request object
     */
    public errorsGetErrorMessageFormat(param: CorecontractsApiErrorsGetErrorMessageFormatRequest, options?: Configuration): Promise<ErrorMessageFormatResponse> {
        return this.api.errorsGetErrorMessageFormat(param.chainID, param.contractHname, param.errorID, param.block,  options).toPromise();
    }

    /**
     * Returns the allowed state controller addresses
     * Get the allowed state controller addresses
     * @param param the request object
     */
    public governanceGetAllowedStateControllerAddressesWithHttpInfo(param: CorecontractsApiGovernanceGetAllowedStateControllerAddressesRequest, options?: Configuration): Promise<HttpInfo<GovAllowedStateControllerAddressesResponse>> {
        return this.api.governanceGetAllowedStateControllerAddressesWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Returns the allowed state controller addresses
     * Get the allowed state controller addresses
     * @param param the request object
     */
    public governanceGetAllowedStateControllerAddresses(param: CorecontractsApiGovernanceGetAllowedStateControllerAddressesRequest, options?: Configuration): Promise<GovAllowedStateControllerAddressesResponse> {
        return this.api.governanceGetAllowedStateControllerAddresses(param.chainID, param.block,  options).toPromise();
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chains/:chainID\' to get information about a chain.
     * Get the chain info
     * @param param the request object
     */
    public governanceGetChainInfoWithHttpInfo(param: CorecontractsApiGovernanceGetChainInfoRequest, options?: Configuration): Promise<HttpInfo<GovChainInfoResponse>> {
        return this.api.governanceGetChainInfoWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chains/:chainID\' to get information about a chain.
     * Get the chain info
     * @param param the request object
     */
    public governanceGetChainInfo(param: CorecontractsApiGovernanceGetChainInfoRequest, options?: Configuration): Promise<GovChainInfoResponse> {
        return this.api.governanceGetChainInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Returns the chain owner
     * Get the chain owner
     * @param param the request object
     */
    public governanceGetChainOwnerWithHttpInfo(param: CorecontractsApiGovernanceGetChainOwnerRequest, options?: Configuration): Promise<HttpInfo<GovChainOwnerResponse>> {
        return this.api.governanceGetChainOwnerWithHttpInfo(param.chainID, param.block,  options).toPromise();
    }

    /**
     * Returns the chain owner
     * Get the chain owner
     * @param param the request object
     */
    public governanceGetChainOwner(param: CorecontractsApiGovernanceGetChainOwnerRequest, options?: Configuration): Promise<GovChainOwnerResponse> {
        return this.api.governanceGetChainOwner(param.chainID, param.block,  options).toPromise();
    }

}

import { ObservableDefaultApi } from "./ObservableAPI";
import { DefaultApiRequestFactory, DefaultApiResponseProcessor} from "../apis/DefaultApi";

export interface DefaultApiGetHealthRequest {
}

export interface DefaultApiV1WsGetRequest {
}

export class ObjectDefaultApi {
    private api: ObservableDefaultApi

    public constructor(configuration: Configuration, requestFactory?: DefaultApiRequestFactory, responseProcessor?: DefaultApiResponseProcessor) {
        this.api = new ObservableDefaultApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Returns 200 if the node is healthy.
     * @param param the request object
     */
    public getHealthWithHttpInfo(param: DefaultApiGetHealthRequest = {}, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.getHealthWithHttpInfo( options).toPromise();
    }

    /**
     * Returns 200 if the node is healthy.
     * @param param the request object
     */
    public getHealth(param: DefaultApiGetHealthRequest = {}, options?: Configuration): Promise<void> {
        return this.api.getHealth( options).toPromise();
    }

    /**
     * The websocket connection service
     * @param param the request object
     */
    public v1WsGetWithHttpInfo(param: DefaultApiV1WsGetRequest = {}, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.v1WsGetWithHttpInfo( options).toPromise();
    }

    /**
     * The websocket connection service
     * @param param the request object
     */
    public v1WsGet(param: DefaultApiV1WsGetRequest = {}, options?: Configuration): Promise<void> {
        return this.api.v1WsGet( options).toPromise();
    }

}

import { ObservableMetricsApi } from "./ObservableAPI";
import { MetricsApiRequestFactory, MetricsApiResponseProcessor} from "../apis/MetricsApi";

export interface MetricsApiGetChainMessageMetricsRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof MetricsApigetChainMessageMetrics
     */
    chainID: string
}

export interface MetricsApiGetChainPipeMetricsRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof MetricsApigetChainPipeMetrics
     */
    chainID: string
}

export interface MetricsApiGetChainWorkflowMetricsRequest {
    /**
     * ChainID (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof MetricsApigetChainWorkflowMetrics
     */
    chainID: string
}

export interface MetricsApiGetNodeMessageMetricsRequest {
}

export class ObjectMetricsApi {
    private api: ObservableMetricsApi

    public constructor(configuration: Configuration, requestFactory?: MetricsApiRequestFactory, responseProcessor?: MetricsApiResponseProcessor) {
        this.api = new ObservableMetricsApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Get chain specific message metrics.
     * @param param the request object
     */
    public getChainMessageMetricsWithHttpInfo(param: MetricsApiGetChainMessageMetricsRequest, options?: Configuration): Promise<HttpInfo<ChainMessageMetrics>> {
        return this.api.getChainMessageMetricsWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * Get chain specific message metrics.
     * @param param the request object
     */
    public getChainMessageMetrics(param: MetricsApiGetChainMessageMetricsRequest, options?: Configuration): Promise<ChainMessageMetrics> {
        return this.api.getChainMessageMetrics(param.chainID,  options).toPromise();
    }

    /**
     * Get chain pipe event metrics.
     * @param param the request object
     */
    public getChainPipeMetricsWithHttpInfo(param: MetricsApiGetChainPipeMetricsRequest, options?: Configuration): Promise<HttpInfo<ConsensusPipeMetrics>> {
        return this.api.getChainPipeMetricsWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * Get chain pipe event metrics.
     * @param param the request object
     */
    public getChainPipeMetrics(param: MetricsApiGetChainPipeMetricsRequest, options?: Configuration): Promise<ConsensusPipeMetrics> {
        return this.api.getChainPipeMetrics(param.chainID,  options).toPromise();
    }

    /**
     * Get chain workflow metrics.
     * @param param the request object
     */
    public getChainWorkflowMetricsWithHttpInfo(param: MetricsApiGetChainWorkflowMetricsRequest, options?: Configuration): Promise<HttpInfo<ConsensusWorkflowMetrics>> {
        return this.api.getChainWorkflowMetricsWithHttpInfo(param.chainID,  options).toPromise();
    }

    /**
     * Get chain workflow metrics.
     * @param param the request object
     */
    public getChainWorkflowMetrics(param: MetricsApiGetChainWorkflowMetricsRequest, options?: Configuration): Promise<ConsensusWorkflowMetrics> {
        return this.api.getChainWorkflowMetrics(param.chainID,  options).toPromise();
    }

    /**
     * Get accumulated message metrics.
     * @param param the request object
     */
    public getNodeMessageMetricsWithHttpInfo(param: MetricsApiGetNodeMessageMetricsRequest = {}, options?: Configuration): Promise<HttpInfo<NodeMessageMetrics>> {
        return this.api.getNodeMessageMetricsWithHttpInfo( options).toPromise();
    }

    /**
     * Get accumulated message metrics.
     * @param param the request object
     */
    public getNodeMessageMetrics(param: MetricsApiGetNodeMessageMetricsRequest = {}, options?: Configuration): Promise<NodeMessageMetrics> {
        return this.api.getNodeMessageMetrics( options).toPromise();
    }

}

import { ObservableNodeApi } from "./ObservableAPI";
import { NodeApiRequestFactory, NodeApiResponseProcessor} from "../apis/NodeApi";

export interface NodeApiDistrustPeerRequest {
    /**
     * Name or PubKey (hex) of the trusted peer
     * Defaults to: undefined
     * @type string
     * @memberof NodeApidistrustPeer
     */
    peer: string
}

export interface NodeApiGenerateDKSRequest {
    /**
     * Request parameters
     * @type DKSharesPostRequest
     * @memberof NodeApigenerateDKS
     */
    dKSharesPostRequest: DKSharesPostRequest
}

export interface NodeApiGetAllPeersRequest {
}

export interface NodeApiGetConfigurationRequest {
}

export interface NodeApiGetDKSInfoRequest {
    /**
     * SharedAddress (Hex Address)
     * Defaults to: undefined
     * @type string
     * @memberof NodeApigetDKSInfo
     */
    sharedAddress: string
}

export interface NodeApiGetInfoRequest {
}

export interface NodeApiGetPeeringIdentityRequest {
}

export interface NodeApiGetTrustedPeersRequest {
}

export interface NodeApiGetVersionRequest {
}

export interface NodeApiOwnerCertificateRequest {
}

export interface NodeApiShutdownNodeRequest {
}

export interface NodeApiTrustPeerRequest {
    /**
     * Info of the peer to trust
     * @type PeeringTrustRequest
     * @memberof NodeApitrustPeer
     */
    peeringTrustRequest: PeeringTrustRequest
}

export class ObjectNodeApi {
    private api: ObservableNodeApi

    public constructor(configuration: Configuration, requestFactory?: NodeApiRequestFactory, responseProcessor?: NodeApiResponseProcessor) {
        this.api = new ObservableNodeApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Distrust a peering node
     * @param param the request object
     */
    public distrustPeerWithHttpInfo(param: NodeApiDistrustPeerRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.distrustPeerWithHttpInfo(param.peer,  options).toPromise();
    }

    /**
     * Distrust a peering node
     * @param param the request object
     */
    public distrustPeer(param: NodeApiDistrustPeerRequest, options?: Configuration): Promise<void> {
        return this.api.distrustPeer(param.peer,  options).toPromise();
    }

    /**
     * Generate a new distributed key
     * @param param the request object
     */
    public generateDKSWithHttpInfo(param: NodeApiGenerateDKSRequest, options?: Configuration): Promise<HttpInfo<DKSharesInfo>> {
        return this.api.generateDKSWithHttpInfo(param.dKSharesPostRequest,  options).toPromise();
    }

    /**
     * Generate a new distributed key
     * @param param the request object
     */
    public generateDKS(param: NodeApiGenerateDKSRequest, options?: Configuration): Promise<DKSharesInfo> {
        return this.api.generateDKS(param.dKSharesPostRequest,  options).toPromise();
    }

    /**
     * Get basic information about all configured peers
     * @param param the request object
     */
    public getAllPeersWithHttpInfo(param: NodeApiGetAllPeersRequest = {}, options?: Configuration): Promise<HttpInfo<Array<PeeringNodeStatusResponse>>> {
        return this.api.getAllPeersWithHttpInfo( options).toPromise();
    }

    /**
     * Get basic information about all configured peers
     * @param param the request object
     */
    public getAllPeers(param: NodeApiGetAllPeersRequest = {}, options?: Configuration): Promise<Array<PeeringNodeStatusResponse>> {
        return this.api.getAllPeers( options).toPromise();
    }

    /**
     * Return the Wasp configuration
     * @param param the request object
     */
    public getConfigurationWithHttpInfo(param: NodeApiGetConfigurationRequest = {}, options?: Configuration): Promise<HttpInfo<{ [key: string]: string; }>> {
        return this.api.getConfigurationWithHttpInfo( options).toPromise();
    }

    /**
     * Return the Wasp configuration
     * @param param the request object
     */
    public getConfiguration(param: NodeApiGetConfigurationRequest = {}, options?: Configuration): Promise<{ [key: string]: string; }> {
        return this.api.getConfiguration( options).toPromise();
    }

    /**
     * Get information about the shared address DKS configuration
     * @param param the request object
     */
    public getDKSInfoWithHttpInfo(param: NodeApiGetDKSInfoRequest, options?: Configuration): Promise<HttpInfo<DKSharesInfo>> {
        return this.api.getDKSInfoWithHttpInfo(param.sharedAddress,  options).toPromise();
    }

    /**
     * Get information about the shared address DKS configuration
     * @param param the request object
     */
    public getDKSInfo(param: NodeApiGetDKSInfoRequest, options?: Configuration): Promise<DKSharesInfo> {
        return this.api.getDKSInfo(param.sharedAddress,  options).toPromise();
    }

    /**
     * Returns private information about this node.
     * @param param the request object
     */
    public getInfoWithHttpInfo(param: NodeApiGetInfoRequest = {}, options?: Configuration): Promise<HttpInfo<InfoResponse>> {
        return this.api.getInfoWithHttpInfo( options).toPromise();
    }

    /**
     * Returns private information about this node.
     * @param param the request object
     */
    public getInfo(param: NodeApiGetInfoRequest = {}, options?: Configuration): Promise<InfoResponse> {
        return this.api.getInfo( options).toPromise();
    }

    /**
     * Get basic peer info of the current node
     * @param param the request object
     */
    public getPeeringIdentityWithHttpInfo(param: NodeApiGetPeeringIdentityRequest = {}, options?: Configuration): Promise<HttpInfo<PeeringNodeIdentityResponse>> {
        return this.api.getPeeringIdentityWithHttpInfo( options).toPromise();
    }

    /**
     * Get basic peer info of the current node
     * @param param the request object
     */
    public getPeeringIdentity(param: NodeApiGetPeeringIdentityRequest = {}, options?: Configuration): Promise<PeeringNodeIdentityResponse> {
        return this.api.getPeeringIdentity( options).toPromise();
    }

    /**
     * Get trusted peers
     * @param param the request object
     */
    public getTrustedPeersWithHttpInfo(param: NodeApiGetTrustedPeersRequest = {}, options?: Configuration): Promise<HttpInfo<Array<PeeringNodeIdentityResponse>>> {
        return this.api.getTrustedPeersWithHttpInfo( options).toPromise();
    }

    /**
     * Get trusted peers
     * @param param the request object
     */
    public getTrustedPeers(param: NodeApiGetTrustedPeersRequest = {}, options?: Configuration): Promise<Array<PeeringNodeIdentityResponse>> {
        return this.api.getTrustedPeers( options).toPromise();
    }

    /**
     * Returns the node version.
     * @param param the request object
     */
    public getVersionWithHttpInfo(param: NodeApiGetVersionRequest = {}, options?: Configuration): Promise<HttpInfo<VersionResponse>> {
        return this.api.getVersionWithHttpInfo( options).toPromise();
    }

    /**
     * Returns the node version.
     * @param param the request object
     */
    public getVersion(param: NodeApiGetVersionRequest = {}, options?: Configuration): Promise<VersionResponse> {
        return this.api.getVersion( options).toPromise();
    }

    /**
     * Gets the node owner
     * @param param the request object
     */
    public ownerCertificateWithHttpInfo(param: NodeApiOwnerCertificateRequest = {}, options?: Configuration): Promise<HttpInfo<NodeOwnerCertificateResponse>> {
        return this.api.ownerCertificateWithHttpInfo( options).toPromise();
    }

    /**
     * Gets the node owner
     * @param param the request object
     */
    public ownerCertificate(param: NodeApiOwnerCertificateRequest = {}, options?: Configuration): Promise<NodeOwnerCertificateResponse> {
        return this.api.ownerCertificate( options).toPromise();
    }

    /**
     * Shut down the node
     * @param param the request object
     */
    public shutdownNodeWithHttpInfo(param: NodeApiShutdownNodeRequest = {}, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.shutdownNodeWithHttpInfo( options).toPromise();
    }

    /**
     * Shut down the node
     * @param param the request object
     */
    public shutdownNode(param: NodeApiShutdownNodeRequest = {}, options?: Configuration): Promise<void> {
        return this.api.shutdownNode( options).toPromise();
    }

    /**
     * Trust a peering node
     * @param param the request object
     */
    public trustPeerWithHttpInfo(param: NodeApiTrustPeerRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.trustPeerWithHttpInfo(param.peeringTrustRequest,  options).toPromise();
    }

    /**
     * Trust a peering node
     * @param param the request object
     */
    public trustPeer(param: NodeApiTrustPeerRequest, options?: Configuration): Promise<void> {
        return this.api.trustPeer(param.peeringTrustRequest,  options).toPromise();
    }

}

import { ObservableRequestsApi } from "./ObservableAPI";
import { RequestsApiRequestFactory, RequestsApiResponseProcessor} from "../apis/RequestsApi";

export interface RequestsApiOffLedgerRequest {
    /**
     * Offledger request as JSON. Request encoded in Hex
     * @type OffLedgerRequest
     * @memberof RequestsApioffLedger
     */
    offLedgerRequest: OffLedgerRequest
}

export class ObjectRequestsApi {
    private api: ObservableRequestsApi

    public constructor(configuration: Configuration, requestFactory?: RequestsApiRequestFactory, responseProcessor?: RequestsApiResponseProcessor) {
        this.api = new ObservableRequestsApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Post an off-ledger request
     * @param param the request object
     */
    public offLedgerWithHttpInfo(param: RequestsApiOffLedgerRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.offLedgerWithHttpInfo(param.offLedgerRequest,  options).toPromise();
    }

    /**
     * Post an off-ledger request
     * @param param the request object
     */
    public offLedger(param: RequestsApiOffLedgerRequest, options?: Configuration): Promise<void> {
        return this.api.offLedger(param.offLedgerRequest,  options).toPromise();
    }

}

import { ObservableUsersApi } from "./ObservableAPI";
import { UsersApiRequestFactory, UsersApiResponseProcessor} from "../apis/UsersApi";

export interface UsersApiAddUserRequest {
    /**
     * The user data
     * @type AddUserRequest
     * @memberof UsersApiaddUser
     */
    addUserRequest: AddUserRequest
}

export interface UsersApiChangeUserPasswordRequest {
    /**
     * The username
     * Defaults to: undefined
     * @type string
     * @memberof UsersApichangeUserPassword
     */
    username: string
    /**
     * The users new password
     * @type UpdateUserPasswordRequest
     * @memberof UsersApichangeUserPassword
     */
    updateUserPasswordRequest: UpdateUserPasswordRequest
}

export interface UsersApiChangeUserPermissionsRequest {
    /**
     * The username
     * Defaults to: undefined
     * @type string
     * @memberof UsersApichangeUserPermissions
     */
    username: string
    /**
     * The users new permissions
     * @type UpdateUserPermissionsRequest
     * @memberof UsersApichangeUserPermissions
     */
    updateUserPermissionsRequest: UpdateUserPermissionsRequest
}

export interface UsersApiDeleteUserRequest {
    /**
     * The username
     * Defaults to: undefined
     * @type string
     * @memberof UsersApideleteUser
     */
    username: string
}

export interface UsersApiGetUserRequest {
    /**
     * The username
     * Defaults to: undefined
     * @type string
     * @memberof UsersApigetUser
     */
    username: string
}

export interface UsersApiGetUsersRequest {
}

export class ObjectUsersApi {
    private api: ObservableUsersApi

    public constructor(configuration: Configuration, requestFactory?: UsersApiRequestFactory, responseProcessor?: UsersApiResponseProcessor) {
        this.api = new ObservableUsersApi(configuration, requestFactory, responseProcessor);
    }

    /**
     * Add a user
     * @param param the request object
     */
    public addUserWithHttpInfo(param: UsersApiAddUserRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.addUserWithHttpInfo(param.addUserRequest,  options).toPromise();
    }

    /**
     * Add a user
     * @param param the request object
     */
    public addUser(param: UsersApiAddUserRequest, options?: Configuration): Promise<void> {
        return this.api.addUser(param.addUserRequest,  options).toPromise();
    }

    /**
     * Change user password
     * @param param the request object
     */
    public changeUserPasswordWithHttpInfo(param: UsersApiChangeUserPasswordRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.changeUserPasswordWithHttpInfo(param.username, param.updateUserPasswordRequest,  options).toPromise();
    }

    /**
     * Change user password
     * @param param the request object
     */
    public changeUserPassword(param: UsersApiChangeUserPasswordRequest, options?: Configuration): Promise<void> {
        return this.api.changeUserPassword(param.username, param.updateUserPasswordRequest,  options).toPromise();
    }

    /**
     * Change user permissions
     * @param param the request object
     */
    public changeUserPermissionsWithHttpInfo(param: UsersApiChangeUserPermissionsRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.changeUserPermissionsWithHttpInfo(param.username, param.updateUserPermissionsRequest,  options).toPromise();
    }

    /**
     * Change user permissions
     * @param param the request object
     */
    public changeUserPermissions(param: UsersApiChangeUserPermissionsRequest, options?: Configuration): Promise<void> {
        return this.api.changeUserPermissions(param.username, param.updateUserPermissionsRequest,  options).toPromise();
    }

    /**
     * Deletes a user
     * @param param the request object
     */
    public deleteUserWithHttpInfo(param: UsersApiDeleteUserRequest, options?: Configuration): Promise<HttpInfo<void>> {
        return this.api.deleteUserWithHttpInfo(param.username,  options).toPromise();
    }

    /**
     * Deletes a user
     * @param param the request object
     */
    public deleteUser(param: UsersApiDeleteUserRequest, options?: Configuration): Promise<void> {
        return this.api.deleteUser(param.username,  options).toPromise();
    }

    /**
     * Get a user
     * @param param the request object
     */
    public getUserWithHttpInfo(param: UsersApiGetUserRequest, options?: Configuration): Promise<HttpInfo<User>> {
        return this.api.getUserWithHttpInfo(param.username,  options).toPromise();
    }

    /**
     * Get a user
     * @param param the request object
     */
    public getUser(param: UsersApiGetUserRequest, options?: Configuration): Promise<User> {
        return this.api.getUser(param.username,  options).toPromise();
    }

    /**
     * Get a list of all users
     * @param param the request object
     */
    public getUsersWithHttpInfo(param: UsersApiGetUsersRequest = {}, options?: Configuration): Promise<HttpInfo<Array<User>>> {
        return this.api.getUsersWithHttpInfo( options).toPromise();
    }

    /**
     * Get a list of all users
     * @param param the request object
     */
    public getUsers(param: UsersApiGetUsersRequest = {}, options?: Configuration): Promise<Array<User>> {
        return this.api.getUsers( options).toPromise();
    }

}
