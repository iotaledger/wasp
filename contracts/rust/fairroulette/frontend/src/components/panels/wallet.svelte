<script lang="ts">
  import type { IPanelDataItem } from './../../models/IPanelDataItem';
  export let data: IPanelDataItem[];
</script>

<div class="panel">
  {#each data as item}
    <div class="panel-data">
      <div class="eyebrow">{item.eyebrow}</div>
      <div class="label">{item.label ?? '-'}</div>
    </div>
  {/each}
</div>

<style lang="scss">
  .panel {
    @media (min-width: 1024px) {
      padding-bottom: 24px;
    }
    .panel-data {
      display: grid;
      grid-template-rows: 1fr;
      grid-template-columns: 1fr 2fr;
      @media (min-width: 1024px) {
        grid-template-columns: 1fr;
      }
      .label {
        width: 100%;
        font-size: 14px;
        letter-spacing: 0.5px;
        color: var(--gray-3);
        line-height: 150%;
        word-break: break-all;
        @media (min-width: 1024px) {
          font-weight: bold;
        }
      }
    }
    .panel-data:not(:last-child) {
      margin-bottom: 4px;

      @media (min-width: 1024px) {
        margin-bottom: 16px;
      }
    }
  }
</style>
