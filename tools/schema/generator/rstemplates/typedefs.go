package rstemplates

var typedefsRs = map[string]string{
	// *******************************
	"typedefs.rs": `
// @formatter:off

#![allow(dead_code)]

use wasmlib::*;
use wasmlib::host::*;
$#if structs useStructs
$#each typedef typedefProxy

// @formatter:on
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
$#set OldType $FldType
$#if typedef typedefProxyArrayNewOtherTypeTypeDef typedefProxyArrayNewOtherTypeStruct
`,
	// *******************************
	"typedefProxyArrayNewOtherTypeTypeDef": `
$#set varType TYPE_MAP
$#if array setVarTypeArray

	pub fn Get$OldType(&self, index: i32) -> $mut$OldType {
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
$#set proxy Map$FldMapKey$+To$mut$FldType
$#if exist else typedefProxyMapNew
`,
	// *******************************
	"typedefProxyMapNew": `

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
    pub fn get_$fld_type(&self, key: $FldMapKeyLangType) -> Sc$mut$FldType {
        Sc$mut$FldType::new(self.obj_id, key.get_key_id())
    }
`,
	// *******************************
	"typedefProxyMapNewOtherType": `
$#set old_type $fld_type
$#set OldType $FldType
$#set OldMapKeyLangType $FldMapKeyLangType
$#set OldMapKeyKey $FldMapKeyKey
$#if typedef typedefProxyMapNewOtherTypeTypeDef typedefProxyMapNewOtherTypeStruct
`,
	// *******************************
	"typedefProxyMapNewOtherTypeTypeDef": `
$#set varType TYPE_MAP
$#if array setVarTypeArray
    pub fn get_$old_type(&self, key: $OldMapKeyLangType) -> $mut$OldType {
        let sub_id = get_object_id(self.obj_id, key.get_key_id(), $varType);
        $mut$OldType { obj_id: sub_id }
    }
`,
	// *******************************
	"typedefProxyMapNewOtherTypeStruct": `
    pub fn get_$old_type(&self, key: $OldMapKeyLangType) -> $mut$OldType {
        $mut$OldType { obj_id: self.obj_id, key_id: key.get_key_id() }
    }
`,
	// *******************************
	"setVarTypeArray": `
$#set varType $ArrayTypeID$space|$space$FldTypeID
`,
}
