// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// encapsulates standard host entities into a simple interface

import {ScFuncCallContext, ScViewCallContext} from './contract';
import {ScSandboxFunc, ScSandboxView} from './sandbox';
import {ScHname} from './wasmtypes/schname';

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract interface with mutable access to state
export class ScFuncContext extends ScSandboxFunc implements ScFuncCallContext {
    // host(): ScHost | null {
    //     return null;
    // }

    initFuncCallContext(): void {
    }

    initViewCallContext(hContract: ScHname): ScHname {
        return hContract;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view interface which has only immutable access to state
export class ScViewContext extends ScSandboxView implements ScViewCallContext {
    initViewCallContext(hContract: ScHname): ScHname {
        return hContract;
    }
}
