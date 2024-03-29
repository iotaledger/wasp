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

import { HttpFile } from '../http/http';

export class ControlAddressesResponse {
    /**
    * The governing address (Bech32)
    */
    'governingAddress': string;
    /**
    * The block index (uint32
    */
    'sinceBlockIndex': number;
    /**
    * The state address (Bech32)
    */
    'stateAddress': string;

    static readonly discriminator: string | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "governingAddress",
            "baseName": "governingAddress",
            "type": "string",
            "format": "string"
        },
        {
            "name": "sinceBlockIndex",
            "baseName": "sinceBlockIndex",
            "type": "number",
            "format": "int32"
        },
        {
            "name": "stateAddress",
            "baseName": "stateAddress",
            "type": "string",
            "format": "string"
        }    ];

    static getAttributeTypeMap() {
        return ControlAddressesResponse.attributeTypeMap;
    }

    public constructor() {
    }
}

