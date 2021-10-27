<script lang="ts">
  import {
    address as currentAddress,
    addressesHistory,
    round,
  } from '../../lib/store';
  import { GOSHIMMER_ADDRESS_EXPLORER_URL } from '../../lib/app';

  const isMyAddress = (addr: string): boolean =>
    $currentAddress === addr || $addressesHistory.includes(addr);
</script>

<div class="panel">
  <h3>Players</h3>
  <div class="players-wrapper">
    <div class="players">
      {#each $round.players as { address, bet, number }, index}
        <div class="player" class:highlight={isMyAddress(address)}>
          <div class="player-index">{index + 1}</div>
          <div>
            <a
              class="player-address"
              href={`${GOSHIMMER_ADDRESS_EXPLORER_URL}/${address}`}
              target="_blank"
              rel="noopener noreferrer">{address}</a
            >
            <div class="player-bet">
              <span>{isMyAddress(address) ? 'You bet: ' : 'Bet: '}</span>
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
    width: 100%;
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
      @media (max-width: 1024px) {
        .players {
          max-height: 400px;
          overflow-y: auto;
        }
      }
      .player {
        display: inline-flex;
        position: relative;
        padding: 4px;
        width: 100%;
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
          margin-right: 24px;
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
