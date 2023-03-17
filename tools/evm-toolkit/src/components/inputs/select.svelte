<script lang="ts">
  import type { NetworkOption } from '$lib/network_option';

  export let options: NetworkOption[] | string[] = [];
  export let displayValue = (option: NetworkOption) => option.text;
  export let index: number = 0;
  export let value: NetworkOption | string;

  $: value = options[index];
</script>

<select-component>
  <select bind:value={index}>
    {#each options as option, i}
      {#if typeof option === 'string'}
        <option value={i}>{option}</option>
      {:else}
        <option value={i}>{displayValue(option)}</option>
      {/if}
    {/each}
  </select>
  <select-arrow />
</select-component>

<style lang="scss">
  select-component {
    @apply relative;
    @apply inline-block;
    @apply w-full;
    select {
      @apply inline-block;
      @apply w-full;
      @apply cursor-pointer;
      @apply p-4;
      @apply outline-0;
      @apply border border-shimmer-background-tertiary;
      @apply rounded-lg;
      @apply bg-shimmer-background-tertiary;
      @apply text-white;
      @apply appearance-none;
    }

    select:hover,
    select:focus {
      @apply bg-shimmer-background-tertiary;
    }
    select:disabled {
      @apply opacity-50;
      @apply cursor-not-allowed;
    }
    select-arrow {
      @apply absolute;
      @apply top-6;
      @apply right-5;
      @apply w-0;
      @apply h-0;
      @apply border-t-0;
      @apply border-r-2;
      @apply border-b-2;
      @apply border-l-0;
      @apply inline-block;
      @apply p-1;
      @apply rotate-45;
    }
  }
</style>
