<script lang="ts">
  import { InputType } from '../../lib/enums';

  export let name: string = '';
  export let required: boolean = false;
  export let label: string = undefined;
  export let type: InputType = InputType.Text;
  export let placeholder: string = '';
  export let value: string = '';
  export let id: string = '';
  export let maxLength: number = undefined;
  export let minLength: number = undefined;
  export let disabled: boolean = false;
  export let stretch: boolean = false;

  const handleInput = (e): void => {
    value = e.target.value;
    // to make sure that type=number only accepts numbers in all browsers
    if (type === InputType.Number && !/[0-9.]/.test(value)) {
      e.target.value = value.slice(0, -1);
    }
    // to make sure that maxlength works in all browsers
    if (maxLength && value.length > maxLength) {
      value = value.slice(0, maxLength);
    }
  };
</script>

<div class:w-full={stretch} class:label>
  {#if label}
    <label for={id}>{label}{required ? '*' : ''}</label>
  {/if}
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
  />
</div>

<style lang="scss">
  div {
    &.label {
      @apply space-y-2;
    }
  }
  input {
    @apply bg-shimmer-background-tertiary;
    @apply border;
    @apply border-shimmer-background-secondary;
    @apply rounded-lg;
    @apply py-2;
    @apply px-3;
    @apply flex;

    &:focus {
      @apply border-shimmer-action-hover;
      @apply outline-none;
    }
    &:disabled {
      @apply bg-shimmer-background-secondary;
      @apply cursor-not-allowed;
    }
  }
</style>
