<script lang="ts">
  import { Box, Tabs, WithdrawNew } from '$components';
  import { FaucetSection } from './sections';

  const TABS = [
    {
      label: 'Faucet',
      value: 1,
      component: FaucetSection,
    },
    {
      label: 'Withdraw',
      // TODO: replace Withdraw component with WithdrawNew
      component: WithdrawNew,
    },
  ];
</script>

<section class="max-w-2xl w-full">
  <Box>
    <Tabs tabs={TABS} />
  </Box>
</section>
