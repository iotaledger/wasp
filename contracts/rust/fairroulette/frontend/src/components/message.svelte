<script lang="ts">
  export let title: string;
  export let subtitle: string = '';
  export let description: string;
</script>

<div class="message">
  <div>{title}</div>
  {#if subtitle}
    <div>{subtitle}</div>
  {/if}
  {#if description}
    <div>{description}</div>
  {/if}
  <div />
</div>

<style lang="scss">
  .message {
    background-color: white;
    text-align: center;
  }
</style>
