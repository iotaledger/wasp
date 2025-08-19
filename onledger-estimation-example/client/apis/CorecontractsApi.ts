// TODO: better import syntax?
import {BaseAPIRequestFactory, RequiredError, COLLECTION_FORMATS} from './baseapi';
import {Configuration} from '../configuration';
import {RequestContext, HttpMethod, ResponseContext, HttpFile, HttpInfo} from '../http/http';
import {ObjectSerializer} from '../models/ObjectSerializer';
import {ApiException} from './exception';
import {canConsumeForm, isCodeInRange} from '../util';
import {SecurityAuthentication} from '../auth/auth';


import { AccountNonceResponse } from '../models/AccountNonceResponse';
import { AssetsResponse } from '../models/AssetsResponse';
import { BlockInfoResponse } from '../models/BlockInfoResponse';
import { ControlAddressesResponse } from '../models/ControlAddressesResponse';
import { ErrorMessageFormatResponse } from '../models/ErrorMessageFormatResponse';
import { EventsResponse } from '../models/EventsResponse';
import { GovChainAdminResponse } from '../models/GovChainAdminResponse';
import { GovChainInfoResponse } from '../models/GovChainInfoResponse';
import { ReceiptResponse } from '../models/ReceiptResponse';
import { RequestIDsResponse } from '../models/RequestIDsResponse';
import { RequestProcessedResponse } from '../models/RequestProcessedResponse';
import { ValidationError } from '../models/ValidationError';

/**
 * no description
 */
export class CorecontractsApiRequestFactory extends BaseAPIRequestFactory {

