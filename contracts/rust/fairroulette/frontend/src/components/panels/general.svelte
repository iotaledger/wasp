<script lang="ts">
  import type { IPanelDataItem } from './../../models/IPanelDataItem';
  export let data: IPanelDataItem[];
</script>

<div class="general-panel">
  {#each data as item}
    <div class="panel-data">
      <span class="eyebrow">{item.eyebrow}</span>
      <span class="label">{item.label ?? '-'}</span>
    </div>
  {/each}
</div>

<style lang="scss">
  .general-panel {
    background: rgba(6, 16, 35, 0.2);
    backdrop-filter: blur(20px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    border-top: 1px solid rgba(255, 255, 255, 0.08);
    padding-top: 10px;
    padding-bottom: 10px;
    width: 100%;
    @media (min-width: 1024px) {
      background: #141e31;
      border: 1px solid rgba(255, 255, 255, 0.12);
      border-radius: 12px;
      width: 275px;
      padding-top: 8px;
      padding-bottom: 8px;
    }
    .panel-data {
      margin-left: 24px;
      line-height: 150%;
      @media (min-width: 1024px) {
        margin-top: 8px;
        margin-bottom: 8px;
      }
      .eyebrow {
        text-transform: capitalize;
        font-weight: bold;
        font-size: 12px;
        letter-spacing: 0.5px;
        color: #36a1ac;
        margin-right: 12px;
        @media (min-width: 900px) {
          display: block;
          margin-right: 0;
        }
      }
      .label {
        font-style: normal;
        font-weight: bold;
        font-size: 14px;
        letter-spacing: 0.5px;
        color: #d8e1f4;
      }
    }
  }
</style>
