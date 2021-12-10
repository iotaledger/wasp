// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// conversion of data bytes to key id

import {getKeyIDFromString} from "./host";

export interface MapKey {
    getKeyID(): Key32;
}

export class Key32 implements MapKey {
    keyID: i32;

    constructor(keyID: i32) {
        this.keyID = keyID;
    }

    static fromString(key: string): Key32 {
        return getKeyIDFromString(key);
    }

    getKeyID(): Key32 {
        return this;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// all predefined key id values should exactly match their counterpart values on the host!
// note that predefined key ids are negative values to distinguish them from indexes

// @formatter:off
export const KEY_ACCOUNT_ID       = new Key32(-1);
export const KEY_ADDRESS          = new Key32(-2);
export const KEY_BALANCES         = new Key32(-3);
export const KEY_BASE58_DECODE    = new Key32(-4);
export const KEY_BASE58_ENCODE    = new Key32(-5);
export const KEY_BLS_ADDRESS      = new Key32(-6);
export const KEY_BLS_AGGREGATE    = new Key32(-7);
export const KEY_BLS_VALID        = new Key32(-8);
export const KEY_CALL             = new Key32(-9);
export const KEY_CALLER           = new Key32(-10);
export const KEY_CHAIN_ID         = new Key32(-11);
export const KEY_CHAIN_OWNER_ID   = new Key32(-12);
export const KEY_COLOR            = new Key32(-13);
export const KEY_CONTRACT         = new Key32(-14);
export const KEY_CONTRACT_CREATOR = new Key32(-15);
export const KEY_DEPLOY           = new Key32(-16);
export const KEY_ED25519_ADDRESS  = new Key32(-17);
export const KEY_ED25519_VALID    = new Key32(-18);
export const KEY_EVENT            = new Key32(-19);
export const KEY_EXPORTS          = new Key32(-20);
export const KEY_HASH_BLAKE2B     = new Key32(-21);
export const KEY_HASH_SHA3        = new Key32(-22);
export const KEY_HNAME            = new Key32(-23);
export const KEY_INCOMING         = new Key32(-24);
export const KEY_LENGTH           = new Key32(-25);
export const KEY_LOG              = new Key32(-26);
export const KEY_MAPS             = new Key32(-27);
export const KEY_MINTED           = new Key32(-28);
export const KEY_PANIC            = new Key32(-29);
export const KEY_PARAMS           = new Key32(-30);
export const KEY_POST             = new Key32(-31);
export const KEY_RANDOM           = new Key32(-32);
export const KEY_REQUEST_ID       = new Key32(-33);
export const KEY_RESULTS          = new Key32(-34);
export const KEY_RETURN           = new Key32(-35);
export const KEY_STATE            = new Key32(-36);
export const KEY_TIMESTAMP        = new Key32(-37);
export const KEY_TRACE            = new Key32(-38);
export const KEY_TRANSFERS        = new Key32(-39);
export const KEY_UTILITY          = new Key32(-40);
export const KEY_ZZZZZZZ          = new Key32(-41);
// @formatter:on
