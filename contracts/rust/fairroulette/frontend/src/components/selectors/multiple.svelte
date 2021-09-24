<script lang="ts">
  import { round } from "../../lib/store";

  export let disabled: boolean = false;

  let activeIndex;

  function onClick(index) {
    activeIndex = activeIndex !== index ? index : undefined;
    round.update(($round) => ({ ...$round, betSelection: activeIndex }));
  }
</script>

<div class="values">
  {#each Array.from({ length: 8 }, (_, i) => i + 1) as number, index}
    <button
      on:click={() => onClick(index)}
      class="cell"
      class:active={activeIndex === index}
      {disabled}
    >
      {number}
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
