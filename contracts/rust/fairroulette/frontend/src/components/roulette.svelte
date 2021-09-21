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
</script>

<div class="roulette">
  <div class="full-roulette">
    <img
      class="roulette-background"
      src="roulette_background.svg"
      alt="roulette"
    />
    {#if mode === 'GAME_STARTED' && winnerNumber === undefined}
      {#each numbers as { url, active }}
        {#if active}
          <img class="active" src={url} alt="active" />
        {/if}
      {/each}
    {:else if mode === 'GAME_STARTED' && winnerNumber > 0 && winnerNumber < 9}
      <img class="active" src={numbers[winnerNumber - 1].url} alt="active" />
    {/if}
  </div>
</div>

<style lang="scss">
  .roulette {
    position: relative;
    width: 75%;
    height: 400px;
    margin: auto;
    .full-roulette {
      max-height: 100%;
      max-width: 100%;
      .roulette-background,
      .active {
        position: absolute;
        width: 100%;
      }
    }
  }
  @media (min-width: 520px) {
    .roulette {
      margin-top: -20px;
      height: 550px;
    }
  }
  @media (min-width: 720px) {
    .roulette {
      margin-top: -20px;
      height: 700px;
    }
  }
  @media (min-width: 900px) {
    .roulette {
      margin-top: -20px;
      height: 800px;
    }
  }
  @media (min-width: 1024px) {
    .roulette {
      height: 500px;
      margin-top: -20px;
    }
  }
  @media (min-width: 1350px) {
    .roulette {
      height: 700px;
      margin-top: -20px;
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
