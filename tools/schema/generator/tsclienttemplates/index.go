package tsclienttemplates

var indexTs = map[string]string{
	// *******************************
	"index.ts": `
export * from "./$package";
export * from "./events";
export * from "./service";
`,
}
