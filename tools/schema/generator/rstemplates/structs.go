// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var structsRs = map[string]string{
	// *******************************
	"structs.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

$#if core useCrate useWasmLib
$#each structs structType
`,
	// *******************************
	"structType": `

#[derive(Clone)]
pub struct $StrName {
$#each struct structField
}

impl $StrName {
    pub fn from_bytes(bytes: &[u8]) -> $StrName {
        let mut dec = WasmDecoder::new(bytes);
        $StrName {
$#each struct structDecode
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut enc = WasmEncoder::new();
$#each struct structEncode
        enc.buf()
    }
}
$#set mut Immutable
$#emit structMethods
$#set mut Mutable
$#emit structMethods
`,
	// *******************************
	"structField": `
    pub $fld_name$fld_pad : $fldLangType, $fldComment
`,
	// *******************************
	"structDecode": `
            $fld_name$fld_pad : $fld_type$+_decode(&mut dec),
`,
	// *******************************
	"structEncode": `
		$fld_type$+_encode(&mut enc, $fldRef$+self.$fld_name);
`,
	// *******************************
	"structMethods": `

#[derive(Clone)]
pub struct $mut$StrName {
    pub(crate) proxy: Proxy,
}

impl $mut$StrName {
$#if mut structMethodDelete
    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }
$#if mut structMethodSetValue

    pub fn value(&self) -> $StrName {
        $StrName::from_bytes(&self.proxy.get())
    }
}
`,
	// *******************************
	"structMethodDelete": `
    pub fn delete(&self) {
        self.proxy.delete();
    }

`,
	// *******************************
	"structMethodSetValue": `

    pub fn set_value(&self, value: &$StrName) {
        self.proxy.set(&value.to_bytes());
    }
`,
}
