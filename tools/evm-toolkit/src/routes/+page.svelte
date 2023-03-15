<script lang="ts">
  import Faucet from '$components/faucet/Faucet.svelte';
  import Withdraw from '$components/withdraw/Withdraw.svelte';
  import { SvelteToast } from '@zerodevx/svelte-toast';
  import { onMount } from 'svelte';
  import { networks, selectedNetwork } from '../store';
  import NetworkSettings from '$components/network_settings/network_settings.svelte';
  import { NETWORKS } from '$lib/networks';

  onMount(async () => {
    networks.set(NETWORKS);
    selectedNetwork.set(NETWORKS[1]);
  });

</script>

<main>
  <div class="item">
    <h2>Network settings</h2>
    <NetworkSettings />
  </div>
  {#if $selectedNetwork}
    <div class="item">
      <div class="control">
        <h2>Faucet</h2>
        <Faucet />
      </div>

      <div class="seperator" />

      <div class="control">
        <h2>Withdraw</h2>
        <Withdraw />
      </div>
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

  .seperator {
    border: 2px solid gray;
    margin: 50px auto 50px auto;
    width: 50%;
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
