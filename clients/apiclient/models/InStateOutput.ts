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

export class InStateOutput {
    'output': Output;
    /**
    * The output ID
    */
    'outputId': string;

    static readonly discriminator: string | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
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
        }    ];

    static getAttributeTypeMap() {
        return InStateOutput.attributeTypeMap;
    }

    public constructor() {
    }
}

