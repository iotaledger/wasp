<script lang="ts">
  import { fade } from 'svelte/transition';
  import type { Notification } from '../lib/notifications';

  export let title: string | undefined = undefined;
  export let message: string;
  export let type: Notification;
  export let id: string = '';
  export let onClose: () => void;
</script>

<div in:fade out:fade class={`toast ${type}`}>
  <div class="toast-content">
    {#if title}
      <div class="title">{title}</div>
    {/if}
    <div class="message">{message}</div>
  </div>
  <button on:click={onClose} class="close">Dismiss</button>
</div>

<style lang="scss">
  .toast {
    border: 1px solid rgba(255, 255, 255, 0.12);
    box-sizing: border-box;
    border-radius: 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 16px;
    font-family: 'DM Sans', sans-serif;
    @media (min-width: 600px) {
      padding: 30px;
    }
    &.error {
      background: rgba(238, 91, 77, 0.2);
    }

    &.win {
      background: rgba(0, 224, 202, 0.6);
    }
    &.info {
      background: rgba(0, 224, 202, 0.2);
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
      line-height: 140%;
      letter-spacing: 0.5px;
      color: var(--white);
      font-weight: 500;
      @media (min-width: 1024px) {
        letter-spacing: 0.75px;
        width: 50%;
      }
    }
    button {
      font-weight: 500;
      font-size: 12px;
      line-height: 12px;
      background: transparent;
      border: none;
      letter-spacing: 0.75px;
      font-weight: 500;

      color: var(--white);
    }
  }
</style>
