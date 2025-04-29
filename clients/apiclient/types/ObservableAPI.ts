import { ResponseContext, RequestContext, HttpFile, HttpInfo } from '../http/http';
import { Configuration} from '../configuration'
import { Observable, of, from } from '../rxjsStub';
import {mergeMap, map} from  '../rxjsStub';
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

import { AuthApiRequestFactory, AuthApiResponseProcessor} from "../apis/AuthApi";
export class ObservableAuthApi {
    private requestFactory: AuthApiRequestFactory;
    private responseProcessor: AuthApiResponseProcessor;
    private configuration: Configuration;

    public constructor(
        configuration: Configuration,
        requestFactory?: AuthApiRequestFactory,
        responseProcessor?: AuthApiResponseProcessor
    ) {
        this.configuration = configuration;
        this.requestFactory = requestFactory || new AuthApiRequestFactory(configuration);
        this.responseProcessor = responseProcessor || new AuthApiResponseProcessor();
    }

    /**
     * Get information about the current authentication mode
     */
    public authInfoWithHttpInfo(_options?: Configuration): Observable<HttpInfo<AuthInfoModel>> {
        const requestContextPromise = this.requestFactory.authInfo(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.authInfoWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get information about the current authentication mode
     */
    public authInfo(_options?: Configuration): Observable<AuthInfoModel> {
        return this.authInfoWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<AuthInfoModel>) => apiResponse.data));
    }

    /**
     * Authenticate towards the node
     * @param loginRequest The login request
     */
    public authenticateWithHttpInfo(loginRequest: LoginRequest, _options?: Configuration): Observable<HttpInfo<LoginResponse>> {
        const requestContextPromise = this.requestFactory.authenticate(loginRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.authenticateWithHttpInfo(rsp)));
            }));
    }

    /**
     * Authenticate towards the node
     * @param loginRequest The login request
     */
    public authenticate(loginRequest: LoginRequest, _options?: Configuration): Observable<LoginResponse> {
        return this.authenticateWithHttpInfo(loginRequest, _options).pipe(map((apiResponse: HttpInfo<LoginResponse>) => apiResponse.data));
    }

}

import { ChainsApiRequestFactory, ChainsApiResponseProcessor} from "../apis/ChainsApi";
export class ObservableChainsApi {
    private requestFactory: ChainsApiRequestFactory;
    private responseProcessor: ChainsApiResponseProcessor;
    private configuration: Configuration;

    public constructor(
        configuration: Configuration,
        requestFactory?: ChainsApiRequestFactory,
        responseProcessor?: ChainsApiResponseProcessor
    ) {
        this.configuration = configuration;
        this.requestFactory = requestFactory || new ChainsApiRequestFactory(configuration);
        this.responseProcessor = responseProcessor || new ChainsApiResponseProcessor();
    }

