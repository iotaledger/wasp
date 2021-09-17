<script lang="ts">
  export let label: string;
  export let onClick: () => void;
</script>

<button on:click={onClick}>{label}</button>

<style lang="scss">
  button {
    background: turquoise;
  }
</style>
