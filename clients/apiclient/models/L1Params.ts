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

import { IotaCoinInfo } from '../models/IotaCoinInfo';
import { Protocol } from '../models/Protocol';
import { HttpFile } from '../http/http';

export class L1Params {
    'baseToken': IotaCoinInfo;
    'protocol': Protocol;

    static readonly discriminator: string | undefined = undefined;

    static readonly mapping: {[index: string]: string} | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "baseToken",
            "baseName": "baseToken",
            "type": "IotaCoinInfo",
            "format": ""
        },
        {
            "name": "protocol",
            "baseName": "protocol",
            "type": "Protocol",
            "format": ""
        }    ];

    static getAttributeTypeMap() {
        return L1Params.attributeTypeMap;
    }

    public constructor() {
    }
}