    /**
     * Activate a chain
     * @param chainID ChainID (Hex Address)
     */
    public activateChainWithHttpInfo(chainID: string, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.activateChain(chainID, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.activateChainWithHttpInfo(rsp)));
            }));
    }

    /**
     * Activate a chain
     * @param chainID ChainID (Hex Address)
     */
    public activateChain(chainID: string, _options?: Configuration): Observable<void> {
        return this.activateChainWithHttpInfo(chainID, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Configure a trusted node to be an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public addAccessNodeWithHttpInfo(peer: string, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.addAccessNode(peer, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.addAccessNodeWithHttpInfo(rsp)));
            }));
    }

    /**
     * Configure a trusted node to be an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public addAccessNode(peer: string, _options?: Configuration): Observable<void> {
        return this.addAccessNodeWithHttpInfo(peer, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param contractCallViewRequest Parameters
     */
    public callViewWithHttpInfo(contractCallViewRequest: ContractCallViewRequest, _options?: Configuration): Observable<HttpInfo<Array<string>>> {
        const requestContextPromise = this.requestFactory.callView(contractCallViewRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.callViewWithHttpInfo(rsp)));
            }));
    }

    /**
     * Execute a view call. Either use HName or Name properties. If both are supplied, HName are used.
     * Call a view function on a contract by Hname
     * @param contractCallViewRequest Parameters
     */
    public callView(contractCallViewRequest: ContractCallViewRequest, _options?: Configuration): Observable<Array<string>> {
        return this.callViewWithHttpInfo(contractCallViewRequest, _options).pipe(map((apiResponse: HttpInfo<Array<string>>) => apiResponse.data));
    }

    /**
     * Deactivate a chain
     */
    public deactivateChainWithHttpInfo(_options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.deactivateChain(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.deactivateChainWithHttpInfo(rsp)));
            }));
    }

    /**
     * Deactivate a chain
     */
    public deactivateChain(_options?: Configuration): Observable<void> {
        return this.deactivateChainWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * dump accounts information into a humanly-readable format
     */
    public dumpAccountsWithHttpInfo(_options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.dumpAccounts(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.dumpAccountsWithHttpInfo(rsp)));
            }));
    }

    /**
     * dump accounts information into a humanly-readable format
     */
    public dumpAccounts(_options?: Configuration): Observable<void> {
        return this.dumpAccountsWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param request Request
     */
    public estimateGasOffledgerWithHttpInfo(request: EstimateGasRequestOffledger, _options?: Configuration): Observable<HttpInfo<ReceiptResponse>> {
        const requestContextPromise = this.requestFactory.estimateGasOffledger(request, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.estimateGasOffledgerWithHttpInfo(rsp)));
            }));
    }

    /**
     * Estimates gas for a given off-ledger ISC request
     * @param request Request
     */
    public estimateGasOffledger(request: EstimateGasRequestOffledger, _options?: Configuration): Observable<ReceiptResponse> {
        return this.estimateGasOffledgerWithHttpInfo(request, _options).pipe(map((apiResponse: HttpInfo<ReceiptResponse>) => apiResponse.data));
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param request Request
     */
    public estimateGasOnledgerWithHttpInfo(request: EstimateGasRequestOnledger, _options?: Configuration): Observable<HttpInfo<ReceiptResponse>> {
        const requestContextPromise = this.requestFactory.estimateGasOnledger(request, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.estimateGasOnledgerWithHttpInfo(rsp)));
            }));
    }

    /**
     * Estimates gas for a given on-ledger ISC request
     * @param request Request
     */
    public estimateGasOnledger(request: EstimateGasRequestOnledger, _options?: Configuration): Observable<ReceiptResponse> {
        return this.estimateGasOnledgerWithHttpInfo(request, _options).pipe(map((apiResponse: HttpInfo<ReceiptResponse>) => apiResponse.data));
    }

    /**
     * Get information about the chain
     * @param [block] Block index or trie root
     */
    public getChainInfoWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<ChainInfoResponse>> {
        const requestContextPromise = this.requestFactory.getChainInfo(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getChainInfoWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get information about the chain
     * @param [block] Block index or trie root
     */
    public getChainInfo(block?: string, _options?: Configuration): Observable<ChainInfoResponse> {
        return this.getChainInfoWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<ChainInfoResponse>) => apiResponse.data));
    }

    /**
     * Get information about the deployed committee
     * @param [block] Block index or trie root
     */
    public getCommitteeInfoWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<CommitteeInfoResponse>> {
        const requestContextPromise = this.requestFactory.getCommitteeInfo(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getCommitteeInfoWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get information about the deployed committee
     * @param [block] Block index or trie root
     */
    public getCommitteeInfo(block?: string, _options?: Configuration): Observable<CommitteeInfoResponse> {
        return this.getCommitteeInfoWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<CommitteeInfoResponse>) => apiResponse.data));
    }

    /**
     * Get the contents of the mempool.
     */
    public getMempoolContentsWithHttpInfo(_options?: Configuration): Observable<HttpInfo<Array<number>>> {
        const requestContextPromise = this.requestFactory.getMempoolContents(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getMempoolContentsWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the contents of the mempool.
     */
    public getMempoolContents(_options?: Configuration): Observable<Array<number>> {
        return this.getMempoolContentsWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<Array<number>>) => apiResponse.data));
    }

    /**
     * Get a receipt from a request ID
     * @param requestID RequestID (Hex)
     */
    public getReceiptWithHttpInfo(requestID: string, _options?: Configuration): Observable<HttpInfo<ReceiptResponse>> {
        const requestContextPromise = this.requestFactory.getReceipt(requestID, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getReceiptWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get a receipt from a request ID
     * @param requestID RequestID (Hex)
     */
    public getReceipt(requestID: string, _options?: Configuration): Observable<ReceiptResponse> {
        return this.getReceiptWithHttpInfo(requestID, _options).pipe(map((apiResponse: HttpInfo<ReceiptResponse>) => apiResponse.data));
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param stateKey State Key (Hex)
     */
    public getStateValueWithHttpInfo(stateKey: string, _options?: Configuration): Observable<HttpInfo<StateResponse>> {
        const requestContextPromise = this.requestFactory.getStateValue(stateKey, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getStateValueWithHttpInfo(rsp)));
            }));
    }

    /**
     * Fetch the raw value associated with the given key in the chain state
     * @param stateKey State Key (Hex)
     */
    public getStateValue(stateKey: string, _options?: Configuration): Observable<StateResponse> {
        return this.getStateValueWithHttpInfo(stateKey, _options).pipe(map((apiResponse: HttpInfo<StateResponse>) => apiResponse.data));
    }

    /**
     * Remove an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public removeAccessNodeWithHttpInfo(peer: string, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.removeAccessNode(peer, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.removeAccessNodeWithHttpInfo(rsp)));
            }));
    }

    /**
     * Remove an access node.
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public removeAccessNode(peer: string, _options?: Configuration): Observable<void> {
        return this.removeAccessNodeWithHttpInfo(peer, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Rotate a chain
     * @param [rotateRequest] RotateRequest
     */
    public rotateChainWithHttpInfo(rotateRequest?: RotateChainRequest, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.rotateChain(rotateRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.rotateChainWithHttpInfo(rsp)));
            }));
    }

    /**
     * Rotate a chain
     * @param [rotateRequest] RotateRequest
     */
    public rotateChain(rotateRequest?: RotateChainRequest, _options?: Configuration): Observable<void> {
        return this.rotateChainWithHttpInfo(rotateRequest, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Sets the chain record.
     * @param chainID ChainID (Hex Address)
     * @param chainRecord Chain Record
     */
    public setChainRecordWithHttpInfo(chainID: string, chainRecord: ChainRecord, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.setChainRecord(chainID, chainRecord, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.setChainRecordWithHttpInfo(rsp)));
            }));
    }

    /**
     * Sets the chain record.
     * @param chainID ChainID (Hex Address)
     * @param chainRecord Chain Record
     */
    public setChainRecord(chainID: string, chainRecord: ChainRecord, _options?: Configuration): Observable<void> {
        return this.setChainRecordWithHttpInfo(chainID, chainRecord, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Ethereum JSON-RPC
     */
    public v1ChainEvmPostWithHttpInfo(_options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.v1ChainEvmPost(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.v1ChainEvmPostWithHttpInfo(rsp)));
            }));
    }

    /**
     * Ethereum JSON-RPC
     */
    public v1ChainEvmPost(_options?: Configuration): Observable<void> {
        return this.v1ChainEvmPostWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     */
    public v1ChainEvmWsGetWithHttpInfo(_options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.v1ChainEvmWsGet(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.v1ChainEvmWsGetWithHttpInfo(rsp)));
            }));
    }

    /**
     * Ethereum JSON-RPC (Websocket transport)
     */
    public v1ChainEvmWsGet(_options?: Configuration): Observable<void> {
        return this.v1ChainEvmWsGetWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Wait until the given request has been processed by the node
     * @param requestID RequestID (Hex)
     * @param [timeoutSeconds] The timeout in seconds, maximum 60s
     * @param [waitForL1Confirmation] Wait for the block to be confirmed on L1
     */
    public waitForRequestWithHttpInfo(requestID: string, timeoutSeconds?: number, waitForL1Confirmation?: boolean, _options?: Configuration): Observable<HttpInfo<ReceiptResponse>> {
        const requestContextPromise = this.requestFactory.waitForRequest(requestID, timeoutSeconds, waitForL1Confirmation, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.waitForRequestWithHttpInfo(rsp)));
            }));
    }

    /**
     * Wait until the given request has been processed by the node
     * @param requestID RequestID (Hex)
     * @param [timeoutSeconds] The timeout in seconds, maximum 60s
     * @param [waitForL1Confirmation] Wait for the block to be confirmed on L1
     */
    public waitForRequest(requestID: string, timeoutSeconds?: number, waitForL1Confirmation?: boolean, _options?: Configuration): Observable<ReceiptResponse> {
        return this.waitForRequestWithHttpInfo(requestID, timeoutSeconds, waitForL1Confirmation, _options).pipe(map((apiResponse: HttpInfo<ReceiptResponse>) => apiResponse.data));
    }

}

