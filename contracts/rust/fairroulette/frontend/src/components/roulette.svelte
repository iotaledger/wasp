<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { randomIntFromInterval } from '../utils/utils';
  import {
    GAME_RUNNING_STATE,
    round,
    START_GAME_STATE,
    state,
  } from './../store';

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
  {#if $state === GAME_RUNNING_STATE}
    {#each numbers as { url, active }}
      {#if active}
        <img class="flashedNumber" src={url} alt="active" />
      {/if}
    {/each}
  {:else if $state === START_GAME_STATE && $round.winningNumber > 0n}
    <img
      class="flashedNumber"
      src={numbers[Number($round.winningNumber - 1n)].url}
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
  }
</style>
