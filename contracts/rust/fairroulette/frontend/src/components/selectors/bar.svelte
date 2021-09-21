<script lang="ts">
  export let minimum: number;
  export let maximum: number;
  export let step: number = undefined;
  export let unit: string = '';
  export let onChange: (val: number) => void;

  let value = (maximum - minimum) / 2;

  onChange((maximum - minimum) / 2);
</script>

<div>
  <div class="value">{value}{' '}{unit}</div>
  <div class="bar-selector">
    <input
      bind:value
      type="range"
      min={minimum}
      max={maximum}
      {step}
      class="slider"
      id="myRange"
      on:change={() => onChange(value)}
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
  .slider {
    height: 26px;
    -webkit-appearance: none;
    margin: 7px 0;
    width: 100%;
    border: 0;
    background-color: transparent;
  }
  .slider:focus {
    outline: none;
  }
  .slider::-webkit-slider-runnable-track {
    width: 100%;
    height: 8px;
    cursor: pointer;
    animate: 0.2s;
    border-radius: 4px;
    border: 4px solid var(--gray-4);
  }
  .slider::-webkit-slider-thumb {
    box-shadow: 0px 0px 0px var(--gray-4);
    border: 0px solid var(--mint-green-light);
    height: 20px;
    width: 20px;
    border-radius: 50px;
    background: var(--mint-green-light);
    cursor: pointer;
    -webkit-appearance: none;
    margin-top: -8px;
  }
  .slider:focus::-webkit-slider-runnable-track {
    background: var(--gray-4);
  }
  input[type='range']::-moz-range-track {
    width: 100%;
    height: 8px;
    cursor: pointer;
    animate: 0.2s;
    box-shadow: 0px 0px 0px var(--gray-4);
    background: var(--gray-4);
    border-radius: 4px;
    border: 4px solid var(--gray-4);
  }
  .slider::-moz-range-thumb {
    box-shadow: 0px 0px 0px var(--gray-4);
    border: 0px solid var(--mint-green-light);
    height: 20px;
    width: 20px;
    border-radius: 50px;
    background: var(--mint-green-light);
    cursor: pointer;
  }
  .slider::-ms-track {
    width: 100%;
    height: 8px;
    cursor: pointer;
    animate: 0.2s;
    background: transparent;
    border-color: transparent;
    color: transparent;
  }
  .slider::-ms-fill-lower {
    background: var(--gray-4);
    border: 4px solid var(--gray-4);
    border-radius: 8px;
    box-shadow: 0px 0px 0px var(--gray-4);
  }
  .slider::-ms-fill-upper {
    background: var(--gray-4);
    border: 4px solid var(--gray-4);
    border-radius: 8px;
    box-shadow: 0px 0px 0px var(--gray-4);
  }
  .slider::-ms-thumb {
    margin-top: 1px;
    box-shadow: 0px 0px 0px var(--gray-4);
    border: 0px solid var(--mint-green-light);
    height: 20px;
    width: 20px;
    border-radius: 50px;
    background: var(--mint-green-light);
    cursor: pointer;
  }
  .slider:focus::-ms-fill-lower {
    background: var(--gray-4);
  }
  .slider:focus::-ms-fill-upper {
    background: var(--gray-4);
  }
</style>
