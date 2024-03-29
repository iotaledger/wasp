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

export class Blob {
    'hash': string;
    'size': number;

    static readonly discriminator: string | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "hash",
            "baseName": "hash",
            "type": "string",
            "format": "string"
        },
        {
            "name": "size",
            "baseName": "size",
            "type": "number",
            "format": "int32"
        }    ];

    static getAttributeTypeMap() {
        return Blob.attributeTypeMap;
    }

    public constructor() {
    }
}

