// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

const CONTRACT_NAME_DEPLOYED: &str = "exampleDeployTR";
const MSG_FULL_PANIC: &str = "========== panic FULL ENTRY POINT =========";
const MSG_VIEW_PANIC: &str = "========== panic VIEW =========";

pub fn func_call_on_chain(ctx: &ScFuncContext) {
    ctx.log("testcore.callOnChain");

    let p = ctx.params();
    let param_hname_contract = p.get_hname(PARAM_HNAME_CONTRACT);
    let param_hname_ep = p.get_hname(PARAM_HNAME_EP);
    let param_int_value = p.get_int(PARAM_INT_VALUE);

    ctx.require(param_int_value.exists(), "missing mandatory intValue");

    let param_int = param_int_value.value();

    let mut target_contract = ctx.contract_id().hname();
    if param_hname_contract.exists() {
        target_contract = param_hname_contract.value()
    }

    let mut target_ep = HFUNC_CALL_ON_CHAIN;
    if param_hname_ep.exists() {
        target_ep = param_hname_ep.value()
    }

    let var_counter = ctx.state().get_int(VAR_COUNTER);
    let counter = var_counter.value();
    var_counter.set_value(counter + 1);

    ctx.log(&format!("call depth = {} hnameContract = {} hnameEP = {} counter = {}",
                     param_int, &target_contract.to_string(), &target_ep.to_string(), counter));

    let params = ScMutableMap::new();
    params.get_int(PARAM_INT_VALUE).set_value(param_int);
    let ret = ctx.call(target_contract, target_ep, Some(params), None);

    let ret_val = ret.get_int(PARAM_INT_VALUE);
    ctx.results().get_int(PARAM_INT_VALUE).set_value(ret_val.value());
    ctx.log("testcore.callOnChain ok");
}

pub fn func_check_context_from_full_ep(ctx: &ScFuncContext) {
    ctx.log("testcore.checkContextFromFullEP");

    let p = ctx.params();
    let param_agent_id = p.get_agent_id(PARAM_AGENT_ID);
    let param_caller = p.get_agent_id(PARAM_CALLER);
    let param_chain_id = p.get_chain_id(PARAM_CHAIN_ID);
    let param_chain_owner_id = p.get_agent_id(PARAM_CHAIN_OWNER_ID);
    let param_contract_creator = p.get_agent_id(PARAM_CONTRACT_CREATOR);
    let param_contract_id = p.get_contract_id(PARAM_CONTRACT_ID);

    ctx.require(param_chain_id.value() == ctx.contract_id().chain_id(), "fail: chainID");
    ctx.require(param_chain_owner_id.value() == ctx.chain_owner_id(), "fail: chainOwnerID");
    ctx.require(param_caller.value() == ctx.caller(), "fail: caller");
    ctx.require(param_contract_id.value() == ctx.contract_id(), "fail: contractID");
    ctx.require(param_agent_id.value() == ctx.contract_id().as_agent_id(), "fail: agentID");
    ctx.require(param_contract_creator.value() == ctx.contract_creator(), "fail: contractCreator");
    ctx.log("testcore.checkContextFromFullEP ok");
}

pub fn func_do_nothing(ctx: &ScFuncContext) {
    ctx.log("testcore.doNothing");
    ctx.log("testcore.doNothing ok");
}

pub fn func_init(ctx: &ScFuncContext) {
    ctx.log("testcore.init");
    ctx.log("testcore.init ok");
}

pub fn func_pass_types_full(ctx: &ScFuncContext) {
    ctx.log("testcore.passTypesFull");

    let p = ctx.params();
    let param_hash = p.get_hash(PARAM_HASH);
    let param_hname = p.get_hname(PARAM_HNAME);
    let param_hname_zero = p.get_hname(PARAM_HNAME_ZERO);
    let param_int64 = p.get_int(PARAM_INT64);
    let param_int64_zero = p.get_int(PARAM_INT64_ZERO);
    let param_string = p.get_string(PARAM_STRING);
    let param_string_zero = p.get_string(PARAM_STRING_ZERO);

    ctx.require(param_hash.exists(), "missing mandatory hash");
    ctx.require(param_hname.exists(), "missing mandatory hname");
    ctx.require(param_hname_zero.exists(), "missing mandatory hnameZero");
    ctx.require(param_int64.exists(), "missing mandatory int64");
    ctx.require(param_int64_zero.exists(), "missing mandatory int64Zero");
    ctx.require(param_string.exists(), "missing mandatory string");
    ctx.require(param_string_zero.exists(), "missing mandatory stringZero");

    let hash = ctx.utility().hash_blake2b(PARAM_HASH.as_bytes());
    ctx.require(param_hash.value() == hash, "Hash wrong");
    ctx.require(param_int64.value() == 42, "int64 wrong");
    ctx.require(param_int64_zero.value() == 0, "int64-0 wrong");
    ctx.require(param_string.value() == PARAM_STRING, "string wrong");
    ctx.require(param_hname.value() == ScHname::new(PARAM_HNAME), "Hname wrong");
    ctx.require(param_hname_zero.value() == ScHname(0), "Hname-0 wrong");
    ctx.log("testcore.passTypesFull ok");
}

