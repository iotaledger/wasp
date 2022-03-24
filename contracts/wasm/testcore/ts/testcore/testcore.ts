// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as wasmtypes from "wasmlib/wasmtypes";
import * as coreaccounts from "wasmlib/coreaccounts"
import * as coregovernance from "wasmlib/coregovernance"
import * as sc from "./index";

const CONTRACT_NAME_DEPLOYED = "exampleDeployTR";
const MSG_CORE_ONLY_PANIC = "========== core only =========";
const MSG_FULL_PANIC = "========== panic FULL ENTRY POINT =========";
const MSG_VIEW_PANIC = "========== panic VIEW =========";

export function funcCallOnChain(ctx: wasmlib.ScFuncContext, f: sc.CallOnChainContext): void {
    let paramInt = f.params.intValue().value();

    let hnameContract = ctx.contract();
    if (f.params.hnameContract().exists()) {
        hnameContract = f.params.hnameContract().value();
    }

    let hnameEP = sc.HFuncCallOnChain;
    if (f.params.hnameEP().exists()) {
        hnameEP = f.params.hnameEP().value();
    }

    let counter = f.state.counter();
    ctx.log("call depth = " + f.params.intValue().toString() +
        ", hnameContract = " + hnameContract.toString() +
        ", hnameEP = " + hnameEP.toString() +
        ", counter = " + counter.toString())

    counter.setValue(counter.value() + 1);

    let params = new wasmlib.ScDict([]);
    const key = wasmtypes.stringToBytes(sc.ParamIntValue);
    params.set(key, wasmtypes.int64ToBytes(paramInt))
    let ret = ctx.call(hnameContract, hnameEP, params, null);
    let retVal = wasmtypes.int64FromBytes(ret.get(key));
    f.results.intValue().setValue(retVal);
}

export function funcCheckContextFromFullEP(ctx: wasmlib.ScFuncContext, f: sc.CheckContextFromFullEPContext): void {
    ctx.require(f.params.agentID().value().equals(ctx.accountID()), "fail: agentID");
    ctx.require(f.params.caller().value().equals(ctx.caller()), "fail: caller");
    ctx.require(f.params.chainID().value().equals(ctx.chainID()), "fail: chainID");
    ctx.require(f.params.chainOwnerID().value().equals(ctx.chainOwnerID()), "fail: chainOwnerID");
    ctx.require(f.params.contractCreator().value().equals(ctx.contractCreator()), "fail: contractCreator");
}

export function funcDoNothing(ctx: wasmlib.ScFuncContext, f: sc.DoNothingContext): void {
    ctx.log("doing nothing...");
}

// export function funcGetMintedSupply(ctx: wasmlib.ScFuncContext, f: sc.GetMintedSupplyContext): void {
//     let minted = ctx.minted();
//     let mintedColors = minted.colors();
//     ctx.require(mintedColors.length == 1, "test only supports one minted color");
//     let color = mintedColors[0];
//     let amount = minted.balance(color);
//     f.results.mintedSupply().setValue(amount);
//     f.results.mintedColor().setValue(color);
// }

export function funcIncCounter(ctx: wasmlib.ScFuncContext, f: sc.IncCounterContext): void {
    let counter = f.state.counter();
    counter.setValue(counter.value() + 1);
}

export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
    if (f.params.fail().exists()) {
        ctx.panic("failing on purpose");
    }
}

export function funcPassTypesFull(ctx: wasmlib.ScFuncContext, f: sc.PassTypesFullContext): void {
    let hash = ctx.utility().hashBlake2b(wasmtypes.stringToBytes(sc.ParamHash));
    ctx.require(f.params.hash().value().equals(hash), "Hash wrong");
    ctx.require(f.params.int64().value() == 42, "int64 wrong");
    ctx.require(f.params.int64Zero().value() == 0, "int64-0 wrong");
    ctx.require(f.params.string().value() == sc.ParamString, "string wrong");
    ctx.require(f.params.stringZero().value() == "", "string-0 wrong");
    ctx.require(f.params.hname().value().equals(ctx.utility().hname(sc.ParamHname)), "Hname wrong");
    ctx.require(f.params.hnameZero().value().equals(new wasmtypes.ScHname(0)), "Hname-0 wrong");
}

export function funcRunRecursion(ctx: wasmlib.ScFuncContext, f: sc.RunRecursionContext): void {
    let depth = f.params.intValue().value();
    if (depth <= 0) {
        return;
    }

    let callOnChain = sc.ScFuncs.callOnChain(ctx);
    callOnChain.params.intValue().setValue(depth - 1);
    callOnChain.params.hnameEP().setValue(sc.HFuncRunRecursion);
    callOnChain.func.call();
    let retVal = callOnChain.results.intValue().value();
    f.results.intValue().setValue(retVal);
}

