import * as wc from "../../index.js"

export function clickMe() {
    window.alert("Hey! Don't click me!");
}

export function testMe() {
    //window.alert("Hey! Don't test me!");
    let client = wc.WasmClientService.DefaultWasmClientService();
}