import { CorecontractsApiRequestFactory, CorecontractsApiResponseProcessor} from "../apis/CorecontractsApi";
export class ObservableCorecontractsApi {
    private requestFactory: CorecontractsApiRequestFactory;
    private responseProcessor: CorecontractsApiResponseProcessor;
    private configuration: Configuration;

    public constructor(
        configuration: Configuration,
        requestFactory?: CorecontractsApiRequestFactory,
        responseProcessor?: CorecontractsApiResponseProcessor
    ) {
        this.configuration = configuration;
        this.requestFactory = requestFactory || new CorecontractsApiRequestFactory(configuration);
        this.responseProcessor = responseProcessor || new CorecontractsApiResponseProcessor();
    }

    /**
     * Get all assets belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountBalanceWithHttpInfo(agentID: string, block?: string, _options?: Configuration): Observable<HttpInfo<AssetsResponse>> {
        const requestContextPromise = this.requestFactory.accountsGetAccountBalance(agentID, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.accountsGetAccountBalanceWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get all assets belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountBalance(agentID: string, block?: string, _options?: Configuration): Observable<AssetsResponse> {
        return this.accountsGetAccountBalanceWithHttpInfo(agentID, block, _options).pipe(map((apiResponse: HttpInfo<AssetsResponse>) => apiResponse.data));
    }

    /**
     * Get the current nonce of an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountNonceWithHttpInfo(agentID: string, block?: string, _options?: Configuration): Observable<HttpInfo<AccountNonceResponse>> {
        const requestContextPromise = this.requestFactory.accountsGetAccountNonce(agentID, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.accountsGetAccountNonceWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the current nonce of an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountNonce(agentID: string, block?: string, _options?: Configuration): Observable<AccountNonceResponse> {
        return this.accountsGetAccountNonceWithHttpInfo(agentID, block, _options).pipe(map((apiResponse: HttpInfo<AccountNonceResponse>) => apiResponse.data));
    }

    /**
     * Get all object ids belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountObjectIDsWithHttpInfo(agentID: string, block?: string, _options?: Configuration): Observable<HttpInfo<AccountObjectsResponse>> {
        const requestContextPromise = this.requestFactory.accountsGetAccountObjectIDs(agentID, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.accountsGetAccountObjectIDsWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get all object ids belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param [block] Block index or trie root
     */
    public accountsGetAccountObjectIDs(agentID: string, block?: string, _options?: Configuration): Observable<AccountObjectsResponse> {
        return this.accountsGetAccountObjectIDsWithHttpInfo(agentID, block, _options).pipe(map((apiResponse: HttpInfo<AccountObjectsResponse>) => apiResponse.data));
    }

