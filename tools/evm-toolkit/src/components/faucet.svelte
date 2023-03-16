<script lang="ts">
  import { IotaWallet } from './faucet/iota_wallet';
  import { SendFundsTransaction } from './faucet/send_funds_transaction';
  import { selectedNetwork, nodeClient, indexerClient } from '../store';
  import { Bech32AddressLength, EVMAddressLength } from '../lib/constants';
  import { Input } from '.';
  import Button from './button.svelte';
  import { NotificationType, showNotification } from '$lib/notification';

  let isSendingFunds: boolean;
  let errorMessage: string;

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

    errorMessage = undefined;
    isSendingFunds = true;

    let wallet: IotaWallet = new IotaWallet(
      $nodeClient,
      $indexerClient,
      $selectedNetwork.faucetEndpoint,
    );

    let toastId: number;

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
      label="Your EVM Address"
      bind:value={evmAddress}
      stretch
    />

    <!-- TODO: when we add notification manager we should replace this with a notification. -->
    {#if errorMessage}
      <div class="error">
        <div class="error_title">Error</div>
        <div class="error_message">
          {errorMessage}
        </div>
      </div>
    {/if}

    <Button
      title="Send funds"
      disabled={!enableSendFunds}
      onClick={sendFunds}
      busy={isSendingFunds}
    />
  {/if}
</faucet-component>

<style>
  .error {
    background-color: #9e534a47;
    border: 2px solid #991c0d78;
    border-radius: 10px;
    padding: 15px;
  }

  .error_title {
    font-weight: bold;
    margin-bottom: 15px;
  }

  component {
    color: rgba(255, 255, 255, 0.87);
    display: flex;
    flex-direction: column;
  }
</style>
