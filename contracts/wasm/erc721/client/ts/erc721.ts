// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as events from "./events"
import * as service from "./service"

const client = new BasicClient(config);
const erc721Service = new service.Erc721Service(client, config.ChainId);

export function onErc721Approval(event: events.EventApproval): void {
}

export function onErc721ApprovalForAll(event: events.EventApprovalForAll): void {
}

export function onErc721Init(event: events.EventInit): void {
}

export function onErc721Mint(event: events.EventMint): void {
}

export function onErc721Transfer(event: events.EventTransfer): void {
}
