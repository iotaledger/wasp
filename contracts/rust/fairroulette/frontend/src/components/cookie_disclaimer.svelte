<script lang="ts">
  export let allow: () => void;
  export let decline: () => void;
</script>

<div class="cookie-disclaimer">
  <button on:click={decline}>Decline cookies</button>
  <button on:click={allow}>Allow cookies</button>
</div>

<style lang="scss">
  .cookie-disclaimer {
    padding: 20px;
    background: turquoise;
    position: fixed;
    bottom: 20px;
    right: 20px;
  }
</style>
