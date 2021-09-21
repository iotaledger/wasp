<script lang="ts">
  import type { ILogEntries } from '../../models/ILogEntries';
  import { LOG_ENTRIES_TYPE } from '../../models/ILogEntries';
  import type { IPlayerEntries } from '../../models/IPlayerEntries';
  import { PLAYER_ENTRIES_TYPE } from './../../models/IPlayerEntries';
  export let title: string;
  export let ordered: boolean = false;
  export let entries: ILogEntries | IPlayerEntries;

  console.log('entries', entries);
</script>

<div class="details-panel">
  <h3>{title}</h3>
  <hr />
  <div class="details-content">
    {#if entries.type === PLAYER_ENTRIES_TYPE}
      <!-- Player panel -->
      {#each entries.data as entry, index}
        <div class="details-tag">
          {#if ordered}
            <span class="item-index">{index + 1}</span>
          {/if}
          <div class="item-eyebrow">{entry.address}</div>
        </div>
        {#each entry.fields as { label, value }}
          <div class="item-description">
            {#if label}
              <span class="description-label">{label}</span>
            {/if}
            <span class="description-value">{value}</span>
          </div>
        {/each}
      {/each}
    {/if}

    {#if entries.type === LOG_ENTRIES_TYPE}
      <!-- Log panel -->
      {#each entries.data as { tag, timestamp, description }, index}
        <div class="details-tag">
          {#if ordered}
            <span class="item-index">{index + 1}</span>
          {/if}
          <div class="item-tag">{tag}</div>
          <span class="description-value">{timestamp}</span>
        </div>
        <span class="description-value">{description}</span>
      {/each}
    {/if}
  </div>
</div>

<style lang="scss">
  .details-panel {
    background: var(--blue-dark);
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
      color: var(--white);
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
      color: var(--gray-5);
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
      margin-left: 26px;
      margin-bottom: 16px;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
    }
    .description-label {
      font-weight: bold;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
    }
    .description-value {
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
    }
  }
</style>
