<script lang="ts">
  import { ButtonHtmlType } from '$lib/enums'
    import { handleEnterKeyDown } from '$lib/utils';

  export let compact: boolean = false
  export let title: string = ''
  export let url: string = ''
  export let htmlType: ButtonHtmlType = ButtonHtmlType.Button
  export let disabled: boolean = false
  export let danger: boolean = false
  export let onClick: () => void = () => {}
  export let isExternal: boolean = false
  export let stretch: boolean = false
  export let busy: boolean = false
  export let secondary: boolean = false

  $: href = htmlType === ButtonHtmlType.Link ? url : undefined
  $: target = htmlType === ButtonHtmlType.Link && isExternal ? '_blank' : undefined
  $: rel = htmlType === ButtonHtmlType.Link && isExternal ? 'noopener noreferrer' : undefined
</script>

<svelte:element
  this={htmlType}
  {href}
  {target}
  {rel}
  disabled={disabled || busy}
  on:click={onClick}
  on:keydown={(event) => handleEnterKeyDown(event, onClick)}
  class:link={htmlType === ButtonHtmlType.Link}
  class:danger
  class:w-full={stretch}
  class:secondary
  class:compact
>
  {busy ? 'Loading...' : title}
</svelte:element>

<style lang="scss">
  button,
  a {
      @apply bg-shimmer-action-primary;
      @apply text-shimmer-text-primary;
      @apply font-semibold;
      @apply p-3;
      @apply rounded-md;
      @apply text-center;

      &:disabled {
          @apply opacity-70;
          @apply cursor-not-allowed;
      }
      &.link {
          @apply bg-transparent;
          @apply text-zinc-800;
          @apply border-2;
          @apply border-zinc-800;
          &:hover {
              @apply bg-zinc-800;
              @apply text-white;
          }
      }

      &.secondary {
          @apply bg-shimmer-background;
          @apply border-0;
          @apply text-shimmer-text-secondary;

      }

      &.compact {
          @apply p-1;
      }
  }
</style>
