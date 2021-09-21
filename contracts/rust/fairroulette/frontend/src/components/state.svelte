<script lang="ts">
  export let title: string;
  export let subtitle: string = '';
  export let description: string;
</script>

<div class="message">
  <h1 class="title">{title}</h1>
  {#if subtitle}
    <div class="subtitle">{subtitle}</div>
  {/if}
  {#if description}
    <div class="description">{description}</div>
  {/if}
  <div />
</div>

<style lang="scss">
  .message {
    font-family: 'Metropolis Bold';
    text-align: center;
    .title {
      text-align: center;
      color: var(--white);
    }
    .subtitle {
      font-family: 'Metropolis Bold';
      font-size: 24px;
      line-height: 120%;
      letter-spacing: 0.02em;
      color: var(--mint-green);
      margin-bottom: 8px;
    }
    .description {
      font-size: 16px;
      line-height: 150%;
      letter-spacing: 0.75px;
      color: var(--gray-5);
    }
  }
</style>
