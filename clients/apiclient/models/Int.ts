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

export class Int {
    'abs'?: Array<number>;
    'neg'?: boolean;

    static readonly discriminator: string | undefined = undefined;

    static readonly mapping: {[index: string]: string} | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "abs",
            "baseName": "abs",
            "type": "Array<number>",
            "format": "int32"
        },
        {
            "name": "neg",
            "baseName": "neg",
            "type": "boolean",
            "format": "boolean"
        }    ];

    static getAttributeTypeMap() {
        return Int.attributeTypeMap;
    }

    public constructor() {
    }
}
