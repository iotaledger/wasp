<script lang="ts">
  export let min: string = '';
  export let max: number = 100;
  export let value: number;
  export let disabled: boolean = false;
  export let label: string = '';

  function handleChange(event) {
    value = event.target.value;
  }
</script>

<div class="flex flex-col space-y-4">
  {#if label}
    <label for="range">{label}</label>
  {/if}
  {#if min || max}
    <div class="w-full flex justify-between">
      <small>{min}</small>
      <small>Max: {(max / 1e6).toFixed(2)}</small>
    </div>
  {/if}
  <input
    type="range"
    bind:value
    on:change={handleChange}
    {min}
    {max}
    {disabled}
  />
</div>

<style lang="scss">
  input {
    -webkit-appearance: none;
    appearance: none;
    background: transparent;
    cursor: pointer;
    width: 100%;

    /* Removes default focus */
    &:focus {
      outline: none;
    }

    /******** Chrome, Safari, Opera and Edge Chromium styles ********/
    /* slider track */
    &::-webkit-slider-runnable-track {
      background-color: #061928;
      border-radius: 0.5rem;
      height: 8px;
    }

    /* slider thumb */
    &::-webkit-slider-thumb {
      -webkit-appearance: none; /* Override default look */
      appearance: none;
      margin-top: -4px; /* Centers thumb on the track */
      background-color: #00f5dd;
      border-radius: 0.5rem;
      height: 1rem;
      width: 1rem;
    }

    &:focus::-webkit-slider-thumb {
      outline: 3px solid #00f5dd;
      outline-offset: 0.125rem;
    }

    /*********** Firefox styles ***********/
    /* slider track */
    &::-moz-range-track {
      background-color: #061928;
      border-radius: 0.5rem;
      height: 8px;
    }

    /* slider thumb */
    &::-moz-range-thumb {
      background-color: #00f5dd;
      border: none; /*Removes extra border that FF applies*/
      border-radius: 0.5rem;
      height: 1rem;
      width: 1rem;
    }

    &:focus::-moz-range-thumb {
      outline: 3px solid #00f5dd;
      outline-offset: 0.125rem;
    }
  }
</style>
