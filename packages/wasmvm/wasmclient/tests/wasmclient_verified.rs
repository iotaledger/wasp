use wasmclient::{self, isc::keypair, wasmclientcontext::*, wasmclientservice::*};
use wasmlib::ScViewCallContext;

const MYCHAIN: &str = "atoi1pq48s2dnudsjpsrljlrahkur6250ey26sqxfv8hsmks23lnhkkqmy0h98sl";
const MYSEED: &str = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3";

fn setup_client() -> wasmclient::wasmclientcontext::WasmClientContext {
    let svc = WasmClientService::new("http://127.0.0.1:19090", "127.0.0.1:15550");
    let mut ctx = WasmClientContext::new(&svc, MYCHAIN, "testwasmlib");
    ctx.sign_requests(&keypair::KeyPair::from_sub_seed(
        &wasmlib::bytes_from_string(MYSEED),
        0,
    ));
    assert!(ctx.error.read().unwrap().is_ok());
    return ctx;
}

#[test]
fn call_view() {
    let ctx = setup_client();
    ctx.fn_chain_id();
    let v = testwasmlib::ScFuncs::get_random(&ctx);
    v.func.call();

    let e = ctx.error.read().unwrap();
    if let Err(e) = &*e {
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
    ctx.fn_chain_id();
    let f = testwasmlib::ScFuncs::random(&ctx);
    f.func.post();
    let e = ctx.error.read().unwrap();
    if let Err(e) = &*e {
        println!("err: {}", e);
    }
    assert!(e.is_ok());

    println!("Waiting");
    ctx.wait_request();
    let e = ctx.error.read().unwrap();
    if let Err(e) = &*e {
        println!("err: {}", e);
    }
    assert!(e.is_ok());

    let v = testwasmlib::ScFuncs::get_random(&ctx);
    v.func.call();

    assert!(ctx.error.read().unwrap().is_ok());
    let rnd = v.results.random().value();
    println!("rnd: {}", rnd);
    assert!(rnd != 0);
}

#[test]
fn event_handling() {}

//     describe('event handling', function () {
//         jest.setTimeout(20000);
//         it('should receive multiple events', async () => {
//             const ctx = setupClient();

//             const events = new testwasmlib.TestWasmLibEventHandlers();
//             let name = '';
//             events.onTestWasmLibTest((e) => {
//                 console.log(e.name);
//                 name = e.name;
//             });
//             ctx.register(events);

//             const event = () => name;

//             await testClientEventsParam(ctx, 'Lala', event);
//             await testClientEventsParam(ctx, 'Trala', event);
//             await testClientEventsParam(ctx, 'Bar|Bar', event);
//             await testClientEventsParam(ctx, 'Bar~|~Bar', event);
//             await testClientEventsParam(ctx, 'Tilde~Tilde', event);
//             await testClientEventsParam(ctx, 'Tilde~~ Bar~/ Space~_', event);

//             ctx.unregister(events);
//             expect(ctx.Err == null).toBeTruthy();
//         });
//     });
// });
