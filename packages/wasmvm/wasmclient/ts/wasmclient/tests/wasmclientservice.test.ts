import { WasmClientService } from '../lib/wasmclientservice';
import * as testwasmlib from "testwasmlib";

describe('wasmclient', function () {

    describe('Create service', function () {
        it('should create service', () => {
            const client = WasmClientService.DefaultWasmClientService();
            expect(client.Err() == null).toBeTruthy();
        });
    });

    describe('Create SC func', function () {
        it('should create SC func', () => {
            const n = testwasmlib.HScName;
            expect(n == testwasmlib.HScName).toBeTruthy();
        });
    });
});
