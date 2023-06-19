// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var eventsRs = map[string]string{
	// *******************************
	"events.rs": `
#![allow(dead_code)]
#![allow(unused_mut)]
$#if core useCrate useWasmLib

$#set TypeName $Package$+Events
pub struct $TypeName {
}

impl $TypeName {
$#each events eventFunc
}
`,
	// *******************************
	"eventFunc": `
$#set endFunc ) {
$#if event eventSetEndFunc

$#each eventComment _eventComment
	pub fn $evt_name(&self$endFunc
$#each event eventParam
$#if event eventEndFunc2
		let mut enc = EventEncoder::new("$package.$evtName");
$#each event eventEmit
		EventEncoder::emit(&enc);
	}
`,
	// *******************************
	"eventParam": `
$#each fldComment _eventParamComment
        $fld_name: $fldRef$fldParamLangType,
`,
	// *******************************
	"eventEmit": `
		$fld_type$+_encode(&mut enc, $fldRef$fld_name);
`,
	// *******************************
	"eventSetEndFunc": `
$#set endFunc ,
`,
	// *******************************
	"eventEndFunc2": `
    ) {
`,
}
