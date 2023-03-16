<script lang="ts">
  import { Box, Tabs, FaucetNew, WithdrawNew } from '$components';

  const TABS = [
    {
      label: 'Faucet',
      value: 1,
      // TODO: replace Faucet component with FaucetNew
      component: FaucetNew,
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
