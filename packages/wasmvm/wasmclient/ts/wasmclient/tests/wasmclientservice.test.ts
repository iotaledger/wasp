import {WasmClientContext, WasmClientService} from '../lib';
import * as testwasmlib from 'testwasmlib';
import {bytesFromString} from 'wasmlib';
import {KeyPair} from '../lib/isc';

const MYCHAIN = 'atoi1prj5xunmvc8uka9qznnpu4yrhn3ftm3ya0wr2jvurwr209llw7xdyztcr6g';
const MYSEED = '0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3';

function setupClient() {
    const svc = new WasmClientService('127.0.0.1:19090', '127.0.0.1:15550');
    const ctx = new WasmClientContext(svc, MYCHAIN, 'testwasmlib');
    ctx.signRequests(KeyPair.fromSubSeed(bytesFromString(MYSEED), 0n));
    expect(ctx.Err == null).toBeTruthy();
    return ctx;
}

// describe('wasmclient unverified', function () {
// });

describe('wasmclient verified', function () {
    describe('call() view', function () {
        it('should call through web API', () => {
            const ctx = setupClient();

            const v = testwasmlib.ScFuncs.getRandom(ctx);
            v.func.call();
            expect(ctx.Err == null).toBeTruthy();
            const rnd = v.results.random().value();
            console.log('Rnd: ' + rnd);
            expect(rnd != 0n).toBeTruthy();
        });
    });

    describe('post() func request', function () {
        it('should post through web API', () => {
            const ctx = setupClient();

            const f = testwasmlib.ScFuncs.random(ctx);
            f.func.post();
            expect(ctx.Err == null).toBeTruthy();

            ctx.waitRequest();
            expect(ctx.Err == null).toBeTruthy();

            const v = testwasmlib.ScFuncs.getRandom(ctx);
            v.func.call();
            expect(ctx.Err == null).toBeTruthy();
            const rnd = v.results.random().value();
            console.log('Rnd: ' + rnd);
            expect(rnd != 0n).toBeTruthy();
        });
    });

    describe('event handling', function () {
        it('should receive multiple events', async () => {
            const ctx = setupClient();

            const events = new testwasmlib.TestWasmLibEventHandlers();
            let name = '';
            events.onTestWasmLibTest((e) => {
                console.log(e.name);
                name = e.name;
            });
            ctx.register(events);

            const event = () => name;

            await testClientEventsParam(ctx, 'Lala', event);
            await testClientEventsParam(ctx, 'Trala', event);
            await testClientEventsParam(ctx, 'Bar|Bar', event);
            await testClientEventsParam(ctx, 'Bar~|~Bar', event);
            await testClientEventsParam(ctx, 'Tilde~Tilde', event);
            await testClientEventsParam(ctx, 'Tilde~~ Bar~/ Space~_', event);

            ctx.unregister(events);
            expect(ctx.Err == null).toBeTruthy();
        });
    });
});

async function testClientEventsParam(ctx: WasmClientContext, name: string, event: () => string) {
    const f = testwasmlib.ScFuncs.triggerEvent(ctx);
    f.params.name().setValue(name);
    f.params.address().setValue(ctx.currentChainID().address());
    f.func.post();
    expect(ctx.Err == null).toBeTruthy();

    ctx.waitRequest();
    expect(ctx.Err == null).toBeTruthy();

    await ctx.waitEvent();
    expect(ctx.Err == null).toBeTruthy();

    expect(name == event()).toBeTruthy();
}
