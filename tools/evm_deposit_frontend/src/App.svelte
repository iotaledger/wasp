<script lang="ts">
  import { IotaWallet } from './lib/iota_wallet';
import { SendFundsTransaction } from './lib/send_funds_transaction';
  const networkOptions = [
    {id: 0, text: "Testnet", apiEndpoint: "https://api.testnet.shimmer.network", faucetEndpoint: "https://faucet.testnet.shimmer.network"},
    {id: 1, text: "Custom settings", default: false, apiEndpoint: "http://localhost:14265", faucetEndpoint: "http://localhost:8091"},
  ];
  let selectedNetworkOption = networkOptions[0];
  let isRequestingFunds: boolean;

  let balance: bigint = 0n;
  let chainId: string = "";
  let evmAddress: string = "";

  async function requestFunds() {
    if (evmAddress.length == 0) {
      return;
    }

    isRequestingFunds = true;

    let wallet: IotaWallet = new IotaWallet(selectedNetworkOption.apiEndpoint, selectedNetworkOption.faucetEndpoint);

    try {
      await wallet.initialize();
      balance = await wallet.requestFunds();
    } catch (ex) {
      console.log(ex);
      return;
    }

    const transaction = new SendFundsTransaction(wallet);

    await transaction.sendFundsToEVMAddress(evmAddress, chainId, balance, 50000000n);
    isRequestingFunds = false;
  }

</script>

<main>
    <div class="border">
      <div class="input_container">
        <select bind:value={selectedNetworkOption}>
          {#each networkOptions as network}
            <option value={network} selected={network.id === 0}>
              {network.text}
            </option>
          {/each}
        </select>
      </div>

      {#if selectedNetworkOption.id == 1}
        <div class="input_container">
          <span class="header">Hornet API endpoint</span>
          <input type="text" bind:value={networkOptions[1].apiEndpoint}>
        </div>

        <div class="input_container">
          <span class="header">Faucet API endpoint</span>
          <input type="text" bind:value={networkOptions[1].faucetEndpoint}>
        </div>
      {/if}

      <div class="input_container">
        <span class="header">Chain ID</span>
        <input type="text" bind:value={chainId}>
      </div>

      <div class="input_container">
        <span class="header">EVM Address</span>
        <input type="text" bind:value={evmAddress}>
      </div>


      <button class="button" on:click="{requestFunds}">Send funds</button>
  </div>
</main>

<style>
  button {
    margin: 10px;
    width: 100%;
  }

  .input_container {
    margin: 15px;
  }

  input, select {
    width: 100%;
    padding: 10px;
  }

  .border {
    display: flex;
    flex-direction:  column;
    border: 1px solid gray;
    padding: 50px;
    border-radius: 25px;
    width: 800px;
  }

</style>