<script lang="ts">
  import { onDestroy } from 'svelte';
  import { placingBet, round, showWinningNumber } from '../lib/store';
  import { generateRandomInt } from '../lib/utils';
  import Animation from './animation.svelte';

  let flashedNumber: number;
  let interval;

  $: $round.active,
    $round.winningNumber,
    $showWinningNumber,
    updateFlashedNumber();

  onDestroy(reset);

  function updateFlashedNumber() {
    if ($round.active) {
      if (!interval) {
        interval = setInterval(() => {
          flashedNumber = generateRandomInt(1, 8, flashedNumber);
        }, 500);
      }
    } else {
      reset();
      if ($showWinningNumber && $round.winningNumber) {
        flashedNumber = Number($round.winningNumber);
      }
    }
  }

  function reset() {
    clearInterval(interval);
    interval = flashedNumber = undefined;
  }
</script>

<div class="roulette">
  {#if !$round.active && ($placingBet || $round.betPlaced)}
    <div class="animation">
      <Animation animation="loading" loop />
    </div>
  {:else}
    <img class="swirl" src="/assets/swirl.svg" alt="IOTA logo" />
  {/if}

  <img
    class="roulette-background"
    src="/assets/roulette_background.svg"
    alt="roulette"
  />
  {#if flashedNumber}
    <img
      class="flashedNumber"
      src={`/assets/${flashedNumber}.svg`}
      alt="active"
    />
  {/if}
</div>

<style lang="scss">
  .roulette {
    position: relative;
    width: 100%;
    .roulette-background,
    .flashedNumber {
      width: 100%;
      height: auto;
    }
    .flashedNumber {
      position: absolute;
      top: 0;
      left: 0;
    }
    .swirl {
      position: absolute;
      max-width: 50%;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
    }
    .animation {
      position: absolute;
      max-width: 50%;
      top: 25%;
      left: 25%;
    }
  }
</style>
