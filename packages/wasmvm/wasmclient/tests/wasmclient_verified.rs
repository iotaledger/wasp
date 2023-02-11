use std::cell::Cell;
use std::rc::Rc;

use wasmclient::{self, isc::keypair, wasmclientcontext::*, wasmclientservice::*};

const MYCHAIN: &str = "atoi1pz2e0rtqje9qc4ksu3khhmm7c62f908e03t3cq35l68m3e7kjr8tjhet7sd";
const MYSEED: &str = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3";

fn check_error(ctx: &WasmClientContext) {
    if let Err(e) = ctx.err() {
        println!("err: {}", e);
        assert!(false);
    }
}

fn setup_client() -> wasmclient::wasmclientcontext::WasmClientContext {
    let svc = WasmClientService::new("http://127.0.0.1:19090", "127.0.0.1:15550");
    let mut ctx = WasmClientContext::new(&svc, MYCHAIN, "testwasmlib");
    ctx.sign_requests(&keypair::KeyPair::from_sub_seed(
        &wasmlib::bytes_from_string(MYSEED),
        0,
    ));
    check_error(&ctx);
    return ctx;
}

#[test]
fn call_view() {
    let ctx = setup_client();
    let v = testwasmlib::ScFuncs::get_random(&ctx);
    v.func.call();
    check_error(&ctx);
    let rnd = v.results.random().value();
    println!("rnd: {}", rnd);
    assert!(rnd != 0);
}

#[test]
fn post_func_request() {
    let ctx = setup_client();
    let f = testwasmlib::ScFuncs::random(&ctx);
    f.func.post();
    check_error(&ctx);

    ctx.wait_request();
    check_error(&ctx);

    let v = testwasmlib::ScFuncs::get_random(&ctx);
    v.func.call();
    check_error(&ctx);
    let rnd = v.results.random().value();
    println!("rnd: {}", rnd);
    assert!(rnd != 0);
}

struct X {
    name: Rc<Cell<String>>,
}

#[test]
fn event_handling() {
    let mut ctx = setup_client();
    let mut events = testwasmlib::TestWasmLibEventHandlers::new("");

    let x = X::new();
    {
        let name = x.name.clone();
        events.on_test_wasm_lib_test(move |e| {
            name.set(e.name.clone());
        });
    }
    ctx.register(Box::new(events));
    check_error(&ctx);

    x.test_client_events_param(&ctx, "Lala");
    x.test_client_events_param(&ctx, "Trala");
    x.test_client_events_param(&ctx, "Bar|Bar");
    x.test_client_events_param(&ctx, "Bar~|~Bar");
    x.test_client_events_param(&ctx, "Tilde~Tilde");
    x.test_client_events_param(&ctx, "Tilde~~ Bar~/ Space~_");

    ctx.unregister("");
    check_error(&ctx);
}

impl X {
    fn new() -> X {
        X {
            name: Rc::new(Cell::new(String::new())),
        }
    }

    fn test_client_events_param(&self, ctx: &WasmClientContext, name: &str) {
        let f = testwasmlib::ScFuncs::trigger_event(ctx);
        f.params.name().set_value(name);
        f.params.address().set_value(&ctx.current_chain_id().address());
        f.func.post();
        check_error(&ctx);

        ctx.wait_request();
        check_error(&ctx);

        ctx.wait_event();
        check_error(&ctx);

        let x = self.name.clone();
        assert!(name == x.take());
    }
}
