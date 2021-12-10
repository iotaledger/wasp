package gotemplates

var paramsGo = map[string]string{
	// *******************************
	"params.go": `
$#emit goHeader
$#each func paramsFunc
`,
	// *******************************
	"paramsFunc": `
$#if params paramsFuncParams
`,
	// *******************************
	"paramsFuncParams": `
$#set Kind Param
$#set mut Immutable
$#if param paramsProxyStruct
$#set mut Mutable
$#if param paramsProxyStruct
`,
	// *******************************
	"paramsProxyStruct": `
$#set TypeName $mut$FuncName$+Params
$#each param proxyContainers

type $TypeName struct {
	id int32
}
$#each param proxyMethods
`,
}
