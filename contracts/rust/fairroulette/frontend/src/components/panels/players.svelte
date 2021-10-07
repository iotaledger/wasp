<script lang="ts">
  import {
    round,
    address as currentAddress,
    addressesHistory,
  } from "../../lib/store";

  const isMyAddress = (addr) =>
    $currentAddress === addr || $addressesHistory.includes(addr);
</script>

<div class="panel">
  <h3>Players</h3>
  <div class="players-wrapper">
    <div>
      {#each $round.players as { address, bet, number }, index}
        <div class="player" class:highlight={isMyAddress(address)}>
          <div class="player-index">{index + 1}</div>
          <div>
            <div class="player-address">{address}</div>
            <div class="player-bet">
              <span>{isMyAddress(address) ? "You bet: " : "Bet: "}</span>
              <span class="bet-value">{bet}i on {number}. </span>
            </div>
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
    .players-wrapper {
      display: flex;
      flex-direction: column-reverse;
      overflow-y: auto;
      margin-top: 16px;

      .player {
        display: inline-flex;
        position: relative;
        padding: 4px;
        width: 100%;

        gap: 24px;
        margin-bottom: 8px;

        &.highlight {
          background: rgba(0, 224, 202, 0.2);
          border-radius: 10px;
        }
        .player-index {
          font-size: 14px;
          line-height: 150%;
          letter-spacing: 0.5px;
          color: var(--gray-5);
          display: inline-block;
        }
        .player-address {
          font-weight: bold;
          font-size: 14px;
          line-height: 150%;
          letter-spacing: 0.5px;
          color: var(--gray-3);
          word-break: break-word;
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
