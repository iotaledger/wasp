<script lang="ts">
  import { BarSelector, MultipleSelector } from './../components';
  import { placeBet } from './../lib';
  import { round } from './../store';
  import Button from './button.svelte';
</script>

<div class="betting-system">
  <MultipleSelector />
  <div>
    <BarSelector />
    <div class="bet-button">
      <Button
        label="Place bet"
        disabled={$round.betSelection === undefined}
        onClick={() => placeBet()}
      />
    </div>
  </div>
</div>

<style lang="scss">
  .betting-system {
    display: flex;
    flex-direction: column;
    align-content: center;
    flex-wrap: wrap;
    gap: 30px;
    @media (min-width: 1024px) {
      flex-direction: row;
      justify-content: center;
      gap: 60px;
      align-items: flex-end;
    }
  }
  .bet-button {
    margin-top: 24px;
    display: flex;
    justify-content: center;
  }
</style>
