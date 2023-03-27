<script lang="ts">
  import { onMount } from 'svelte';

  import { Navbar, NotificationManager, PopupManager } from '$components';

  import { networks, NETWORKS } from '$lib/evm-toolkit';

  import '../app.scss';

  onMount(() => {
    // Initialize networks updating not-configurable data
    networks?.update($networks => {
      const updatedNetworks = $networks.map(network => {
        const matchedDefaultNetwork = NETWORKS.find(
          _network => _network?.id === network?.id,
        );
        if (network?.id === 0) {
          return matchedDefaultNetwork;
        } else {
          return network;
        }
      });
      return updatedNetworks;
    });
  });
</script>

<Navbar />
<main class="w-full flex flex-1 items-center justify-center">
  <background-decorator>
    <div />
  </background-decorator>
  <slot />
</main>
<PopupManager />
<NotificationManager />

<style lang="scss">
  main {
    @apply w-full flex flex-1 items-center justify-center;
    min-height: calc(100vh - (var(--navbar-height)));
  }
  background-decorator {
    @apply bg-shimmer-background;
    @apply w-screen h-screen fixed z-[-1] top-0 left-0;

    div {
      @apply w-screen h-screen;
      @apply bg-no-repeat bg-cover;
      background-position: 50% center;
      background-image: url('/bg-shapes.svg');
    }
  }
</style>
