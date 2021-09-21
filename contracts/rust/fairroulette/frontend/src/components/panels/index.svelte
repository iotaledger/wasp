<script lang="ts">
  import { GeneralPanel, BalancePanel, DetailsPanel } from './';

  export let type: string;

  const types = {
    general: GeneralPanel,
    details: DetailsPanel,
    value: BalancePanel,
  };
</script>

{#if types[type]}
  <svelte:component this={types[type]} {...$$restProps} />
{/if}
