package rstemplates

var typedefsRs = map[string]string{
	// *******************************
	"typedefs.rs": `
#![allow(dead_code)]

use wasmlib::*;
use wasmlib::host::*;
$#if structs useStructs
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

#[derive(Clone, Copy)]
pub struct $proxy {
	pub(crate) obj_id: i32,
}

impl $proxy {
$#if mut typedefProxyArrayClear
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }

$#if basetype typedefProxyArrayNewBaseType typedefProxyArrayNewOtherType
}
$#set exist $proxy
`,
	// *******************************
	"typedefProxyArrayClear": `
    pub fn clear(&self) {
        clear(self.obj_id);
    }

`,
	// *******************************
	"typedefProxyArrayNewBaseType": `
    pub fn get_$fld_type(&self, index: i32) -> Sc$mut$FldType {
        Sc$mut$FldType::new(self.obj_id, Key32(index))
    }
`,
	// *******************************
	"typedefProxyArrayNewOtherType": `
$#if typedef typedefProxyArrayNewOtherTypeTypeDef typedefProxyArrayNewOtherTypeStruct
`,
	// *******************************
	"typedefProxyArrayNewOtherTypeTypeDef": `
$#emit setVarType
	pub fn get_$old_type(&self, index: i32) -> $mut$OldType {
		let sub_id = get_object_id(self.obj_id, Key32(index), $varType);
		$mut$OldType { obj_id: sub_id }
	}
`,
	// *******************************
	"typedefProxyArrayNewOtherTypeStruct": `
	pub fn get_$fld_type(&self, index: i32) -> $mut$FldType {
		$mut$FldType { obj_id: self.obj_id, key_id: Key32(index) }
	}
`,
	// *******************************
	"typedefProxyMap": `
$#set proxy Map$fldMapKey$+To$mut$FldType
$#if exist else typedefProxyMapNew
`,
	// *******************************
	"typedefProxyMapNew": `

#[derive(Clone, Copy)]
pub struct $proxy {
	pub(crate) obj_id: i32,
}

impl $proxy {
$#if mut typedefProxyMapClear
$#if basetype typedefProxyMapNewBaseType typedefProxyMapNewOtherType
}
$#set exist $proxy
`,
	// *******************************
	"typedefProxyMapClear": `
    pub fn clear(&self) {
        clear(self.obj_id);
    }

`,
	// *******************************
	"typedefProxyMapNewBaseType": `
    pub fn get_$fld_type(&self, key: $fldKeyRef$fldKeyParamLangType) -> Sc$mut$FldType {
        Sc$mut$FldType::new(self.obj_id, $fldKeyToKey32)
    }
`,
	// *******************************
	"typedefProxyMapNewOtherType": `
$#if typedef typedefProxyMapNewOtherTypeTypeDef typedefProxyMapNewOtherTypeStruct
`,
	// *******************************
	"typedefProxyMapNewOtherTypeTypeDef": `
$#emit setVarType
    pub fn get_$old_type(&self, key: $oldKeyRef$oldKeyParamLangType) -> $mut$OldType {
        let sub_id = get_object_id(self.obj_id, $oldKeyToKey32, $varType);
        $mut$OldType { obj_id: sub_id }
    }
`,
	// *******************************
	"typedefProxyMapNewOtherTypeStruct": `
    pub fn get_$fld_type(&self, key: $fldKeyRef$fldKeyParamLangType) -> $mut$FldType {
        $mut$FldType { obj_id: self.obj_id, key_id: $fldKeyToKey32 }
    }
`,
}
