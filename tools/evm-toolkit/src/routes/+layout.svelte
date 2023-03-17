<script lang="ts">
  import { Navbar, PopupManager, NotificationManager } from '$components';
  import '../app.scss';
  import { onMount } from 'svelte';

  import { NETWORKS } from '$lib/networks';
  import { networks, selectedNetwork } from '../store';

  import '../app.scss';

  onMount(async () => {
    networks.set(NETWORKS);
    selectedNetwork.set(NETWORKS[1]);
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
