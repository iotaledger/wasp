<script lang="ts">
  import { BALANCE_PANEL_TYPE } from '../../models/IBalancePanel';

  import { ENTRIES_PANEL_TYPE } from '../../models/IEntriesPanel';
  import { WALLET_PANEL_TYPE } from '../../models/IWalletPanel';

  import { WalletPanel, BalancePanel, EntriesPanel } from './';

  export let type: number;

  const types = {
    [WALLET_PANEL_TYPE]: WalletPanel,
    [ENTRIES_PANEL_TYPE]: EntriesPanel,
    [BALANCE_PANEL_TYPE]: BalancePanel,
  };
</script>

{#if types[type]}
  <svelte:component this={types[type]} {...$$restProps} />
{/if}