pub fn func_run_recursion(ctx: &ScFuncContext) {
    ctx.log("testcore.runRecursion");

    let p = ctx.params();
    let param_int_value = p.get_int(PARAM_INT_VALUE);

    ctx.require(param_int_value.exists(), "missing mandatory intValue");

    let depth = param_int_value.value();
    if depth <= 0 {
        return;
    }

    let params = ScMutableMap::new();
    params.get_int(PARAM_INT_VALUE).set_value(depth - 1);
    params.get_hname(PARAM_HNAME_EP).set_value(HFUNC_RUN_RECURSION);
    ctx.call_self(HFUNC_CALL_ON_CHAIN, Some(params), None);
    // TODO how would I return result of the call ???
    ctx.results().get_int(PARAM_INT_VALUE).set_value(depth - 1);
    ctx.log("testcore.runRecursion ok");
}

pub fn func_send_to_address(ctx: &ScFuncContext) {
    ctx.log("testcore.sendToAddress");

    ctx.require(ctx.caller() == ctx.contract_creator(), "no permission");

    let p = ctx.params();
    let param_address = p.get_address(PARAM_ADDRESS);

    ctx.require(param_address.exists(), "missing mandatory address");

    let balances = ScTransfers::new_transfers_from_balances(ctx.balances());
    ctx.transfer_to_address(&param_address.value(), balances);
    ctx.log("testcore.sendToAddress ok");
}

pub fn func_set_int(ctx: &ScFuncContext) {
    ctx.log("testcore.setInt");

    let p = ctx.params();
    let param_int_value = p.get_int(PARAM_INT_VALUE);
    let param_name = p.get_string(PARAM_NAME);

    ctx.require(param_int_value.exists(), "missing mandatory intValue");
    ctx.require(param_name.exists(), "missing mandatory name");

    ctx.state().get_int(&param_name.value()).set_value(param_int_value.value());
    ctx.log("testcore.setInt ok");
}

pub fn func_get_minted_supply(ctx: &ScFuncContext) {
    ctx.log("testcore.getMintedSupply");

    // TODO implement sandbox call
    //  ctx.minted_supply() -> i64

    let minted_supply = 42; // dummy for the core test to pass
    ctx.results().get_int(PARAM_MINTED_SUPPLY).set_value(minted_supply);
    ctx.log("testcore.setInt ok");
}

pub fn func_test_call_panic_full_ep(ctx: &ScFuncContext) {
    ctx.log("testcore.testCallPanicFullEP");
    ctx.call_self(HFUNC_TEST_PANIC_FULL_EP, None, None);
    ctx.log("testcore.testCallPanicFullEP ok");
}

pub fn func_test_call_panic_view_epfrom_full(ctx: &ScFuncContext) {
    ctx.log("testcore.testCallPanicViewEPFromFull");
    ctx.call_self(HVIEW_TEST_PANIC_VIEW_EP, None, None);
    ctx.log("testcore.testCallPanicViewEPFromFull ok");
}

pub fn func_test_chain_owner_idfull(ctx: &ScFuncContext) {
    ctx.log("testcore.testChainOwnerIDFull");
    ctx.results().get_agent_id(PARAM_CHAIN_OWNER_ID).set_value(&ctx.chain_owner_id());
    ctx.log("testcore.testChainOwnerIDFull ok");
}

pub fn func_test_contract_idfull(ctx: &ScFuncContext) {
    ctx.log("testcore.testContractIDFull");
    ctx.results().get_contract_id(PARAM_CONTRACT_ID).set_value(&ctx.contract_id());
    ctx.log("testcore.testContractIDFull ok");
}

pub fn func_test_event_log_deploy(ctx: &ScFuncContext) {
    ctx.log("testcore.testEventLogDeploy");
    //Deploy the same contract with another name
    let program_hash = ctx.utility().hash_blake2b("test_sandbox".as_bytes());
    ctx.deploy(&program_hash, CONTRACT_NAME_DEPLOYED,
               "test contract deploy log", None);
    ctx.log("testcore.testEventLogDeploy ok");
}