    /**
     * Get all stored assets
     * @param [block] Block index or trie root
     */
    public accountsGetTotalAssetsWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<AssetsResponse>> {
        const requestContextPromise = this.requestFactory.accountsGetTotalAssets(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.accountsGetTotalAssetsWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get all stored assets
     * @param [block] Block index or trie root
     */
    public accountsGetTotalAssets(block?: string, _options?: Configuration): Observable<AssetsResponse> {
        return this.accountsGetTotalAssetsWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<AssetsResponse>) => apiResponse.data));
    }

    /**
     * Get the block info of a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetBlockInfoWithHttpInfo(blockIndex: number, block?: string, _options?: Configuration): Observable<HttpInfo<BlockInfoResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetBlockInfo(blockIndex, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetBlockInfoWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the block info of a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetBlockInfo(blockIndex: number, block?: string, _options?: Configuration): Observable<BlockInfoResponse> {
        return this.blocklogGetBlockInfoWithHttpInfo(blockIndex, block, _options).pipe(map((apiResponse: HttpInfo<BlockInfoResponse>) => apiResponse.data));
    }

    /**
     * Get the control addresses
     * @param [block] Block index or trie root
     */
    public blocklogGetControlAddressesWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<ControlAddressesResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetControlAddresses(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetControlAddressesWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the control addresses
     * @param [block] Block index or trie root
     */
    public blocklogGetControlAddresses(block?: string, _options?: Configuration): Observable<ControlAddressesResponse> {
        return this.blocklogGetControlAddressesWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<ControlAddressesResponse>) => apiResponse.data));
    }

    /**
     * Get events of a block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfBlockWithHttpInfo(blockIndex: number, block?: string, _options?: Configuration): Observable<HttpInfo<EventsResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetEventsOfBlock(blockIndex, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetEventsOfBlockWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get events of a block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfBlock(blockIndex: number, block?: string, _options?: Configuration): Observable<EventsResponse> {
        return this.blocklogGetEventsOfBlockWithHttpInfo(blockIndex, block, _options).pipe(map((apiResponse: HttpInfo<EventsResponse>) => apiResponse.data));
    }

    /**
     * Get events of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfLatestBlockWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<EventsResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetEventsOfLatestBlock(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetEventsOfLatestBlockWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get events of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfLatestBlock(block?: string, _options?: Configuration): Observable<EventsResponse> {
        return this.blocklogGetEventsOfLatestBlockWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<EventsResponse>) => apiResponse.data));
    }

    /**
     * Get events of a request
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfRequestWithHttpInfo(requestID: string, block?: string, _options?: Configuration): Observable<HttpInfo<EventsResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetEventsOfRequest(requestID, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetEventsOfRequestWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get events of a request
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetEventsOfRequest(requestID: string, block?: string, _options?: Configuration): Observable<EventsResponse> {
        return this.blocklogGetEventsOfRequestWithHttpInfo(requestID, block, _options).pipe(map((apiResponse: HttpInfo<EventsResponse>) => apiResponse.data));
    }

    /**
     * Get the block info of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetLatestBlockInfoWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<BlockInfoResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetLatestBlockInfo(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetLatestBlockInfoWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the block info of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetLatestBlockInfo(block?: string, _options?: Configuration): Observable<BlockInfoResponse> {
        return this.blocklogGetLatestBlockInfoWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<BlockInfoResponse>) => apiResponse.data));
    }

    /**
     * Get the request ids for a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForBlockWithHttpInfo(blockIndex: number, block?: string, _options?: Configuration): Observable<HttpInfo<RequestIDsResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetRequestIDsForBlock(blockIndex, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetRequestIDsForBlockWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the request ids for a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForBlock(blockIndex: number, block?: string, _options?: Configuration): Observable<RequestIDsResponse> {
        return this.blocklogGetRequestIDsForBlockWithHttpInfo(blockIndex, block, _options).pipe(map((apiResponse: HttpInfo<RequestIDsResponse>) => apiResponse.data));
    }

    /**
     * Get the request ids for the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForLatestBlockWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<RequestIDsResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetRequestIDsForLatestBlock(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetRequestIDsForLatestBlockWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the request ids for the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIDsForLatestBlock(block?: string, _options?: Configuration): Observable<RequestIDsResponse> {
        return this.blocklogGetRequestIDsForLatestBlockWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<RequestIDsResponse>) => apiResponse.data));
    }

    /**
     * Get the request processing status
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIsProcessedWithHttpInfo(requestID: string, block?: string, _options?: Configuration): Observable<HttpInfo<RequestProcessedResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetRequestIsProcessed(requestID, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetRequestIsProcessedWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the request processing status
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestIsProcessed(requestID: string, block?: string, _options?: Configuration): Observable<RequestProcessedResponse> {
        return this.blocklogGetRequestIsProcessedWithHttpInfo(requestID, block, _options).pipe(map((apiResponse: HttpInfo<RequestProcessedResponse>) => apiResponse.data));
    }

    /**
     * Get the receipt of a certain request id
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptWithHttpInfo(requestID: string, block?: string, _options?: Configuration): Observable<HttpInfo<ReceiptResponse>> {
        const requestContextPromise = this.requestFactory.blocklogGetRequestReceipt(requestID, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetRequestReceiptWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the receipt of a certain request id
     * @param requestID RequestID (Hex)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceipt(requestID: string, block?: string, _options?: Configuration): Observable<ReceiptResponse> {
        return this.blocklogGetRequestReceiptWithHttpInfo(requestID, block, _options).pipe(map((apiResponse: HttpInfo<ReceiptResponse>) => apiResponse.data));
    }

    /**
     * Get all receipts of a certain block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfBlockWithHttpInfo(blockIndex: number, block?: string, _options?: Configuration): Observable<HttpInfo<Array<ReceiptResponse>>> {
        const requestContextPromise = this.requestFactory.blocklogGetRequestReceiptsOfBlock(blockIndex, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetRequestReceiptsOfBlockWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get all receipts of a certain block
     * @param blockIndex BlockIndex (uint32)
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfBlock(blockIndex: number, block?: string, _options?: Configuration): Observable<Array<ReceiptResponse>> {
        return this.blocklogGetRequestReceiptsOfBlockWithHttpInfo(blockIndex, block, _options).pipe(map((apiResponse: HttpInfo<Array<ReceiptResponse>>) => apiResponse.data));
    }

    /**
     * Get all receipts of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<Array<ReceiptResponse>>> {
        const requestContextPromise = this.requestFactory.blocklogGetRequestReceiptsOfLatestBlock(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get all receipts of the latest block
     * @param [block] Block index or trie root
     */
    public blocklogGetRequestReceiptsOfLatestBlock(block?: string, _options?: Configuration): Observable<Array<ReceiptResponse>> {
        return this.blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<Array<ReceiptResponse>>) => apiResponse.data));
    }

    /**
     * Get the error message format of a specific error id
     * @param chainID ChainID (Hex Address)
     * @param contractHname Contract (Hname as Hex)
     * @param errorID Error Id (uint16)
     * @param [block] Block index or trie root
     */
    public errorsGetErrorMessageFormatWithHttpInfo(chainID: string, contractHname: string, errorID: number, block?: string, _options?: Configuration): Observable<HttpInfo<ErrorMessageFormatResponse>> {
        const requestContextPromise = this.requestFactory.errorsGetErrorMessageFormat(chainID, contractHname, errorID, block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.errorsGetErrorMessageFormatWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get the error message format of a specific error id
     * @param chainID ChainID (Hex Address)
     * @param contractHname Contract (Hname as Hex)
     * @param errorID Error Id (uint16)
     * @param [block] Block index or trie root
     */
    public errorsGetErrorMessageFormat(chainID: string, contractHname: string, errorID: number, block?: string, _options?: Configuration): Observable<ErrorMessageFormatResponse> {
        return this.errorsGetErrorMessageFormatWithHttpInfo(chainID, contractHname, errorID, block, _options).pipe(map((apiResponse: HttpInfo<ErrorMessageFormatResponse>) => apiResponse.data));
    }

    /**
     * Returns the chain admin
     * Get the chain admin
     * @param [block] Block index or trie root
     */
    public governanceGetChainAdminWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<GovChainAdminResponse>> {
        const requestContextPromise = this.requestFactory.governanceGetChainAdmin(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.governanceGetChainAdminWithHttpInfo(rsp)));
            }));
    }

    /**
     * Returns the chain admin
     * Get the chain admin
     * @param [block] Block index or trie root
     */
    public governanceGetChainAdmin(block?: string, _options?: Configuration): Observable<GovChainAdminResponse> {
        return this.governanceGetChainAdminWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<GovChainAdminResponse>) => apiResponse.data));
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chain\' to get information about the chain.
     * Get the chain info
     * @param [block] Block index or trie root
     */
    public governanceGetChainInfoWithHttpInfo(block?: string, _options?: Configuration): Observable<HttpInfo<GovChainInfoResponse>> {
        const requestContextPromise = this.requestFactory.governanceGetChainInfo(block, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.governanceGetChainInfoWithHttpInfo(rsp)));
            }));
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chain\' to get information about the chain.
     * Get the chain info
     * @param [block] Block index or trie root
     */
    public governanceGetChainInfo(block?: string, _options?: Configuration): Observable<GovChainInfoResponse> {
        return this.governanceGetChainInfoWithHttpInfo(block, _options).pipe(map((apiResponse: HttpInfo<GovChainInfoResponse>) => apiResponse.data));
    }

}

