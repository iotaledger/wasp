// encapsulates standard host entities into a simple interface

use super::host::set_string;
use super::immutable::ScImmutableMap;
use super::keys::key_log;
use super::keys::key_trace;
use super::mutable::ScMutableMap;
use super::mutable::ScMutableMapArray;
use super::mutable::ScMutableString;
use crate::wasp::immutable::ScImmutableStringArray;

pub struct ScContext {
    root: ScMutableMap,
}

impl ScContext {
    pub fn new() -> ScContext {
        ScContext { root: ScMutableMap::new_root_map() }
    }

    pub fn balance(&self, color: &str) -> i64 {
        let color = if color.is_empty() { "iota" } else { color };
        self.root.get_map("balance").get_int(color).value()
    }

    pub fn colors(&self) -> ScImmutableStringArray {
        self.root.get_string_array("colors").immutable()
    }

    pub fn config(&self) -> ScImmutableMap {
        self.root.get_map("config").immutable()
    }

    pub fn error(&self) -> ScMutableString {
        self.root.get_string("error")
    }

    pub fn log(&self, text: &str) {
        set_string(1, key_log(), text)
    }

    pub fn owner(&self) -> String {
        self.root.get_string("owner").value()
    }

    pub fn params(&self) -> ScImmutableMap {
        self.root.get_map("params").immutable()
    }

    pub fn random(&self, max: i64) -> i64 {
        (self.root.get_int("random").value() as u64 % max as u64) as i64
    }

    pub fn request_balance(&self, color: &str) -> i64 {
        let color = if color.is_empty() { "iota" } else { color };
        self.root.get_map("reqBalance").get_int(color).value()
    }

    pub fn request_colors(&self) -> ScImmutableStringArray {
        self.root.get_string_array("reqColors").immutable()
    }

    pub fn request_hash(&self) -> String {
        self.root.get_string("reqHash").value()
    }

    pub fn requests(&self) -> ScMutableMapArray {
        self.root.get_map_array("requests")
    }

    pub fn sc_address(&self) -> String {
        self.root.get_string("scAddress").value()
    }

    pub fn sender(&self) -> String {
        self.root.get_string("sender").value()
    }

    pub fn state(&self) -> ScMutableMap {
        self.root.get_map("state")
    }

    pub fn timestamp(&self) -> i64 {
        self.root.get_int("timestamp").value()
    }

    pub fn trace(&self, text: &str) {
        set_string(1, key_trace(), text)
    }

    pub fn transfers(&self) -> ScMutableMapArray {
        self.root.get_map_array("transfers")
    }
}
