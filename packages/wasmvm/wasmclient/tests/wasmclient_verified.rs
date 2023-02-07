use wasmclient::{self, isc::keypair, wasmclientcontext::*, wasmclientservice::*};

const MYCHAIN: &str = "atoi1pr2y08306jpza3fd6d3fzgzggpy88jtqvkyjjt05ulv048ukry2xwf8hx4v";
const MYSEED: &str = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3";

fn setup_client() -> wasmclient::wasmclientcontext::WasmClientContext {
    let svc = WasmClientService::new("http://127.0.0.1:19090", "127.0.0.1:15550");
    let mut ctx = WasmClientContext::new(&svc, MYCHAIN, "testwasmlib");
    ctx.sign_requests(&keypair::KeyPair::from_sub_seed(
        &wasmlib::bytes_from_string(MYSEED),
        0,
    ));
    assert!(ctx.err().is_ok());
    return ctx;
}

#[test]
fn call_view() {
    let ctx = setup_client();
    let v = testwasmlib::ScFuncs::get_random(&ctx);
    v.func.call();

    let e = ctx.err();
    if let Err(e) = e.clone() {
        println!("err: {}", e);
    }
    assert!(e.is_ok());
    let rnd = v.results.random().value();
    println!("rnd: {}", rnd);
    assert!(rnd != 0);
}

#[test]
fn post_func_request() {
    let ctx = setup_client();
    let f = testwasmlib::ScFuncs::random(&ctx);
    f.func.post();
    let e = ctx.err();
    if let Err(e) = e.clone() {
        println!("err: {}", e);
    }
    assert!(e.is_ok());

    ctx.wait_request();
    let e = ctx.err();
    if let Err(e) = e.clone() {
        println!("err: {}", e);
    }
    assert!(e.is_ok());

    let v = testwasmlib::ScFuncs::get_random(&ctx);
    v.func.call();

    assert!(ctx.err().is_ok());
    let rnd = v.results.random().value();
    println!("rnd: {}", rnd);
    assert!(rnd != 0);
}

#[test]
fn event_handling() {
    let ctx = setup_client();
    let mut events = testwasmlib::TestWasmLibEventHandlers::new();

    let mut name = String::new();
    // events.on_test_wasm_lib_test(|e|  name = e.name);
    ctx.register(&events);

    let event = || name;

    test_client_events_param(&ctx, "Lala", event);
    test_client_events_param(&ctx, "Trala", event);
    test_client_events_param(&ctx, "Bar|Bar", event);
    test_client_events_param(&ctx, "Bar~|~Bar", event);
    test_client_events_param(&ctx, "Tilde~Tilde", event);
    test_client_events_param(&ctx, "Tilde~~ Bar~/ Space~_", event);

    ctx.unregister(&events);
    assert!(ctx.err().is_ok());
}

fn test_client_events_param(ctx: &WasmClientContext, name: &str, event: fn() -> String) {
    let f = testwasmlib::ScFuncs::trigger_event(ctx);
    f.params.name().set_value(name);
    f.params.address().set_value(&ctx.current_chain_id().address());
    f.func.post();
    assert!(ctx.err().is_ok());

    ctx.wait_request();
    assert!(ctx.err().is_ok());

    ctx.wait_event();
    assert!(ctx.err().is_ok());

    assert!(name == event());
}

