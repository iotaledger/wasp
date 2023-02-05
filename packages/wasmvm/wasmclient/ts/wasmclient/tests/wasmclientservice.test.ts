import {WasmClientContext, WasmClientService} from '../lib';
import * as testwasmlib from 'testwasmlib';
import {bytesFromString, bytesToString} from 'wasmlib';
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

describe('keypair tests', function () {
    const mySeed = bytesFromString(MYSEED);
    it('construct proper sub-seed 0', () => {
        const subSeed = KeyPair.subSeed(mySeed, 0n);
        console.log('Seed: ' + bytesToString(subSeed));
        expect(bytesToString(subSeed) == '0x24642f47bd363fbd4e05f13ed6c60b04c8a4cf1d295f76fc16917532bc4cd0af').toBeTruthy();
    });

    it('construct proper sub-seed 1', () => {
        const subSeed = KeyPair.subSeed(mySeed, 1n);
        console.log('Seed: ' + bytesToString(subSeed));
        expect(bytesToString(subSeed) == '0xb83d28550d9ee5651796eeb36027e737f0d79495b56d3d8931c716f2141017c8').toBeTruthy();
    });

    it('should construct a proper pair', () => {
        const pair = new KeyPair(mySeed);
        console.log('Publ: ' + bytesToString(pair.publicKey));
        console.log('Priv: ' + bytesToString(pair.privateKey));
        expect(bytesToString(pair.publicKey) == '0x30adc0bd555d56ed51895528e47dcb403e36e0026fe49b6ae59e9adcea5f9a87').toBeTruthy();
        expect(bytesToString(pair.privateKey) == '0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f330adc0bd555d56ed51895528e47dcb403e36e0026fe49b6ae59e9adcea5f9a87').toBeTruthy();
    });

    it('should construct sub-seed pair 0', () => {
        const pair = KeyPair.fromSubSeed(mySeed, 0n);
        console.log('Publ: ' + bytesToString(pair.publicKey));
        console.log('Priv: ' + bytesToString(pair.privateKey));
        expect(bytesToString(pair.publicKey) == '0x40a757d26f6ef94dccee5b4f947faa78532286fe18117f2150a80acf2a95a8e2').toBeTruthy();
        expect(bytesToString(pair.privateKey) == '0x24642f47bd363fbd4e05f13ed6c60b04c8a4cf1d295f76fc16917532bc4cd0af40a757d26f6ef94dccee5b4f947faa78532286fe18117f2150a80acf2a95a8e2').toBeTruthy();
    });

    it('should construct sub-seed pair 1', () => {
        const pair = KeyPair.fromSubSeed(mySeed, 1n);
        console.log('Publ: ' + bytesToString(pair.publicKey));
        console.log('Priv: ' + bytesToString(pair.privateKey));
        expect(bytesToString(pair.publicKey) == '0x120d2b26fc1b1d53bb916b8a277bcc2efa09e92c95be1a8fd5c6b3adbc795679').toBeTruthy();
        expect(bytesToString(pair.privateKey) == '0xb83d28550d9ee5651796eeb36027e737f0d79495b56d3d8931c716f2141017c8120d2b26fc1b1d53bb916b8a277bcc2efa09e92c95be1a8fd5c6b3adbc795679').toBeTruthy();
    });
});

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
