<script lang="ts">
  import {
    GAME_RUNNING_STATE,
    START_GAME_STATE,
    state,
    fundsRequested,
  } from './../lib/store';
  const MESSAGES = {
    [START_GAME_STATE]: {
      title: 'Start game',
      description: 'Start game!',
    },

    [GAME_RUNNING_STATE]: {
      title: 'GAME RUNNING!',
      description: 'The round ends in 50 seconds.',
    },
  };

  const ADD_FUNDS_MESSAGE = {
    title: 'Add funds',
    description:
      'To start the demo, you first need to request funds for your wallet. Those coins are generated from the dev-net.',
  };
</script>

<div class="message">
  <h2 class="title">
    {$fundsRequested ? MESSAGES[$state].title : ADD_FUNDS_MESSAGE.title}
  </h2>
  <div class="description">
    {$fundsRequested
      ? MESSAGES[$state].description
      : ADD_FUNDS_MESSAGE.description}
  </div>
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
      padding: 16px;
      font-size: 16px;
      line-height: 150%;
      letter-spacing: 0.75px;
      color: var(--gray-5);
    }
  }
</style>
