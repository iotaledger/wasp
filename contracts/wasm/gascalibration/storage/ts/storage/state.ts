// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

import * as wasmtypes from "wasmlib/wasmtypes";
import * as sc from "./index";

export class ArrayOfImmutableUint32 extends wasmtypes.ScProxy {

	length(): u32 {
		return this.proxy.length();
	}

	getUint32(index: u32): wasmtypes.ScImmutableUint32 {
		return new wasmtypes.ScImmutableUint32(this.proxy.index(index));
	}
}

export class ImmutablestorageState extends wasmtypes.ScProxy {
	v(): sc.ArrayOfImmutableUint32 {
		return new sc.ArrayOfImmutableUint32(this.proxy.root(sc.StateV));
	}
}

export class ArrayOfMutableUint32 extends wasmtypes.ScProxy {

	appendUint32(): wasmtypes.ScMutableUint32 {
		return new wasmtypes.ScMutableUint32(this.proxy.append());
	}

	clear(): void {
		this.proxy.clearArray();
	}

	length(): u32 {
		return this.proxy.length();
	}

	getUint32(index: u32): wasmtypes.ScMutableUint32 {
		return new wasmtypes.ScMutableUint32(this.proxy.index(index));
	}
}

export class MutablestorageState extends wasmtypes.ScProxy {
	asImmutable(): sc.ImmutablestorageState {
		return new sc.ImmutablestorageState(this.proxy);
	}

	v(): sc.ArrayOfMutableUint32 {
		return new sc.ArrayOfMutableUint32(this.proxy.root(sc.StateV));
	}
}
