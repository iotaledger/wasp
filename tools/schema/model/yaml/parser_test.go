// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl,unparam
package yaml_test

import (
	"io"
	"os"
	"testing"

	yaml "github.com/iotaledger/wasp/tools/schema/model/yaml"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	type args struct {
		path string
	}
	type wants struct {
		out *yaml.Node
	}
	type test struct {
		args  args
		wants wants
	}

	tests := map[string]func(t *testing.T) test{
		"successfully test1": func(t *testing.T) test {
			return test{
				args: args{
					path: "testdata/test1.yaml",
				},
				wants: wants{
					&yaml.Node{
						Contents: []*yaml.Node{
							&yaml.Node{
								Val:  "name",
								Line: 1,
								Contents: []*yaml.Node{
									{
										Val:  "SchemaComment",
										Line: 1,
									},
								},
							},
							&yaml.Node{
								Val:  "description",
								Line: 2,
								Contents: []*yaml.Node{
									{
										Val:  "test description",
										Line: 2,
									},
								},
							},
							&yaml.Node{
								Val:         "events",
								Line:        6,
								HeadComment: "// header comment for event 1 \"this comment block will be ignored in yaml.Convert()\"\n// header comment for event 2\n",
								Contents: []*yaml.Node{
									{
										Val:         "TestEvent1",
										Line:        7,
										LineComment: "// line comment for TestEvent1\n",
										Contents: []*yaml.Node{
											{
												Val:         "eventParam11",
												Line:        8,
												LineComment: "// line comment for eventParam11\n",
												Contents: []*yaml.Node{
													{
														Val:  "String",
														Line: 8,
													},
												},
											},
										},
									},
									{
										Val:         "TestEvent2",
										Line:        9,
										LineComment: "// line comment for TestEvent2\n",
										Contents: []*yaml.Node{
											{
												Val:         "eventParam21",
												Line:        10,
												LineComment: "// line comment for eventParam21\n",
												Contents: []*yaml.Node{
													{
														Val:  "String",
														Line: 10,
													},
												},
											},
											{
												Val:         "eventParam22",
												Line:        11,
												LineComment: "// line comment for eventParam22\n",
												Contents: []*yaml.Node{
													{
														Val:  "String",
														Line: 11,
													},
												},
											},
										},
									},
								},
							},
							&yaml.Node{
								Val:  "structs",
								Line: 15,
								Contents: []*yaml.Node{
									{
										Val:  "TestStruct1",
										Line: 16,
										Contents: []*yaml.Node{
											{
												Val:  "x1",
												Line: 17,
												Contents: []*yaml.Node{
													{
														Val:  "Int32",
														Line: 17,
													},
												},
											},
											{
												Val:  "y1",
												Line: 18,
												Contents: []*yaml.Node{
													{
														Val:  "Int32",
														Line: 18,
													},
												},
											},
										},
									},
									{
										Val:         "TestStruct2",
										Line:        20,
										LineComment: "// comment for TestStruct2\n",
										Contents: []*yaml.Node{
											{
												Val:         "x2",
												Line:        21,
												LineComment: "// comment for x2\n",
												Contents: []*yaml.Node{
													{
														Val:  "Int32",
														Line: 21,
													},
												},
											},
											{
												Val:         "y2",
												Line:        25,
												HeadComment: "// comment for y2 1\n// comment for y2 2\n",
												Contents: []*yaml.Node{
													{
														Val:  "Int32",
														Line: 26,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
		},
		"successfully test2": func(t *testing.T) test {
			return test{
				args: args{
					path: "testdata/test2.yaml",
				},
				wants: wants{
					&yaml.Node{
						Contents: []*yaml.Node{
							&yaml.Node{
								Val:  "name",
								Line: 1,
								Contents: []*yaml.Node{
									{
										Val:  "SchemaComment",
										Line: 1,
									},
								},
							},
							&yaml.Node{
								Val:  "description",
								Line: 2,
								Contents: []*yaml.Node{
									{
										Val:  "test description",
										Line: 2,
									},
								},
							},
							&yaml.Node{
								Val:         "events",
								Line:        8,
								HeadComment: "// header comment for event 1 \"this comment block will be ignored in yaml.Convert()\"\n// header comment for event 2\n",
								Contents: []*yaml.Node{
									{
										Val:         "TestEvent1",
										Line:        15,
										HeadComment: "// header comment for TestEvent1 1\n// header comment for TestEvent1 2\n",
										Contents: []*yaml.Node{
											{
												Val:         "eventParam1",
												Line:        22,
												HeadComment: "// header comment for eventParam1 1\n// header comment for eventParam1 2\n",
												Contents: []*yaml.Node{
													{
														Val:  "String",
														Line: 22,
													},
												},
											},
										},
									},
									{
										Val:         "TestEvent2",
										Line:        34,
										LineComment: "// line comment for TestEvent2 1\n",
										Contents: []*yaml.Node{
											{
												Val:         "eventParam2",
												Line:        38,
												HeadComment: "// header comment for eventParam2 1\n// header comment for eventParam2 2\n",
												Contents: []*yaml.Node{
													{
														Val:         "String",
														Line:        41,
														HeadComment: "// line comment for eventParam2 2\n// line comment for eventParam2 3\n",
													},
												},
											},
										},
									},
								},
							},
							&yaml.Node{
								Val:  "structs",
								Line: 47,
								Contents: []*yaml.Node{
									{
										Val:         "TestStruct",
										Line:        49,
										HeadComment: "// comment for TestStruct 1\n",
										Contents: []*yaml.Node{
											{
												Val:         "x",
												Line:        54,
												HeadComment: "// comment for x 1\n// comment for x 2\n",
												Contents: []*yaml.Node{
													{
														Val:  "Int32",
														Line: 54,
													},
												},
											},
											{
												Val:  "y",
												Line: 57,
												Contents: []*yaml.Node{
													{
														Val:         "Int32",
														Line:        58,
														LineComment: "// comment for y 1\n",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
		},
		"successfully test3": func(t *testing.T) test {
			return test{
				args: args{
					path: "testdata/test3.yaml",
				},
				wants: wants{
					&yaml.Node{
						Contents: []*yaml.Node{
							&yaml.Node{
								Val:  "name",
								Line: 1,
								Contents: []*yaml.Node{
									{
										Val:  "SchemaComment",
										Line: 1,
									},
								},
							},
							&yaml.Node{
								Val:  "description",
								Line: 2,
								Contents: []*yaml.Node{
									{
										Val:  "test description",
										Line: 2,
									},
								},
							},
							&yaml.Node{
								Val:  "events",
								Line: 6,
								Contents: []*yaml.Node{
									{
										Val:  "TestEvent1",
										Line: 7,
										Contents: []*yaml.Node{
											{
												Val:  "eventParam11",
												Line: 8,
												Contents: []*yaml.Node{
													{
														Val:  "String",
														Line: 8,
													},
												},
											},
										},
									},
									{
										Val:  "TestEvent2",
										Line: 9,
										Contents: []*yaml.Node{
											{
												Val:  "eventParam21",
												Line: 10,
												Contents: []*yaml.Node{
													{
														Val:  "String",
														Line: 10,
													},
												},
											},
											{
												Val:  "eventParam22",
												Line: 11,
												Contents: []*yaml.Node{
													{
														Val:  "String",
														Line: 11,
													},
												},
											},
										},
									},
								},
							},
							&yaml.Node{
								Val:  "structs",
								Line: 15,
								Contents: []*yaml.Node{
									{
										Val:  "TestStruct1",
										Line: 16,
										Contents: []*yaml.Node{
											{
												Val:  "x",
												Line: 17,
												Contents: []*yaml.Node{
													{
														Val:  "Int32",
														Line: 17,
													},
												},
											},
											{
												Val:  "y",
												Line: 18,
												Contents: []*yaml.Node{
													{
														Val:  "Int32",
														Line: 18,
													},
												},
											},
										},
									},
									{
										Val:  "TestStruct2",
										Line: 20,
										Contents: []*yaml.Node{
											{
												Val:  "x",
												Line: 21,
												Contents: []*yaml.Node{
													{
														Val:  "Int32",
														Line: 22,
													},
												},
											},
										},
									},
								},
							},
							&yaml.Node{
								Val:  "funcs",
								Line: 24,
								Contents: []*yaml.Node{
									{
										Val:  "TestFunc1",
										Line: 25,
										Contents: []*yaml.Node{
											{
												Val:  "access",
												Line: 26,
												Contents: []*yaml.Node{
													{
														Val:  "owner",
														Line: 26,
													},
												},
											},
											{
												Val:  "params",
												Line: 27,
												Contents: []*yaml.Node{
													{
														Val:  "name",
														Line: 28,
														Contents: []*yaml.Node{
															{
																Val:  "String",
																Line: 28,
															},
														},
													},
													{
														Val:  "value",
														Line: 29,
														Contents: []*yaml.Node{
															{
																Val:  "String",
																Line: 29,
															},
														},
													},
												},
											},
											{
												Val:  "results",
												Line: 30,
												Contents: []*yaml.Node{
													{
														Val:  "length",
														Line: 31,
														Contents: []*yaml.Node{
															{
																Val:  "Uint32",
																Line: 31,
															},
														},
													},
												},
											},
										},
									},
									{
										Val:  "TestFunc2",
										Line: 32,
										Contents: []*yaml.Node{
											{
												Val:  "access",
												Line: 33,
												Contents: []*yaml.Node{
													{
														Val:  "owner",
														Line: 33,
													},
												},
											},
											{
												Val:  "params",
												Line: 34,
												Contents: []*yaml.Node{
													{
														Val:  "name",
														Line: 35,
														Contents: []*yaml.Node{
															{
																Val:  "String",
																Line: 35,
															},
														},
													},
													{
														Val:  "value",
														Line: 36,
														Contents: []*yaml.Node{
															{
																Val:  "String",
																Line: 36,
															},
														},
													},
												},
											},
											{
												Val:  "results",
												Line: 37,
												Contents: []*yaml.Node{
													{
														Val:  "length",
														Line: 38,
														Contents: []*yaml.Node{
															{
																Val:  "Uint32",
																Line: 38,
															},
														},
													},
												},
											},
										},
									},
								},
							},
							&yaml.Node{
								Val:  "views",
								Line: 41,
								Contents: []*yaml.Node{
									{
										Val:  "TestView1",
										Line: 42,
										Contents: []*yaml.Node{
											{
												Val:  "access",
												Line: 43,
												Contents: []*yaml.Node{
													{
														Val:  "owner",
														Line: 43,
													},
												},
											},
											{
												Val:  "params",
												Line: 44,
												Contents: []*yaml.Node{
													{
														Val:  "name",
														Line: 45,
														Contents: []*yaml.Node{
															{
																Val:  "String",
																Line: 45,
															},
														},
													},
													{
														Val:  "id",
														Line: 46,
														Contents: []*yaml.Node{
															{
																Val:  "Int32",
																Line: 46,
															},
														},
													},
												},
											},
											{
												Val:  "results",
												Line: 47,
												Contents: []*yaml.Node{
													{
														Val:  "length",
														Line: 48,
														Contents: []*yaml.Node{
															{
																Val:  "Uint32",
																Line: 48,
															},
														},
													},
												},
											},
										},
									},
									{
										Val:  "TestView2",
										Line: 49,
										Contents: []*yaml.Node{
											{
												Val:  "access",
												Line: 50,
												Contents: []*yaml.Node{
													{
														Val:  "owner",
														Line: 50,
													},
												},
											},
											{
												Val:  "params",
												Line: 51,
												Contents: []*yaml.Node{
													{
														Val:  "name",
														Line: 52,
														Contents: []*yaml.Node{
															{
																Val:  "String",
																Line: 52,
															},
														},
													},
													{
														Val:  "id",
														Line: 53,
														Contents: []*yaml.Node{
															{
																Val:  "Int32",
																Line: 53,
															},
														},
													},
												},
											},
											{
												Val:  "results",
												Line: 54,
												Contents: []*yaml.Node{
													{
														Val:  "length",
														Line: 55,
														Contents: []*yaml.Node{
															{
																Val:  "Uint32",
																Line: 55,
															},
														},
													},
												},
											},
										},
									},
								},
							},
							&yaml.Node{
								Val:  "typedefs",
								Line: 57,
								Contents: []*yaml.Node{
									{
										Val:  "TestTypedef",
										Line: 58,
										Contents: []*yaml.Node{
											{
												Val:  "String[]",
												Line: 58,
											},
										},
									},
								},
							},
							&yaml.Node{
								Val:  "state",
								Line: 62,
								Contents: []*yaml.Node{
									{
										Val:  "TestState",
										Line: 64,
										Contents: []*yaml.Node{
											{
												Val:  "Int64[]",
												Line: 64,
											},
										},
									},
								},
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
			in, err := io.ReadAll(file)
			assert.NoError(t, err)

			got := yaml.Parse(in)
			assert.Equal(t, tt.wants.out, got)
		})
	}
}
