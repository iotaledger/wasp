package rstemplates

var structsRs = map[string]string{
	// *******************************
	"structs.rs": `
#![allow(dead_code)]
#![allow(unused_imports)]

use wasmlib::*;
use wasmlib::host::*;
$#if typedefs useTypeDefs
$#each structs structType
`,
	// *******************************
	"structType": `

pub struct $StrName {
$#each struct structField
}

impl $StrName {
    pub fn from_bytes(bytes: &[u8]) -> $StrName {
        let mut decode = BytesDecoder::new(bytes);
        $StrName {
$#each struct structDecode
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut encode = BytesEncoder::new();
$#each struct structEncode
        return encode.data();
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
            $fld_name$fld_pad : decode.$fld_type(),
`,
	// *******************************
	"structEncode": `
		encode.$fld_type($fldRef$+self.$fld_name);
`,
	// *******************************
	"structMethods": `

pub struct $mut$StrName {
    pub(crate) obj_id: i32,
    pub(crate) key_id: Key32,
}

impl $mut$StrName {
$#if mut structMethodDelete
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BYTES)
    }
$#if mut structMethodSetValue

    pub fn value(&self) -> $StrName {
        $StrName::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_BYTES))
    }
}
`,
	// *******************************
	"structMethodDelete": `
    pub fn delete(&self) {
        del_key(self.obj_id, self.key_id, TYPE_BYTES);
    }

`,
	// *******************************
	"structMethodSetValue": `

    pub fn set_value(&self, value: &$StrName) {
        set_bytes(self.obj_id, self.key_id, TYPE_BYTES, &value.to_bytes());
    }
`,
}
