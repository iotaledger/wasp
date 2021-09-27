<script lang="ts">
  import { fade } from 'svelte/transition';

  export let title: string;
  export let message: string;
  export let type: 'winner' | 'error';
  export let id: string = '';
  export let timeout: number = undefined;
  export let onClose: () => void;
</script>

<div in:fade out:fade class={`toast ${type}`}>
  <div class="toast-content">
    <div class="title">{title}</div>
    <div class="message">{message}</div>
  </div>
  <button on:click={onClose} class="close"
    ><img src="close.svg" alt="close" /></button
  >
</div>

<style lang="scss">
  .toast {
    border: 1px solid rgba(255, 255, 255, 0.12);
    box-sizing: border-box;
    border-radius: 12px;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 20px 16px;
    @media (min-width: 1024px) {
      padding: 20px 40px;
      align-items: center;
    }
    &.error {
      background: rgba(238, 91, 77, 0.24);
    }

    &.win {
      background: rgba(0, 224, 202, 0.4);
    }
    .title {
      font-family: 'Metropolis Bold';
      font-size: 24px;
      line-height: 120%;
      letter-spacing: 0.02em;
      color: var(--gray-1);
    }
    .message {
      padding-top: 8px;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-3);
      @media (min-width: 1024px) {
        font-size: 16px;
        line-height: 150%;
        letter-spacing: 0.75px;
      }
    }
    button {
      background: transparent;
      border: none;
      color: var(--white);
    }
  }
</style>