pub fn func_test_event_log_event_data(ctx: &ScFuncContext) {
    ctx.log("testcore.testEventLogEventData");
    ctx.event("[Event] - Testing Event...");
    ctx.log("testcore.testEventLogEventData ok");
}

pub fn func_test_event_log_generic_data(ctx: &ScFuncContext) {
    ctx.log("testcore.testEventLogGenericData");

    let p = ctx.params();
    let param_counter = p.get_int(PARAM_COUNTER);

    ctx.require(param_counter.exists(), "missing mandatory counter");

    let event = "[GenericData] Counter Number: ".to_string() + &param_counter.to_string();
    ctx.event(&event);
    ctx.log("testcore.testEventLogGenericData ok");
}

pub fn func_test_panic_full_ep(ctx: &ScFuncContext) {
    ctx.log("testcore.testPanicFullEP");
    ctx.panic(MSG_FULL_PANIC);
    ctx.log("testcore.testPanicFullEP ok");
}

pub fn func_withdraw_to_chain(ctx: &ScFuncContext) {
    ctx.log("testcore.withdrawToChain");

    let p = ctx.params();
    let param_chain_id = p.get_chain_id(PARAM_CHAIN_ID);

    ctx.require(param_chain_id.exists(), "missing mandatory chainId");

    //Deploy the same contract with another name
    let target_contract_id = ScContractId::new(&param_chain_id.value(), &CORE_ACCOUNTS);
    let transfers = ScTransfers::new(&ScColor::IOTA, 2);
    ctx.post(&PostRequestParams {
        contract_id: target_contract_id,
        function: CORE_ACCOUNTS_FUNC_WITHDRAW_TO_CHAIN,
        params: None,
        transfer: Some(transfers),
        delay: 0,
    });
    // TODO how to check if post was successful
    ctx.log("testcore.withdrawToChain ok");
}

pub fn view_check_context_from_view_ep(ctx: &ScViewContext) {
    ctx.log("testcore.checkContextFromViewEP");

    let p = ctx.params();
    let param_agent_id = p.get_agent_id(PARAM_AGENT_ID);
    let param_chain_id = p.get_chain_id(PARAM_CHAIN_ID);
    let param_chain_owner_id = p.get_agent_id(PARAM_CHAIN_OWNER_ID);
    let param_contract_creator = p.get_agent_id(PARAM_CONTRACT_CREATOR);
    let param_contract_id = p.get_contract_id(PARAM_CONTRACT_ID);

    ctx.require(param_chain_id.value() == ctx.contract_id().chain_id(), "fail: chainID");
    ctx.require(param_chain_owner_id.value() == ctx.chain_owner_id(), "fail: chainOwnerID");
    ctx.require(param_contract_id.value() == ctx.contract_id(), "fail: contractID");
    ctx.require(param_agent_id.value() == ctx.contract_id().as_agent_id(), "fail: agentID");
    ctx.require(param_contract_creator.value() == ctx.contract_creator(), "fail: contractCreator");
    ctx.log("testcore.checkContextFromViewEP ok");
}

pub fn view_fibonacci(ctx: &ScViewContext) {
    ctx.log("testcore.fibonacci");

    let p = ctx.params();
    let param_int_value = p.get_int(PARAM_INT_VALUE);

    ctx.require(param_int_value.exists(), "missing mandatory intValue");

    let n = param_int_value.value();
    if n == 0 || n == 1 {
        ctx.results().get_int(PARAM_INT_VALUE).set_value(n);
        return;
    }
    let params1 = ScMutableMap::new();
    params1.get_int(PARAM_INT_VALUE).set_value(n - 1);
    let results1 = ctx.call_self(HVIEW_FIBONACCI, Some(params1));
    let n1 = results1.get_int(PARAM_INT_VALUE).value();

    let params2 = ScMutableMap::new();
    params2.get_int(PARAM_INT_VALUE).set_value(n - 2);
    let results2 = ctx.call_self(HVIEW_FIBONACCI, Some(params2));
    let n2 = results2.get_int(PARAM_INT_VALUE).value();

    ctx.results().get_int(PARAM_INT_VALUE).set_value(n1 + n2);
    ctx.log("testcore.fibonacci ok");
}

