<script lang="ts">
  import { players } from '../../store';

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
  <div class="players-wrapper">
    {#each $players as { address, bet }, index}
      <div class="player">
        <div class="player-index">{index + 1}</div>
        <div>
          <div class="player-address">{address}</div>
          <div class="player-bet">
            <span>Bet:</span>
            <span class="bet-value">{bet}</span>
          </div>
        </div>
      </div>
    {/each}
  </div>
</div>

<style lang="scss">
  .panel {
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
    .players-wrapper {
      flex: 1;
      overflow-y: auto;
      padding-right: 16px;
      height: 520px;
      .player {
        display: inline-flex;
        width: 100%;
        margin-top: 10px;
        .player-index {
          font-size: 14px;
          line-height: 150%;
          letter-spacing: 0.5px;
          color: var(--gray-5);
          margin-right: 16px;
          display: inline-block;
        }
        .player-address {
          font-weight: bold;
          font-size: 14px;
          line-height: 150%;
          letter-spacing: 0.5px;
          color: var(--gray-3);
        }
        .player-bet {
          font-weight: bold;
          font-size: 14px;
          line-height: 150%;
          letter-spacing: 0.5px;
          color: var(--gray-5);
          .bet-value {
            font-weight: normal;
            margin-right: 16px;
          }
        }
      }
    }
  }
</style>
