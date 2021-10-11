<script lang="ts">
  import { round } from '../../lib/store';
</script>

<div class="panel">
  <h3>Logs</h3>
  <div class="logs-wrapper">
    <div>
      {#each $round?.logs as { tag, timestamp, description }, index}
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
</div>

<style lang="scss">
  .panel {
    padding: 16px;
    display: flex;
    flex-direction: column;
    height: 100%;
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
    .logs-wrapper {
      // flex: 1;
      overflow-y: auto;
      padding-right: 16px;
      margin-top: 16px;
      display: flex;
      flex-direction: column-reverse;
      .log {
        display: flex;
        flex-direction: row;
        position: relative;
        padding: 0 0 17px 40px;
        .log-index {
          font-size: 14px;
          line-height: 150%;
          letter-spacing: 0.5px;
          color: var(--gray-5);
          position: absolute;
          left: 0;
        }
        .log-content {
          width: 100%;
          .log-content-header {
            display: flex;
            justify-content: space-between;
            width: 100%;
            margin-bottom: 6px;
            .log-tag {
              font-weight: bold;
              font-size: 12px;
              line-height: 150%;
              background: rgba(0, 224, 202, 0.2);
              border-radius: 6px;
              letter-spacing: 0.5px;
              color: var(--mint-green-dark);
              padding: 2px 6px;
              word-break: break-word;
            }
            .log-timestamp {
              font-weight: 500;
              font-size: 12px;
              line-height: 150%;
              letter-spacing: 0.5px;
              color: var(--gray-6);
              flex-shrink: 0;
              margin-left: 4px;
            }
          }

          .log-description {
            font-size: 14px;
            line-height: 150%;
            letter-spacing: 0.5px;
            color: var(--gray-3);
            word-break: break-word;
          }
        }
      }
    }
  }
</style>
