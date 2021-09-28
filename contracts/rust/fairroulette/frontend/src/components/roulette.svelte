<script lang="ts">
  import { onDestroy } from 'svelte';
  import {
    placingBet,
    round,
    showWinningNumber,
    timestamp,
    timeToFinished,
  } from '../lib/store';
  import { generateRandomInt } from '../lib/utils';
  import Animation from './animation.svelte';

  import { ROUND_LENGTH } from './../lib/app';
  import { calculateRoundLengthLeft } from '../lib/app';

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

  const radius = 250;
  const stroke = 1;
  const normalizedRadius = radius - stroke;
  const circumference = normalizedRadius * 2 * Math.PI;
  $: progress = $timeToFinished === 0 ? 0 : 100;
  $: console.log(
    'progress',
    progress,
    'calculateRoundLengthLeft',
    calculateRoundLengthLeft($timestamp),
    'round length;',
    ROUND_LENGTH
  );

  $: strokeDashoffset = circumference - (progress / 100) * circumference;
</script>

<div class="arrow-button">
  <div class="rag">
    <svg
      class="circle-animated"
      viewBox="0 0 500 500"
      preserveAspectRatio="xMinYMin meet"
    >
      <circle
        class:animate={progress !== 0}
        stroke="#00E0CA"
        fill="transparent"
        stroke-dasharray={circumference + ' ' + circumference}
        style={`stroke-dashoffset: ${strokeDashoffset};`}
        stroke-width={stroke}
        r={normalizedRadius}
        cx={radius}
        cy={radius}
      />
    </svg>
  </div>

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

  .arrow-button {
    position: relative;
    top: 5px;
    display: flex;
    align-items: center;
    outline: none;
    svg {
      position: absolute;
      top: 3px;
      padding: 15px;
      &.circle-animated {
        circle {
          &.animate {
            transition: stroke-dashoffset 30s linear;
          }
          transform: rotate(-90deg);
          transform-origin: 50% 50%;
        }
      }

      path {
        fill: transparent;
      }
    }
    &.left {
      svg {
        left: 0;
      }
    }
    &.right {
      svg {
        right: 0;
      }
    }
    @keyframes offsettozero {
      to {
        stroke-dashoffset: 0;
      }
    }

    @screen md {
      top: -25px;
    }
  }
</style>
