// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Buffer} from "./buffer";

export const TYPE_ADDRESS = 1;
export const TYPE_AGENT_ID = 2;
export const TYPE_BOOL = 3;
export const TYPE_BYTES = 4;
export const TYPE_CHAIN_ID = 5;
export const TYPE_COLOR = 6;
export const TYPE_HASH = 7;
export const TYPE_HNAME = 8;
export const TYPE_INT8 = 9;
export const TYPE_INT16 = 10;
export const TYPE_INT32 = 11;
export const TYPE_INT64 = 12;
export const TYPE_MAP = 13;
export const TYPE_REQUEST_ID = 14;
export const TYPE_STRING = 15;

export type Address = string;
export type AgentID = string;
export type Bool = boolean;
export type Bytes = Buffer;
export type ChainID = string;
export type Color = string;
export type Hash = string;
export type Hname = number;
export type Int8 = number;
export type Int16 = number;
export type Int32 = number;
export type Int64 = bigint;
export type RequestID = string;
export type String = string;
export type Uint8 = number;
export type Uint16 = number;
export type Uint32 = number;
export type Uint64 = bigint;

export const TYPE_SIZES = new Uint8Array([0, 33, 37, 1, 0, 33, 32, 32, 4, 1, 2, 4, 8, 0, 34, 0]);

export function panic(err: string) {
    throw new Error(err);
}