import { DefaultApiRequestFactory, DefaultApiResponseProcessor} from "../apis/DefaultApi";
export class ObservableDefaultApi {
    private requestFactory: DefaultApiRequestFactory;
    private responseProcessor: DefaultApiResponseProcessor;
    private configuration: Configuration;

    public constructor(
        configuration: Configuration,
        requestFactory?: DefaultApiRequestFactory,
        responseProcessor?: DefaultApiResponseProcessor
    ) {
        this.configuration = configuration;
        this.requestFactory = requestFactory || new DefaultApiRequestFactory(configuration);
        this.responseProcessor = responseProcessor || new DefaultApiResponseProcessor();
    }

    /**
     * Returns 200 if the node is healthy.
     */
    public getHealthWithHttpInfo(_options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.getHealth(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getHealthWithHttpInfo(rsp)));
            }));
    }

    /**
     * Returns 200 if the node is healthy.
     */
    public getHealth(_options?: Configuration): Observable<void> {
        return this.getHealthWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * The websocket connection service
     */
    public v1WsGetWithHttpInfo(_options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.v1WsGet(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.v1WsGetWithHttpInfo(rsp)));
            }));
    }

    /**
     * The websocket connection service
     */
    public v1WsGet(_options?: Configuration): Observable<void> {
        return this.v1WsGetWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

}

