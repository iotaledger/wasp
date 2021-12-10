package gotemplates

var resultsGo = map[string]string{
	// *******************************
	"results.go": `
$#emit goHeader
$#each func resultsFunc
`,
	// *******************************
	"resultsFunc": `
$#if results resultsFuncResults
`,
	// *******************************
	"resultsFuncResults": `
$#set Kind Result
$#set mut Immutable
$#if result resultsProxyStruct
$#set mut Mutable
$#if result resultsProxyStruct
`,
	// *******************************
	"resultsProxyStruct": `
$#set TypeName $mut$FuncName$+Results
$#each result proxyContainers

type $TypeName struct {
	id int32
}
$#each result proxyMethods
`,
}
