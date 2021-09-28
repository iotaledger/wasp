<script lang="ts">
  import { updateFunds } from '../../lib/app';

  import { balance } from '../../lib/store';

  export let value: number;
  export let disabled: boolean = false;

  let invalidMessage, textValue;

  $: value, updateLabel();

  $: textValue, validate();

  function updateLabel() {
    textValue = value;
  }

  const VALIDATION_ERRORS = {
    UNDER_RANGE_OF_BALANCE: `Value must be less than or equal to ${$balance}.`,
    MUST_BE_INTEGER: 'Value must be an integer.',
  };

  function validate(): void {
    let regex = new RegExp(/^\d+$/);
    invalidMessage = undefined;

    if (textValue < 0 || textValue > $balance) {
      invalidMessage = VALIDATION_ERRORS.UNDER_RANGE_OF_BALANCE;
    } else if (!regex.test(textValue)) {
      invalidMessage = VALIDATION_ERRORS.MUST_BE_INTEGER;
    } else {
      value = textValue;
    }
  }
</script>

<div>
  <div class="value">
    <input bind:value={textValue} on:keydown={() => validate()} />{' '}i
  </div>
  {#if invalidMessage}
    <div class="invalid-message">{invalidMessage}</div>
  {/if}
  <div class="bar-selector">
    <input
      bind:value
      type="range"
      min={1}
      max={Number($balance)}
      id="myRange"
      {disabled}
    />
  </div>
</div>

<style lang="scss">
  .value {
    text-align: end;
    font-size: 14px;
    line-height: 150%;
    letter-spacing: 0.5px;
    color: var(--white);
  }
  .invalid-message {
    font-size: 11px;
    color: tomato;
    margin-top: 8px;
  }
  .bar-selector {
    position: relative;
    margin: 7px 0;
    width: 100%;
    input {
      height: 26px;
      -webkit-appearance: none;
      width: 100%;
      border: 0;
      background-color: transparent;
      &::-webkit-slider-runnable-track {
        width: 100%;
        height: 8px;
        animate: 0.2s;
        border-radius: 4px;
        border: 4px solid var(--gray-7);
      }
      &::-webkit-slider-thumb {
        box-shadow: 0px 0px 0px var(--gray-7);
        border: 0px solid var(--mint-green-light);
        height: 20px;
        width: 20px;
        border-radius: 50px;
        background: var(--mint-green-light);
        -webkit-appearance: none;
        margin-top: -8px;
      }
    }
  }
</style>
