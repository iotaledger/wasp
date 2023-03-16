<script lang="ts">
  import { Box, Tabs, FaucetNew, Withdraw } from '$components';
  import { NETWORKS } from '$lib/networks';
  import { networks, selectedNetwork } from '../../store';
  import { onMount } from 'svelte';

  onMount(async () => {
    networks.set(NETWORKS);
    selectedNetwork.set(NETWORKS[1]);
  });

  const TABS = [
    {
      label: 'Faucet',
      value: 1,
      // TODO: replace Faucet component with FaucetNew
      component: FaucetNew,
    },
    {
      label: 'Withdraw',
      component: Withdraw,
    },
  ];
</script>

<section class="max-w-2xl w-full">
  <Box>
    <Tabs tabs={TABS} />
  </Box>
</section>
