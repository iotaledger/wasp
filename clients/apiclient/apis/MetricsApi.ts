// TODO: better import syntax?
import {BaseAPIRequestFactory, RequiredError, COLLECTION_FORMATS} from './baseapi';
import {Configuration} from '../configuration';
import {RequestContext, HttpMethod, ResponseContext, HttpFile} from '../http/http';
import {ObjectSerializer} from '../models/ObjectSerializer';
import {ApiException} from './exception';
import {canConsumeForm, isCodeInRange} from '../util';
import {SecurityAuthentication} from '../auth/auth';


import { ChainMessageMetrics } from '../models/ChainMessageMetrics';
import { ConsensusPipeMetrics } from '../models/ConsensusPipeMetrics';
import { ConsensusWorkflowMetrics } from '../models/ConsensusWorkflowMetrics';
import { NodeMessageMetrics } from '../models/NodeMessageMetrics';
import { ValidationError } from '../models/ValidationError';

/**
 * no description
 */
export class MetricsApiRequestFactory extends BaseAPIRequestFactory {

    /**
     * Get chain specific message metrics.
     * @param chainID ChainID (Bech32)
     */
    public async getChainMessageMetrics(chainID: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("MetricsApi", "getChainMessageMetrics", "chainID");
        }


        // Path Params
        const localVarPath = '/v1/metrics/chain/{chainID}/messages'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")


        let authMethod: SecurityAuthentication | undefined;
        // Apply auth methods
        authMethod = _config.authMethods["Authorization"]
        if (authMethod?.applySecurityAuthentication) {
            await authMethod?.applySecurityAuthentication(requestContext);
        }
        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get chain pipe event metrics.
     * @param chainID ChainID (Bech32)
     */
    public async getChainPipeMetrics(chainID: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("MetricsApi", "getChainPipeMetrics", "chainID");
        }


        // Path Params
        const localVarPath = '/v1/metrics/chain/{chainID}/pipe'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")


        let authMethod: SecurityAuthentication | undefined;
        // Apply auth methods
        authMethod = _config.authMethods["Authorization"]
        if (authMethod?.applySecurityAuthentication) {
            await authMethod?.applySecurityAuthentication(requestContext);
        }
        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get chain workflow metrics.
     * @param chainID ChainID (Bech32)
     */
    public async getChainWorkflowMetrics(chainID: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("MetricsApi", "getChainWorkflowMetrics", "chainID");
        }


        // Path Params
        const localVarPath = '/v1/metrics/chain/{chainID}/workflow'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")


        let authMethod: SecurityAuthentication | undefined;
        // Apply auth methods
        authMethod = _config.authMethods["Authorization"]
        if (authMethod?.applySecurityAuthentication) {
            await authMethod?.applySecurityAuthentication(requestContext);
        }
        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

    /**
     * Get accumulated message metrics.
     */
    public async getNodeMessageMetrics(_options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // Path Params
        const localVarPath = '/v1/metrics/node/messages';

        // Make Request Context
        const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET);
        requestContext.setHeaderParam("Accept", "application/json, */*;q=0.8")


        let authMethod: SecurityAuthentication | undefined;
        // Apply auth methods
        authMethod = _config.authMethods["Authorization"]
        if (authMethod?.applySecurityAuthentication) {
            await authMethod?.applySecurityAuthentication(requestContext);
        }
        
        const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
        if (defaultAuth?.applySecurityAuthentication) {
            await defaultAuth?.applySecurityAuthentication(requestContext);
        }

        return requestContext;
    }

}

export class MetricsApiResponseProcessor {

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to getChainMessageMetrics
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async getChainMessageMetrics(response: ResponseContext): Promise<ChainMessageMetrics > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ChainMessageMetrics = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ChainMessageMetrics", ""
            ) as ChainMessageMetrics;
            return body;
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }
        if (isCodeInRange("404", response.httpStatusCode)) {
            throw new ApiException<undefined>(response.httpStatusCode, "Chain not found", undefined, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: ChainMessageMetrics = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ChainMessageMetrics", ""
            ) as ChainMessageMetrics;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to getChainPipeMetrics
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async getChainPipeMetrics(response: ResponseContext): Promise<ConsensusPipeMetrics > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ConsensusPipeMetrics = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ConsensusPipeMetrics", ""
            ) as ConsensusPipeMetrics;
            return body;
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }
        if (isCodeInRange("404", response.httpStatusCode)) {
            throw new ApiException<undefined>(response.httpStatusCode, "Chain not found", undefined, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: ConsensusPipeMetrics = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ConsensusPipeMetrics", ""
            ) as ConsensusPipeMetrics;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to getChainWorkflowMetrics
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async getChainWorkflowMetrics(response: ResponseContext): Promise<ConsensusWorkflowMetrics > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ConsensusWorkflowMetrics = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ConsensusWorkflowMetrics", ""
            ) as ConsensusWorkflowMetrics;
            return body;
        }
        if (isCodeInRange("401", response.httpStatusCode)) {
            const body: ValidationError = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ValidationError", ""
            ) as ValidationError;
            throw new ApiException<ValidationError>(response.httpStatusCode, "Unauthorized (Wrong permissions, missing token)", body, response.headers);
        }
        if (isCodeInRange("404", response.httpStatusCode)) {
            throw new ApiException<undefined>(response.httpStatusCode, "Chain not found", undefined, response.headers);
        }

        // Work around for missing responses in specification, e.g. for petstore.yaml
        if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
            const body: ConsensusWorkflowMetrics = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ConsensusWorkflowMetrics", ""
            ) as ConsensusWorkflowMetrics;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to getNodeMessageMetrics
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async getNodeMessageMetrics(response: ResponseContext): Promise<NodeMessageMetrics > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: NodeMessageMetrics = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "NodeMessageMetrics", ""
            ) as NodeMessageMetrics;
            return body;
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
            const body: NodeMessageMetrics = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "NodeMessageMetrics", ""
            ) as NodeMessageMetrics;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

}
