// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Buffer} from "./buffer";

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


export function panic(err: string) {
    throw new Error(err);
}