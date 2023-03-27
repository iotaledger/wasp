<script lang="ts">
  import { Box, Tabs } from '$components';
  import { FaucetSection, WithdrawSection } from './sections';

  const TABS = [
    {
      label: 'Faucet',
      value: 1,
      component: FaucetSection,
    },
    {
      label: 'Withdraw',
      component: WithdrawSection,
    },
  ];
</script>

<section class="max-w-2xl w-full">
  <Box>
    <Tabs tabs={TABS} />
  </Box>
</section>