export function funcSendToAddress(ctx: wasmlib.ScFuncContext, f: sc.SendToAddressContext): void {
    // let transfer = wasmlib.ScTransfers.fromBalances(ctx.balances());
    // ctx.send(f.params.address().value(), transfer);
}

export function funcSetInt(ctx: wasmlib.ScFuncContext, f: sc.SetIntContext): void {
    f.state.ints().getInt64(f.params.name().value()).setValue(f.params.intValue().value());
}

export function funcSpawn(ctx: wasmlib.ScFuncContext, f: sc.SpawnContext): void {
    let programHash = f.params.progHash().value();
    let spawnName = sc.ScName + "_spawned";
    let spawnDescr = "spawned contract description";
    ctx.deployContract(programHash, spawnName, spawnDescr, null);

    let spawnHname = ctx.utility().hname(spawnName);
    for (let i = 0; i < 5; i++) {
        ctx.call(spawnHname, sc.HFuncIncCounter, null, null);
    }
}

export function funcTestBlockContext1(ctx: wasmlib.ScFuncContext, f: sc.TestBlockContext1Context): void {
    ctx.panic(MSG_CORE_ONLY_PANIC);
}

export function funcTestBlockContext2(ctx: wasmlib.ScFuncContext, f: sc.TestBlockContext2Context): void {
    ctx.panic(MSG_CORE_ONLY_PANIC);
}

export function funcTestCallPanicFullEP(ctx: wasmlib.ScFuncContext, f: sc.TestCallPanicFullEPContext): void {
    sc.ScFuncs.testPanicFullEP(ctx).func.call();
}

export function funcTestCallPanicViewEPFromFull(ctx: wasmlib.ScFuncContext, f: sc.TestCallPanicViewEPFromFullContext): void {
    sc.ScFuncs.testPanicViewEP(ctx).func.call();
}

export function funcTestChainOwnerIDFull(ctx: wasmlib.ScFuncContext, f: sc.TestChainOwnerIDFullContext): void {
    f.results.chainOwnerID().setValue(ctx.chainOwnerID());
}

export function funcTestEventLogDeploy(ctx: wasmlib.ScFuncContext, f: sc.TestEventLogDeployContext): void {
    // deploy the same contract with another name
    let programHash = ctx.utility().hashBlake2b(wasmtypes.stringToBytes("testcore"));
    ctx.deployContract(programHash, CONTRACT_NAME_DEPLOYED, "test contract deploy log", null);
}

export function funcTestEventLogEventData(ctx: wasmlib.ScFuncContext, f: sc.TestEventLogEventDataContext): void {
    ctx.event("[Event] - Testing Event...");
}

export function funcTestEventLogGenericData(ctx: wasmlib.ScFuncContext, f: sc.TestEventLogGenericDataContext): void {
    let event = "[GenericData] Counter Number: ".toString() + f.params.counter().toString();
    ctx.event(event);
}

export function funcTestPanicFullEP(ctx: wasmlib.ScFuncContext, f: sc.TestPanicFullEPContext): void {
    ctx.panic(MSG_FULL_PANIC);
}

// export function funcWithdrawToChain(ctx: wasmlib.ScFuncContext, f: sc.WithdrawToChainContext): void {
//     let xx = coreaccounts.ScFuncs.withdraw(ctx);
//     xx.func.postToChain(f.params.chainID().value());
// }

export function viewCheckContextFromViewEP(ctx: wasmlib.ScViewContext, f: sc.CheckContextFromViewEPContext): void {
    ctx.require(f.params.agentID().value().equals(ctx.accountID()), "fail: agentID");
    ctx.require(f.params.chainID().value().equals(ctx.chainID()), "fail: chainID");
    ctx.require(f.params.chainOwnerID().value().equals(ctx.chainOwnerID()), "fail: chainOwnerID");
    ctx.require(f.params.contractCreator().value().equals(ctx.contractCreator()), "fail: contractCreator");
}

export function viewFibonacci(ctx: wasmlib.ScViewContext, f: sc.FibonacciContext): void {
    let n = f.params.intValue().value();
    if (n == 0 || n == 1) {
        f.results.intValue().setValue(n);
        return;
    }

    let fib = sc.ScFuncs.fibonacci(ctx);
    fib.params.intValue().setValue(n - 1);
    fib.func.call();
    let n1 = fib.results.intValue().value();

    fib.params.intValue().setValue(n - 2);
    fib.func.call();
    let n2 = fib.results.intValue().value();

    f.results.intValue().setValue(n1 + n2);
}

export function viewGetCounter(ctx: wasmlib.ScViewContext, f: sc.GetCounterContext): void {
    f.results.counter().setValue(f.state.counter().value());
}

export function viewGetInt(ctx: wasmlib.ScViewContext, f: sc.GetIntContext): void {
    let name = f.params.name().value();
    let value = f.state.ints().getInt64(name);
    ctx.require(value.exists(), "param '" + name + "' not found");
    f.results.values().getInt64(name).setValue(value.value());
}

