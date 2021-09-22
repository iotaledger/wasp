<script lang="ts">
  import { players } from '../store';

  //   TODO: Delete. Only a demo.
  if ($players.length < 1) {
    players.set([
      {
        address: 'demo address',
        bet: 15,
      },
    ]);
  }
</script>

<div class="panel">
  <h3>Players</h3>
  <div class="players-wraper">
    {#each $players as { address, bet }, index}
      <div class="player">
        <div class="player-index">{index + 1}</div>
        <div class="player-content">
          <div class="player-address">{address}</div>
          <div class="player-bet">Bet: {bet}</div>
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
    .player-index {
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
      margin-bottom: 6px;
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
    .player-address {
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
