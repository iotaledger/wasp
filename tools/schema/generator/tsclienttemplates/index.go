package tsclienttemplates

var indexTs = map[string]string{
	// *******************************
	"index.ts": `
$#if events exportEvents
export * from "./service";
`,
	// *******************************
	"exportEvents": `
export * from "./events";
`,
}
