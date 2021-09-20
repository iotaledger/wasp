<script lang="ts">
  export let values: string[];
  export let onClick: (index: number) => void;

  let active;

  $: onClick(active);
</script>

<div class="values">
  {#each values as value, i}
    <button
      on:click={() => {
        active === i ? (active = undefined) : (active = i);
      }}
      class="cell"
      class:active={active === i}
    >
      {value}
    </button>
  {/each}
</div>

<style lang="scss">
  .values {
    display: grid;
    grid-template-columns: auto auto auto auto;
    grid-column-gap: 4px;
    grid-row-gap: 8px;
    width: fit-content;

    .cell {
      background-color: tomato;
      width: 60px;
      height: 60px;
      &.active {
        background-color: turquoise;
      }
    }
  }
</style>
