<script lang="ts">
  import { IotaWallet } from './lib/iota_wallet';
  import { SendFundsTransaction } from './lib/send_funds_transaction';

  import { SvelteToast, toast } from '@zerodevx/svelte-toast'
  import { networkOptions } from '../networks';

  const ChainIdLength: number = 63;
  const EVMAddressLength: number = 42;

  let selectedNetworkOption = networkOptions[0];
  let isSendingFunds: boolean;
  let errorMessage: string;

  let balance: bigint = 0n;
  let chainId: string = '';
  let evmAddress: string = '';

  $: enableSendFunds =
    chainId.length == ChainIdLength &&
    evmAddress.length == EVMAddressLength &&
    selectedNetworkOption != null &&
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
        chainId,
        balance,
        50000000n
      );
      toast.pop(toastId);
    } catch (ex) {
      errorMessage = ex;
      toast.pop(toastId);
      toast.push(ex.message);
    }

    isSendingFunds = false;
  }
</script>

<main>
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
    <span class="header">Chain ID</span>
    <input type="text" bind:value={chainId} />
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
</main>
<SvelteToast/>
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

  button {
    background: rgba(16, 140, 255, 0.12);
    border: 1px solid #57aeff;
    border-radius: 10px;
    cursor: pointer;
    color: rgba(255, 255, 255, 0.87);
    font-size: 1em;
    font-weight: 500;
    padding: 0.6em 1.2em;
    transition: border-color 0.25s;
    width: 100%;
  }

  button:hover {
    border-color: #646cff;
  }

  button:focus,
  button:focus-visible {
    outline: 4px auto -webkit-focus-ring-color;
  }

  button:disabled {
    background-color: #192742;
    border: 1px solid #405985;
    border-radius: 10px;
    color: #9aadce;
  }

  .input_container {
    margin: 15px;
  }

  input,
  select {
    background: #1b2d4b;
    box-sizing: border-box;
    border: 1px solid #ffffff;
    border-radius: 10px;
    color: rgba(255, 255, 255, 0.87);
    padding: 10px;
    width: 100%;
  }

  main {
    border: 1px solid gray;
    border-radius: 25px;
    color: rgba(255, 255, 255, 0.87);
    display: flex;
    flex-direction: column;
    padding: 15px;
    min-width: 450px;
  }
</style>
