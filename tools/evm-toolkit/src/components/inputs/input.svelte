<script lang="ts">
  import { InputType } from '$lib/common';

  export let name: string = '';
  export let required: boolean = false;
  export let label: string = undefined;
  export let type: InputType = InputType.Text;
  export let placeholder: string = '';
  export let value: string | number = '';
  export let id: string = '';
  export let maxLength: number = undefined;
  export let minLength: number = undefined;
  export let disabled: boolean = false;
  export let stretch: boolean = false;
  export let autofocus: boolean = false;

  function handleInput(e): void {
    value = type.match(/^(number|range)$/) ? +e.target.value : e.target.value;
    // to make sure that maxlength works in all browsers
    let stringValue = value?.toString() ?? '';
    if (maxLength && stringValue.length > maxLength) {
      value = stringValue.slice(0, maxLength);
      if (type === InputType.Number) {
        value = parseInt(value);
      }
    }
  }
</script>

<input-component class:w-full={stretch} class:label class:disabled>
  {#if label}
    <label for={id}>{label}{required ? '*' : ''}</label>
  {/if}
  <!-- svelte-ignore a11y-autofocus -->
  <input
    {name}
    {id}
    {type}
    {placeholder}
    {value}
    on:input={handleInput}
    maxlength={maxLength}
    minlength={minLength}
    {disabled}
    class:w-full={stretch}
    {required}
    {autofocus}
  />
</input-component>

<style lang="scss">
  input-component {
    @apply text-white;
    @apply flex flex-col space-y-1;
    @apply text-base;
    @apply bg-shimmer-background-tertiary;
    @apply border border-shimmer-background-secondary;
    @apply rounded-lg;
    @apply p-4;
    &.disabled {
      @apply opacity-50;
    }
  }
  input {
    @apply outline-none;
    @apply border-none;
    @apply bg-transparent;
    @apply p-0;
    &::placeholder {
      @apply text-gray-500;
    }
    &:disabled {
      @apply pointer-events-none;
    }
  }
  label {
    @apply text-sm;
    @apply font-semibold;
    @apply text-gray-400;
  }
</style>