import { MetricsApiRequestFactory, MetricsApiResponseProcessor} from "../apis/MetricsApi";
export class ObservableMetricsApi {
    private requestFactory: MetricsApiRequestFactory;
    private responseProcessor: MetricsApiResponseProcessor;
    private configuration: Configuration;

    public constructor(
        configuration: Configuration,
        requestFactory?: MetricsApiRequestFactory,
        responseProcessor?: MetricsApiResponseProcessor
    ) {
        this.configuration = configuration;
        this.requestFactory = requestFactory || new MetricsApiRequestFactory(configuration);
        this.responseProcessor = responseProcessor || new MetricsApiResponseProcessor();
    }

    /**
     * Get chain specific message metrics.
     */
    public getChainMessageMetricsWithHttpInfo(_options?: Configuration): Observable<HttpInfo<ChainMessageMetrics>> {
        const requestContextPromise = this.requestFactory.getChainMessageMetrics(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getChainMessageMetricsWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get chain specific message metrics.
     */
    public getChainMessageMetrics(_options?: Configuration): Observable<ChainMessageMetrics> {
        return this.getChainMessageMetricsWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<ChainMessageMetrics>) => apiResponse.data));
    }

    /**
     * Get chain pipe event metrics.
     */
    public getChainPipeMetricsWithHttpInfo(_options?: Configuration): Observable<HttpInfo<ConsensusPipeMetrics>> {
        const requestContextPromise = this.requestFactory.getChainPipeMetrics(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getChainPipeMetricsWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get chain pipe event metrics.
     */
    public getChainPipeMetrics(_options?: Configuration): Observable<ConsensusPipeMetrics> {
        return this.getChainPipeMetricsWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<ConsensusPipeMetrics>) => apiResponse.data));
    }

    /**
     * Get chain workflow metrics.
     */
    public getChainWorkflowMetricsWithHttpInfo(_options?: Configuration): Observable<HttpInfo<ConsensusWorkflowMetrics>> {
        const requestContextPromise = this.requestFactory.getChainWorkflowMetrics(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getChainWorkflowMetricsWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get chain workflow metrics.
     */
    public getChainWorkflowMetrics(_options?: Configuration): Observable<ConsensusWorkflowMetrics> {
        return this.getChainWorkflowMetricsWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<ConsensusWorkflowMetrics>) => apiResponse.data));
    }

}

import { NodeApiRequestFactory, NodeApiResponseProcessor} from "../apis/NodeApi";
export class ObservableNodeApi {
    private requestFactory: NodeApiRequestFactory;
    private responseProcessor: NodeApiResponseProcessor;
    private configuration: Configuration;

    public constructor(
        configuration: Configuration,
        requestFactory?: NodeApiRequestFactory,
        responseProcessor?: NodeApiResponseProcessor
    ) {
        this.configuration = configuration;
        this.requestFactory = requestFactory || new NodeApiRequestFactory(configuration);
        this.responseProcessor = responseProcessor || new NodeApiResponseProcessor();
    }

