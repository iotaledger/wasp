import { writable, type Writable } from 'svelte/store';

import type { WithdrawState } from './interfaces';

export const withdrawStateStore: Writable<WithdrawState> = writable({
  availableBaseTokens: 0,
  availableNativeTokens: [],
  availableNFTs: [],
  contract: undefined,
  evmChainID: 0,

  balancePollingHandle: undefined,
  isMetamaskConnected: false,
  isLoading: true,
});

export function updateWithdrawStateStore(keys: {}) {
  withdrawStateStore.update(_withdrawStateStore => ({
    ..._withdrawStateStore,
    ...keys,
  }));
}
