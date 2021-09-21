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
  <img
    class="roulette-background"
    src="roulette_background.svg"
    alt="roulette"
  />
  {#if mode === 'GAME_STARTED' && winnerNumber === undefined}
    {#each numbers as { url, active }}
      {#if active}
        <img class="flashedNumber" src={url} alt="active" />
      {/if}
    {/each}
  {:else if mode === 'GAME_STARTED' && winnerNumber > 0 && winnerNumber < 9}
    <img class="flashedNumber" src={numbers[winnerNumber - 1].url} alt="active" />
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
  }
</style>
