<script lang="ts">
  import type { ToastType } from './../lib/utils';

  export let title: string;
  export let message: string;
  export let type: ToastType;

  let show: boolean = true;
</script>

{#if show}
  <div class={`toast ${type}`}>
    <div class="toast-content">
      <div class="title">{title}</div>
      <div class="message">{message}</div>
    </div>
    <button
      on:click={() => {
        show = false;
      }}
      class="close">X</button
    >
  </div>
{/if}

<style lang="scss">
  .toast {
    display: flex;
    justify-content: space-between;
    align-items: center;

    &.error {
      background: tomato;
    }

    &.win {
      background: turquoise;
    }
  }
</style>
