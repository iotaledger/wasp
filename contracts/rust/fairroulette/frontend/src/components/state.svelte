<script lang="ts">
  export let phase: string = 'GAME_STARTED';

  const MESSAGES = {
    GAME_STARTED: {
      title: 'Start game',
      subtitle: 'This is a subtitle',
      description: 'The round starts in 50 seconds.',
    },
    GAME_STOPPED: {
      title: 'Start game',
      subtitle: 'This is a subtitle',
      description: 'The round starts in 50 seconds.',
    },
    WINNER: {
      title: 'Start game',
      subtitle: 'This is a subtitle',
      description: 'The round starts in 50 seconds.',
    },
    ADD_FUNDS: {
      title: 'Start game',
      subtitle: 'This is a subtitle',
      description: 'The round starts in 50 seconds.',
    },
  };
</script>

<div class="message">
  <h2 class="title">{MESSAGES[phase].title}</h2>
  {#if MESSAGES[phase].subtitle}
    <div class="subtitle">{MESSAGES[phase].subtitle}</div>
  {/if}
  {#if MESSAGES[phase].description}
    <div class="description">{MESSAGES[phase].description}</div>
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
