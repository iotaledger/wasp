<script lang="ts">
  import { onDestroy } from 'svelte';
  import {
    placingBet,
    round,
    showWinningNumber,
    timeToFinished,
  } from '../lib/store';
  import { generateRandomInt } from '../lib/utils';
  import Animation from './animation.svelte';

  import { ROUND_LENGTH } from './../lib/app';
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

  // Progress bar
  const radius = 250;
  const stroke = 4;
  const normalizedRadius = radius - stroke;
  const circumference = normalizedRadius * 2 * Math.PI;

  let start = 0;
  let progress = 0;

  let startLoaded = false;
  let strokeDashOffsetLoaded = false;

  let timeout;

  $: $timeToFinished, initializeProgressBar();

  function initializeProgressBar() {
    if (!isNaN($timeToFinished) && $timeToFinished !== 0 && !startLoaded) {
      start = ROUND_LENGTH - $timeToFinished;
      startLoaded = true;
      progress = start;
      timeout = setTimeout(() => {
        progress = ROUND_LENGTH;
        strokeDashOffsetLoaded = true;
      }, 100);
    }
  }

  $: $round.active, resetProgressBar();

  function resetProgressBar() {
    if (!$round.active) {
      progress = 0;
      start = 0;
      startLoaded = false;
      strokeDashOffsetLoaded = false;
    }
  }

  onDestroy(() => clearTimeout(timeout));

  $: strokeDashoffset = $round.active
    ? circumference - (progress / ROUND_LENGTH) * circumference
    : circumference;
</script>

<div class="roulette-wrapper">
  <div class="progress-bar">
    <svg
      class="circle-animated"
      viewBox="0 0 500 500"
      preserveAspectRatio="xMinYMin meet"
    >
      <circle
        stroke="#00E0CA"
        fill="transparent"
        stroke-dasharray={circumference + ' ' + circumference}
        style={`stroke-dashoffset: ${strokeDashoffset ?? 0}; transition: ${
          progress !== 0 && strokeDashOffsetLoaded
            ? `stroke-dashoffset ${ROUND_LENGTH - start}s linear;`
            : ''
        }`}
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
  .roulette-wrapper {
    .progress-bar {
      position: relative;
      display: flex;
      align-items: center;
      outline: none;
      svg {
        position: absolute;
        top: 3px;
        padding: 13px;
        &.circle-animated {
          circle {
            &.animate {
              transition: stroke-dashoffset 0.5s linear;
            }
            transform: rotate(-90deg);
            transform-origin: 50% 50%;
          }
        }

        path {
          fill: transparent;
        }
      }
    }
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
  }
</style>
