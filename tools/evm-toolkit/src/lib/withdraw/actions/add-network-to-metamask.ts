import { get } from 'svelte/store';

import { selectedNetwork } from '$lib/evm-toolkit';
import { NotificationType, showNotification } from '$lib/notification';

export async function addSelectedNetworkToMetamask(): Promise<void> {
    const { ethereum } = window as any;
    const $selectedNetwork = get(selectedNetwork);
    if ($selectedNetwork) {
        if (ethereum && ethereum.isMetaMask) {
            try {
                await ethereum.request({
                    method: 'wallet_addEthereumChain',
                    params: [
                        {
                            chainId: `0x${$selectedNetwork.chainID?.toString(16)}`,
                            chainName: $selectedNetwork.text,
                            nativeCurrency: {
                                name: 'SMR',
                                symbol: 'SMR',
                                decimals: 18,
                            },
                            ...($selectedNetwork.networkUrl && { rpcUrls: [$selectedNetwork.networkUrl] }),
                            ...($selectedNetwork.blockExplorer && { blockExplorerUrls: [$selectedNetwork.blockExplorer] }),
                        },
                    ],
                });
            }
            catch (ex) {
                console.error(ex?.message);
                throw new Error(ex?.message);
            }
        } else {
            showNotification({
                type: NotificationType.Warning,
                message: "Could not add the selected network to your wallet. Please add it network manually.",
            });
        }
    } else {
        showNotification({
            type: NotificationType.Warning,
            message: "Please select a network first.",
        });
    }
}
