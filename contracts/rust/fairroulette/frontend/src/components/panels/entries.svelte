<script lang="ts">
  import type { IEntriesPanel } from '../../models/IEntriesPanel';
  import type { ILogEntries } from '../../models/ILogEntries';
  import { LOG_ENTRIES_TYPE } from '../../models/ILogEntries';
  import type { IPlayerEntries } from '../../models/IPlayerEntries';

  import type { IPanelDataItem } from './../../models/IPanelDataItem';

  import { PLAYER_ENTRIES_TYPE } from './../../models/IPlayerEntries';
  export let title: string;
  export let ordered: boolean = false;
  export let entries: ILogEntries | IPlayerEntries;

  console.log('entries', entries);
</script>

<div class="panel">
  <h3>{title}</h3>
  <div>
    {#if entries.type === PLAYER_ENTRIES_TYPE}
      <!-- Player panel -->
      {#each entries.data as entry, index}
        <div class="details-tag">
          {#if ordered}
            <span class="item-index">{index}</span>
          {/if}
          <div class="item-eyebrow">{entry.address}</div>
        </div>
        <div class="item-description">
          {#each entry.fields as { label, value }}
            {#if label}
              <span class="description-label">{label}</span>
            {/if}
            <span class="description-value ">{value}</span>
          {/each}
        </div>
      {/each}
    {/if}

    {#if entries.type === LOG_ENTRIES_TYPE}
      <!-- Log panel -->
      {#each entries.data as { tag, timestamp, description }, index}
        <div class="details-tag">
          {#if ordered}
            <span class="item-index">{index}</span>
          {/if}
          <div class="tag">
            <div class="item-tag">{tag}</div>
            <span class="description-value">{timestamp}</span>
          </div>
        </div>

        <span class="description-value">{description}</span>
      {/each}
    {/if}
  </div>
</div>

<style lang="scss">
  .panel {
    background: var(--blue-dark);
    padding: 16px;
    h3 {
      font-weight: bold;
      font-size: 18px;
      line-height: 150%;
      letter-spacing: 0.03em;
      color: var(--white);
      border-bottom: 1px solid var(--border-color);
      padding-bottom: 14px;
      margin: 0;
    }
    .item-index {
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
      margin-right: 16px;
    }
    .tag {
      display: flex;
      justify-content: space-between;
      width: 100%;
    }
    .details-tag {
      display: flex;
      margin-top: 12px;
    }
    .item-tag {
      font-weight: bold;
      font-size: 12px;
      line-height: 150%;
      background: rgba(0, 224, 202, 0.2);
      border-radius: 6px;
      letter-spacing: 0.5px;
      color: var(--mint-green-dark);
      padding: 2px 6px;
    }
    .item-eyebrow {
      font-weight: bold;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-3);
    }
    .tag-description {
      display: flex;
    }
    .item-description {
      margin: 0 0 16px 0;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
      .description-value {
        margin-left: 0;
      }
    }
    .description-label {
      font-weight: bold;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
      margin-left: 26px;
    }
    .description-value {
      margin-left: 26px;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
    }
  }
</style>
