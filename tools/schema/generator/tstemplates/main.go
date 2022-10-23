// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tstemplates

var mainTs = map[string]string{
	// *******************************
	"../main.ts": `
import * as sc from "./$package";

export function on_call(index: i32): void {
    sc.onLoad(index);
}

export function on_load(): void {
    sc.onLoad(-1);
}
`,
}
