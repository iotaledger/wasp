use wasmclient::{self, isc::keypair, wasmclientcontext::*, wasmclientservice::*};
use wasmlib::ScViewCallContext;

const MYCHAIN: &str = "tst1ppx8hf6vl7ak6xk2phxhx3xf6vd2r5zyulkgaf20kmfev9xusy4t2tku6he";
const MYSEED: &str = "0x927fc9c0502ca9acc4a2ae15fabb7248d054b319423862873088c92d9b835c15";

fn setup_client() -> wasmclient::wasmclientcontext::WasmClientContext {
    let svc = WasmClientService::new("http://127.0.0.1:14265", "127.0.0.1:15550");
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
    assert!(ctx.error.read().unwrap().is_ok());
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
    assert!(ctx.error.read().unwrap().is_ok());

    ctx.wait_request();
    assert!(ctx.error.read().unwrap().is_ok());

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
