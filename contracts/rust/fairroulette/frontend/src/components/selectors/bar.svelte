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
    color: white;
  }
</style>
