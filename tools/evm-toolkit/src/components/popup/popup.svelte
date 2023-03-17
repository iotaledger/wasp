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
<popup-component>
  <popup-overlay on:click|stopPropagation={handleClose} on:keydown />
  <popup-main>
    <popup-header>
      <span class="capitalize">{title}</span>
      <Button onClick={handleClose} title="x" secondary />
    </popup-header>
    <popup-body>
      <slot />
    </popup-body>
    {#if actions?.length > 0}
      <popup-footer>
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
</popup-component>

<style lang="scss">
  popup-component {
    @apply fixed top-0 left-0 z-50;
    @apply w-screen h-screen;
    @apply flex flex-col justify-center items-center;
  }
  popup-overlay {
    @apply absolute top-0 left-0;
    @apply w-screen h-screen;
    @apply bg-black;
    @apply bg-opacity-10;
    @apply backdrop-blur-lg;
  }
  popup-main {
    @apply p-6;
    @apply flex flex-col;
    @apply w-11/12 md:w-full;
    @apply bg-shimmer-background;
    @apply rounded-xl;
    @apply z-50;
    max-width: 500px;
    popup-header {
      @apply flex justify-between items-center;
      @apply pb-4;
      @apply text-2xl font-semibold;
    }

    popup-footer {
      @apply flex justify-end space-x-4;
      @apply py-4;
    }
  }
</style>
