import {WasmClientContext, WasmClientService} from '../lib';
import * as testwasmlib from "testwasmlib";
import {bytesFromString, hexEncode} from "wasmlib";
import {KeyPair} from "../lib/isc";
import * as net from "net";

var nano = require('nanomsg');

const MYCHAIN = "tst1pqqf4qxh2w9x7rz2z4qqcvd0y8n22axsx82gqzmncvtsjqzwmhnjs438rhk";
const MYSEED = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3";

function setupClient() {
    const svc = new WasmClientService('127.0.0.1:9090', '127.0.0.1:5550');
    const ctx = new WasmClientContext(svc, MYCHAIN, "testwasmlib");
    ctx.signRequests(KeyPair.fromSubSeed(bytesFromString(MYSEED), 0n));
    expect(ctx.Err == null).toBeTruthy();
    return ctx;
}

function waitForPortState(socket: net.Socket, state: string, msec: number) : Promise<void> {
    return new Promise(function (resolve) {
        setTimeout(function () {
            if (socket.readyState === state || msec == 0) {
                resolve();
            } else {
                waitForPortState(socket, state, msec - 100).then(resolve);
            }
        }, 100);
    });
}

function waitForSocketState(socket: WebSocket, state: number, msec: number) : Promise<void> {
    return new Promise(function (resolve) {
        setTimeout(function () {
            if (socket.readyState === state || msec == 0) {
                resolve();
            } else {
                waitForSocketState(socket, state,msec - 100).then(resolve);
            }
        }, 5);
    });
}

var message = false;

function waitForNanoState(msec: number) : Promise<void> {
    return new Promise(function (resolve) {
        setTimeout(function () {
            if (message || msec == 0) {
                resolve();
            } else {
                waitForNanoState(msec - 100).then(resolve);
            }
        }, 5);
    });
}

describe('wasmclient unverified', function () {
    describe('Create nanomsg listener', function () {
        test('should connect to 127.0.0.1:5550', async () => {
            console.log('Starting');
            const ctx = setupClient();

            var sub = nano.socket('sub');

            var addr = 'tcp://127.0.0.1:5550'
            sub.connect(addr);

            sub.on('error', function (err:any) {
                console.log('Error: ' + err);
                sub.close();
            });

            sub.on('data', function (buf:any) {
                console.log('Data: ' + buf);
                const msg = buf.toString().split(' ');
                if (msg[0] != 'contract') {
                    return;
                }
                message = true;
                sub.close();
            });

            sub.on('close', function () {
                console.log('Close');
            });

            // await waitForPortState(client, "open", 1000);
           //  console.log('Wait state: ' + client.readyState);

            // get new triggerEvent interface, pass params, and post the request
            const f = testwasmlib.ScFuncs.triggerEvent(ctx);
            f.params.name().setValue("Lala");
            f.params.address().setValue(ctx.currentChainID().address());
            f.func.post();
            expect(ctx.Err == null).toBeTruthy();

            await waitForNanoState(2000);

            console.log('Stopping');
        });
    });

    describe('Create TCP listener', function () {
        test('should connect to 127.0.0.1:5550', async () => {
            console.log('Starting');
            const ctx = setupClient();

            const client = net.createConnection(5550, '127.0.0.1', function() {
                console.log('Connected');
                console.log('State: ' + client.readyState);
            });

            client.on('error',function(err: any) {
                console.log('Error: ' + err);
                console.log('State: ' + client.readyState);
            } );

            let msgs = 0;
            client.on('data', function(data: Uint8Array) {
                console.log('Received: ' + hexEncode(data));
                console.log('State: ' + client.readyState);
                msgs++;
                if (msgs == 2) {
                    client.end();
                }
            });

            client.on('close', function() {
                console.log('Connection closed');
                console.log('State: ' + client.readyState);
            });

            client.on('end', function() {
                console.log('Connection ended');
                console.log('State: ' + client.readyState);
            });
            await waitForPortState(client, "open", 1000);
            console.log('Wait state: ' + client.readyState);

            // get new triggerEvent interface, pass params, and post the request
            const f = testwasmlib.ScFuncs.triggerEvent(ctx);
            f.params.name().setValue("Lala");
            f.params.address().setValue(ctx.currentChainID().address());
            f.func.post();
            expect(ctx.Err == null).toBeTruthy();

            await waitForPortState(client, "closed", 2000);
            console.log('Wait state: ' + client.readyState);

            console.log('Stopping');
        });
    });

    describe('Create web socket listener', function () {
        test('should connect to wasp chain websocket', async () => {
            console.log('Starting');
            const webSocketUrl = "ws://127.0.0.1:9090/chain/" + MYCHAIN + "/ws";
            const webSocket = new WebSocket(webSocketUrl);
            webSocket.addEventListener('open', () => {
                console.log("Open");
            });
            webSocket.addEventListener('error', (x: any) => {
                console.log("Error " + x.toString());
                console.log(x);
                webSocket.close();
            });
            webSocket.addEventListener('message', (x: MessageEvent<any>) => {
                console.log("Message " + x);
             });
            webSocket.addEventListener('close', () => {
                console.log("Close");
            });
            await waitForSocketState(webSocket, webSocket.OPEN, 2000);
            await waitForSocketState(webSocket, webSocket.CLOSED, 2000);
            console.log('Stopping');
       });
    });

    describe('Event handling', function () {
        it('should receive events', () => {
            console.log('Starting');
            const ctx = setupClient();

            const events = new testwasmlib.TestWasmLibEventHandlers();
            let name = "";
            events.onTestWasmLibTest((e) => {
                name = e.name;
            })
            ctx.register(events);

            // get new triggerEvent interface, pass params, and post the request
            const f = testwasmlib.ScFuncs.triggerEvent(ctx);
            f.params.name().setValue("Lala");
            f.params.address().setValue(ctx.currentChainID().address());
            f.func.post();
            expect(ctx.Err == null).toBeTruthy();

            ctx.waitRequest();
            expect(ctx.Err == null).toBeTruthy();

            // make sure we wait for the event to show up
            ctx.waitEvent(() => {
                expect(ctx.Err == null).toBeTruthy();

//                expect(name == "Lala").toBeTruthy();
            });
        });
    });
});

describe('wasmclient verified', function () {
    describe('Create service', function () {
        it('should create service', () => {
            const client = WasmClientService.DefaultWasmClientService();
            expect(client != null).toBeTruthy();
        });
    });

    describe('Create SC func', function () {
        it('should create SC func', () => {
            const n = testwasmlib.HScName;
            expect(n == testwasmlib.HScName).toBeTruthy();
        });
    });

    describe('Call web API', function () {
        it('should call web API', () => {
            const ctx = setupClient();

            const v = testwasmlib.ScFuncs.getRandom(ctx);
            v.func.call();
            expect(ctx.Err == null).toBeTruthy();
            const rnd = v.results.random().value();
            console.log("Rnd: " + rnd);
            expect(rnd != 0n).toBeTruthy();
        });
    });

    describe('Post web API', function () {
        it('should post to web API', () => {
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
            console.log("Rnd: " + rnd);
            expect(rnd != 0n).toBeTruthy();
        });
    });
});
