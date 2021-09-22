<script lang="ts">
  import { round } from '../../store';
  import { get } from 'svelte/store';

  //   TODO: Delete. Only a demo.
  if ($round?.players.length < 1) {
    round.update((_round) => {
      if (_round) {
        _round.players.push({
          address: 'demo address',
          bet: 15,
        });
      }
      return _round;
    });
  }
</script>

<div class="panel">
  <h3>Players</h3>
  <div class="players-wrapper">
    {#each $round?.players as { address, bet }, index}
      <div class="player">
        <div class="player-index">{index + 1}</div>
        <div class="player-content">
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
    .players-wrapper {
      margin-top: 10px;
    }
    .player {
      display: inline-flex;
    }
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
</style>
