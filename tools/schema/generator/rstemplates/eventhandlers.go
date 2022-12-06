// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var eventhandlersRs = map[string]string{
	// *******************************
	"eventhandlers.rs": `
use std::collections::HashMap;

use wasmlib::*;

use crate::*;

pub struct $PkgName$+EventHandlers {
    $pkg_name$+_handlers: HashMap<&'static str, fn(evt: &$PkgName$+EventHandlers, msg: &Vec<String>)>,

$#each events eventHandlerMember
}

impl IEventHandlers for $PkgName$+EventHandlers {
    fn call_handler(&self, topic: &str, params: &Vec<String>) {
        if let Some(handler) = self.$pkg_name$+_handlers.get(topic) {
            handler(self, params);
        }
    }
}

impl $PkgName$+EventHandlers {
    pub fn new() -> $PkgName$+EventHandlers {
        let mut handlers: HashMap<&str, fn(evt: &$PkgName$+EventHandlers, msg: &Vec<String>)> = HashMap::new();
$#each events eventHandler
        return $PkgName$+EventHandlers {
            $pkg_name$+_handlers: handlers,
$#each events eventHandlerMemberInit
        };
    }
$#each events eventFuncSignature
}
$#each events eventClass
`,
	// *******************************
	"eventHandlerMember": `
    pub $evt_name: fn(e: &Event$EvtName),
`,
	// *******************************
	"eventHandlerMemberInit": `
            $evt_name: |_e| {},
`,
	// *******************************
	"eventFuncSignature": `

    pub fn on_$pkg_name$+_$evt_name(&mut self, handler: fn(e: &Event$EvtName)) {
        self.$evt_name = handler;
    }
`,
	// *******************************
	"eventHandler": `
        handlers.insert("$package.$evtName", |e, m| { (e.$evt_name)(&Event$EvtName::new(m)); });
`,
	// *******************************
	"eventClass": `

pub struct Event$EvtName {
    pub timestamp: u64,
$#each event eventClassField
}

impl Event$EvtName {
    pub fn new(msg: &Vec<String>) -> Event$EvtName {
        let mut evt = EventDecoder::new(msg);
        Event$EvtName {
            timestamp: evt.timestamp(),
$#each event eventHandlerField
        }
    }
}
`,
	// *******************************
	"eventClassField": `
    pub $fld_name: $fldLangType,
`,
	// *******************************
	"eventHandlerField": `
            $fld_name: $fld_type$+_from_string(&evt.decode()),
`,
}
