import { WasmClientService } from '../lib/wasmclientservice';
import * as testwasmlib from "testwasmlib";
import * as syncRequest from 'ts-sync-request';
import {SyncRequestClient} from "ts-sync-request";

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

    describe('Call web API', function () {
        it('should call web API', () => {
            // define the URL of the API
            const API_URL = 'https://sc.testnet.shimmer.network/chains';

            const client = new SyncRequestClient();
            client.addHeader('Content-Type', 'application/json')
            const response = client.get<string>(API_URL);
            console.log(response);

            // // check if the response is successful
            // if (response.statusCode >= 200 && response.statusCode < 300) {
            //     // parse the response as JSON
            //     const data = JSON.parse(response.getBody());
            //
            //     // use the data from the API
            //     console.log(data);
            // } else {
            //     // throw an error if the response is not successful
            //     throw new Error(response.statusMessage);
            // }
            // const n = testwasmlib.HScName;
            // expect(n == testwasmlib.HScName).toBeTruthy();
        });
    });
});
