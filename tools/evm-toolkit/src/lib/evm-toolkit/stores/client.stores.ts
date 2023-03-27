import { IndexerPluginClient, SingleNodeClient } from '@iota/iota.js';
import { BrowserPowProvider } from '@iota/pow-browser.js';
import { writable } from 'svelte/store';

import { selectedNetwork } from '.';

export const indexerClient = writable<IndexerPluginClient>();
export const nodeClient = writable<SingleNodeClient>();

selectedNetwork?.subscribe(network => {
    if (!network) {
        return;
    } else {
        console.log(`Creating new client for: ${network?.apiEndpoint}`);
        const client = new SingleNodeClient(network.apiEndpoint, {
            powProvider: new BrowserPowProvider(),
        });
        nodeClient?.set(client);
        indexerClient?.set(new IndexerPluginClient(client));
    }
});
