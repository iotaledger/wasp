<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import { handleEnterKeyDown } from '$lib/utils';

  export let message: string = null;
  export let show: boolean = false;

  const dispatch = createEventDispatcher();

  let left: number = 0;
  let top: number = 0;
  let parentRef = null;

  let hideTooltipTimeout = null;

  onMount(() => {
    const rect = parentRef.getBoundingClientRect();
    left = rect.left + rect.width / 2;
    top = rect.top;
  });

  $: showTooltip = show && message;

  function handleClick(event) {
    clearTimeout(hideTooltipTimeout);
    dispatch('click', event);
    hideTooltipTimeout = setTimeout(() => {
      show = false;
    }, 2000);
  }
</script>

<tooltip-component
  class="relative"
  bind:this={parentRef}
  on:click={handleClick}
  on:keydown={event => handleEnterKeyDown(event, handleClick)}
>
  <tooltip class:show={showTooltip}>
    {#if showTooltip}
      <tooltip-text>
        {message}
      </tooltip-text>
    {/if}
  </tooltip>
  <slot />
</tooltip-component>

<style lang="scss">
  tooltip {
    @apply opacity-0;
    transition: opacity 0.2s ease-in-out;
    &.show {
      @apply opacity-100;
    }
    tooltip-text {
      @apply text-shimmer-action-primary;
      @apply fixed;
      @apply bg-shimmer-background;
      @apply rounded;
      @apply flex-nowrap;
      @apply text-sm;
      transform: translate(-50%, calc(-100% - 8px));
      padding: 2px 8px;
      flex-wrap: nowrap;
      min-width: max-content;
    }
  }
</style>
