<script lang="ts">
  export let primary: boolean = true;
</script>

<dot-component class:primary />

<style lang="scss">
  dot-component {
    @apply block w-2 h-2;
    @apply rounded-full;
    @apply bg-shimmer-action-primary;
  }
</style>
