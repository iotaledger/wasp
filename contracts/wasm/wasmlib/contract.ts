// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// base contract objects

import {BytesEncoder} from "./bytes";
import {ROOT,ScTransfers} from "./context";
import {ScChainID,ScHname} from "./hashtypes";
import {getObjectID, panic, TYPE_MAP} from "./host";
import * as keys from "./keys";
import {ScMutableMap} from "./mutable";

export interface ScFuncCallContext {
    canCallFunc():void;
}

export interface ScViewCallContext {
    canCallView():void;
}

export class ScMapID {
    mapID: i32 = 0;
}

type NullableScMapID = ScMapID | null;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScView {
    hContract: ScHname;
    hFunction: ScHname;
    paramsID: NullableScMapID;
    resultsID: NullableScMapID;

    constructor(hContract: ScHname, hFunction: ScHname) {
        this.hContract = hContract;
        this.hFunction = hFunction;
        this.paramsID = null;
        this.resultsID = null;
    }

    setPtrs(paramsID: NullableScMapID, resultsID: NullableScMapID): void {
        this.paramsID = paramsID;
        this.resultsID = resultsID;

        if (paramsID === null) {
        } else {
            paramsID.mapID = ScMutableMap.create().mapID();
        }
    }

    call(): void {
        this.callWithTransfer(0);
    }

    callWithTransfer(transferID: i32): void {
        let encode = new BytesEncoder();
        encode.hname(this.hContract);
        encode.hname(this.hFunction);
        encode.int32(this.id(this.paramsID));
        encode.int32(transferID);
        ROOT.getBytes(keys.KEY_CALL).setValue(encode.data());

        let resultsID = this.resultsID;
        if (resultsID === null) {
        } else {
            resultsID.mapID = getObjectID(1, keys.KEY_RETURN, TYPE_MAP);
        }
    }

    clone(): ScView {
        let o = new ScView(this.hContract, this.hFunction);
        o.paramsID = this.paramsID;
        o.resultsID = this.resultsID;
        return o;
    }

    ofContract(hContract: ScHname): ScView {
        let ret = this.clone();
        ret.hContract = hContract;
        return ret;
    }

    id(paramsID: NullableScMapID): i32 {
        if (paramsID === null) {
            return 0;
        }
        return paramsID.mapID;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScInitFunc {
    view: ScView;

    constructor(hContract: ScHname, hFunction: ScHname) {
        this.view = new ScView(hContract, hFunction);
    }

    setPtrs(paramsID: NullableScMapID, resultsID: NullableScMapID): void {
        this.view.setPtrs(paramsID, resultsID);
    }

    call(): void {
        return panic("cannot call init");
    }

    clone(): ScInitFunc {
        let o = new ScInitFunc(this.view.hContract, this.view.hFunction);
        o.view = this.view.clone();
        return o;
    }

    ofContract(hContract: ScHname): ScInitFunc {
        let ret = this.clone();
        ret.view.hContract = hContract;
        return ret;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScFunc {
    view: ScView;
    delaySeconds: i32 = 0;
    transferID: i32 = 0;

    constructor(hContract: ScHname, hFunction: ScHname) {
        this.view = new ScView(hContract, hFunction);
    }

    setPtrs(paramsID: NullableScMapID, resultsID: NullableScMapID): void {
        this.view.setPtrs(paramsID, resultsID);
    }

    call(): void {
        if (this.delaySeconds != 0) {
            return panic("cannot delay a call");
        }
        this.view.callWithTransfer(this.transferID);
    }

    clone(): ScFunc {
        let o = new ScFunc(this.view.hContract, this.view.hFunction);
        o.view = this.view.clone();
        o.delaySeconds = this.delaySeconds;
        o.transferID = this.transferID;
        return o;
    }

    delay(seconds: i32): ScFunc {
        let ret = this.clone();
        ret.delaySeconds = seconds;
        return ret;
    }

    ofContract(hContract: ScHname): ScFunc {
        let ret = this.clone();
        ret.view.hContract = hContract;
        return ret;
    }

    post(): void {
        return this.postToChain(ROOT.getChainID(keys.KEY_CHAIN_ID).value());
    }

    postToChain(chainID: ScChainID): void {
        let encode = new BytesEncoder();
        encode.chainID(chainID);
        encode.hname(this.view.hContract);
        encode.hname(this.view.hFunction);
        encode.int32(this.view.id(this.view.paramsID));
        encode.int32(this.transferID);
        encode.int32(this.delaySeconds);
        ROOT.getBytes(keys.KEY_POST).setValue(encode.data());
    }

    transfer(transfer: ScTransfers): ScFunc {
        let ret = this.clone();
        ret.transferID = transfer.transfers.objID;
        return ret;
    }

    transferIotas(amount: i64): ScFunc {
        return this.transfer(ScTransfers.iotas(amount));
    }
}
