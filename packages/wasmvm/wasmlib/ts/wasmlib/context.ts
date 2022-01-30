// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// encapsulates standard host entities into a simple interface

import * as wasmtypes from "./wasmtypes"
import {ScBalances, ScTransfers} from "./assets";
import {ScFuncCallContext, ScViewCallContext} from "./contract";
import {panic, ScSandboxFunc, ScSandboxView} from "./sandbox";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract interface with mutable access to state
export class ScFuncContext extends ScSandboxFunc implements ScViewCallContext, ScFuncCallContext {
    canCallFunc(): void {
        panic("canCallFunc");
    }

    canCallView(): void {
        panic("canCallView");
    }

    // TODO deprecated
    incoming(): ScBalances {
        return super.incomingTransfer();
    }

    // TODO deprecated
    transferToAddress(address: wasmtypes.ScAddress, transfer: ScTransfers): void {
        super.send(address, transfer);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view interface which has only immutable access to state
export class ScViewContext extends ScSandboxView implements ScViewCallContext {
    canCallView(): void {
        panic("canCallView");
    }
}
