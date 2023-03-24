<script lang="ts">
  import { connected, selectedAccount } from 'svelte-web3';

  import { Button, Input } from '$components';

  import { Bech32AddressLength, EVMAddressLength } from '$lib/constants';
  import { indexerClient, nodeClient, selectedNetwork } from '$lib/evm-toolkit';
  import { IotaWallet, SendFundsTransaction } from '$lib/faucet';
  import { NotificationType, showNotification } from '$lib/notification';
  import { handleEnterKeyDown } from '$lib/common';

  let isSendingFunds: boolean;

  let balance: bigint = BigInt(0);
  let evmAddress: string = '';

  $: enableSendFunds =
    evmAddress.length == EVMAddressLength &&
    $selectedNetwork != null &&
    $selectedNetwork.chainAddress.length == Bech32AddressLength &&
    !isSendingFunds;
  $: allowUseSelectedAddress =
    $connected && $selectedAccount && $selectedAccount !== evmAddress;

  async function sendFunds() {
    if (!enableSendFunds) {
      return;
    }

    isSendingFunds = true;

    let wallet: IotaWallet = new IotaWallet(
      $nodeClient,
      $indexerClient,
      $selectedNetwork.faucetEndpoint,
    );

    try {
      await wallet.initialize();
      showNotification({
        type: NotificationType.Warning,
        message: 'Initializing wallet',
      });

      balance = await wallet.requestFunds();
      showNotification({
        type: NotificationType.Warning,
        message: 'Requesting funds from the faucet',
        duration: 20 * 2000, // 20 retries, 2s delay each.
      });

      showNotification({
        type: NotificationType.Warning,
        message: 'Sending funds',
      });

      const transaction = new SendFundsTransaction(wallet);
      await transaction.sendFundsToEVMAddress(
        evmAddress,
        $selectedNetwork.chainAddress,
        balance,
        BigInt(5000000),
      );

      showNotification({
        type: NotificationType.Success,
        message:
          'Funds successfully sent! It may take 10-30 seconds to arrive.',
        duration: 10 * 1000,
      });
    } catch (ex) {
      showNotification({
        type: NotificationType.Error,
        message: ex.message,
      });
    }

    isSendingFunds = false;
  }

  function onUseMyAddress() {
    evmAddress = $selectedAccount;
  }
</script>

<faucet-component class="flex flex-col space-y-6 mt-6">
  {#if $selectedNetwork}
    <div class="flex flex-col space-y-2">
      <Input
        id="evmAddress"
        label="EVM Address"
        bind:value={evmAddress}
        placeholder="0x..."
        stretch
        autofocus
      />
      <div
        on:click={onUseMyAddress}
        on:keydown={event => handleEnterKeyDown(event, onUseMyAddress)}
        class="cursor-pointer text-shimmer-action-primary"
        class:opacity-50={!allowUseSelectedAddress}
        class:pointer-events-none={!allowUseSelectedAddress}
      >
        Use my own address
      </div>
    </div>

    <Button
      title="Send funds"
      busyMessage="Sending funds..."
      disabled={!enableSendFunds}
      onClick={sendFunds}
      busy={isSendingFunds}
    />
  {:else}
    <span>Loading Network Configuration...</span>
  {/if}
</faucet-component>

<style lang="scss">
</style>
