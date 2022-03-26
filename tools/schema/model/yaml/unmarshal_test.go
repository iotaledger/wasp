package yaml_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/iotaledger/wasp/tools/schema/model"
	"github.com/iotaledger/wasp/tools/schema/model/yaml"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	type args struct {
		path string
	}
	type wants struct {
		out *model.SchemaDef
	}
	type test struct {
		args  args
		wants wants
	}

	tests := map[string]func(t *testing.T) test{
		"successfully test3": func(t *testing.T) test {
			return test{
				args: args{
					path: "testdata/test3.yaml",
				},
				wants: wants{
					out: &model.SchemaDef{
						Name: model.DefElt{
							Val:  "SchemaComment",
							Line: 1,
						},
						Description: model.DefElt{
							Val:  "test description",
							Line: 2,
						},
						Events: model.DefMapMap{
							model.DefElt{Val: "TestEvent1", Line: 7}: &model.DefMap{
								model.DefElt{Val: "eventParam11", Line: 8}: &model.DefElt{
									Val:  "String",
									Line: 8,
								},
							},
							model.DefElt{Val: "TestEvent2", Line: 9}: &model.DefMap{
								model.DefElt{Val: "eventParam21", Line: 10}: &model.DefElt{
									Val:  "String",
									Line: 10,
								},
								model.DefElt{Val: "eventParam22", Line: 11}: &model.DefElt{
									Val:  "String",
									Line: 11,
								},
							},
						},
						Structs: model.DefMapMap{
							model.DefElt{Val: "TestStruct1", Line: 16}: &model.DefMap{
								model.DefElt{Val: "x", Line: 17}: &model.DefElt{
									Val:  "Int32",
									Line: 17,
								},
								model.DefElt{Val: "y", Line: 18}: &model.DefElt{
									Val:  "Int32",
									Line: 18,
								},
							},
							model.DefElt{Val: "TestStruct2", Line: 20}: &model.DefMap{
								model.DefElt{Val: "x", Line: 21}: &model.DefElt{
									Val:  "Int32",
									Line: 22,
								},
							},
						},
						Funcs: model.FuncDefMap{
							"TestFunc1": &model.FuncDef{
								Access: model.DefElt{
									Val:  "owner",
									Line: 26,
								},
								Params: model.DefMap{
									model.DefElt{Val: "name", Line: 28}: &model.DefElt{
										Val:  "String",
										Line: 28,
									},
									model.DefElt{Val: "value", Line: 29}: &model.DefElt{
										Val:  "String",
										Line: 29,
									},
								},
								Results: model.DefMap{
									model.DefElt{Val: "length", Line: 31}: &model.DefElt{
										Val:  "Uint32",
										Line: 31,
									},
								},
							},
							"TestFunc2": &model.FuncDef{
								Access: model.DefElt{
									Val:  "owner",
									Line: 33,
								},
								Params: model.DefMap{
									model.DefElt{Val: "name", Line: 35}: &model.DefElt{
										Val:  "String",
										Line: 35,
									},
									model.DefElt{Val: "value", Line: 36}: &model.DefElt{
										Val:  "String",
										Line: 36,
									},
								},
								Results: model.DefMap{
									model.DefElt{Val: "length", Line: 38}: &model.DefElt{
										Val:  "Uint32",
										Line: 38,
									},
								},
							},
						},
						Views: model.FuncDefMap{
							"TestView1": &model.FuncDef{
								Access: model.DefElt{
									Val:  "owner",
									Line: 43,
								},
								Params: model.DefMap{
									model.DefElt{Val: "name", Line: 45}: &model.DefElt{
										Val:  "String",
										Line: 45,
									},
									model.DefElt{Val: "id", Line: 46}: &model.DefElt{
										Val:  "Int32",
										Line: 46,
									},
								},
								Results: model.DefMap{
									model.DefElt{Val: "length", Line: 48}: &model.DefElt{
										Val:  "Uint32",
										Line: 48,
									},
								},
							},
							"TestView2": &model.FuncDef{
								Access: model.DefElt{
									Val:  "owner",
									Line: 50,
								},
								Params: model.DefMap{
									model.DefElt{Val: "name", Line: 52}: &model.DefElt{
										Val:  "String",
										Line: 52,
									},
									model.DefElt{Val: "id", Line: 53}: &model.DefElt{
										Val:  "Int32",
										Line: 53,
									},
								},
								Results: model.DefMap{
									model.DefElt{Val: "length", Line: 55}: &model.DefElt{
										Val:  "Uint32",
										Line: 55,
									},
								},
							},
						},
						Typedefs: model.DefMap{
							model.DefElt{Val: "TestTypedef", Line: 58}: &model.DefElt{
								Val:  "String[]",
								Line: 58,
							},
						},
						State: model.DefMap{
							model.DefElt{Val: "TestState", Line: 64}: &model.DefElt{
								Val:  "Int64[]",
								Line: 64,
							},
						},
					},
				},
			}
		},
	}

	for name, fn := range tests {
		t.Run(name, func(t *testing.T) {
			tt := fn(t)

			file, err := os.Open(tt.args.path)
			assert.NoError(t, err)
			in, err := ioutil.ReadAll(file)
			assert.NoError(t, err)

			def := &model.SchemaDef{}
			root := yaml.Parse(in)
			assert.NotNil(t, root)
			yaml.Unmarshal(root, def)

			assert.Equal(t, tt.wants.out.Funcs, def.Funcs)
		})
	}
}
