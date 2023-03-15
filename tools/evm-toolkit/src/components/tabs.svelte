<script lang="ts">
  import type { ITab } from '$lib/interfaces';

  export let tabs: ITab[];
  export let activeTabValue: number = 1;

  const handleClick = (tabValue: number) => () => (activeTabValue = tabValue);
</script>

<tabs-wrapper>
  {#each tabs as tab}
    <tab-item class:active={activeTabValue === tab.value}>
      <tab-label
        on:click={handleClick(tab.value)}
        on:keypress={handleClick(tab.value)}
      >
        {tab.label}
      </tab-label>
    </tab-item>
  {/each}
</tabs-wrapper>

{#each tabs as tab}
  {#if activeTabValue == tab.value}
    <svelte:component this={tab.component} />
  {/if}
{/each}

<style lang="scss">
  tabs-wrapper {
    @apply flex;
    @apply flex-wrap;
    @apply list-none;
  }

  tab-item {
    @apply flex-grow-0;
    @apply py-2 px-6;
    &.active {
      @apply border-b-2;
      @apply border-b-shimmer-action-primary;
    }
  }

  tab-label {
    @apply font-semibold;
    @apply flex;
    &:hover {
      @apply cursor-pointer;
    }
  }
</style>
