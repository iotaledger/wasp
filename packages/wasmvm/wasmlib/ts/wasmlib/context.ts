// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// encapsulates standard host entities into a simple interface

import {ScFuncClientContext, ScViewClientContext} from './contract';
import {ScSandboxFunc, ScSandboxView} from './sandbox';
import {ScHname} from './wasmtypes/schname';
import {CallRequest, PostRequest} from './wasmrequests';
import {ScChainID} from './wasmtypes';

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract interface with mutable access to state
export class ScFuncContext extends ScSandboxFunc implements ScFuncClientContext {
    clientContract(hContract: ScHname): ScHname {
        return hContract;
    }

    fnCall(req: CallRequest): Uint8Array {
        return super.fnCall(req);
    }

    fnChainID(): ScChainID {
        return super.fnChainID()
    }

    fnPost(req: PostRequest): Uint8Array {
        return super.fnPost(req);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view interface which has only immutable access to state
export class ScViewContext extends ScSandboxView implements ScViewClientContext {
    clientContract(hContract: ScHname): ScHname {
        return hContract;
    }

    fnCall(req: CallRequest): Uint8Array {
        return super.fnCall(req);
    }

    fnChainID(): ScChainID {
        return super.fnChainID()
    }
}
