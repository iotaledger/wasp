<script lang="ts">
  export let winnerNumber: number = undefined;
  export let mode: 'GAME_STARTED' | 'GAME_FINISHED';

  let numbers: {
    number: number;
    url: string;
    active?: boolean;
  }[] = [
    { number: 1, url: './1.svg' },
    { number: 2, url: './2.svg' },
    { number: 3, url: './3.svg' },
    { number: 4, url: './4.svg' },
    { number: 5, url: './5.svg' },
    { number: 6, url: './6.svg' },
    { number: 7, url: './7.svg' },
    { number: 8, url: './8.svg' },
  ];

  const TIME_ACTIVE = 500;

  function randomIntFromInterval(min, max) {
    return Math.floor(Math.random() * (max - min + 1) + min);
  }

  function disableAll() {
    for (let i = 0; i < numbers.length; i++) {
      numbers[i].active = false;
    }
  }

  //   $: if (mode === 'GAME_STARTED' && winnerNumber === undefined) {
  //     setInterval(() => {
  //       let randomNumber = randomIntFromInterval(0, 7);
  //       disableAll();
  //       numbers[randomNumber].active = true;
  //       console.log('active', numbers);
  //     }, 500);
  //   }
</script>

<div class="roulette">
  <img
    class="roulette-background"
    src="roulette_background.svg"
    alt="roulette"
  />
  {#each numbers as { number, url, active }}
    {#if active}
      <img class="active" src={url} alt="active" />
    {/if}
  {/each}
</div>

<style lang="scss">
  .roulette {
    position: relative;
    .roulette-background,
    .active {
      position: absolute;
    }
  }
</style>
