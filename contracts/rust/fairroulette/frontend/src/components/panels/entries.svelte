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

<div class="panel">
  <h3>{title}</h3>
  <div>
    {#if entries.type === PLAYER_ENTRIES_TYPE}
      <!-- Player panel -->
      {#each entries.data as entry, index}
        <div>
          {#if ordered}
            <div class="item-index">{index + 1}</div>
          {/if}
          <div class="details-tag">
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
        </div>
      {/each}
    {/if}

    {#if entries.type === LOG_ENTRIES_TYPE}
      <!-- Log panel -->
      {#each entries.data as { tag, timestamp, description }, index}
        <div class="ITEM">
          <div class="details-tag">
            {#if ordered}
              <div class="item-index">{index + 1}</div>
            {/if}
            <div class="tag">
              <span class="item-tag">{tag}</span>
              <span class="description-value">{timestamp}</span>
            </div>
          </div>
          <div class="description-value">{description}</div>
        </div>
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
      display: inline-block;
    }
    .tag {
      display: flex;
      justify-content: space-between;
      width: 100%;
      margin-bottom: 6px;
    }
    .details-tag {
      display: inline-flex;
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
      height: min-content;
    }
    .item-eyebrow {
      font-weight: bold;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-3);
    }
    .tag-description {
      display: inline-flex;
    }
    .item-description {
      display: block;
      margin: 0 0 16px 0;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
      @media (min-width: 1024px) {
        display: inline-flex;
      }
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
      @media (min-width: 1024px) {
        display: inline-flex;
      }
    }
  }
</style>
