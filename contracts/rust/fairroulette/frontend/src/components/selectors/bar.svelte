<script lang="ts">
  import { balance } from '../../lib/store';

  export let value: number;
  export let disabled: boolean = false;
</script>

<div>
  <div class="value">{value}{' '}i</div>
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
      &::-moz-range-track {
        height: 8px;
        border-radius: 4px;
        background: var(--gray-7);
      }
      &::-moz-range-thumb {
        border: 0px solid var(--mint-green-light);
        background: var(--mint-green-light);
        height: 20px;
        width: 20px;
        border-radius: 50px;
      }
    }
  }
</style>
