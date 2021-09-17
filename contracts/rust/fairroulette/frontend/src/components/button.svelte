<script lang="ts">
  export let label: string;
  export let onClick: () => void;
</script>

<button on:click={onClick}>{label}</button>

<style lang="scss">
  button {
    background: #14cabf;
    border: 1px solid #14cabf;
    border-radius: 6px;
    font-weight: bold;
    font-size: 14px;
    line-height: 120%;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: #ffffff;
    padding: 12px;
    @media (min-width: 1024px) {
      padding-top: 12px;
      padding-bottom: 12px;
      padding-left: 31px;
      padding-right: 31px;
    }
  }
</style>
