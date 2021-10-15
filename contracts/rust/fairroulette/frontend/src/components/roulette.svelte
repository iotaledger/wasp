<script lang="ts">
  import { onDestroy } from 'svelte';
  import {
    placingBet,
    round,
    showBettingSystem,
    showWinningNumber,
    timeToFinished,
  } from '../lib/store';
  import { generateRandomInt } from '../lib/utils';
  import { ROUND_LENGTH } from './../lib/app';
  import Animation from './animation.svelte';
  import BettingSystem from './betting_system.svelte';
  import { fade } from 'svelte/transition';

  // Highlighted number
  let flashedNumber: number;
  let interval: number | undefined;

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

  let timeout: number | undefined;

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
        `transition: stroke-dashoffset ${ROUND_LENGTH - roundTimeAgo}s linear`}"
      stroke-width={stroke}
      r={normalizedRadius}
      cx={radius}
      cy={radius}
    />
  </svg>
  <div class="roulette-aspect-ratio">
    <div class="roulette">
      {#if $showBettingSystem}
        <img
          class="roulette-progress-road"
          src="/assets/progress.svg"
          alt="progress"
        />
      {:else}
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
            class:blink={$showWinningNumber}
            src={`/assets/${flashedNumber}.svg`}
            alt="active"
          />
        {/if}
      {/if}
    </div>
  </div>
  {#if $showBettingSystem}
    <div in:fade>
      <BettingSystem />
    </div>
  {/if}
</div>

<style lang="scss">
  .roulette-wrapper {
    position: relative;
    width: 100%;
    @media (min-width: 1024px) {
      margin: 0 auto;
    }
    .circle-animated {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, calc(-50% + 3px));
      width: calc(100% - 25px);
      height: calc(100% - 25px);
      filter: drop-shadow(0px 0px 10px rgba(0, 245, 221, 0.5));
      z-index: -1;
      circle {
        &.animate {
          transition: stroke-dashoffset 0.5s linear;
        }
        transform: rotate(-90deg);
        transform-origin: 50% 50%;
      }
    }
    .roulette-aspect-ratio {
      width: 100%;
      padding-bottom: 100%;
      aspect-ratio: 1/1;
    }
    .roulette {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      .roulette-progress-road {
        width: 100%;
        height: auto;
        position: relative;
      }
      .roulette-background,
      .flashedNumber {
        width: 100%;
        height: auto;
      }
      .flashedNumber {
        position: absolute;
        top: 0;
        left: 0;

        &.blink {
          animation: 1s blink linear 5;
        }
        @keyframes blink {
          50% {
            opacity: 0;
          }
        }
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
