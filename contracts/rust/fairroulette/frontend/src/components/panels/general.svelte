<script lang="ts">
  import type { IPanelDataItem } from './../../models/IPanelDataItem';
  export let data: IPanelDataItem[];
</script>

<div class="general-panel">
  {#each data as item}
    <div>
      <div class="eyebrow">{item.eyebrow}</div>
      <div class="label">{item.label ?? '-'}</div>
    </div>
  {/each}
</div>

<style lang="scss">
  .general-panel {
    background: yellow;
    max-width: 400px;
  }
</style>
