// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as events from "./events"
import * as service from "./service"

const client = new BasicClient(config);
const testWasmLibService = new service.TestWasmLibService(client, config.ChainId);

export function onTestWasmLibTest(event: events.EventTest): void {
}