    /**
     * Distrust a peering node
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public distrustPeerWithHttpInfo(peer: string, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.distrustPeer(peer, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.distrustPeerWithHttpInfo(rsp)));
            }));
    }

    /**
     * Distrust a peering node
     * @param peer Name or PubKey (hex) of the trusted peer
     */
    public distrustPeer(peer: string, _options?: Configuration): Observable<void> {
        return this.distrustPeerWithHttpInfo(peer, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Generate a new distributed key
     * @param dKSharesPostRequest Request parameters
     */
    public generateDKSWithHttpInfo(dKSharesPostRequest: DKSharesPostRequest, _options?: Configuration): Observable<HttpInfo<DKSharesInfo>> {
        const requestContextPromise = this.requestFactory.generateDKS(dKSharesPostRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.generateDKSWithHttpInfo(rsp)));
            }));
    }

    /**
     * Generate a new distributed key
     * @param dKSharesPostRequest Request parameters
     */
    public generateDKS(dKSharesPostRequest: DKSharesPostRequest, _options?: Configuration): Observable<DKSharesInfo> {
        return this.generateDKSWithHttpInfo(dKSharesPostRequest, _options).pipe(map((apiResponse: HttpInfo<DKSharesInfo>) => apiResponse.data));
    }

    /**
     * Get basic information about all configured peers
     */
    public getAllPeersWithHttpInfo(_options?: Configuration): Observable<HttpInfo<Array<PeeringNodeStatusResponse>>> {
        const requestContextPromise = this.requestFactory.getAllPeers(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getAllPeersWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get basic information about all configured peers
     */
    public getAllPeers(_options?: Configuration): Observable<Array<PeeringNodeStatusResponse>> {
        return this.getAllPeersWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<Array<PeeringNodeStatusResponse>>) => apiResponse.data));
    }

    /**
     * Return the Wasp configuration
     */
    public getConfigurationWithHttpInfo(_options?: Configuration): Observable<HttpInfo<{ [key: string]: string; }>> {
        const requestContextPromise = this.requestFactory.getConfiguration(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getConfigurationWithHttpInfo(rsp)));
            }));
    }

    /**
     * Return the Wasp configuration
     */
    public getConfiguration(_options?: Configuration): Observable<{ [key: string]: string; }> {
        return this.getConfigurationWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<{ [key: string]: string; }>) => apiResponse.data));
    }

    /**
     * Get information about the shared address DKS configuration
     * @param sharedAddress SharedAddress (Hex Address)
     */
    public getDKSInfoWithHttpInfo(sharedAddress: string, _options?: Configuration): Observable<HttpInfo<DKSharesInfo>> {
        const requestContextPromise = this.requestFactory.getDKSInfo(sharedAddress, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getDKSInfoWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get information about the shared address DKS configuration
     * @param sharedAddress SharedAddress (Hex Address)
     */
    public getDKSInfo(sharedAddress: string, _options?: Configuration): Observable<DKSharesInfo> {
        return this.getDKSInfoWithHttpInfo(sharedAddress, _options).pipe(map((apiResponse: HttpInfo<DKSharesInfo>) => apiResponse.data));
    }

    /**
     * Returns private information about this node.
     */
    public getInfoWithHttpInfo(_options?: Configuration): Observable<HttpInfo<InfoResponse>> {
        const requestContextPromise = this.requestFactory.getInfo(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getInfoWithHttpInfo(rsp)));
            }));
    }

    /**
     * Returns private information about this node.
     */
    public getInfo(_options?: Configuration): Observable<InfoResponse> {
        return this.getInfoWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<InfoResponse>) => apiResponse.data));
    }

    /**
     * Get basic peer info of the current node
     */
    public getPeeringIdentityWithHttpInfo(_options?: Configuration): Observable<HttpInfo<PeeringNodeIdentityResponse>> {
        const requestContextPromise = this.requestFactory.getPeeringIdentity(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getPeeringIdentityWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get basic peer info of the current node
     */
    public getPeeringIdentity(_options?: Configuration): Observable<PeeringNodeIdentityResponse> {
        return this.getPeeringIdentityWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<PeeringNodeIdentityResponse>) => apiResponse.data));
    }

    /**
     * Get trusted peers
     */
    public getTrustedPeersWithHttpInfo(_options?: Configuration): Observable<HttpInfo<Array<PeeringNodeIdentityResponse>>> {
        const requestContextPromise = this.requestFactory.getTrustedPeers(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getTrustedPeersWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get trusted peers
     */
    public getTrustedPeers(_options?: Configuration): Observable<Array<PeeringNodeIdentityResponse>> {
        return this.getTrustedPeersWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<Array<PeeringNodeIdentityResponse>>) => apiResponse.data));
    }

    /**
     * Returns the node version.
     */
    public getVersionWithHttpInfo(_options?: Configuration): Observable<HttpInfo<VersionResponse>> {
        const requestContextPromise = this.requestFactory.getVersion(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getVersionWithHttpInfo(rsp)));
            }));
    }

    /**
     * Returns the node version.
     */
    public getVersion(_options?: Configuration): Observable<VersionResponse> {
        return this.getVersionWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<VersionResponse>) => apiResponse.data));
    }

    /**
     * Gets the node owner
     */
    public ownerCertificateWithHttpInfo(_options?: Configuration): Observable<HttpInfo<NodeOwnerCertificateResponse>> {
        const requestContextPromise = this.requestFactory.ownerCertificate(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.ownerCertificateWithHttpInfo(rsp)));
            }));
    }

    /**
     * Gets the node owner
     */
    public ownerCertificate(_options?: Configuration): Observable<NodeOwnerCertificateResponse> {
        return this.ownerCertificateWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<NodeOwnerCertificateResponse>) => apiResponse.data));
    }

    /**
     * Shut down the node
     */
    public shutdownNodeWithHttpInfo(_options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.shutdownNode(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.shutdownNodeWithHttpInfo(rsp)));
            }));
    }

    /**
     * Shut down the node
     */
    public shutdownNode(_options?: Configuration): Observable<void> {
        return this.shutdownNodeWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Trust a peering node
     * @param peeringTrustRequest Info of the peer to trust
     */
    public trustPeerWithHttpInfo(peeringTrustRequest: PeeringTrustRequest, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.trustPeer(peeringTrustRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.trustPeerWithHttpInfo(rsp)));
            }));
    }

    /**
     * Trust a peering node
     * @param peeringTrustRequest Info of the peer to trust
     */
    public trustPeer(peeringTrustRequest: PeeringTrustRequest, _options?: Configuration): Observable<void> {
        return this.trustPeerWithHttpInfo(peeringTrustRequest, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

}

import { RequestsApiRequestFactory, RequestsApiResponseProcessor} from "../apis/RequestsApi";
export class ObservableRequestsApi {
    private requestFactory: RequestsApiRequestFactory;
    private responseProcessor: RequestsApiResponseProcessor;
    private configuration: Configuration;

    public constructor(
        configuration: Configuration,
        requestFactory?: RequestsApiRequestFactory,
        responseProcessor?: RequestsApiResponseProcessor
    ) {
        this.configuration = configuration;
        this.requestFactory = requestFactory || new RequestsApiRequestFactory(configuration);
        this.responseProcessor = responseProcessor || new RequestsApiResponseProcessor();
    }

    /**
     * Post an off-ledger request
     * @param offLedgerRequest Offledger request as JSON. Request encoded in Hex
     */
    public offLedgerWithHttpInfo(offLedgerRequest: OffLedgerRequest, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.offLedger(offLedgerRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.offLedgerWithHttpInfo(rsp)));
            }));
    }

    /**
     * Post an off-ledger request
     * @param offLedgerRequest Offledger request as JSON. Request encoded in Hex
     */
    public offLedger(offLedgerRequest: OffLedgerRequest, _options?: Configuration): Observable<void> {
        return this.offLedgerWithHttpInfo(offLedgerRequest, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

}

import { UsersApiRequestFactory, UsersApiResponseProcessor} from "../apis/UsersApi";
export class ObservableUsersApi {
    private requestFactory: UsersApiRequestFactory;
    private responseProcessor: UsersApiResponseProcessor;
    private configuration: Configuration;

    public constructor(
        configuration: Configuration,
        requestFactory?: UsersApiRequestFactory,
        responseProcessor?: UsersApiResponseProcessor
    ) {
        this.configuration = configuration;
        this.requestFactory = requestFactory || new UsersApiRequestFactory(configuration);
        this.responseProcessor = responseProcessor || new UsersApiResponseProcessor();
    }

    /**
     * Add a user
     * @param addUserRequest The user data
     */
    public addUserWithHttpInfo(addUserRequest: AddUserRequest, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.addUser(addUserRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.addUserWithHttpInfo(rsp)));
            }));
    }

    /**
     * Add a user
     * @param addUserRequest The user data
     */
    public addUser(addUserRequest: AddUserRequest, _options?: Configuration): Observable<void> {
        return this.addUserWithHttpInfo(addUserRequest, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Change user password
     * @param username The username
     * @param updateUserPasswordRequest The users new password
     */
    public changeUserPasswordWithHttpInfo(username: string, updateUserPasswordRequest: UpdateUserPasswordRequest, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.changeUserPassword(username, updateUserPasswordRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.changeUserPasswordWithHttpInfo(rsp)));
            }));
    }

    /**
     * Change user password
     * @param username The username
     * @param updateUserPasswordRequest The users new password
     */
    public changeUserPassword(username: string, updateUserPasswordRequest: UpdateUserPasswordRequest, _options?: Configuration): Observable<void> {
        return this.changeUserPasswordWithHttpInfo(username, updateUserPasswordRequest, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Change user permissions
     * @param username The username
     * @param updateUserPermissionsRequest The users new permissions
     */
    public changeUserPermissionsWithHttpInfo(username: string, updateUserPermissionsRequest: UpdateUserPermissionsRequest, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.changeUserPermissions(username, updateUserPermissionsRequest, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.changeUserPermissionsWithHttpInfo(rsp)));
            }));
    }

    /**
     * Change user permissions
     * @param username The username
     * @param updateUserPermissionsRequest The users new permissions
     */
    public changeUserPermissions(username: string, updateUserPermissionsRequest: UpdateUserPermissionsRequest, _options?: Configuration): Observable<void> {
        return this.changeUserPermissionsWithHttpInfo(username, updateUserPermissionsRequest, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Deletes a user
     * @param username The username
     */
    public deleteUserWithHttpInfo(username: string, _options?: Configuration): Observable<HttpInfo<void>> {
        const requestContextPromise = this.requestFactory.deleteUser(username, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.deleteUserWithHttpInfo(rsp)));
            }));
    }

    /**
     * Deletes a user
     * @param username The username
     */
    public deleteUser(username: string, _options?: Configuration): Observable<void> {
        return this.deleteUserWithHttpInfo(username, _options).pipe(map((apiResponse: HttpInfo<void>) => apiResponse.data));
    }

    /**
     * Get a user
     * @param username The username
     */
    public getUserWithHttpInfo(username: string, _options?: Configuration): Observable<HttpInfo<User>> {
        const requestContextPromise = this.requestFactory.getUser(username, _options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getUserWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get a user
     * @param username The username
     */
    public getUser(username: string, _options?: Configuration): Observable<User> {
        return this.getUserWithHttpInfo(username, _options).pipe(map((apiResponse: HttpInfo<User>) => apiResponse.data));
    }

    /**
     * Get a list of all users
     */
    public getUsersWithHttpInfo(_options?: Configuration): Observable<HttpInfo<Array<User>>> {
        const requestContextPromise = this.requestFactory.getUsers(_options);

        // build promise chain
        let middlewarePreObservable = from<RequestContext>(requestContextPromise);
        for (const middleware of this.configuration.middleware) {
            middlewarePreObservable = middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => middleware.pre(ctx)));
        }

        return middlewarePreObservable.pipe(mergeMap((ctx: RequestContext) => this.configuration.httpApi.send(ctx))).
            pipe(mergeMap((response: ResponseContext) => {
                let middlewarePostObservable = of(response);
                for (const middleware of this.configuration.middleware) {
                    middlewarePostObservable = middlewarePostObservable.pipe(mergeMap((rsp: ResponseContext) => middleware.post(rsp)));
                }
                return middlewarePostObservable.pipe(map((rsp: ResponseContext) => this.responseProcessor.getUsersWithHttpInfo(rsp)));
            }));
    }

    /**
     * Get a list of all users
     */
    public getUsers(_options?: Configuration): Observable<Array<User>> {
        return this.getUsersWithHttpInfo(_options).pipe(map((apiResponse: HttpInfo<Array<User>>) => apiResponse.data));
    }

}
