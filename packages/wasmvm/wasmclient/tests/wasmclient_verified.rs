use std::sync::{Arc, Mutex};
use wasmlib::IEventHandlers;

use wasmclient::{self, isc::keypair, wasmclientcontext::*, wasmclientservice::*};

const MYCHAIN: &str = "atoi1pzyjf9ex6wmw2uk3936vrtq5rc6d8fkltmyvfdwnnkpqxnu6lj5uugva0qj";
const MYSEED: &str = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3";

const PARAMS: &[&str] = &[
    "Lala",
    "Trala",
    "Bar|Bar",
    "Bar~|~Bar",
    "Tilde~Tilde",
    "Tilde~~ Bar~/ Space~_",
];

struct EventProcessor {
    name: Arc<Mutex<String>>,
}

impl EventProcessor {
    fn new() -> EventProcessor {
        EventProcessor {
            name: Arc::new(Mutex::new(String::new())),
        }
    }

    fn send_client_events_param(&self, ctx: &WasmClientContext, param: &str) {
        let f = testwasmlib::ScFuncs::trigger_event(ctx);
        f.params.name().set_value(param);
        f.params.address().set_value(&ctx.current_chain_id().address());
        f.func.post();
        check_error(ctx);
    }

    fn wait_client_events_param(&self, ctx: &WasmClientContext, param: &str) {
        for _ in 0..100 {
            {
                let name = self.name.lock().unwrap();
                if (*name).len() != 0 || ctx.err().is_err() {
                    break;
                }
            }
            std::thread::sleep(std::time::Duration::from_millis(100));
        }
        check_error(ctx);

        let mut name = self.name.lock().unwrap();
        assert_eq!(*name, param);
        name.clear();
    }
}

fn check_error(ctx: &WasmClientContext) {
    if let Err(e) = ctx.err() {
        println!("err: {}", e);
        assert!(false);
    }
}

fn setup_client() -> WasmClientContext {
    let svc = Arc::new(WasmClientService::new("http://localhost:19090", MYCHAIN));
    let mut ctx = WasmClientContext::new(svc.clone(), "testwasmlib");
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
    assert_ne!(rnd, 0);
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
    assert_ne!(rnd, 0);
}

#[test]
fn event_handling() {
    let ctx = setup_client();
    let mut events = testwasmlib::TestWasmLibEventHandlers::new();

    let proc = EventProcessor::new();
    {
        let name = proc.name.clone();
        events.on_test_wasm_lib_test(move |e| {
            let mut name = name.lock().unwrap();
            *name = e.name.clone();
        });
    }
    let events_id = events.id();
    ctx.register(Box::new(events));
    check_error(&ctx);

    for param in PARAMS {
        proc.send_client_events_param(&ctx, &param);
        proc.wait_client_events_param(&ctx, &param);
    }

    ctx.unregister(events_id);
    check_error(&ctx);
}
