<script lang="ts">
  import { onDestroy } from "svelte";
  import { round } from "../lib/store";
  import { generateRandomInt } from "../lib/utils";

  let flashedNumber: number;
  let interval;

  // TODO: review this logic when websockets are working
  $: if ($round.winningNumber) {
    clearInterval(interval);
    interval = undefined;
    flashedNumber = Number($round.winningNumber);
  }

  // TODO: review this logic when websockets are working
  $: if ($round.active) {
    if (!interval) {
      interval = setInterval(() => {
        flashedNumber = generateRandomInt(1, 8, flashedNumber);
      }, 500);
    }
  } else {
    if (interval) {
      flashedNumber = undefined;
      clearInterval(interval);
      interval = undefined;
    }
  }

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
  {#if flashedNumber}
    <img class="flashedNumber" src={`./${flashedNumber}.svg`} alt="active" />
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
