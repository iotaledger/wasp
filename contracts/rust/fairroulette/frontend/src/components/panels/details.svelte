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
  <div>{title}</div>
  <div>
    {#each data as item, index}
      {#if ordered}
        <div>{index}</div>
      {/if}

      <div>
        {#if item.tag}
          <div>{item.tag}</div>
        {/if}

        <div>{item.eyebrow}</div>
      </div>

      {#if item.label}
        <div>{item.label}</div>
      {/if}

      {#each item.description as { label, value }}
        <div>
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
    background-color: pink;
  }
</style>
