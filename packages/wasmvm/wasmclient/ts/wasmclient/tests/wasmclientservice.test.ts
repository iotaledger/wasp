import {WasmClientContext, WasmClientService} from '../lib';
import * as testwasmlib from "testwasmlib";
import {bytesFromString} from "wasmlib";
import {KeyPair} from "../lib/isc";
import * as net from "net";
import {WebSocket} from "ws"

const MYCHAIN = "tst1pqsaz75y2wp66f3qmvez2fv9jtsvgwxgnsytv3uxlzpj9uc2n6zwyahepcw";
const MYSEED = "0x925d2270b9088c46b91d124b3de2b6731e75aaa1296e75d16130040e505f6d87";

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

function waitForSocketState(socket: WebSocket, state: number) : Promise<void> {
    return new Promise(function (resolve) {
        setTimeout(function () {
            if (socket.readyState === state) {
                resolve();
            } else {
                waitForSocketState(socket, state).then(resolve);
            }
        }, 5);
    });
}

describe('wasmclient', function () {
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
            client.on('data', function(data: any) {
                console.log('Received: ' + data.toString());
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
                webSocket.close();
            });
            webSocket.addEventListener('error', (x: any) => {
                console.log("Error " + x);
            });
            webSocket.addEventListener('message', (x: any) => {
                console.log("Message " + x);
             });
            webSocket.addEventListener('close', () => {
                console.log("Close");
            });
            await waitForSocketState(webSocket, webSocket.OPEN);
            await waitForSocketState(webSocket, webSocket.CLOSED);
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
            expect(rnd != 0n).toBeTruthy();
        });
    });
});
