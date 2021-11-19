package rstemplates

var keysRs = map[string]string{
	// *******************************
	"keys.rs": `
#![allow(dead_code)]

use wasmlib::*;

use crate::*;

$#set constPrefix PARAM_
$#each params constFieldIdx

$#set constPrefix RESULT_
$#each results constFieldIdx

$#set constPrefix STATE_
$#each state constFieldIdx

pub const KEY_MAP_LEN: usize = $maxIndex;

pub const KEY_MAP: [&str; KEY_MAP_LEN] = [
$#set constPrefix PARAM_
$#each params constFieldKey
$#set constPrefix RESULT_
$#each results constFieldKey
$#set constPrefix STATE_
$#each state constFieldKey
];

pub static mut IDX_MAP: [Key32; KEY_MAP_LEN] = [Key32(0); KEY_MAP_LEN];

pub fn idx_map(idx: usize) -> Key32 {
    unsafe {
        IDX_MAP[idx]
    }
}
`,
	// *******************************
	"constFieldIdx": `
pub(crate) const IDX_$constPrefix$FLD_NAME$fld_pad : usize = $fldIndex;
`,
	// *******************************
	"constFieldKey": `
	$constPrefix$FLD_NAME,
`,
}