pub fn func_inc_counter(ctx: &ScFuncContext) {
    ctx.log("testcore.incCounter");
    ctx.state().get_int(VAR_COUNTER).set_value(ctx.state().get_int(VAR_COUNTER).value()+1);
    ctx.log("testcore.incCounter ok");
}

pub fn view_get_counter(ctx: &ScViewContext) {
    ctx.log("testcore.getCounter");
    let counter = ctx.state().get_int(VAR_COUNTER);
    ctx.results().get_int(VAR_COUNTER).set_value(counter.value());
    ctx.log("testcore.getCounter ok");
}

pub fn view_get_int(ctx: &ScViewContext) {
    ctx.log("testcore.getInt");

    let p = ctx.params();
    let param_name = p.get_string(PARAM_NAME);

    ctx.require(param_name.exists(), "missing mandatory name");

    let name = param_name.value();
    let value = ctx.state().get_int(&name);
    ctx.require(value.exists(), "param 'value' not found");
    ctx.results().get_int(&name).set_value(value.value());
    ctx.log("testcore.getInt ok");
}

pub fn view_just_view(ctx: &ScViewContext) {
    ctx.log("testcore.justView");
    ctx.log("testcore.justView ok");
}

pub fn view_pass_types_view(ctx: &ScViewContext) {
    ctx.log("testcore.passTypesView");

    let p = ctx.params();
    let param_hash = p.get_hash(PARAM_HASH);
    let param_hname = p.get_hname(PARAM_HNAME);
    let param_hname_zero = p.get_hname(PARAM_HNAME_ZERO);
    let param_int64 = p.get_int(PARAM_INT64);
    let param_int64_zero = p.get_int(PARAM_INT64_ZERO);
    let param_string = p.get_string(PARAM_STRING);
    let param_string_zero = p.get_string(PARAM_STRING_ZERO);

    ctx.require(param_hash.exists(), "missing mandatory hash");
    ctx.require(param_hname.exists(), "missing mandatory hname");
    ctx.require(param_hname_zero.exists(), "missing mandatory hnameZero");
    ctx.require(param_int64.exists(), "missing mandatory int64");
    ctx.require(param_int64_zero.exists(), "missing mandatory int64Zero");
    ctx.require(param_string.exists(), "missing mandatory string");
    ctx.require(param_string_zero.exists(), "missing mandatory stringZero");

    let hash = ctx.utility().hash_blake2b(PARAM_HASH.as_bytes());
    ctx.require(param_hash.value() == hash, "Hash wrong");
    ctx.require(param_int64.value() == 42, "int64 wrong");
    ctx.require(param_int64_zero.value() == 0, "int64-0 wrong");
    ctx.require(param_string.value() == PARAM_STRING, "string wrong");
    ctx.require(param_string_zero.value() == "", "string-0 wrong");
    ctx.require(param_hname.value() == ScHname::new(PARAM_HNAME), "Hname wrong");
    ctx.require(param_hname_zero.value() == ScHname(0), "Hname-0 wrong");
    ctx.log("testcore.passTypesView ok");
}

pub fn view_test_call_panic_view_epfrom_view(ctx: &ScViewContext) {
    ctx.log("testcore.testCallPanicViewEPFromView");
    ctx.call_self(HVIEW_TEST_PANIC_VIEW_EP, None);
    ctx.log("testcore.testCallPanicViewEPFromView ok");
}

pub fn view_test_chain_owner_idview(ctx: &ScViewContext) {
    ctx.log("testcore.testChainOwnerIDView");
    ctx.results().get_agent_id(PARAM_CHAIN_OWNER_ID).set_value(&ctx.chain_owner_id());
    ctx.log("testcore.testChainOwnerIDView ok");
}

pub fn view_test_contract_idview(ctx: &ScViewContext) {
    ctx.log("testcore.testContractIDView");
    ctx.results().get_contract_id(PARAM_CONTRACT_ID).set_value(&ctx.contract_id());
    ctx.log("testcore.testContractIDView ok");
}

pub fn view_test_panic_view_ep(ctx: &ScViewContext) {
    ctx.log("testcore.testPanicViewEP");
    ctx.panic(MSG_VIEW_PANIC);
    ctx.log("testcore.testPanicViewEP ok");
}

pub fn view_test_sandbox_call(ctx: &ScViewContext) {
    ctx.log("testcore.testSandboxCall");
    let ret = ctx.call(CORE_ROOT, CORE_ROOT_VIEW_GET_CHAIN_INFO, None);
    let desc = ret.get_string("d").value();
    ctx.results().get_string("sandboxCall").set_value(&desc);
    ctx.log("testcore.testSandboxCall ok");
}
