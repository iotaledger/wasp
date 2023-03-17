<script lang="ts">
  import { PopupId } from '$lib/popup';
  import { openPopup } from '$lib/popup/actions';
  import { handleEnterKeyDown } from '$lib/utils';
  import { NetworkSettings } from '.';
  import { selectedAccount, connected } from 'svelte-web3';
  import { Button, AccountButton } from '$components';
  import { truncateText } from '$lib/utils';

  function handleSettings() {
    openPopup(PopupId.Settings, {
      component: NetworkSettings,
    });
  }
  function handleAccount() {
    openPopup(PopupId.Account, {
      account: $selectedAccount,
      actions: [],
    });
  }
</script>

<nav class="flex justify-between items center">
  <image-wrapper class="h-full flex items-center">
    <img src="/logo.svg" alt="Shimmer logo" />
    <span class="text-24 ml-4 text-white font-semibold">shimmer</span>
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
      <Button onClick={handleAccount} title="Connect wallet" />
    {:else}
      <AccountButton
        title={truncateText($selectedAccount)}
        onClick={handleAccount}
      />
    {/if}
  </items-wrapper>
</nav>
