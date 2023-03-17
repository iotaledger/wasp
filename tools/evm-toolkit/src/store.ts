import { writable } from 'svelte/store';
import type { NetworkOption } from '$lib/network_option';
import { SingleNodeClient, IndexerPluginClient } from '@iota/iota.js';
import { BrowserPowProvider } from '@iota/pow-browser.js';

export const networks = writable<NetworkOption[]>([]);
export const selectedNetwork = writable<NetworkOption>();

selectedNetwork.subscribe(network => {
  if (!network) {
    return;
  }

  console.log(`Creating new client for: ${network.apiEndpoint}`);
  const client = new SingleNodeClient(network.apiEndpoint, {
    powProvider: new BrowserPowProvider(),
  });
  nodeClient.set(client);
  indexerClient.set(new IndexerPluginClient(client));
});

export const indexerClient = writable<IndexerPluginClient>();
export const nodeClient = writable<SingleNodeClient>();
