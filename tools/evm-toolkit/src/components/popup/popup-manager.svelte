<script lang="ts">
  import { closePopup, PopupId, popupStore } from '$lib/popup'
  import { SettingsPopup, Popup } from '.'

  const COMPONENTS = {
      [PopupId.Settings]: SettingsPopup,
  }

  function getTitle(popupId: PopupId): string | undefined {
      switch (popupId) {
          case PopupId.Settings:
              return 'settings'
          default:
              return undefined
      }
  }
</script>

{#each $popupStore as popup}
  {@const popupId = popup.id}
  <Popup title={getTitle(popupId)} {...popup.props} onClose={() => closePopup(popupId)}>
      <svelte:component this={COMPONENTS[popupId]} {...popup.props} />
  </Popup>
{/each}
