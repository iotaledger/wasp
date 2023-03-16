import { defaultEvmStores } from 'svelte-web3';
import { unsubscribeBalance } from './subscriptions';

export async function disconnectWallet(): Promise<void> {
  await defaultEvmStores.disconnect();
  unsubscribeBalance();
}
