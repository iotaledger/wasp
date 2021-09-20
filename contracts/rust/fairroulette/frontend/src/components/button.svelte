<script lang="ts">
  export let label: string;
  export let disabled: boolean = false;
  export let onClick: () => void;
</script>

<button {disabled} class:disabled on:click={onClick}>{label}</button>

<style lang="scss">
  button {
    background: var(--mint-green-light);
    border: 1px solid var(--mint-green-light);
    border-radius: 6px;
    font-weight: bold;
    font-size: 14px;
    line-height: 120%;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--white);
    padding: 12px;
    @media (min-width: 1024px) {
      padding-top: 12px;
      padding-bottom: 12px;
      padding-left: 31px;
      padding-right: 31px;
    }
    &.disabled {
      background: transparent;
      color: var(--gray-5);
    }
  }
</style>
