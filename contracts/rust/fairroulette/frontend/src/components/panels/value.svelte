<script lang="ts">
  import type { IPanelDataItem } from './../../models/IPanelDataItem';
  import type { IButton } from './../../models/IButton';
  import Button from '../button.svelte';

  export let data: IPanelDataItem;
  export let buttons: IButton[] = [];
</script>

<div class="value-panel">
  <div>
    <div>{data.eyebrow}</div>
    {#if data.label}
      <div>{data.label}</div>
    {/if}
  </div>
  <div>
    {#if buttons}
      {#each buttons as button}
        <Button {...button} />
      {/each}
    {/if}
  </div>
</div>

<style lang="scss">
  .value-panel {
    background-color: blue;
    max-width: 400px;
  }
</style>
