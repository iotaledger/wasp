// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;
use crate::contract::*;

const CONTRACT_NAME_DEPLOYED: &str = "exampleDeployTR";
const MSG_CORE_ONLY_PANIC: &str = "========== core only =========";
const MSG_FULL_PANIC: &str = "========== panic FULL ENTRY POINT =========";
const MSG_VIEW_PANIC: &str = "========== panic VIEW =========";

pub fn func_call_on_chain(ctx: &ScFuncContext, f: &CallOnChainContext) {
    let param_int = f.params.n().value();

    let mut hname_contract = ctx.contract();
    if f.params.hname_contract().exists() {
        hname_contract = f.params.hname_contract().value();
    }

    let mut hname_ep = HFUNC_CALL_ON_CHAIN;
    if f.params.hname_ep().exists() {
        hname_ep = f.params.hname_ep().value();
    }

    let counter = f.state.counter();
    ctx.log(&format!("call depth = {}, hnameContract = {}, hnameEP = {}, counter = {}",
                     &f.params.n().to_string(),
                     &hname_contract.to_string(),
                     &hname_ep.to_string(),
                     &counter.to_string()));

    counter.set_value(counter.value() + 1);

    let parms = ScDict::new(&[]);
    let key = string_to_bytes(PARAM_N);
    parms.set(&key, &uint64_to_bytes(param_int));
    let ret = ctx.call(hname_contract, hname_ep, Some(parms), None);
    let ret_val = uint64_from_bytes(&ret.get(&key));
    f.results.n().set_value(ret_val);
}

pub fn func_check_context_from_full_ep(ctx: &ScFuncContext, f: &CheckContextFromFullEPContext) {
    ctx.require(f.params.agent_id().value() == ctx.account_id(), "fail: agentID");
    ctx.require(f.params.caller().value() == ctx.caller(), "fail: caller");
    ctx.require(f.params.chain_id().value() == ctx.current_chain_id(), "fail: chainID");
    ctx.require(f.params.chain_owner_id().value() == ctx.chain_owner_id(), "fail: chainOwnerID");
}

pub fn func_claim_allowance(ctx: &ScFuncContext, _f: &ClaimAllowanceContext) {
    let allowance = ctx.allowance();
    let transfer = wasmlib::ScTransfer::from_balances(&allowance);
    ctx.transfer_allowed(&ctx.account_id(), &transfer, false);
}

pub fn func_do_nothing(ctx: &ScFuncContext, _f: &DoNothingContext) {
    ctx.log("doing nothing...");
}

pub fn func_estimate_min_dust(ctx: &ScFuncContext, _f: &EstimateMinDustContext) {
    let provided = ctx.allowance().iotas();
    let dummy = ScFuncs::estimate_min_dust(ctx);
    let required = ctx.estimate_dust(&dummy.func);
    ctx.require(provided >= required, "not enough funds");
}

pub fn func_inc_counter(_ctx: &ScFuncContext, f: &IncCounterContext) {
    let counter = f.state.counter();
    counter.set_value(counter.value() + 1);
}

pub fn func_infinite_loop(_ctx: &ScFuncContext, _f: &InfiniteLoopContext) {
    loop {
        // do nothing, just waste gas
    }
}

pub fn func_init(ctx: &ScFuncContext, f: &InitContext) {
    if f.params.fail().exists() {
        ctx.panic("failing on purpose");
    }
}

pub fn func_pass_types_full(ctx: &ScFuncContext, f: &PassTypesFullContext) {
    let hash = ctx.utility().hash_blake2b(PARAM_HASH.as_bytes());
    ctx.require(f.params.hash().value() == hash, "Hash wrong");
    ctx.require(f.params.int64().value() == 42, "int64 wrong");
    ctx.require(f.params.int64_zero().value() == 0, "int64-0 wrong");
    ctx.require(f.params.string().value() == PARAM_STRING, "string wrong");
    ctx.require(f.params.string_zero().value() == "", "string-0 wrong");
    ctx.require(f.params.hname().value() == ctx.utility().hash_name(PARAM_HNAME), "Hname wrong");
    ctx.require(f.params.hname_zero().value() == ScHname(0), "Hname-0 wrong");
}

pub fn func_ping_allowance_back(ctx: &ScFuncContext, _f: &PingAllowanceBackContext) {
    let caller = ctx.caller();
    ctx.require(caller.is_address(), "pingAllowanceBack: caller expected to be a L1 address");
    let transfer = wasmlib::ScTransfer::from_balances(&ctx.allowance());
    ctx.transfer_allowed(&ctx.account_id(), &transfer, false);
    ctx.send(&caller.address(), &transfer);
}

