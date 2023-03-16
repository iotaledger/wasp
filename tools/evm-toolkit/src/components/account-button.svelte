<script lang="ts">
  export let disabled: boolean = false;
  export let onClick: () => void = () => {};
  export let title: string = '';
</script>

<button class="dot-primary" on:click={onClick} {disabled}>{title}</button>

<style lang="scss">
  button {
    @apply rounded-md;
    cursor: pointer;
    width: 100%;
    @apply bg-shimmer-background;
    font-weight: 600;
    font-size: 16px;
    line-height: 24px;
    @apply text-white;
    padding: 6px 12px 6px 36px;
    height: 48px;
    display: flex;
    flex-direction: row;
    align-items: center;
    border: none;
    position: relative;
  }
</style>
