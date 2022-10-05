<script lang="ts">
  import { IotaWallet } from './iota_wallet';
  import { SendFundsTransaction } from './send_funds_transaction';

  import { toast } from '@zerodevx/svelte-toast'
  import { networkOptions } from '../../../networks';

  const ChainAddressLength: number = 63;
  const EVMAddressLength: number = 42;

  let selectedNetworkOption = networkOptions[0];
  let isSendingFunds: boolean;
  let errorMessage: string;

  let balance: bigint = BigInt(0);
  let evmAddress: string = '';

  $: enableSendFunds =
    evmAddress.length == EVMAddressLength &&
    selectedNetworkOption != null &&
    selectedNetworkOption.chainAddress.length == ChainAddressLength &&
    !isSendingFunds;

  async function sendFunds() {
    if (!enableSendFunds) {
      return;
    }

    errorMessage = undefined;
    isSendingFunds = true;

    let wallet: IotaWallet = new IotaWallet(
      selectedNetworkOption.apiEndpoint,
      selectedNetworkOption.faucetEndpoint
    );

    let toastId: number;

    try {
      toastId = toast.push('Initializing wallet');
      await wallet.initialize();
      toast.pop(toastId)

      toastId = toast.push('Requesting funds from the faucet', {
        duration: 20*2000, // 20 retries, 2s delay each.
      });
      balance = await wallet.requestFunds();
      toast.pop(toastId)

      toastId = toast.push('Sending funds');
      const transaction = new SendFundsTransaction(wallet);
      await transaction.sendFundsToEVMAddress(
        evmAddress,
        selectedNetworkOption.chainAddress,
        balance,
        BigInt(500000)
      );
      toast.pop(toastId);

      toast.push('Funds successfully sent! It may take 10-30 seconds to arive.', {
        duration: 10 * 1000
      });
    } catch (ex) {
      errorMessage = ex;
      toast.pop(toastId);
      toast.push(ex.message);
    }

    isSendingFunds = false;
  }
</script>

<component>
  <div class="input_container">
    <span class="header">Network</span>
    <select bind:value={selectedNetworkOption}>
      {#each networkOptions as network}
        <option value={network} >
          {network.text}
        </option>
      {/each}
    </select>
  </div>

  {#if selectedNetworkOption.id == 1}
    <div class="input_container">
      <span class="header">Hornet API endpoint</span>
      <input type="text" bind:value={selectedNetworkOption.apiEndpoint} />
    </div>

    <div class="input_container">
      <span class="header">Faucet API endpoint</span>
      <input type="text" bind:value={selectedNetworkOption.faucetEndpoint} />
    </div>
  {/if}

  <div class="input_container">
    <span class="header">Chain Address</span>
    <input type="text" bind:value={selectedNetworkOption.chainAddress} />
  </div>

  <div class="input_container">
    <span class="header">EVM Address</span>
    <input type="text" bind:value={evmAddress} />
  </div>

  {#if errorMessage}
    <div class="input_container">
      <div class="error">
        <div class="error_title">Error</div>
        <div class="error_message">
          {errorMessage}
        </div>
      </div>
    </div>
  {/if}

  <div class="input_container">
    <button class="button" disabled={!enableSendFunds} on:click={sendFunds}>
      {#if !isSendingFunds}
        Send funds
      {:else}
        Sending ..
      {/if}
    </button>
  </div>
</component>
<style>
  .error {
    background-color: #9e534a47;
    border: 2px solid #991c0d78;
    border-radius: 10px;
    padding: 15px;
    max-width: 386px;
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
