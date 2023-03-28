<script lang="ts">
  export let decimals: number = 0;
  export let min: number = 0;
  export let max: number = 0;
  export let value: number = 0;
  export let disabled: boolean = false;
  export let label: string = '';
  export let valid: boolean = true;

  let maxValueFormatted = 0;
  let valueFormatted = 0;

  $: maxValueFormatted = max / 10 ** decimals;
  $: valid = value >= min && value <= max;

  function handleRangeChange(event): void {
    value = event.target.value;
    valueFormatted = value / 10 ** decimals;
  }
  function handleInputChange(event): void {
    valueFormatted = event.target.value;
    value = valueFormatted * 10 ** decimals;
  }
</script>

<div class="flex flex-col space-y-4">
  {#if label}
    <div class="flex space-x-2">
      <label for="formatted" class="flex flex-shrink-0">{label}</label>
      <input
        type="text"
        id="formatted"
        bind:value={valueFormatted}
        on:input={handleInputChange}
        class:error={!valid}
        {disabled}
      />
    </div>
  {/if}
  <input
    type="range"
    bind:value
    on:input={handleRangeChange}
    {min}
    {max}
    {disabled}
  />
  <div class="w-full flex justify-between">
    <small>Min: {min}</small>
    <small>Max: {maxValueFormatted}</small>
  </div>
</div>

<style lang="scss">
  input {
    -webkit-appearance: none;
    appearance: none;
    background: transparent;
    cursor: pointer;
    width: 100%;

    &.error {
      @apply text-red-500;
    }
    
    &:disabled {
      @apply pointer-events-none;
      @apply opacity-50;
    }

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
  input[type='text'] {
    &::placeholder {
      @apply text-xs;
      @apply text-shimmer-text-secondary;
    }
  }
</style>
