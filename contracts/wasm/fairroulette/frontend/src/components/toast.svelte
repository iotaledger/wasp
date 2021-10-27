<script lang="ts">
  import { fade } from 'svelte/transition';
  import { Notification } from '../lib/notifications';

  export let title: string | undefined = undefined;
  export let message: string;
  export let type: Notification;
  export let id: string = '';
  export let onClose: () => void;

  const EMOJIS: {
    [Notification.Info]: string | undefined;
    [Notification.Win]: string | undefined;
    [Notification.Lose]: string | undefined;
    [Notification.Error]: string | undefined;
  } = {
    [Notification.Info]: '🙂',
    [Notification.Win]: '🥳',
    [Notification.Lose]: '😩',
    [Notification.Error]: undefined,
  };
</script>

<div in:fade out:fade class="toast {type} {EMOJIS[type] && 'with-emoji'}">
  <div class="toast-content">
    {#if title}
      <div class="title">{title}</div>
    {/if}
    <div class="message">{message}</div>
  </div>
  {#if EMOJIS[type]}
    <div class="emoji">
      {EMOJIS[type]}
    </div>
  {/if}
  <button on:click={onClose} class="close">
    <svg
      width="14"
      height="14"
      viewBox="0 0 14 14"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M14 1.41L12.59 0L7 5.59L1.41 0L0 1.41L5.59 7L0 12.59L1.41 14L7 8.41L12.59 14L14 12.59L8.41 7L14 1.41Z"
        fill="#F6F8FC"
      />
    </svg>
  </button>
</div>

<style lang="scss">
  .toast {
    border: 1px solid rgba(255, 255, 255, 0.12);
    box-sizing: border-box;
    border-radius: 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 40px 60px 40px 40px;
    font-family: 'DM Sans', sans-serif;
    position: relative;
    &.with-emoji {
      padding: 40px 60px 40px 70px;
    }
    &.error {
      background: rgba(255, 103, 85, 0.95);
    }
    &.info,
    &.win,
    &.lose {
      background: rgba(7, 60, 70, 0.95);
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
      }
    }
    button {
      background: transparent;
      border: none;
      position: absolute;
      top: 24px;
      right: 24px;
    }
    .emoji {
      position: absolute;
      top: 50%;
      transform: translateY(-50%);
      left: 24px;
      font-size: 24px;
    }
  }
</style>
