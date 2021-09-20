<script lang="ts">
  import type { IBarSelector } from '../models/IBarSelector';
  import type { IButton } from '../models/IButton';
  import type { IMultipleSelector } from '../models/IMultipleSelector';
  import { Selector } from './';
  import Button from './button.svelte';

  let betAmount;
  let selectedNumber;

  const BET_NUMBER_SELECTOR: IMultipleSelector = {
    type: 'multiple',
    values: [...Array(8).keys()].map((i) => (i + 1).toString()), // ["1", "2", ... ,"8"]
    onClick: (indexSelected) =>
      (selectedNumber = BET_NUMBER_SELECTOR.values[indexSelected]),
  };

  const BET_IOTA_AMOUNT_SELECTOR: IBarSelector = {
    type: 'bar',
    minimum: 1,
    maximum: 200,
    unit: 'i',
    onChange: (value) => {
      betAmount = value;
    },
  };

  const PLACE_BET_BUTTON: IButton = {
    label: 'Place Bet',
    onClick: () => {
      console.log('¡¡¡¡Place bet!!!!');
      console.log('Number chosen: ', selectedNumber);
      console.log('IOTA Amount: ', betAmount);
    },
  };
</script>

<div class="betting-system">
  <Selector {...BET_NUMBER_SELECTOR} />
  <div>
    <Selector {...BET_IOTA_AMOUNT_SELECTOR} />
    <div class="bet-button">
      <Button {...PLACE_BET_BUTTON} disabled={selectedNumber === undefined} />
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
