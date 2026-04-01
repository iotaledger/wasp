package iotago_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

func TestNewResourceType(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    *iotago.ResourceType
		wantErr bool
	}{
		{
			name: "no subtype",
			str:  "0x23::coin::Xxxx",
			want: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x23"),
				Module:     "coin",
				ObjectName: "Xxxx",
				SubTypes:   nil,
			},
		},
		{
			name: "one subtype",
			str:  "0x1::m1::f1<0x2::m2::f2>",
			want: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x1"),
				Module:     "m1",
				ObjectName: "f1",
				SubTypes: []*iotago.ResourceType{
					{
						Address:    iotago.MustAddressFromHex("0x2"),
						Module:     "m2",
						ObjectName: "f2",
						SubTypes:   nil,
					},
				},
			},
		},
		{
			name: "multiple subtypes",
			str:  "0x111::aaa::AAA<0x222::bbb::BBB, 0x333::ccc::CCC, 0x444::ddd::DDD>",
			want: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x111"),
				Module:     "aaa",
				ObjectName: "AAA",
				SubTypes: []*iotago.ResourceType{
					{
						Address:    iotago.MustAddressFromHex("0x222"),
						Module:     "bbb",
						ObjectName: "BBB",
						SubTypes:   nil,
					},
					{
						Address:    iotago.MustAddressFromHex("0x333"),
						Module:     "ccc",
						ObjectName: "CCC",
						SubTypes:   nil,
					},
					{
						Address:    iotago.MustAddressFromHex("0x444"),
						Module:     "ddd",
						ObjectName: "DDD",
						SubTypes:   nil,
					},
				},
			},
		},
		{
			name: "nested generics two levels",
			str:  "0x2::dynamic_field::Field<0x1::ascii::String, 0x2::balance::Balance<0x2::iota::IOTA>>",
			want: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x2"),
				Module:     "dynamic_field",
				ObjectName: "Field",
				SubTypes: []*iotago.ResourceType{
					{
						Address:    iotago.MustAddressFromHex("0x1"),
						Module:     "ascii",
						ObjectName: "String",
						SubTypes:   nil,
					},
					{
						Address:    iotago.MustAddressFromHex("0x2"),
						Module:     "balance",
						ObjectName: "Balance",
						SubTypes: []*iotago.ResourceType{
							{
								Address:    iotago.MustAddressFromHex("0x2"),
								Module:     "iota",
								ObjectName: "IOTA",
								SubTypes:   nil,
							},
						},
					},
				},
			},
		},
		{
			name: "deeply nested generics",
			str:  "0x1::outer::O<0x2::mid::M<0x3::inner::I, 0x4::k::K>, 0x5::leaf::L>",
			want: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x1"),
				Module:     "outer",
				ObjectName: "O",
				SubTypes: []*iotago.ResourceType{
					{
						Address:    iotago.MustAddressFromHex("0x2"),
						Module:     "mid",
						ObjectName: "M",
						SubTypes: []*iotago.ResourceType{
							{
								Address:    iotago.MustAddressFromHex("0x3"),
								Module:     "inner",
								ObjectName: "I",
								SubTypes:   nil,
							},
							{
								Address:    iotago.MustAddressFromHex("0x4"),
								Module:     "k",
								ObjectName: "K",
								SubTypes:   nil,
							},
						},
					},
					{
						Address:    iotago.MustAddressFromHex("0x5"),
						Module:     "leaf",
						ObjectName: "L",
						SubTypes:   nil,
					},
				},
			},
		},
		{
			name: "whitespace tolerance",
			str:  "0x1::m1::f1< 0x2::m2::f2 , 0x3::m3::f3 >",
			want: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x1"),
				Module:     "m1",
				ObjectName: "f1",
				SubTypes: []*iotago.ResourceType{
					{
						Address:    iotago.MustAddressFromHex("0x2"),
						Module:     "m2",
						ObjectName: "f2",
						SubTypes:   nil,
					},
					{
						Address:    iotago.MustAddressFromHex("0x3"),
						Module:     "m3",
						ObjectName: "f3",
						SubTypes:   nil,
					},
				},
			},
		},
		{
			name:    "error address",
			str:     "0x123abcg::coin::Xxxx",
			wantErr: true,
		},
		{
			name:    "error format",
			str:     "0x1::m1::f1<0x2::m2::f2>x",
			wantErr: true,
		},
		{
			name:    "error format2",
			str:     "0x1::m1::f1<0x2::m2::f2<0x3::m3::f3>",
			wantErr: true,
		},
		{
			name:    "error format3",
			str:     "<0x3::m3::f3>0x1::m1::f1<0x2::m2::f2>",
			wantErr: true,
		},
		{
			name:    "error format - empty subtype",
			str:     "0x1::m1::f1<0x2::m2::f2,>",
			wantErr: true,
		},
		{
			name:    "error format - empty generic",
			str:     "0x1::m1::f1<>",
			wantErr: true,
		},
		{
			name:    "error format - wrong parts count",
			str:     "0x1::m1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := iotago.NewResourceType(tt.str)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewResourceType() error: %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewResourceType(): %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		target *iotago.ResourceType
		want   bool
	}{
		{
			name:   "successful, matches outer type",
			str:    "0xe87e::swap::Pool<0x2f63::testcoin::TESTCOIN>",
			target: &iotago.ResourceType{Module: "swap", ObjectName: "Pool"},
			want:   true,
		},
		{
			name:   "successful, matches single subtype",
			str:    "0xe87e::swap::Pool<0x2f63::testcoin::TESTCOIN>",
			target: &iotago.ResourceType{Module: "testcoin", ObjectName: "TESTCOIN"},
			want:   true,
		},
		{
			name:   "successful, matches first subtype in multiple subtypes",
			str:    "0x2::dynamic_field::Field<0x1::ascii::String, 0x2::balance::Balance<0x2::iota::IOTA>>",
			target: &iotago.ResourceType{Module: "ascii", ObjectName: "String"},
			want:   true,
		},
		{
			name:   "successful, matches second subtype in multiple subtypes",
			str:    "0x2::dynamic_field::Field<0x1::ascii::String, 0x2::balance::Balance<0x2::iota::IOTA>>",
			target: &iotago.ResourceType{Module: "balance", ObjectName: "Balance"},
			want:   true,
		},
		{
			name:   "successful, matches deeply nested subtype",
			str:    "0x2::dynamic_field::Field<0x1::ascii::String, 0x2::balance::Balance<0x2::iota::IOTA>>",
			target: &iotago.ResourceType{Module: "iota", ObjectName: "IOTA"},
			want:   true,
		},
		{
			name:   "successful, matches in complex nested structure",
			str:    "0x111::aaa::AAA<0x222::bbb::BBB, 0x333::ccc::CCC, 0x444::ddd::DDD>",
			target: &iotago.ResourceType{Module: "ccc", ObjectName: "CCC"},
			want:   true,
		},
		{
			name:   "successful, matches deeply nested type",
			str:    "0x1::outer::O<0x2::mid::M<0x3::inner::I, 0x4::k::K>, 0x5::leaf::L>",
			target: &iotago.ResourceType{Module: "inner", ObjectName: "I"},
			want:   true,
		},
		{
			name:   "successful, with address match",
			str:    "0xe87e::swap::Pool<0x2f63::testcoin::TESTCOIN>",
			target: &iotago.ResourceType{Address: iotago.MustAddressFromHex("0xe87e"), Module: "swap", ObjectName: "Pool"},
			want:   true,
		},
		{
			name:   "failed, wrong module name",
			str:    "0xe87e::swap::Pool<0x2f63::testcoin::TESTCOIN>",
			target: &iotago.ResourceType{Module: "name", ObjectName: "Pool"},
			want:   false,
		},
		{
			name:   "failed, wrong address",
			str:    "0xe87e::swap::Pool<0x2f63::testcoin::TESTCOIN>",
			target: &iotago.ResourceType{Address: iotago.MustAddressFromHex("0x1"), Module: "swap", ObjectName: "Pool"},
			want:   false,
		},
		{
			name:   "failed, not found anywhere",
			str:    "0x111::aaa::AAA<0x222::bbb::BBB, 0x333::ccc::CCC>",
			target: &iotago.ResourceType{Module: "nonexistent", ObjectName: "NotFound"},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				src, err := iotago.NewResourceType(tt.str)
				require.NoError(t, err)
				require.Equal(t, tt.want, src.Contains(tt.target.Address, tt.target.Module, tt.target.ObjectName))
			},
		)
	}
}

