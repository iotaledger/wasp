<script lang="ts">
  import { onMount } from 'svelte';
  import { network } from '../../store';

  let networkOptions: any;
  let selectedNetworkOption: any;

  onMount(async () => {
    const networkOptionsFile = await fetch('./networks.json');
    networkOptions = await networkOptionsFile.json();
    selectedNetworkOption = networkOptions[1];
  });

  $: {
    if (selectedNetworkOption && selectedNetworkOption.apiEndpoint) {
      network.set(selectedNetworkOption);
    }
  }
</script>

<component>
  {#if selectedNetworkOption}
    <div class="input_container">
      <span class="header">Network</span>
      <select bind:value={selectedNetworkOption}>
        {#each networkOptions as network}
          <option value={network}>
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

      <div class="input_container">
        <span class="header">Chain Address</span>
        <input type="text" bind:value={selectedNetworkOption.chainAddress} />
      </div>
    {/if}
  {:else}
    <p>Loading Network Config File...</p>
  {/if}
</component>

<style>
  component {
    color: rgba(255, 255, 255, 0.87);
    display: flex;
    flex-direction: column;
  }
</style>
