<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { randomIntFromInterval } from '../utils/utils';

  export let winnerNumber: 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 = undefined;
  export let mode: 'GAME_STARTED' | 'GAME_STOPPED';

  let numbers: {
    number: number;
    url: string;
    active?: boolean;
  }[] = new Array(8).fill({}).map((item, i) => {
    return { ...item, number: i + 1, url: `./${i + 1}.svg` };
  });

  // Progress bar animation
  const radius = 25;
  const stroke = 1;
  const normalizedRadius = radius - stroke;
  const circumference = normalizedRadius * 2 * Math.PI;
  let progress = 0;

  $: strokeDashoffset = circumference - (progress / 100) * circumference;

  function disableAll() {
    numbers = numbers.map((n) => ({ ...n, active: false }));
  }

  let interval;
  let previousRandomNumber: number = -1;

  onMount(() => {
    interval = setInterval(() => {
      let randomNumber;

      do {
        randomNumber = randomIntFromInterval(0, 7);
      } while (randomNumber === previousRandomNumber);

      disableAll();
      numbers[randomNumber].active = true;
      previousRandomNumber = randomNumber;
    }, 500);
  });

  onDestroy(() => {
    clearInterval(interval);
  });

  const startCountdown = () => {
    progress === 0 ? (progress = 100) : (progress = 0);
    console.log('clicked!');
  };
</script>

<button on:click={() => startCountdown()}>Start</button>

<div class="roulette">
  <svg class="circle-animated" height={radius * 2} width={radius * 2}>
    <circle
      class:animate={progress !== 0}
      stroke="#00E0CA"
      fill="transparent"
      stroke-width={stroke}
      stroke-dasharray={circumference + ' ' + circumference}
      style={`stroke-dashoffset: ${strokeDashoffset};`}
      r={normalizedRadius}
      cx={radius}
      cy={radius}
    />
  </svg>
  <img
    class="roulette-background"
    src="roulette_background.svg"
    alt="roulette"
  />
  {#if mode === 'GAME_STARTED' && winnerNumber === undefined}
    {#each new Array(1200) as _}
      {#each numbers as { url, active }}
        {#if active}
          <img class="active" src={url} alt="active" />
        {/if}
      {/each}
    {/each}
  {:else if mode === 'GAME_STARTED' && winnerNumber > 0 && winnerNumber < 9}
    <img class="active" src={numbers[winnerNumber - 1].url} alt="active" />
  {/if}
</div>

<style lang="scss">
  .roulette {
    position: relative;
    max-width: 800px;
    height: 100%;
    margin: auto;
    .roulette-background,
    .active {
      position: absolute;
      width: 100%;
    }
  }
  .circle-animated {
    position: absolute;
    transform: translate(-50%, calc(-50% + 20px));
    top: 50%;
    left: 50%;
    opacity: 1;

    circle {
      &.animate {
        transition: stroke-dashoffset 3s linear;
      }
      transform: rotate(0);
      transform-origin: 50% 50%;
    }
  }
</style>
