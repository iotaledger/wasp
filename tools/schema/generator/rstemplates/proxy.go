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
$#set varID $Kind$FLD_NAME.get_key_id()
$#if init setInitVarID
$#if core setCoreVarID
$#if array proxyArray proxyMethods2
`,
	// *******************************
	"setInitVarID": `
$#set varID idx_map(IDX_$Kind$FLD_NAME)
`,
	// *******************************
	"proxyMethods2": `
$#if map proxyMap proxyMethods3
`,
	// *******************************
	"proxyMethods3": `
$#if basetype proxyBaseType proxyMethods4
`,
	// *******************************
	"proxyMethods4": `
$#if typedef proxyTypeDef proxyStruct
`,
	// *******************************
	"setCoreVarID": `
$#set varID $Kind$FLD_NAME.get_key_id()
`,
	// *******************************
	"proxyArray": `
    pub fn $fld_name(&self) -> ArrayOf$mut$FldType {
		let arr_id = get_object_id(self.id, $varID, $arrayTypeID | $fldTypeID);
		ArrayOf$mut$FldType { obj_id: arr_id }
	}
`,
	// *******************************
	"proxyMap": `
$#if this proxyMapThis proxyMapOther
`,
	// *******************************
	"proxyMapThis": `
    pub fn $fld_name(&self) -> Map$fldMapKey$+To$mut$FldType {
		Map$fldMapKey$+To$mut$FldType { obj_id: self.id }
	}
`,
	// *******************************
	"proxyMapOther": `
    pub fn $fld_name(&self) -> Map$fldMapKey$+To$mut$FldType {
		let map_id = get_object_id(self.id, $varID, TYPE_MAP);
		Map$fldMapKey$+To$mut$FldType { obj_id: map_id }
	}
`,
	// *******************************
	"proxyBaseType": `
    pub fn $fld_name(&self) -> Sc$mut$FldType {
		Sc$mut$FldType::new(self.id, $varID)
	}
`,
	// *******************************
	"proxyTypeDef": `
$#emit setVarType
    pub fn $old_name(&self) -> $mut$OldType {
		let sub_id = get_object_id(self.id, $varID, $varType);
		$mut$OldType { obj_id: sub_id }
	}
`,
	// *******************************
	"proxyStruct": `
    pub fn $fld_name(&self) -> $mut$FldType {
		$mut$FldType { obj_id: self.id, key_id: $varID }
	}
`,
}
