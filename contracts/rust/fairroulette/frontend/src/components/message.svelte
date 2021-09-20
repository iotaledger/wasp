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
    .title {
      font-weight: 600;
      font-size: 32px;
      line-height: 36px;
      text-align: center;
      letter-spacing: 0.75px;
      color: #ffffff;
      text-align: center;
      @media (min-width: 1024px) {
        font-size: 48px;
      }
    }
    .subtitle {
      font-weight: 600;
      font-size: 24px;
      line-height: 120%;
      text-align: center;
      letter-spacing: 0.02em;
      color: #00e0ca;
    }
    .description {
    font-size: 16px;
    line-height: 150%;
    text-align: center;
    letter-spacing: 0.75px;
    color: #909fbe;
  }
</style>
