<script lang="ts">
  import { Button, Dot, Tooltip } from '$components';

  import {
    copyToClipboard,
    handleEnterKeyDown,
    truncateText,
  } from '$lib/common';
  import { NotificationType, showNotification } from '$lib/notification';
  import { closePopup, PopupId } from '$lib/popup';
  import { connectToWallet, disconnectWallet } from '$lib/withdraw';

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
    <account-box>
      <account-address class="flex items-center justify-between w-full">
        <div class="flex items-center justify-between space-x-2">
          <div><Dot /></div>
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
        <Button title="Disconnect" compact ghost onClick={onDisconnectClick} />
      </account-address>
    </account-box>
  {:else}
    <wallet-box
      on:click={onConnectClick}
      on:keydown={event => handleEnterKeyDown(event, onConnectClick)}
    >
      <div class="flex items-center justify-center space-x-4">
        <span class="text-white font-semibold">Connect your wallet</span>
      </div></wallet-box
    >
  {/if}
</div>

<style lang="scss">
  account-box,
  wallet-box {
    @apply relative;
    @apply w-full;
    @apply flex justify-between;
    @apply bg-shimmer-background-tertiary;
    @apply rounded-xl;
    @apply p-4;
    @apply transition-all duration-200 ease-in-out;
  }
  wallet-box {
    @apply cursor-pointer;
    @apply hover:bg-shimmer-background-tertiary-hover;
  }
</style>
