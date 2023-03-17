<script lang="ts">
  import { Button, Input } from '$components';

  import { indexerClient, nodeClient, selectedNetwork } from '$lib/../store';
  import { Bech32AddressLength, EVMAddressLength } from '$lib/constants';
  import { IotaWallet, SendFundsTransaction } from '$lib/faucet';
  import { NotificationType, showNotification } from '$lib/notification';

  let isSendingFunds: boolean;

  let balance: bigint = BigInt(0);
  let evmAddress: string = '';

  $: enableSendFunds =
    evmAddress.length == EVMAddressLength &&
    $selectedNetwork != null &&
    $selectedNetwork.chainAddress.length == Bech32AddressLength &&
    !isSendingFunds;

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
        message: 'Funds successfully sent! It may take 10-30 seconds to arive.',
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
</script>

<faucet-component class="flex flex-col space-y-6 mt-6">
  {#if $selectedNetwork}
    <Input
      id="evmAddress"
      label="EVM Address"
      bind:value={evmAddress}
      placeholder="0x..."
      stretch
      autofocus
    />

    <Button
      title="Send funds"
      disabled={!enableSendFunds}
      onClick={sendFunds}
      busy={isSendingFunds}
    />
  {:else}
    <span> Please select a network first. </span>
  {/if}
</faucet-component>

<style lang="scss">
</style>