pub fn func_run_recursion(ctx: &ScFuncContext, f: &RunRecursionContext) {
    let depth = f.params.n().value();
    if depth <= 0 {
        return;
    }

    let call_on_chain = ScFuncs::call_on_chain(ctx);
    call_on_chain.params.n().set_value(depth - 1);
    call_on_chain.params.hname_ep().set_value(HFUNC_RUN_RECURSION);
    call_on_chain.func.call();
    let ret_val = call_on_chain.results.n().value();
    f.results.n().set_value(ret_val);
}

pub fn func_send_large_request(_ctx: &ScFuncContext, _f: &SendLargeRequestContext) {
}

pub fn func_send_nf_ts_back(ctx: &ScFuncContext, _f: &SendNFTsBackContext) {
    let address = ctx.caller().address();
    let allowance = ctx.allowance();
    let transfer = wasmlib::ScTransfer::from_balances(&allowance);
    ctx.transfer_allowed(&ctx.account_id(), &transfer, false);
    for nft_id in allowance.nft_ids() {
        let transfer = ScTransfer::nft(nft_id);
        ctx.send(&address, &transfer);
    }
}

pub fn func_send_to_address(_ctx: &ScFuncContext, _f: &SendToAddressContext) {
    // let balances = ScTransfers::from_balances(ctx.balances());
    // ctx.send(&f.params.address().value(), &balances);
}

pub fn func_set_int(_ctx: &ScFuncContext, f: &SetIntContext) {
    f.state.ints().get_int64(&f.params.name().value()).set_value(f.params.int_value().value());
}

pub fn func_spawn(ctx: &ScFuncContext, f: &SpawnContext) {
    let program_hash = f.params.prog_hash().value();
    let spawn_name = SC_NAME.to_string() + "_spawned";
    let spawn_descr = "spawned contract description";
    ctx.deploy_contract(&program_hash, &spawn_name, spawn_descr, None);

    let spawn_hname = ctx.utility().hash_name(&spawn_name);
    for _i in 0..5 {
        ctx.call(spawn_hname, HFUNC_INC_COUNTER, None, None);
    }
}

pub fn func_split_funds(ctx: &ScFuncContext, _f: &SplitFundsContext) {
    let mut iotas = ctx.allowance().iotas();
    let address = ctx.caller().address();
    let iotas_to_transfer : u64 = 1_000_000;
    let transfer = wasmlib::ScTransfer::iotas(iotas_to_transfer);
    while iotas >= iotas_to_transfer {
        ctx.transfer_allowed(&ctx.account_id(), &transfer, false);
        ctx.send(&address, &transfer);
        iotas -= iotas_to_transfer;
    }
}

pub fn func_split_funds_native_tokens(ctx: &ScFuncContext, _f: &SplitFundsNativeTokensContext) {
    let iotas = ctx.allowance().iotas();
    let address = ctx.caller().address();
    let transfer = wasmlib::ScTransfer::iotas(iotas);
    ctx.transfer_allowed(&ctx.account_id(), &transfer, false);
    for token in ctx.allowance().token_ids() {
        let one = ScBigInt::from_uint64(1);
        let transfer = wasmlib::ScTransfer::tokens(&token, &one);
        let mut tokens = ctx.allowance().balance(&token);
        while tokens.cmp(&one) >= 0 {
            ctx.transfer_allowed(&ctx.account_id(), &transfer, false);
            ctx.send(&address, &transfer);
            tokens = tokens.sub(&one);
        }
    }
}

pub fn func_test_block_context1(ctx: &ScFuncContext, _f: &TestBlockContext1Context) {
    ctx.panic(MSG_CORE_ONLY_PANIC);
}

pub fn func_test_block_context2(ctx: &ScFuncContext, _f: &TestBlockContext2Context) {
    ctx.panic(MSG_CORE_ONLY_PANIC);
}

pub fn func_test_call_panic_full_ep(ctx: &ScFuncContext, _f: &TestCallPanicFullEPContext) {
    ScFuncs::test_panic_full_ep(ctx).func.call();
}

pub fn func_test_call_panic_view_ep_from_full(ctx: &ScFuncContext, _f: &TestCallPanicViewEPFromFullContext) {
    ScFuncs::test_panic_view_ep(ctx).func.call();
}

pub fn func_test_chain_owner_id_full(ctx: &ScFuncContext, f: &TestChainOwnerIDFullContext) {
    f.results.chain_owner_id().set_value(&ctx.chain_owner_id());
}

pub fn func_test_event_log_deploy(ctx: &ScFuncContext, _f: &TestEventLogDeployContext) {
    // deploy the same contract with another name
    let program_hash = ctx.utility().hash_blake2b("testcore".as_bytes());
    ctx.deploy_contract(&program_hash, CONTRACT_NAME_DEPLOYED, "test contract deploy log", None);
}

