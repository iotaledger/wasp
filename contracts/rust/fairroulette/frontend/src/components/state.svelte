<script lang="ts">
  export let title: string;
  export let subtitle: string = '';
  export let description: string;
</script>

<div class="message">
  <h1 class="title">{title}</h1>
  {#if subtitle}
    <span class="subtitle">{subtitle}</span>
  {/if}
  {#if description}
    <span class="description">{description}</span>
  {/if}
  <div />
</div>

<style lang="scss">
  .message {
    text-align: center;
    .title {
      text-align: center;
      color: var(--white);
    }
    .subtitle {
      font-weight: 600;
      font-size: 24px;
      line-height: 120%;
      letter-spacing: 0.02em;
      color: #00e0ca;
    }
    .description {
      font-size: 16px;
      line-height: 150%;
      letter-spacing: 0.75px;
      color: #909fbe;
    }
  }
</style>
