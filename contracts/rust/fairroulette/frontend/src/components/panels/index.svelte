<script lang="ts">
  import { GeneralPanel, ValuePanel, DetailsPanel } from "./";

  export let type: string;

  const types = {
    general: GeneralPanel,
    details: DetailsPanel,
    value: ValuePanel,
  };
</script>

{#if types[type]}
  <svelte:component this={types[type]} {...$$restProps} />
{/if}