pub fn func_test_event_log_event_data(ctx: &ScFuncContext, _f: &TestEventLogEventDataContext) {
    ctx.event("[Event] - Testing Event...");
}

pub fn func_test_event_log_generic_data(ctx: &ScFuncContext, f: &TestEventLogGenericDataContext) {
    let event = "[GenericData] Counter Number: ".to_string() + &f.params.counter().to_string();
    ctx.event(&event);
}

pub fn func_test_panic_full_ep(ctx: &ScFuncContext, _f: &TestPanicFullEPContext) {
    ctx.panic(MSG_FULL_PANIC);
}

pub fn func_withdraw_from_chain(_ctx: &ScFuncContext, _f: &WithdrawFromChainContext) {
}

pub fn view_check_context_from_view_ep(ctx: &ScViewContext, f: &CheckContextFromViewEPContext) {
    ctx.require(f.params.agent_id().value() == ctx.account_id(), "fail: agentID");
    ctx.require(f.params.chain_id().value() == ctx.current_chain_id(), "fail: chainID");
    ctx.require(f.params.chain_owner_id().value() == ctx.chain_owner_id(), "fail: chainOwnerID");
}

fn fibonacci(n: u64) -> u64 {
    if n <= 1 {
        return n;
    }
    fibonacci(n - 1) + fibonacci(n - 2)
}

pub fn view_fibonacci(_ctx: &ScViewContext, f: &FibonacciContext) {
    let n = f.params.n().value();
    let result = fibonacci(n);
    f.results.n().set_value(result);
}

pub fn view_fibonacci_indirect(ctx: &ScViewContext, f: &FibonacciIndirectContext) {
    let n = f.params.n().value();
    if n == 0 || n == 1 {
        f.results.n().set_value(n);
        return;
    }

    let fib = ScFuncs::fibonacci_indirect(ctx);
    fib.params.n().set_value(n - 1);
    fib.func.call();
    let n1 = fib.results.n().value();

    fib.params.n().set_value(n - 2);
    fib.func.call();
    let n2 = fib.results.n().value();

    f.results.n().set_value(n1 + n2);
}

pub fn view_get_counter(_ctx: &ScViewContext, f: &GetCounterContext) {
    f.results.counter().set_value(f.state.counter().value());
}

pub fn view_get_int(ctx: &ScViewContext, f: &GetIntContext) {
    let name = f.params.name().value();
    let value = f.state.ints().get_int64(&name);
    ctx.require(value.exists(), &("param '".to_string() + &name + "' not found"));
    f.results.values().get_int64(&name).set_value(value.value());
}

pub fn view_get_string_value(ctx: &ScViewContext, _f: &GetStringValueContext) {
    ctx.panic(MSG_CORE_ONLY_PANIC);
}

pub fn view_infinite_loop_view(_ctx: &ScViewContext, _f: &InfiniteLoopViewContext) {
    loop {
        // do nothing, just waste gas
    }
}

pub fn view_just_view(ctx: &ScViewContext, _f: &JustViewContext) {
    ctx.log("doing nothing...");
}

pub fn view_pass_types_view(ctx: &ScViewContext, f: &PassTypesViewContext) {
    let hash = ctx.utility().hash_blake2b(PARAM_HASH.as_bytes());
    ctx.require(f.params.hash().value() == hash, "Hash wrong");
    ctx.require(f.params.int64().value() == 42, "int64 wrong");
    ctx.require(f.params.int64_zero().value() == 0, "int64-0 wrong");
    ctx.require(f.params.string().value() == PARAM_STRING, "string wrong");
    ctx.require(f.params.string_zero().value() == "", "string-0 wrong");
    ctx.require(f.params.hname().value() == ctx.utility().hash_name(PARAM_HNAME), "Hname wrong");
    ctx.require(f.params.hname_zero().value() == ScHname(0), "Hname-0 wrong");
}

pub fn view_test_call_panic_view_ep_from_view(ctx: &ScViewContext, _f: &TestCallPanicViewEPFromViewContext) {
    ScFuncs::test_panic_view_ep(ctx).func.call();
}

pub fn view_test_chain_owner_id_view(ctx: &ScViewContext, f: &TestChainOwnerIDViewContext) {
    f.results.chain_owner_id().set_value(&ctx.chain_owner_id());
}

pub fn view_test_panic_view_ep(ctx: &ScViewContext, _f: &TestPanicViewEPContext) {
    ctx.panic(MSG_VIEW_PANIC);
}

pub fn view_test_sandbox_call(ctx: &ScViewContext, f: &TestSandboxCallContext) {
    let get_chain_info = coregovernance::ScFuncs::get_chain_info(ctx);
    get_chain_info.func.call();
    f.results.sandbox_call().set_value(&get_chain_info.results.description().value());
}
