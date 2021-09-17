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
  <div>
    {#each data as item, index}
      {#if ordered}
        <span class="item-index">{index}</span>
      {/if}

      <div class="tag">
        {#if item.tag}
          <span class="item-tag">{item.tag}</span>
        {/if}

        <span class="item-eyebrow">{item.eyebrow}</span>
      </div>

      {#if item.label}
        <div class="item-label">{item.label}</div>
      {/if}

      {#each item.description as { label, value }}
        <div class="item-description">
          {#if label}
            <span>{label}</span>
          {/if}

          <span>{value}</span>
        </div>
      {/each}
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
    .item-index {
      color: yellow;
    }
    .tag {
      color: violet;
    }
    .item-label {
      color: red;
    }
    .item-description {
      color: turquoise;
    }
  }
</style>
