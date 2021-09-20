<script lang="ts">
  import type { IPanelDataItem } from './../../models/IPanelDataItem';

  export let title: string;
  export let ordered: boolean = false;
  export let data: (IPanelDataItem & {
    tag?: string;
    description: {
      label?: string;
      value: string;
    }[];
  })[];
</script>

<div class="details-panel">
  <h3>{title}</h3>
  <hr />
  <div class="details-content">
    {#each data as item, index}
      <div class="details-tag">
        {#if ordered}
          <span class="item-index">{index}</span>
        {/if}

        <div class="tag">
          {#if item.tag}
            <span class="item-tag">{item.tag}</span>
          {/if}

          <span class="item-eyebrow">{item.eyebrow}</span>
        </div>
      </div>
      {#if item.label}
        <div class="item-label">{item.label}</div>
      {/if}
      <div class="tag-description">
        {#each item.description as { label, value }}
          <div class="item-description">
            {#if label}
              <span class="description-label">{label}</span>
            {/if}

            <span class="description-value">{value}</span>
          </div>
        {/each}
      </div>
    {/each}
  </div>
</div>

<style lang="scss">
  .details-panel {
    background: #141e31;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 12px;
    padding-top: 16px;
    padding-bottom: 32px;
    width: 100%;
    @media (min-width: 1024px) {
      width: 275px;
      padding-bottom: 80px;
    }
    h3 {
      font-weight: bold;
      font-size: 18px;
      line-height: 150%;
      letter-spacing: 0.03em;
      color: #ffffff;
      padding-left: 16px;
    }
    hr {
      margin-left: 16px;
      margin-right: 16px;
      border-color: rgba(255, 255, 255, 0.12);
    }
    .details-content {
      padding: 0 16px;
    }
    .item-index {
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: #909fbe;
      margin-right: 16px;
    }
    .tag {
      display: flex;
      justify-content: space-between;
      margin-bottom: 6px;
      width: 100%;
    }
    .details-tag {
      display: flex;
    }
    .item-tag {
      font-weight: bold;
      font-size: 12px;
      line-height: 150%;
      background: rgba(0, 224, 202, 0.2);
      border-radius: 6px;
      letter-spacing: 0.5px;
      color: #36a1ac;
      padding: 2px 6px;
    }
    .item-eyebrow {
      font-weight: bold;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: #d8e1f4;
    }
    .tag-description {
      display: flex;
    }
    .item-description {
      margin-left: 26px;
      margin-bottom: 16px;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: #909fbe;
    }
    .description-label {
      font-weight: bold;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: #909fbe;
    }
    .description-value {
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: #909fbe;
    }
  }
</style>
