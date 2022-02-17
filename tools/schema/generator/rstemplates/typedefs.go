package rstemplates

var typedefsRs = map[string]string{
	// *******************************
	"typedefs.rs": `
#![allow(dead_code)]

$#if core else useWasmLib
use crate::*;
$#each typedef typedefProxy
`,
	// *******************************
	"typedefProxy": `
$#set mut Immutable
$#if array typedefProxyArray
$#if array typedefProxyAlias
$#if map typedefProxyMap
$#if map typedefProxyAlias
$#set mut Mutable
$#if array typedefProxyArray
$#if array typedefProxyAlias
$#if map typedefProxyMap
$#if map typedefProxyAlias
`,
	// *******************************
	"typedefProxyAlias": `

pub type $mut$FldName = $proxy;
`,
	// *******************************
	"typedefProxyArray": `
$#set proxy ArrayOf$mut$FldType
$#if exist else typedefProxyArrayNew
`,
	// *******************************
	"typedefProxyArrayNew": `

#[derive(Clone)]
pub struct $proxy {
	pub(crate) proxy: Proxy,
}

impl $proxy {
$#if mut typedefProxyArrayMut
    pub fn length(&self) -> u32 {
        self.proxy.length()
    }

$#if basetype typedefProxyArrayNewBaseType typedefProxyArrayNewOtherType
}
$#set exist $proxy
`,
	// *******************************
	"typedefProxyArrayMut": `
$#if basetype typedefProxyArrayAppendBaseType typedefProxyArrayAppendOtherType
	pub fn clear(&self) {
        self.proxy.clear_array();
    }

`,
	// *******************************
	"typedefProxyArrayAppendBaseType": `
	pub fn append_$fld_type(&self) -> Sc$mut$FldType {
		Sc$mut$FldType::new(self.proxy.append())
	}

`,
	// *******************************
	"typedefProxyArrayAppendOtherType": `

	pub fn append_$fld_type(&self) -> $mut$FldType {
		$mut$FldType { proxy: self.proxy.append() }
	}
`,
	// *******************************
	"typedefProxyArrayNewBaseType": `
    pub fn get_$fld_type(&self, index: u32) -> Sc$mut$FldType {
        Sc$mut$FldType::new(self.proxy.index(index))
    }
`,
	// *******************************
	"typedefProxyArrayNewOtherType": `

	pub fn get_$fld_type(&self, index: u32) -> $mut$FldType {
		$mut$FldType { proxy: self.proxy.index(index) }
	}
`,
	// *******************************
	"typedefProxyMap": `
$#set proxy Map$FldMapKey$+To$mut$FldType
$#if exist else typedefProxyMapNew
`,
	// *******************************
	"typedefProxyMapNew": `

#[derive(Clone)]
pub struct $proxy {
	pub(crate) proxy: Proxy,
}

impl $proxy {
$#if mut typedefProxyMapMut
$#if basetype typedefProxyMapNewBaseType typedefProxyMapNewOtherType
}
$#set exist $proxy
`,
	// *******************************
	"typedefProxyMapMut": `
    pub fn clear(&self) {
        self.proxy.clear_map();
    }

`,
	// *******************************
	"typedefProxyMapNewBaseType": `
    pub fn get_$fld_type(&self, key: $fldKeyRef$fldKeyParamLangType) -> Sc$mut$FldType {
        Sc$mut$FldType::new(self.proxy.key(&$fld_map_key$+_to_bytes(key)))
    }
`,
	// *******************************
	"typedefProxyMapNewOtherType": `
    pub fn get_$fld_type(&self, key: $fldKeyRef$fldKeyParamLangType) -> $mut$FldType {
        $mut$FldType { proxy: self.proxy.key(&$fld_map_key$+_to_bytes(key)) }
    }
`,
}
