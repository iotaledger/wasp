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

import { Output } from '../models/Output';
import { HttpFile } from '../http/http';

export class OnLedgerRequest {
    /**
    * The request ID
    */
    'id': string;
    'output': Output;
    /**
    * The output ID
    */
    'outputId': string;
    /**
    * The raw data of the request (Hex)
    */
    'raw': string;

    static readonly discriminator: string | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "id",
            "baseName": "id",
            "type": "string",
            "format": "string"
        },
        {
            "name": "output",
            "baseName": "output",
            "type": "Output",
            "format": ""
        },
        {
            "name": "outputId",
            "baseName": "outputId",
            "type": "string",
            "format": "string"
        },
        {
            "name": "raw",
            "baseName": "raw",
            "type": "string",
            "format": "string"
        }    ];

    static getAttributeTypeMap() {
        return OnLedgerRequest.attributeTypeMap;
    }

    public constructor() {
    }
}