export function viewGetStringValue(ctx: wasmlib.ScViewContext, f: sc.GetStringValueContext): void {
    ctx.panic(MSG_CORE_ONLY_PANIC);
}

export function viewJustView(ctx: wasmlib.ScViewContext, f: sc.JustViewContext): void {
    ctx.log("doing nothing...");
}

export function viewPassTypesView(ctx: wasmlib.ScViewContext, f: sc.PassTypesViewContext): void {
    let hash = ctx.utility().hashBlake2b(wasmtypes.stringToBytes(sc.ParamHash));
    ctx.require(f.params.hash().value().equals(hash), "Hash wrong");
    ctx.require(f.params.int64().value() == 42, "int64 wrong");
    ctx.require(f.params.int64Zero().value() == 0, "int64-0 wrong");
    ctx.require(f.params.string().value() == sc.ParamString, "string wrong");
    ctx.require(f.params.stringZero().value() == "", "string-0 wrong");
    ctx.require(f.params.hname().value().equals(ctx.utility().hname(sc.ParamHname)), "Hname wrong");
    ctx.require(f.params.hnameZero().value().equals(new wasmtypes.ScHname(0)), "Hname-0 wrong");
}

export function viewTestCallPanicViewEPFromView(ctx: wasmlib.ScViewContext, f: sc.TestCallPanicViewEPFromViewContext): void {
    sc.ScFuncs.testPanicViewEP(ctx).func.call();
}

export function viewTestChainOwnerIDView(ctx: wasmlib.ScViewContext, f: sc.TestChainOwnerIDViewContext): void {
    f.results.chainOwnerID().setValue(ctx.chainOwnerID());
}

export function viewTestPanicViewEP(ctx: wasmlib.ScViewContext, f: sc.TestPanicViewEPContext): void {
    ctx.panic(MSG_VIEW_PANIC);
}

export function viewTestSandboxCall(ctx: wasmlib.ScViewContext, f: sc.TestSandboxCallContext): void {
    let getChainInfo = coregovernance.ScFuncs.getChainInfo(ctx);
    getChainInfo.func.call();
    f.results.sandboxCall().setValue(getChainInfo.results.description().value());
}

export function funcClaimAllowance(ctx: wasmlib.ScFuncContext, f: sc.ClaimAllowanceContext): void {
}

export function funcEstimateMinDust(ctx: wasmlib.ScFuncContext, f: sc.EstimateMinDustContext): void {
}

export function funcInfiniteLoop(ctx: wasmlib.ScFuncContext, f: sc.InfiniteLoopContext): void {
    for (; ;) {
        // do nothing, just waste gas
    }
}

export function funcPingAllowanceBack(ctx: wasmlib.ScFuncContext, f: sc.PingAllowanceBackContext): void {
}

export function funcSendLargeRequest(ctx: wasmlib.ScFuncContext, f: sc.SendLargeRequestContext): void {
}

export function funcSendNFTsBack(ctx: wasmlib.ScFuncContext, f: sc.SendNFTsBackContext): void {
}

export function funcSplitFunds(ctx: wasmlib.ScFuncContext, f: sc.SplitFundsContext): void {
    let iotas = ctx.allowance().balance(wasmtypes.IOTA);
    const address = ctx.caller().address();
    const transfer = wasmlib.ScTransfers.iotas(200);
    for (; iotas >= 200; iotas -= 200) {
        ctx.transferAllowed(ctx.accountID(), transfer, false);
        ctx.send(address, transfer);
    }
}

export function funcSplitFundsNativeTokens(ctx: wasmlib.ScFuncContext, f: sc.SplitFundsNativeTokensContext): void {
    let iotas = ctx.allowance().balance(wasmtypes.IOTA);
    const address = ctx.caller().address();
    let transfer = wasmlib.ScTransfers.iotas(iotas);
    ctx.transferAllowed(ctx.accountID(), transfer, false);
    const colors = ctx.allowance().colors();
    for (let i = 0; i < colors.length; i++) {
        const token = colors[i];
        if (token.equals(wasmtypes.IOTA)) {
            continue;
        }
        transfer = wasmlib.ScTransfers.transfer(token, 1);
        let tokens = ctx.allowance().balance(token);
        for (; tokens >= 1; tokens--) {
            ctx.transferAllowed(ctx.accountID(), transfer, false);
            ctx.send(address, transfer);
        }
    }
}

export function funcWithdrawFromChain(ctx: wasmlib.ScFuncContext, f: sc.WithdrawFromChainContext): void {
}

export function viewInfiniteLoopView(ctx: wasmlib.ScViewContext, f: sc.InfiniteLoopViewContext): void {
    for (; ;) {
        // do nothing, just waste gas
    }
}
