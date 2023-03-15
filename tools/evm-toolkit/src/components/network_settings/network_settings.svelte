<script lang="ts">
  import { onMount } from 'svelte';
  import { selectedNetwork, networks } from '../../store';
</script>

<component>
  {#if $selectedNetwork}
    <div class="input_container">
      <span class="header">Network</span>
      <select bind:value={$selectedNetwork}>
        {#each $networks as network}
          <option value={network}>
            {network.text}
          </option>
        {/each}
      </select>
    </div>

      <div class="input_container">
        <span class="header">Hornet API endpoint</span>
        <input type="text" bind:value={$selectedNetwork.apiEndpoint} disabled={$selectedNetwork.id !== 1} />
      </div>

      <div class="input_container">
        <span class="header">Faucet API endpoint</span>
        <input type="text" bind:value={$selectedNetwork.faucetEndpoint} disabled={$selectedNetwork.id !== 1} />
      </div>

      <div class="input_container">
        <span class="header">Chain Address</span>
        <input type="text" bind:value={$selectedNetwork.chainAddress} disabled={$selectedNetwork.id !== 1} />
      </div>
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
