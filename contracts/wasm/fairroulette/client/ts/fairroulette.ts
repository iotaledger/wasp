// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as events from "./events"
import * as service from "./service"

const client = new BasicClient(config);
const fairRouletteService = new service.FairRouletteService(client, config.ChainId);

export function onFairRouletteBet(event: events.EventBet): void {
}

export function onFairRoulettePayout(event: events.EventPayout): void {
}

export function onFairRouletteRound(event: events.EventRound): void {
}

export function onFairRouletteStart(event: events.EventStart): void {
}

export function onFairRouletteStop(event: events.EventStop): void {
}

export function onFairRouletteWinner(event: events.EventWinner): void {
}
