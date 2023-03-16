<script lang="ts">
  import { Navbar, PopupManager } from '$components';
  import '../app.scss';
  import { onMount } from 'svelte';
  import { networks, selectedNetwork } from '../store';
  import { NETWORKS } from '$lib/networks';

  onMount(async () => {
    networks.set(NETWORKS);
    selectedNetwork.set(NETWORKS[1]);
  });

</script>

<Navbar />
<main class="w-full flex items-center justify-center">
  <background-radial />
  <slot />
</main>
<PopupManager />

<style lang="scss">
  background-radial {
    @apply absolute;
    @apply top-0;
    @apply w-full;
    @apply h-full;
    @apply z-[-1];
    background: radial-gradient(
      farthest-corner at top right,
      #00e0c7 -44%,
      #061928 40%,
      #061928 100%
    );
  }
</style>
