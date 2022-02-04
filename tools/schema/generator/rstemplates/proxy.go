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
	// TODO when will this be called, and if so, fix it
	"proxyOtherType": `
$#if typedef proxyTypeDef proxyStruct
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
	// TODO when will this be called, and if so, fix it
	"proxyTypeDef": `
$#emit setVarType
    pub fn $old_name(&self) -> $mut$OldType {
		let sub_id = get_object_id(self.id, $varID, $varType);
		$mut$OldType { obj_id: sub_id }
	}
`,
	// *******************************
	// TODO when will this be called, and if so, fix it
	"proxyStruct": `
    pub fn $fld_name(&self) -> $mut$FldType {
		$mut$FldType { obj_id: self.id, key_id: $varID }
	}
`,
}
