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
      background-color: transparent;
      width: 56px;
      height: 56px;
      border: 1px solid #677695;
      box-sizing: border-box;
      border-radius: 6px;
      color: var(--white);
      &.active {
        background-color: rgba(20, 202, 191, 0.08);
        border: 1px solid var(--mint-green-light);
      }
    }
  }
</style>
