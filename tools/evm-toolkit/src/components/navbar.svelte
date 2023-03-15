<script lang="ts">
  import { PopupId } from '$lib/popup';
  import { openPopup } from '$lib/popup/actions';
  import { closePopup } from '$lib/popup/actions';
  import { handleEnterKeyDown } from '$lib/utils';
  import { NetworkSettings } from '.';

  function handleSettings() {
    function _closeSettings() {
      closePopup(PopupId.Settings);
    }
    openPopup(PopupId.Settings, {
      title: 'Network settings',
      component: NetworkSettings,
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
  </items-wrapper>
</nav>
