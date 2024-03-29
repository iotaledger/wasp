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

import { Item } from '../models/Item';
import { HttpFile } from '../http/http';

export class JSONDict {
    'items'?: Array<Item>;

    static readonly discriminator: string | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "items",
            "baseName": "Items",
            "type": "Array<Item>",
            "format": ""
        }    ];

    static getAttributeTypeMap() {
        return JSONDict.attributeTypeMap;
    }

    public constructor() {
    }
}

