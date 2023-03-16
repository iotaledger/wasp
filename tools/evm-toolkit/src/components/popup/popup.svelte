<script lang="ts">
  import type { PopupAction } from '$lib/popup';
  import { handleEscapeKeyDown } from '$lib/utils';
  import { Button } from '..';

  export let onClose: () => unknown = () => {};
  export let title: string | undefined = undefined;
  export let actions: PopupAction[] = undefined;

  const actionsBusyState: boolean[] = [];

  function handleClose(): void {
    if (!actionsBusyState.some(busy => busy)) onClose();
  }

  async function handleActionClick(
    index: number,
    action: (() => void) | (() => Promise<void>),
  ): Promise<void> {
    if (action) {
      try {
        actionsBusyState[index] = true;
        await action();
        onClose();
      } catch (error) {
        console.error(error);
      } finally {
        actionsBusyState[index] = false;
      }
    }
  }
</script>

<svelte:window on:keydown={event => handleEscapeKeyDown(event, handleClose)} />
<popup-overlay on:click={handleClose} on:keydown />
<popup-main>
  <popup-header>
    <h3 class="capitalize">{title}</h3>
    <Button onClick={handleClose} title="X" secondary />
  </popup-header>
  <popup-body class="p-4">
    <slot />
  </popup-body>
  {#if actions.length > 0}
    <popup-footer class="space-x-4">
      {#each actions as { action, title }, index}
        <Button
          {title}
          onClick={() => handleActionClick(index, action)}
          busy={actionsBusyState[index]}
        />
      {/each}
    </popup-footer>
  {/if}
</popup-main>

<style lang="scss">
  popup-overlay {
    @apply fixed;
    @apply top-0;
    @apply left-0;
    @apply w-full;
    @apply h-full;
    @apply flex;
    @apply flex-col;
    @apply justify-center;
    @apply items-center;
    @apply bg-black;
    @apply bg-opacity-50;
    @apply z-50;
  }
  popup-main {
    @apply absolute;
    @apply top-0;
    @apply left-1/2;
    @apply -translate-x-1/2;
    @apply translate-y-1/4;
    @apply md:translate-y-1/2;
    @apply flex;
    @apply flex-col;
    @apply justify-between;
    @apply w-11/12;
    @apply md:w-full;
    @apply bg-shimmer-background;
    @apply rounded-xl;
    @apply z-50;
    max-width: 500px;
    popup-header {
      @apply flex;
      @apply justify-between;
      @apply items-center;
      @apply p-4;
      @apply text-2xl;
      @apply font-semibold;
      @apply border-b;
      @apply border-shimmer-background-tertiary;
    }

    popup-footer {
      @apply flex;
      @apply justify-end;
      @apply p-4;
      @apply border-t;
      @apply border-solid;
      @apply border-shimmer-background-tertiary;
    }
  }
</style>
