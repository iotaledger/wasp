<script lang="ts">
  import Faucet from './components/faucet/Faucet.svelte';
  import Withdraw from './components/withdraw/Withdraw.svelte';
  import { SvelteToast } from '@zerodevx/svelte-toast';
  import { onMount } from 'svelte';
  import type { NetworkOption } from './lib/network_option';
  import { network } from './store';
    import NetworkSettings from './components/network_settings/network_settings.svelte';
    import { IndexerPluginClient, SingleNodeClient } from '@iota/iota.js';

  onMount(async () => {
    const networkOptionsFile = await fetch('./networks.json');
    const networkOptions: NetworkOption[] = await networkOptionsFile.json();
    network.set(networkOptions[0]);
  });
</script>

<main>
  <div class="item">
    <h2>Network settings</h2>
    <NetworkSettings />
  </div>
  {#if $network}
    <div class="item">
      <h2>Faucet</h2>
      <Faucet />

      <h2>Withdraw</h2>
      <Withdraw />
    </div>

  {/if}
</main>
<SvelteToast />

<style>
  h2 {
    margin: 15px;
  }
  main {
    display: flex;
    flex: 1 1 auto;
    flex-direction: row;
    justify-content: center;
  }

  .item {
    flex-grow: 1;
    max-width: 900px;
  }

  @media (max-width: 900px) {
    main {
      flex-direction: column;
    }
    .item {
      max-width: 450x;
    }
  }
</style>
