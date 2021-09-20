<script lang="ts">
  import { MultipleSelector, BarSelector } from './';

  export let type: string;

  const types = {
    multiple: MultipleSelector,
    bar: BarSelector,
  };
</script>

{#if types[type]}
  <svelte:component this={types[type]} {...$$restProps} />
{/if}
