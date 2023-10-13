/**
 * Wasp API
 * REST API for the Wasp node
 *
 * OpenAPI spec version: 0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { FeePolicy } from '../models/FeePolicy';
import { Limits } from '../models/Limits';
import { PublicChainMetadata } from '../models/PublicChainMetadata';
import { HttpFile } from '../http/http';

export class ChainInfoResponse {
    /**
    * ChainID (Bech32-encoded)
    */
    'chainID': string;
    /**
    * The chain owner address (Bech32-encoded)
    */
    'chainOwnerId': string;
    /**
    * The EVM chain ID
    */
    'evmChainId': number;
    'gasFeePolicy': FeePolicy;
    'gasLimits': Limits;
    /**
    * Whether or not the chain is active
    */
    'isActive': boolean;
    'metadata': PublicChainMetadata;
    /**
    * The fully qualified public url leading to the chains metadata
    */
    'publicURL': string;

    static readonly discriminator: string | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "chainID",
            "baseName": "chainID",
            "type": "string",
            "format": "string"
        },
        {
            "name": "chainOwnerId",
            "baseName": "chainOwnerId",
            "type": "string",
            "format": "string"
        },
        {
            "name": "evmChainId",
            "baseName": "evmChainId",
            "type": "number",
            "format": "int32"
        },
        {
            "name": "gasFeePolicy",
            "baseName": "gasFeePolicy",
            "type": "FeePolicy",
            "format": ""
        },
        {
            "name": "gasLimits",
            "baseName": "gasLimits",
            "type": "Limits",
            "format": ""
        },
        {
            "name": "isActive",
            "baseName": "isActive",
            "type": "boolean",
            "format": "boolean"
        },
        {
            "name": "metadata",
            "baseName": "metadata",
            "type": "PublicChainMetadata",
            "format": ""
        },
        {
            "name": "publicURL",
            "baseName": "publicURL",
            "type": "string",
            "format": "string"
        }    ];

    static getAttributeTypeMap() {
        return ChainInfoResponse.attributeTypeMap;
    }

    public constructor() {
    }
}
