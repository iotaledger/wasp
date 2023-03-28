import { clearIntervalAsync, setIntervalAsync } from 'set-interval-async';
import { get } from 'svelte/store';

import { updateWithdrawStateStore, withdrawStateStore } from '../stores';
import { pollAccount } from './polls';

export async function subscribeBalance() {
  await unsubscribeBalance();
  const $withdrawStateStore = get(withdrawStateStore);
  if ($withdrawStateStore.balancePollingHandle) {
    return;
  }
  await pollAccount();
  updateWithdrawStateStore({
    balancePollingHandle: setIntervalAsync(pollAccount, 2500),
  });
}

export async function unsubscribeBalance() {
  const $withdrawStateStore = get(withdrawStateStore);

  if (!$withdrawStateStore.balancePollingHandle) {
    return;
  }

  await clearIntervalAsync($withdrawStateStore.balancePollingHandle);
  updateWithdrawStateStore({
    balancePollingHandle: undefined,
  });
}
