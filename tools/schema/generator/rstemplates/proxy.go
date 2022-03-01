// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rstemplates

var proxyRs = map[string]string{
	// *******************************
	"proxyContainers": `
$#if array typedefProxyArray
$#if map typedefProxyMap
`,
	// *******************************
	"proxyMethods": `
$#if separator newline
$#set separator $true
$#if array proxyArray proxyMethods2
`,
	// *******************************
	"proxyMethods2": `
$#if map proxyMap proxyMethods3
`,
	// *******************************
	"proxyMethods3": `
$#if basetype proxyBaseType proxyOtherType
`,
	// *******************************
	"proxyArray": `
    pub fn $fld_name(&self) -> ArrayOf$mut$FldType {
		ArrayOf$mut$FldType { proxy: self.proxy.root($Kind$FLD_NAME) }
	}
`,
	// *******************************
	"proxyMap": `
$#if this proxyMapThis proxyMapOther
`,
	// *******************************
	"proxyMapThis": `
    pub fn $fld_name(&self) -> Map$FldMapKey$+To$mut$FldType {
		Map$FldMapKey$+To$mut$FldType { proxy: self.proxy.clone() }
	}
`,
	// *******************************
	"proxyMapOther": `
    pub fn $fld_name(&self) -> Map$FldMapKey$+To$mut$FldType {
		Map$FldMapKey$+To$mut$FldType { proxy: self.proxy.root($Kind$FLD_NAME) }
	}
`,
	// *******************************
	"proxyBaseType": `
    pub fn $fld_name(&self) -> Sc$mut$FldType {
		Sc$mut$FldType::new(self.proxy.root($Kind$FLD_NAME))
	}
`,
	// *******************************
	"proxyOtherType": `
    pub fn $fld_name(&self) -> $mut$FldType {
		$mut$FldType { proxy: self.proxy.root($Kind$FLD_NAME) }
	}
`,
}
