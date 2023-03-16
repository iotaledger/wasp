import { iscAbi, iscContractAddress } from '$components/withdraw/constants';
import { ISCMagic } from '$components/withdraw/iscmagic/iscmagic';
import { NotificationType, showNotification } from '$lib/notification';
import { defaultEvmStores, selectedAccount, web3 } from 'svelte-web3';
import { get } from 'svelte/store';
import { subscribeBalance } from '.';
import { updateWithdrawStateStore, withdrawStateStore } from '../stores';
import { pollAccount } from './polls';

export async function connectToWallet() {
  updateWithdrawStateStore({ isLoading: true });

  try {
    await defaultEvmStores.setProvider();

    const evmChainID = await get(web3).eth.getChainId();
    updateWithdrawStateStore({ evmChainID });

    const EthContract = get(web3).eth.Contract;
    const contract = new EthContract(iscAbi, iscContractAddress, {
      from: get(selectedAccount),
    });

    updateWithdrawStateStore({ contract });

    const iscMagic = new ISCMagic(get(withdrawStateStore).contract, null);
    updateWithdrawStateStore({ iscMagic });

    await pollAccount();
    await subscribeBalance();
  } catch (ex) {
    showNotification({
      type: NotificationType.Error,
      message: `Failed to connect to wallet: ${ex.message}`,
    });
    console.error('Failed to connect to wallet: ', ex.message);
  }

  updateWithdrawStateStore({ isLoading: false });
}
