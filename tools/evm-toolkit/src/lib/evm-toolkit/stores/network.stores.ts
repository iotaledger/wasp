import { derived, type Readable, type Writable } from 'svelte/store';

import { persistent } from '$lib/stores';

import { NETWORKS } from '../constants';
import type { INetwork } from '../interfaces';

const SELECTED_NETWORK_KEY = 'selectedNetworkId';
const NETWORKS_KEY = 'networks';

export const networks: Writable<INetwork[]> = persistent<INetwork[]>(NETWORKS_KEY, NETWORKS);
export const selectedNetworkId: Writable<number> = persistent(
    SELECTED_NETWORK_KEY,
    0,
);

export const selectedNetwork: Readable<INetwork> = derived(
    ([networks, selectedNetworkId]), ([$networks, $selectedNetworkId]) => {
        if (!$networks?.length || !($selectedNetworkId >= 0)) {
            return null;
        }
        return $networks.find(network => network.id === $selectedNetworkId);
    }
);

export function updateNetwork(network: INetwork) {
    networks.update($networks => {
        const index = $networks.findIndex(_network => _network?.id === network?.id);
        if (index !== -1) {
            $networks[index] = network;
        }
        return $networks;
    })
}