    /**
     * Get all assets belonging to an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param block Block index or trie root
     */
    public async accountsGetAccountBalance(agentID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'agentID' is not null or undefined
        if (agentID === null || agentID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountBalance", "agentID");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/accounts/account/{agentID}/balance'
            .replace('{' + 'agentID' + '}', encodeURIComponent(String(agentID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the current nonce of an account
     * @param agentID AgentID (Hex Address for L1 accounts | Hex for EVM)
     * @param block Block index or trie root
     */
    public async accountsGetAccountNonce(agentID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'agentID' is not null or undefined
        if (agentID === null || agentID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountNonce", "agentID");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/accounts/account/{agentID}/nonce'
            .replace('{' + 'agentID' + '}', encodeURIComponent(String(agentID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get all stored assets
     * @param block Block index or trie root
     */
    public async accountsGetTotalAssets(block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;


        // Path Params
        const localVarPath = '/v1/chain/core/accounts/total_assets';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the block info of a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public async blocklogGetBlockInfo(blockIndex: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'blockIndex' is not null or undefined
        if (blockIndex === null || blockIndex === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetBlockInfo", "blockIndex");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/blocks/{blockIndex}'
            .replace('{' + 'blockIndex' + '}', encodeURIComponent(String(blockIndex)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the control addresses
     * @param block Block index or trie root
     */
    public async blocklogGetControlAddresses(block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;


        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/controladdresses';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get events of a block
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public async blocklogGetEventsOfBlock(blockIndex: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'blockIndex' is not null or undefined
        if (blockIndex === null || blockIndex === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetEventsOfBlock", "blockIndex");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/events/block/{blockIndex}'
            .replace('{' + 'blockIndex' + '}', encodeURIComponent(String(blockIndex)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get events of the latest block
     * @param block Block index or trie root
     */
    public async blocklogGetEventsOfLatestBlock(block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;


        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/events/block/latest';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get events of a request
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public async blocklogGetEventsOfRequest(requestID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'requestID' is not null or undefined
        if (requestID === null || requestID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetEventsOfRequest", "requestID");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/events/request/{requestID}'
            .replace('{' + 'requestID' + '}', encodeURIComponent(String(requestID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the block info of the latest block
     * @param block Block index or trie root
     */
    public async blocklogGetLatestBlockInfo(block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;


        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/blocks/latest';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the request ids for a certain block index
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestIDsForBlock(blockIndex: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'blockIndex' is not null or undefined
        if (blockIndex === null || blockIndex === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestIDsForBlock", "blockIndex");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/blocks/{blockIndex}/requestids'
            .replace('{' + 'blockIndex' + '}', encodeURIComponent(String(blockIndex)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the request ids for the latest block
     * @param block Block index or trie root
     */
    public async blocklogGetRequestIDsForLatestBlock(block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;


        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/blocks/latest/requestids';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the request processing status
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestIsProcessed(requestID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'requestID' is not null or undefined
        if (requestID === null || requestID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestIsProcessed", "requestID");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/requests/{requestID}/is_processed'
            .replace('{' + 'requestID' + '}', encodeURIComponent(String(requestID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the receipt of a certain request id
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestReceipt(requestID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'requestID' is not null or undefined
        if (requestID === null || requestID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestReceipt", "requestID");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/requests/{requestID}'
            .replace('{' + 'requestID' + '}', encodeURIComponent(String(requestID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get all receipts of a certain block
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestReceiptsOfBlock(blockIndex: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'blockIndex' is not null or undefined
        if (blockIndex === null || blockIndex === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestReceiptsOfBlock", "blockIndex");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/blocks/{blockIndex}/receipts'
            .replace('{' + 'blockIndex' + '}', encodeURIComponent(String(blockIndex)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get all receipts of the latest block
     * @param block Block index or trie root
     */
    public async blocklogGetRequestReceiptsOfLatestBlock(block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;


        // Path Params
        const localVarPath = '/v1/chain/core/blocklog/blocks/latest/receipts';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get the error message format of a specific error id
     * @param chainID ChainID (Hex Address)
     * @param contractHname Contract (Hname as Hex)
     * @param errorID Error Id (uint16)
     * @param block Block index or trie root
     */
    public async errorsGetErrorMessageFormat(chainID: string, contractHname: string, errorID: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "errorsGetErrorMessageFormat", "chainID");
        }


        // verify required parameter 'contractHname' is not null or undefined
        if (contractHname === null || contractHname === undefined) {
            throw new RequiredError("CorecontractsApi", "errorsGetErrorMessageFormat", "contractHname");
        }


        // verify required parameter 'errorID' is not null or undefined
        if (errorID === null || errorID === undefined) {
            throw new RequiredError("CorecontractsApi", "errorsGetErrorMessageFormat", "errorID");
        }



        // Path Params
        const localVarPath = '/v1/chain/core/errors/{contractHname}/message/{errorID}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
            .replace('{' + 'contractHname' + '}', encodeURIComponent(String(contractHname)))
            .replace('{' + 'errorID' + '}', encodeURIComponent(String(errorID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Returns the chain admin
     * Get the chain admin
     * @param block Block index or trie root
     */
    public async governanceGetChainAdmin(block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;


        // Path Params
        const localVarPath = '/v1/chain/core/governance/chainadmin';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * If you are using the common API functions, you most likely rather want to use \'/v1/chains/:chainID\' to get information about a chain.
     * Get the chain info
     * @param block Block index or trie root
     */
    public async governanceGetChainInfo(block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;


        // Path Params
        const localVarPath = '/v1/chain/core/governance/chaininfo';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")

        // Query Params
        if (block !== undefined) {
            requestContext.setQueryParam("block", ObjectSerializer.serialize(block, "string", "string"));
        }


        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

}

export class CorecontractsApiResponseProcessor {

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to accountsGetAccountBalance
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async accountsGetAccountBalanceWithHttpInfo(response: ResponseContext): Promise<HttpInfo<AssetsResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: AssetsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AssetsResponse", ""
            ) as AssetsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: AssetsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AssetsResponse", ""
            ) as AssetsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to accountsGetAccountNonce
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async accountsGetAccountNonceWithHttpInfo(response: ResponseContext): Promise<HttpInfo<AccountNonceResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: AccountNonceResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AccountNonceResponse", ""
            ) as AccountNonceResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: AccountNonceResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AccountNonceResponse", ""
            ) as AccountNonceResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to accountsGetTotalAssets
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async accountsGetTotalAssetsWithHttpInfo(response: ResponseContext): Promise<HttpInfo<AssetsResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: AssetsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AssetsResponse", ""
            ) as AssetsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: AssetsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AssetsResponse", ""
            ) as AssetsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetBlockInfo
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetBlockInfoWithHttpInfo(response: ResponseContext): Promise<HttpInfo<BlockInfoResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: BlockInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlockInfoResponse", ""
            ) as BlockInfoResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: BlockInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlockInfoResponse", ""
            ) as BlockInfoResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetControlAddresses
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetControlAddressesWithHttpInfo(response: ResponseContext): Promise<HttpInfo<ControlAddressesResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ControlAddressesResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ControlAddressesResponse", ""
            ) as ControlAddressesResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: ControlAddressesResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ControlAddressesResponse", ""
            ) as ControlAddressesResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetEventsOfBlock
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetEventsOfBlockWithHttpInfo(response: ResponseContext): Promise<HttpInfo<EventsResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetEventsOfLatestBlock
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetEventsOfLatestBlockWithHttpInfo(response: ResponseContext): Promise<HttpInfo<EventsResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetEventsOfRequest
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetEventsOfRequestWithHttpInfo(response: ResponseContext): Promise<HttpInfo<EventsResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetLatestBlockInfo
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetLatestBlockInfoWithHttpInfo(response: ResponseContext): Promise<HttpInfo<BlockInfoResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: BlockInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlockInfoResponse", ""
            ) as BlockInfoResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: BlockInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlockInfoResponse", ""
            ) as BlockInfoResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetRequestIDsForBlock
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetRequestIDsForBlockWithHttpInfo(response: ResponseContext): Promise<HttpInfo<RequestIDsResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: RequestIDsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestIDsResponse", ""
            ) as RequestIDsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: RequestIDsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestIDsResponse", ""
            ) as RequestIDsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetRequestIDsForLatestBlock
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetRequestIDsForLatestBlockWithHttpInfo(response: ResponseContext): Promise<HttpInfo<RequestIDsResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: RequestIDsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestIDsResponse", ""
            ) as RequestIDsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: RequestIDsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestIDsResponse", ""
            ) as RequestIDsResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetRequestIsProcessed
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetRequestIsProcessedWithHttpInfo(response: ResponseContext): Promise<HttpInfo<RequestProcessedResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: RequestProcessedResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestProcessedResponse", ""
            ) as RequestProcessedResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: RequestProcessedResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestProcessedResponse", ""
            ) as RequestProcessedResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetRequestReceipt
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetRequestReceiptWithHttpInfo(response: ResponseContext): Promise<HttpInfo<ReceiptResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ReceiptResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ReceiptResponse", ""
            ) as ReceiptResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: ReceiptResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ReceiptResponse", ""
            ) as ReceiptResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetRequestReceiptsOfBlock
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetRequestReceiptsOfBlockWithHttpInfo(response: ResponseContext): Promise<HttpInfo<Array<ReceiptResponse> >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: Array<ReceiptResponse> = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "Array<ReceiptResponse>", ""
            ) as Array<ReceiptResponse>;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: Array<ReceiptResponse> = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "Array<ReceiptResponse>", ""
            ) as Array<ReceiptResponse>;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blocklogGetRequestReceiptsOfLatestBlock
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blocklogGetRequestReceiptsOfLatestBlockWithHttpInfo(response: ResponseContext): Promise<HttpInfo<Array<ReceiptResponse> >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: Array<ReceiptResponse> = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "Array<ReceiptResponse>", ""
            ) as Array<ReceiptResponse>;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: Array<ReceiptResponse> = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "Array<ReceiptResponse>", ""
            ) as Array<ReceiptResponse>;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to errorsGetErrorMessageFormat
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async errorsGetErrorMessageFormatWithHttpInfo(response: ResponseContext): Promise<HttpInfo<ErrorMessageFormatResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ErrorMessageFormatResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ErrorMessageFormatResponse", ""
            ) as ErrorMessageFormatResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: ErrorMessageFormatResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ErrorMessageFormatResponse", ""
            ) as ErrorMessageFormatResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to governanceGetChainAdmin
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async governanceGetChainAdminWithHttpInfo(response: ResponseContext): Promise<HttpInfo<GovChainAdminResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: GovChainAdminResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovChainAdminResponse", ""
            ) as GovChainAdminResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: GovChainAdminResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovChainAdminResponse", ""
            ) as GovChainAdminResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to governanceGetChainInfo
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async governanceGetChainInfoWithHttpInfo(response: ResponseContext): Promise<HttpInfo<GovChainInfoResponse >> {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: GovChainInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovChainInfoResponse", ""
            ) as GovChainInfoResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: GovChainInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovChainInfoResponse", ""
            ) as GovChainInfoResponse;
            return new HttpInfo(response.httpStatusCode, response.headers, response.body, body);
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

}
