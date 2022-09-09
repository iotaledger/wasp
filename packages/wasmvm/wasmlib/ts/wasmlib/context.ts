// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// encapsulates standard host entities into a simple interface

import * as wasmtypes from "./wasmtypes"
import {ScFuncCallContext, ScViewCallContext} from "./contract";
import {ScSandboxFunc, ScSandboxView} from "./sandbox";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract interface with mutable access to state
export class ScFuncContext extends ScSandboxFunc implements ScFuncCallContext {
    // host(): ScHost | null {
    //     return null;
    // }

    chainID(): wasmtypes.ScChainID {
        return super.currentChainID();
    }

    initFuncCallContext(): void {
    }

    initViewCallContext(hContract: wasmtypes.ScHname): wasmtypes.ScHname {
        return hContract;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view interface which has only immutable access to state
export class ScViewContext extends ScSandboxView implements ScViewCallContext {

    chainID(): wasmtypes.ScChainID {
        return super.currentChainID();
    }
    initViewCallContext(hContract: wasmtypes.ScHname): wasmtypes.ScHname {
        return hContract;
    }
}
