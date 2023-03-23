<script lang="ts">
  import { connected, selectedAccount } from 'svelte-web3';

  import { AccountButton, Button } from '$components';

  import { handleEnterKeyDown, truncateText } from '$lib/common';
  import { NotificationType, showNotification } from '$lib/notification';
  import { PopupId } from '$lib/popup';
  import { openPopup } from '$lib/popup/actions';
  import { connectToWallet } from '$lib/withdraw';

  function handleSettings() {
    openPopup(PopupId.Settings);
  }

  function onAccountClick() {
    openPopup(PopupId.Account, {
      account: $selectedAccount,
      actions: [],
    });
  }

  async function onConnectClick(): Promise<void> {
    try {
      await connectToWallet();
    } catch (e) {
      showNotification({
        type: NotificationType.Error,
        message: e,
      });
      console.error(e);
    }
  }
</script>

<nav class="flex justify-between items center">
  <image-wrapper class="h-full flex items-center">
    <img src="/logo.svg" alt="Shimmer logo" />
    <h1 class="text-md md:text-2xl ml-4 text-white font-semibold">shimmer</h1>
  </image-wrapper>
  <items-wrapper class="flex items-center space-x-4 mr-4">
    <img
      src="/settings-icon.svg"
      alt="Shimmer logo"
      class="cursor-pointer p-2"
      on:click={handleSettings}
      on:keydown={event => handleEnterKeyDown(event, handleSettings)}
    />
    {#if !$connected || !$selectedAccount}
      <Button onClick={onConnectClick} title="Connect wallet" />
    {:else}
      <AccountButton
        title={truncateText($selectedAccount)}
        onClick={onAccountClick}
      />
    {/if}
  </items-wrapper>
</nav>
