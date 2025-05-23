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
import { GovPublicChainMetadata } from '../models/GovPublicChainMetadata';
import { Limits } from '../models/Limits';
import { HttpFile } from '../http/http';

export class GovChainInfoResponse {
    /**
    * The chain admin address (Hex Address).
    */
    'chainAdmin': string;
    /**
    * ChainID (Hex Address).
    */
    'chainID': string;
    'gasFeePolicy': FeePolicy;
    'gasLimits': Limits;
    'metadata': GovPublicChainMetadata;
    /**
    * The fully qualified public url leading to the chains metadata
    */
    'publicURL': string;

    static readonly discriminator: string | undefined = undefined;

    static readonly mapping: {[index: string]: string} | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "chainAdmin",
            "baseName": "chainAdmin",
            "type": "string",
            "format": "string"
        },
        {
            "name": "chainID",
            "baseName": "chainID",
            "type": "string",
            "format": "string"
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
            "name": "metadata",
            "baseName": "metadata",
            "type": "GovPublicChainMetadata",
            "format": ""
        },
        {
            "name": "publicURL",
            "baseName": "publicURL",
            "type": "string",
            "format": "string"
        }    ];

    static getAttributeTypeMap() {
        return GovChainInfoResponse.attributeTypeMap;
    }

    public constructor() {
    }
}
