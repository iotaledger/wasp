<script lang="ts">
  import { Button, Tooltip } from '$components';
  import { closePopup, PopupId } from '$lib/popup';
  import {
    copyToClipboard,
    handleEnterKeyDown,
    truncateText,
  } from '$lib/utils';
  import { connectToWallet, disconnectWallet } from '$lib/withdraw';
  import { NotificationType, showNotification } from '$lib/notification';

  export let account = undefined;

  let showCopiedTooltip = false;

  async function handleCopyToClipboard(copyValue: string): Promise<void> {
    try {
      await copyToClipboard(copyValue);
      showCopiedTooltip = true;
    } catch (e) {
      console.error(e);
    }
  }

  async function onConnectClick(): Promise<void> {
    try {
      await connectToWallet();
      closePopup(PopupId.Account);
    } catch (e) {
      showNotification({
        type: NotificationType.Error,
        message: e,
      });
      console.error(e);
    }
  }

  async function onDisconnectClick(): Promise<void> {
    try {
      await disconnectWallet();
      closePopup(PopupId.Account);
    } catch (e) {
      showNotification({
        type: NotificationType.Error,
        message: e,
      });
      console.error(e);
    }
  }
</script>

<div class="flex flex-row items-center py-6">
  {#if account}
    <box-item class:connected={account}>
      <account-address class="flex items-center justify-between w-full">
        <div class="dot-primary flex items-center justify-between space-x-2">
          <span class="font-semibold">{truncateText(account)}</span>
          <Tooltip message="Copied" bind:show={showCopiedTooltip}>
            <img
              src="/copy-icon.svg"
              alt="Copy address"
              class="relative cursor-pointer"
              on:click={() => handleCopyToClipboard(account)}
              on:keydown={event =>
                handleEnterKeyDown(event, () => handleCopyToClipboard(account))}
            />
          </Tooltip>
        </div>
        <Button
          title="Disconnect"
          outline
          compact
          onClick={onDisconnectClick}
        />
      </account-address>
    </box-item>
  {:else}
    <box-item
      on:click={onConnectClick}
      on:keydown={event => handleEnterKeyDown(event, onConnectClick)}
      class:disconnected={!account}
    >
      <metamask-box class="flex items-center justify-center space-x-4">
        <img src="/metamask-logo.png" alt="metamask icon" />
        <span class="text-white font-semibold">Connect your wallet</span>
      </metamask-box>
    </box-item>
  {/if}
</div>

<style lang="scss">
  box-item {
    @apply relative;
    @apply w-full;
    @apply flex;
    @apply justify-between;
    @apply bg-shimmer-background-tertiary;
    @apply rounded-xl;
    @apply p-4;
    @apply transition-all;
    @apply duration-200;
    @apply ease-in-out;
    &.connected {
      @apply pl-8;
    }
    &.disconnected {
      @apply cursor-pointer;
      @apply hover:bg-shimmer-background-tertiary-hover;
    }
  }
</style>
