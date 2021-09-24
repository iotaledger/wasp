<script lang="ts">
  import { BETTING_NUMBERS } from "../../lib/app";
  import { round } from "../../lib/store";

  export let onClick: (number: number) => void = () => {};
  export let disabled: boolean = false;
</script>

<div class="values">
  {#each Array.from({ length: BETTING_NUMBERS }, (_, i) => i + 1) as number, index}
    <button
      on:click={() => onClick(index + 1)}
      class="cell"
      class:active={$round?.betSelection - 1 === index}
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
