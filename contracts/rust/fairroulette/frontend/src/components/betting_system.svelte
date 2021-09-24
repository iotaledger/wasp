<script lang="ts">
  import { BarSelector, MultipleSelector } from "../components";
  import { placeBet, sendFaucetRequest } from "../lib/app";
  import {
    address,
    balance,
    placingBet,
    requestingFunds,
    round,
  } from "../lib/store";
  import Button from "./button.svelte";

  //TODO: Improve disable condition
  $: disabled = $placingBet;

  $: console.log(disabled);
  $: console.log(
    "$round.players.filter((_player) => _player.address === $address).length > 0",
    $round.players.filter((_player) => _player.address === $address).length > 0
  );
</script>

<div class="betting-system" class:disabled>
  <MultipleSelector {disabled} />
  <div>
    <BarSelector {disabled} />
    <div class="bet-button">
      {#if $balance > 1n}
        <Button
          label="Place bet"
          disabled={$round.betSelection === undefined || $placingBet}
          onClick={placeBet}
          loading={$placingBet}
        />
      {:else}
        <Button
          label={$requestingFunds ? "Requesting..." : "Request funds"}
          onClick={sendFaucetRequest}
          disabled={$requestingFunds}
          loading={$requestingFunds}
        />
      {/if}
    </div>
  </div>
</div>

<style lang="scss">
  .betting-system {
    display: flex;
    flex-direction: column;
    align-content: center;
    flex-wrap: wrap;
    gap: 30px;
    @media (min-width: 1024px) {
      flex-direction: row;
      justify-content: center;
      gap: 60px;
      align-items: flex-end;
    }
    &.disabled {
      opacity: 0.5;
    }
    .bet-button {
      margin-top: 24px;
      display: flex;
      justify-content: center;
    }
  }
</style>
