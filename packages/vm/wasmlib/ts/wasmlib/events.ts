// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {ScAddress, ScAgentID, ScChainID, ScColor, ScHash, ScHname, ScRequestID} from "./hashtypes";
import * as keys from "./keys";
import {base58Encode, ROOT} from "./context";

// encodes separate entities into a byte buffer
export class EventEncoder {
    event: string;

    // constructs an encoder
    constructor(eventName: string) {
        this.event = eventName;
        let timestamp = ROOT.getInt64(keys.KEY_TIMESTAMP).value();
        this.int64(timestamp / 1_000_000_000);
    }

    // encodes an ScAddress into the byte buffer
    address(value: ScAddress): EventEncoder {
        return this.string(value.toString());
    }

    // encodes an ScAgentID into the byte buffer
    agentID(value: ScAgentID): EventEncoder {
        return this.string(value.toString());
    }

    // encodes a substring of bytes into the byte buffer
    bytes(value: u8[]): EventEncoder {
        return this.string(base58Encode(value));
    }

    // encodes an ScChainID into the byte buffer
    chainID(value: ScChainID): EventEncoder {
        return this.string(value.toString());
    }

    // encodes an ScColor into the byte buffer
    color(value: ScColor): EventEncoder {
        return this.string(value.toString());
    }

    // retrieve the encoded byte buffer
    emit(): void {
        ROOT.getString(keys.KEY_EVENT).setValue(this.event);
    }

    // encodes an ScHash into the byte buffer
    hash(value: ScHash): EventEncoder {
        return this.string(value.toString());
    }

    // encodes an ScHname into the byte buffer
    hname(value: ScHname): EventEncoder {
        return this.string(value.toString());
    }

    // encodes an int16 into the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int16(value: i16): EventEncoder {
        return this.string(value.toString());
    }

    // encodes an int32 into the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int32(value: i32): EventEncoder {
        return this.string(value.toString());
    }

    // encodes an int64 into the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int64(value: i64): EventEncoder {
        return this.string(value.toString());
    }

    // encodes an ScRequestID into the byte buffer
    requestID(value: ScRequestID): EventEncoder {
        return this.string(value.toString());
    }

    // encodes an UTF-8 text string into the byte buffer
    string(value: string): EventEncoder {
        //TODO encode potential vertical bars that are present in the value string
        this.event += "|" + value;
        return this;
    }
}
