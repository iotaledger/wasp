<script lang="ts">
    import type { ComponentType } from "svelte";
  export let component: ComponentType
</script>

<slot>
  <svelte:component this={component} />
</slot>