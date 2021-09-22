<script lang="ts">
  import { logs } from '../../store';
</script>

<div class="panel">
  <h3>Logs</h3>
  <div class="logs-wraper">
    {#each $logs as { tag, timestamp, description }, index}
      <div class="log">
        <div class="log-index">{index + 1}</div>
        <div class="log-content">
          <div class="log-content-header">
            <div class="log-tag">{tag}</div>
            <div class="log-timestamp">{timestamp}</div>
          </div>
          <div class="log-description">{description}</div>
        </div>
      </div>
    {/each}
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
    .log-index {
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-5);
      margin-right: 16px;
    }
    .log-content-header {
      display: flex;
      justify-content: space-between;
      width: 100%;
      margin-bottom: 6px;
    }
    .details-tag {
      display: flex;
      margin-top: 12px;
    }
    .log-tag {
      font-weight: bold;
      font-size: 12px;
      line-height: 150%;
      background: rgba(0, 224, 202, 0.2);
      border-radius: 6px;
      letter-spacing: 0.5px;
      color: var(--mint-green-dark);
      padding: 2px 6px;
    }
    .log-eyebrow {
      font-weight: bold;
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.5px;
      color: var(--gray-3);
    }
    .tag-description {
      display: flex;
    }
    .log-description {
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
