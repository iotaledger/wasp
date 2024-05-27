// TODO: better import syntax?
import {BaseAPIRequestFactory, RequiredError, COLLECTION_FORMATS} from './baseapi';
import {Configuration} from '../configuration';
import {RequestContext, HttpMethod, ResponseContext, HttpFile} from '../http/http';
import {ObjectSerializer} from '../models/ObjectSerializer';
import {ApiException} from './exception';
import {canConsumeForm, isCodeInRange} from '../util';
import {SecurityAuthentication} from '../auth/auth';


import { AccountFoundriesResponse } from '../models/AccountFoundriesResponse';
import { AccountNFTsResponse } from '../models/AccountNFTsResponse';
import { AccountNonceResponse } from '../models/AccountNonceResponse';
import { AssetsResponse } from '../models/AssetsResponse';
import { BlobInfoResponse } from '../models/BlobInfoResponse';
import { BlobValueResponse } from '../models/BlobValueResponse';
import { BlockInfoResponse } from '../models/BlockInfoResponse';
import { ControlAddressesResponse } from '../models/ControlAddressesResponse';
import { ErrorMessageFormatResponse } from '../models/ErrorMessageFormatResponse';
import { EventsResponse } from '../models/EventsResponse';
import { FoundryOutputResponse } from '../models/FoundryOutputResponse';
import { GovAllowedStateControllerAddressesResponse } from '../models/GovAllowedStateControllerAddressesResponse';
import { GovChainInfoResponse } from '../models/GovChainInfoResponse';
import { GovChainOwnerResponse } from '../models/GovChainOwnerResponse';
import { NFTJSON } from '../models/NFTJSON';
import { NativeTokenIDRegistryResponse } from '../models/NativeTokenIDRegistryResponse';
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
     * @param chainID ChainID (Bech32)
     * @param agentID AgentID (Bech32 for WasmVM | Hex for EVM)
     * @param block Block index or trie root
     */
    public async accountsGetAccountBalance(chainID: string, agentID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountBalance", "chainID");
        }


        // verify required parameter 'agentID' is not null or undefined
        if (agentID === null || agentID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountBalance", "agentID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/accounts/account/{agentID}/balance'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * Get all foundries owned by an account
     * @param chainID ChainID (Bech32)
     * @param agentID AgentID (Bech32 for WasmVM | Hex for EVM)
     * @param block Block index or trie root
     */
    public async accountsGetAccountFoundries(chainID: string, agentID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountFoundries", "chainID");
        }


        // verify required parameter 'agentID' is not null or undefined
        if (agentID === null || agentID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountFoundries", "agentID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/accounts/account/{agentID}/foundries'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * Get all NFT ids belonging to an account
     * @param chainID ChainID (Bech32)
     * @param agentID AgentID (Bech32 for WasmVM | Hex for EVM)
     * @param block Block index or trie root
     */
    public async accountsGetAccountNFTIDs(chainID: string, agentID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountNFTIDs", "chainID");
        }


        // verify required parameter 'agentID' is not null or undefined
        if (agentID === null || agentID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountNFTIDs", "agentID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/accounts/account/{agentID}/nfts'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * @param chainID ChainID (Bech32)
     * @param agentID AgentID (Bech32 for WasmVM | Hex for EVM)
     * @param block Block index or trie root
     */
    public async accountsGetAccountNonce(chainID: string, agentID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountNonce", "chainID");
        }


        // verify required parameter 'agentID' is not null or undefined
        if (agentID === null || agentID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetAccountNonce", "agentID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/accounts/account/{agentID}/nonce'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * Get the foundry output
     * @param chainID ChainID (Bech32)
     * @param serialNumber Serial Number (uint32)
     * @param block Block index or trie root
     */
    public async accountsGetFoundryOutput(chainID: string, serialNumber: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetFoundryOutput", "chainID");
        }


        // verify required parameter 'serialNumber' is not null or undefined
        if (serialNumber === null || serialNumber === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetFoundryOutput", "serialNumber");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/accounts/foundry_output/{serialNumber}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
            .replace('{' + 'serialNumber' + '}', encodeURIComponent(String(serialNumber)));

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
     * Get the NFT data by an ID
     * @param chainID ChainID (Bech32)
     * @param nftID NFT ID (Hex)
     * @param block Block index or trie root
     */
    public async accountsGetNFTData(chainID: string, nftID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetNFTData", "chainID");
        }


        // verify required parameter 'nftID' is not null or undefined
        if (nftID === null || nftID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetNFTData", "nftID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/accounts/nftdata/{nftID}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
            .replace('{' + 'nftID' + '}', encodeURIComponent(String(nftID)));

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
     * Get a list of all registries
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async accountsGetNativeTokenIDRegistry(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetNativeTokenIDRegistry", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/accounts/token_registry'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async accountsGetTotalAssets(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "accountsGetTotalAssets", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/accounts/total_assets'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * Get all fields of a blob
     * @param chainID ChainID (Bech32)
     * @param blobHash BlobHash (Hex)
     * @param block Block index or trie root
     */
    public async blobsGetBlobInfo(chainID: string, blobHash: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blobsGetBlobInfo", "chainID");
        }


        // verify required parameter 'blobHash' is not null or undefined
        if (blobHash === null || blobHash === undefined) {
            throw new RequiredError("CorecontractsApi", "blobsGetBlobInfo", "blobHash");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blobs/{blobHash}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
            .replace('{' + 'blobHash' + '}', encodeURIComponent(String(blobHash)));

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
     * Get the value of the supplied field (key)
     * @param chainID ChainID (Bech32)
     * @param blobHash BlobHash (Hex)
     * @param fieldKey FieldKey (String)
     * @param block Block index or trie root
     */
    public async blobsGetBlobValue(chainID: string, blobHash: string, fieldKey: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blobsGetBlobValue", "chainID");
        }


        // verify required parameter 'blobHash' is not null or undefined
        if (blobHash === null || blobHash === undefined) {
            throw new RequiredError("CorecontractsApi", "blobsGetBlobValue", "blobHash");
        }


        // verify required parameter 'fieldKey' is not null or undefined
        if (fieldKey === null || fieldKey === undefined) {
            throw new RequiredError("CorecontractsApi", "blobsGetBlobValue", "fieldKey");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blobs/{blobHash}/data/{fieldKey}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
            .replace('{' + 'blobHash' + '}', encodeURIComponent(String(blobHash)))
            .replace('{' + 'fieldKey' + '}', encodeURIComponent(String(fieldKey)));

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
     * @param chainID ChainID (Bech32)
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public async blocklogGetBlockInfo(chainID: string, blockIndex: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetBlockInfo", "chainID");
        }


        // verify required parameter 'blockIndex' is not null or undefined
        if (blockIndex === null || blockIndex === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetBlockInfo", "blockIndex");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/blocks/{blockIndex}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async blocklogGetControlAddresses(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetControlAddresses", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/controladdresses'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * @param chainID ChainID (Bech32)
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public async blocklogGetEventsOfBlock(chainID: string, blockIndex: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetEventsOfBlock", "chainID");
        }


        // verify required parameter 'blockIndex' is not null or undefined
        if (blockIndex === null || blockIndex === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetEventsOfBlock", "blockIndex");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/events/block/{blockIndex}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async blocklogGetEventsOfLatestBlock(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetEventsOfLatestBlock", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/events/block/latest'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * @param chainID ChainID (Bech32)
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public async blocklogGetEventsOfRequest(chainID: string, requestID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetEventsOfRequest", "chainID");
        }


        // verify required parameter 'requestID' is not null or undefined
        if (requestID === null || requestID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetEventsOfRequest", "requestID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/events/request/{requestID}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async blocklogGetLatestBlockInfo(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetLatestBlockInfo", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/blocks/latest'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * @param chainID ChainID (Bech32)
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestIDsForBlock(chainID: string, blockIndex: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestIDsForBlock", "chainID");
        }


        // verify required parameter 'blockIndex' is not null or undefined
        if (blockIndex === null || blockIndex === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestIDsForBlock", "blockIndex");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/blocks/{blockIndex}/requestids'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestIDsForLatestBlock(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestIDsForLatestBlock", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/blocks/latest/requestids'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * @param chainID ChainID (Bech32)
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestIsProcessed(chainID: string, requestID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestIsProcessed", "chainID");
        }


        // verify required parameter 'requestID' is not null or undefined
        if (requestID === null || requestID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestIsProcessed", "requestID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/requests/{requestID}/is_processed'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * @param chainID ChainID (Bech32)
     * @param requestID RequestID (Hex)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestReceipt(chainID: string, requestID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestReceipt", "chainID");
        }


        // verify required parameter 'requestID' is not null or undefined
        if (requestID === null || requestID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestReceipt", "requestID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/requests/{requestID}'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * @param chainID ChainID (Bech32)
     * @param blockIndex BlockIndex (uint32)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestReceiptsOfBlock(chainID: string, blockIndex: number, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestReceiptsOfBlock", "chainID");
        }


        // verify required parameter 'blockIndex' is not null or undefined
        if (blockIndex === null || blockIndex === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestReceiptsOfBlock", "blockIndex");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/blocks/{blockIndex}/receipts'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)))
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
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async blocklogGetRequestReceiptsOfLatestBlock(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "blocklogGetRequestReceiptsOfLatestBlock", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/blocklog/blocks/latest/receipts'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * @param chainID ChainID (Bech32)
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
        const localVarPath = '/v1/chains/{chainID}/core/errors/{contractHname}/message/{errorID}'
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
     * Returns the allowed state controller addresses
     * Get the allowed state controller addresses
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async governanceGetAllowedStateControllerAddresses(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "governanceGetAllowedStateControllerAddresses", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/governance/allowedstatecontrollers'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * If you are using the common API functions, you most likely rather want to use '/v1/chains/:chainID' to get information about a chain.
     * Get the chain info
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async governanceGetChainInfo(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "governanceGetChainInfo", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/governance/chaininfo'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     * Returns the chain owner
     * Get the chain owner
     * @param chainID ChainID (Bech32)
     * @param block Block index or trie root
     */
    public async governanceGetChainOwner(chainID: string, block?: string, _options?: Configuration): Promise<RequestContext> {
        let _config = _options || this.configuration;

        // verify required parameter 'chainID' is not null or undefined
        if (chainID === null || chainID === undefined) {
            throw new RequiredError("CorecontractsApi", "governanceGetChainOwner", "chainID");
        }



        // Path Params
        const localVarPath = '/v1/chains/{chainID}/core/governance/chainowner'
            .replace('{' + 'chainID' + '}', encodeURIComponent(String(chainID)));

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
     public async accountsGetAccountBalance(response: ResponseContext): Promise<AssetsResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: AssetsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AssetsResponse", ""
            ) as AssetsResponse;
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
            const body: AssetsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AssetsResponse", ""
            ) as AssetsResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to accountsGetAccountFoundries
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async accountsGetAccountFoundries(response: ResponseContext): Promise<AccountFoundriesResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: AccountFoundriesResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AccountFoundriesResponse", ""
            ) as AccountFoundriesResponse;
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
            const body: AccountFoundriesResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AccountFoundriesResponse", ""
            ) as AccountFoundriesResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to accountsGetAccountNFTIDs
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async accountsGetAccountNFTIDs(response: ResponseContext): Promise<AccountNFTsResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: AccountNFTsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AccountNFTsResponse", ""
            ) as AccountNFTsResponse;
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
            const body: AccountNFTsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AccountNFTsResponse", ""
            ) as AccountNFTsResponse;
            return body;
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
     public async accountsGetAccountNonce(response: ResponseContext): Promise<AccountNonceResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: AccountNonceResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AccountNonceResponse", ""
            ) as AccountNonceResponse;
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
            const body: AccountNonceResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AccountNonceResponse", ""
            ) as AccountNonceResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to accountsGetFoundryOutput
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async accountsGetFoundryOutput(response: ResponseContext): Promise<FoundryOutputResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: FoundryOutputResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "FoundryOutputResponse", ""
            ) as FoundryOutputResponse;
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
            const body: FoundryOutputResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "FoundryOutputResponse", ""
            ) as FoundryOutputResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to accountsGetNFTData
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async accountsGetNFTData(response: ResponseContext): Promise<NFTJSON > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: NFTJSON = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "NFTJSON", ""
            ) as NFTJSON;
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
            const body: NFTJSON = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "NFTJSON", ""
            ) as NFTJSON;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to accountsGetNativeTokenIDRegistry
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async accountsGetNativeTokenIDRegistry(response: ResponseContext): Promise<NativeTokenIDRegistryResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: NativeTokenIDRegistryResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "NativeTokenIDRegistryResponse", ""
            ) as NativeTokenIDRegistryResponse;
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
            const body: NativeTokenIDRegistryResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "NativeTokenIDRegistryResponse", ""
            ) as NativeTokenIDRegistryResponse;
            return body;
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
     public async accountsGetTotalAssets(response: ResponseContext): Promise<AssetsResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: AssetsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AssetsResponse", ""
            ) as AssetsResponse;
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
            const body: AssetsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "AssetsResponse", ""
            ) as AssetsResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blobsGetBlobInfo
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blobsGetBlobInfo(response: ResponseContext): Promise<BlobInfoResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: BlobInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlobInfoResponse", ""
            ) as BlobInfoResponse;
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
            const body: BlobInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlobInfoResponse", ""
            ) as BlobInfoResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to blobsGetBlobValue
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async blobsGetBlobValue(response: ResponseContext): Promise<BlobValueResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: BlobValueResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlobValueResponse", ""
            ) as BlobValueResponse;
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
            const body: BlobValueResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlobValueResponse", ""
            ) as BlobValueResponse;
            return body;
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
     public async blocklogGetBlockInfo(response: ResponseContext): Promise<BlockInfoResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: BlockInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlockInfoResponse", ""
            ) as BlockInfoResponse;
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
            const body: BlockInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlockInfoResponse", ""
            ) as BlockInfoResponse;
            return body;
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
     public async blocklogGetControlAddresses(response: ResponseContext): Promise<ControlAddressesResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ControlAddressesResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ControlAddressesResponse", ""
            ) as ControlAddressesResponse;
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
            const body: ControlAddressesResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ControlAddressesResponse", ""
            ) as ControlAddressesResponse;
            return body;
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
     public async blocklogGetEventsOfBlock(response: ResponseContext): Promise<EventsResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
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
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return body;
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
     public async blocklogGetEventsOfLatestBlock(response: ResponseContext): Promise<EventsResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
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
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return body;
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
     public async blocklogGetEventsOfRequest(response: ResponseContext): Promise<EventsResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
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
            const body: EventsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "EventsResponse", ""
            ) as EventsResponse;
            return body;
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
     public async blocklogGetLatestBlockInfo(response: ResponseContext): Promise<BlockInfoResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: BlockInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlockInfoResponse", ""
            ) as BlockInfoResponse;
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
            const body: BlockInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "BlockInfoResponse", ""
            ) as BlockInfoResponse;
            return body;
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
     public async blocklogGetRequestIDsForBlock(response: ResponseContext): Promise<RequestIDsResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: RequestIDsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestIDsResponse", ""
            ) as RequestIDsResponse;
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
            const body: RequestIDsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestIDsResponse", ""
            ) as RequestIDsResponse;
            return body;
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
     public async blocklogGetRequestIDsForLatestBlock(response: ResponseContext): Promise<RequestIDsResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: RequestIDsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestIDsResponse", ""
            ) as RequestIDsResponse;
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
            const body: RequestIDsResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestIDsResponse", ""
            ) as RequestIDsResponse;
            return body;
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
     public async blocklogGetRequestIsProcessed(response: ResponseContext): Promise<RequestProcessedResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: RequestProcessedResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestProcessedResponse", ""
            ) as RequestProcessedResponse;
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
            const body: RequestProcessedResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "RequestProcessedResponse", ""
            ) as RequestProcessedResponse;
            return body;
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
     public async blocklogGetRequestReceipt(response: ResponseContext): Promise<ReceiptResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ReceiptResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ReceiptResponse", ""
            ) as ReceiptResponse;
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
            const body: ReceiptResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ReceiptResponse", ""
            ) as ReceiptResponse;
            return body;
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
     public async blocklogGetRequestReceiptsOfBlock(response: ResponseContext): Promise<Array<ReceiptResponse> > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: Array<ReceiptResponse> = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "Array<ReceiptResponse>", ""
            ) as Array<ReceiptResponse>;
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
            const body: Array<ReceiptResponse> = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "Array<ReceiptResponse>", ""
            ) as Array<ReceiptResponse>;
            return body;
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
     public async blocklogGetRequestReceiptsOfLatestBlock(response: ResponseContext): Promise<Array<ReceiptResponse> > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: Array<ReceiptResponse> = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "Array<ReceiptResponse>", ""
            ) as Array<ReceiptResponse>;
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
            const body: Array<ReceiptResponse> = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "Array<ReceiptResponse>", ""
            ) as Array<ReceiptResponse>;
            return body;
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
     public async errorsGetErrorMessageFormat(response: ResponseContext): Promise<ErrorMessageFormatResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: ErrorMessageFormatResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ErrorMessageFormatResponse", ""
            ) as ErrorMessageFormatResponse;
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
            const body: ErrorMessageFormatResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "ErrorMessageFormatResponse", ""
            ) as ErrorMessageFormatResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to governanceGetAllowedStateControllerAddresses
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async governanceGetAllowedStateControllerAddresses(response: ResponseContext): Promise<GovAllowedStateControllerAddressesResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: GovAllowedStateControllerAddressesResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovAllowedStateControllerAddressesResponse", ""
            ) as GovAllowedStateControllerAddressesResponse;
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
            const body: GovAllowedStateControllerAddressesResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovAllowedStateControllerAddressesResponse", ""
            ) as GovAllowedStateControllerAddressesResponse;
            return body;
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
     public async governanceGetChainInfo(response: ResponseContext): Promise<GovChainInfoResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: GovChainInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovChainInfoResponse", ""
            ) as GovChainInfoResponse;
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
            const body: GovChainInfoResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovChainInfoResponse", ""
            ) as GovChainInfoResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

    /**
     * Unwraps the actual response sent by the server from the response context and deserializes the response content
     * to the expected objects
     *
     * @params response Response returned by the server for a request to governanceGetChainOwner
     * @throws ApiException if the response code was not in [200, 299]
     */
     public async governanceGetChainOwner(response: ResponseContext): Promise<GovChainOwnerResponse > {
        const contentType = ObjectSerializer.normalizeMediaType(response.headers["content-type"]);
        if (isCodeInRange("200", response.httpStatusCode)) {
            const body: GovChainOwnerResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovChainOwnerResponse", ""
            ) as GovChainOwnerResponse;
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
            const body: GovChainOwnerResponse = ObjectSerializer.deserialize(
                ObjectSerializer.parse(await response.body.text(), contentType),
                "GovChainOwnerResponse", ""
            ) as GovChainOwnerResponse;
            return body;
        }

        throw new ApiException<string | Blob | undefined>(response.httpStatusCode, "Unknown API Status Code!", await response.getBodyAsAny(), response.headers);
    }

}
