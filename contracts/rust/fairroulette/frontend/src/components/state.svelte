<script lang="ts">
  import { round, showAddFunds, timestamp, timeToFinished } from '../lib/store';
  enum State {
    Running = 'Running',
    Start = 'Start',
    AddFunds = 'AddFunds',
  }

  let state: State;

  $: MESSAGES = {
    [State.Running]: {
      title: 'Game Running!',
      description: `The round ends in ${$timeToFinished} seconds.`,
    },
    [State.Start]: {
      title: 'Start game',
      description: `Lorem ipsum`,
    },
    [State.AddFunds]: {
      title: 'Add funds',
      description:
        'To start the demo, you first need to request funds for your wallet. Those coins are generated from the dev-net.',
    },
  };

  $: state = $showAddFunds
    ? State.AddFunds
    : $round.active
    ? State.Running
    : State.Start;
</script>

{#if state}
  <div class="message">
    <h2 class="title">
      {MESSAGES[state].title}
    </h2>
    <div class="description">
      {MESSAGES[state].description}
    </div>
    <div />
  </div>
{/if}

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
