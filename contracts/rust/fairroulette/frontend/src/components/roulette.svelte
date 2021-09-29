<script lang="ts">
  import { onDestroy } from 'svelte';
  import {
    placingBet,
    round,
    showWinningNumber,
    timeToFinished,
  } from '../lib/store';
  import { generateRandomInt } from '../lib/utils';
  import { ROUND_LENGTH } from './../lib/app';
  import Animation from './animation.svelte';

  // Highlighted number
  let flashedNumber: number;
  let interval;

  $: $round.active,
    $round.winningNumber,
    $showWinningNumber,
    updateFlashedNumber();
  // --

  // Progress bar
  const radius = 250;
  const stroke = 4;
  const normalizedRadius = radius - stroke;
  const circumference = normalizedRadius * 2 * Math.PI;

  let roundTimeAgo = 0;
  let progress = 0;
  let strokeDashoffset = 0;
  let animateProgressBar = false;

  let timeout;

  $: if ($round.active) {
    if ($timeToFinished > 0) {
      initializeProgressBar();
    }
  } else {
    resetProgressBar();
  }

  $: strokeDashoffset = $round.active
    ? circumference - (progress / ROUND_LENGTH) * circumference
    : circumference;
  // --

  function updateFlashedNumber() {
    if ($round.active) {
      if (!interval) {
        interval = setInterval(() => {
          flashedNumber = generateRandomInt(1, 8, flashedNumber);
        }, 500);
      }
    } else {
      resetFlashedNumber();
      if ($showWinningNumber && $round.winningNumber) {
        flashedNumber = Number($round.winningNumber);
      }
    }
  }

  function initializeProgressBar() {
    if (!timeout) {
      roundTimeAgo = ROUND_LENGTH - $timeToFinished;
      progress = roundTimeAgo;
      timeout = setTimeout(() => {
        progress = ROUND_LENGTH;
        animateProgressBar = true;
      }, 100);
    }
  }

  function resetFlashedNumber() {
    clearInterval(interval);
    interval = timeout = flashedNumber = undefined;
  }

  function resetProgressBar() {
    clearTimeout(timeout);
    animateProgressBar = false;
    progress = roundTimeAgo = 0;
  }

  onDestroy(() => {
    resetFlashedNumber();
    resetProgressBar();
  });
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
        stroke-dasharray="{circumference} {circumference}"
        style="stroke-dashoffset: {strokeDashoffset};
        {animateProgressBar &&
          `transition: stroke-dashoffset ${
            ROUND_LENGTH - roundTimeAgo
          }s linear`}"
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