func TestResourceTypeString(t *testing.T) {
	typeString := "0x1::mmm1::fff1<0x123abcdef::mm2::ff3>"

	resourceType, err := iotago.NewResourceType(typeString)
	require.NoError(t, err)
	res := "0x0000000000000000000000000000000000000000000000000000000000000001::mmm1::fff1<0x0000000000000000000000000000000000000000000000000000000123abcdef::mm2::ff3>"
	require.Equal(t, resourceType.String(), res)
}

func TestResourceTypeShortString(t *testing.T) {
	tests := []struct {
		name string
		arg  *iotago.ResourceType
		want string
	}{
		{
			name: "no subtypes",
			arg: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x1"),
				Module:     "m1",
				ObjectName: "f1",
				SubTypes:   nil,
			},
			want: "0x1::m1::f1",
		},
		{
			name: "nested generics",
			arg: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x1"),
				Module:     "m1",
				ObjectName: "f1",
				SubTypes: []*iotago.ResourceType{
					{
						Address:    iotago.MustAddressFromHex("0x2"),
						Module:     "m2",
						ObjectName: "f2",
						SubTypes: []*iotago.ResourceType{
							{
								Address:    iotago.MustAddressFromHex("0x123abcdef"),
								Module:     "m3",
								ObjectName: "f3",
								SubTypes:   nil,
							},
						},
					},
				},
			},
			want: "0x1::m1::f1<0x2::m2::f2<0x123abcdef::m3::f3>>",
		},
		{
			name: "multiple subtypes",
			arg: &iotago.ResourceType{
				Address:    iotago.MustAddressFromHex("0x1"),
				Module:     "outer",
				ObjectName: "Type",
				SubTypes: []*iotago.ResourceType{
					{
						Address:    iotago.MustAddressFromHex("0x2"),
						Module:     "mod1",
						ObjectName: "Type1",
						SubTypes:   nil,
					},
					{
						Address:    iotago.MustAddressFromHex("0x3"),
						Module:     "mod2",
						ObjectName: "Type2",
						SubTypes:   nil,
					},
				},
			},
			want: "0x1::outer::Type<0x2::mod1::Type1, 0x3::mod2::Type2>",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := tt.arg.ShortString(); got != tt.want {
					t.Errorf("ResourceType.ShortString(): %v, want %v", got, tt.want)
				}
			},
		)
	}
}
