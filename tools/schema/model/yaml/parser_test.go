package yaml_test

import (
	"io/ioutil"
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
								Val:     "events",
								Line:    6,
								Comment: " header comment for event 1\n header comment for event 2\n line comment for event\n line comment for TestEvent1\n line comment for eventParam11\n line comment for TestEvent2\n line comment for eventParam21\n line comment for eventParam21\n",
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
										Val:  "TestStruct",
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

			got := yaml.Parse(in)
			assert.Equal(t, tt.wants.out, got)
		})
	}
}
