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
			name: "with array",
			str:  "0x2::dynamic_field::Field<0x1::ascii::String, 0x2::balance::Balance<0x2::iota::IOTA>>",
			want: &iotago.ResourceType{
				iotago.MustAddressFromHex("0x2"), "dynamic_field", "Field",
				&iotago.ResourceType{
					iotago.MustAddressFromHex("0x1"), "ascii", "String",
					nil,
					nil,
				},
				&iotago.ResourceType{
					iotago.MustAddressFromHex("0x2"), "balance", "Balance",
					&iotago.ResourceType{
						iotago.MustAddressFromHex("0x2"), "iota", "IOTA",
						nil,
						nil,
					},
					nil,
				},
			},
		},
		{
			name: "sample",
			str:  "0x23::coin::Xxxx",
			want: &iotago.ResourceType{iotago.MustAddressFromHex("0x23"), "coin", "Xxxx", nil, nil},
		},
		{
			name: "three level",
			str:  "0xabc::Coin::Xxxx<0x789::AAA::ppp<0x111::mod3::func3>>",
			want: &iotago.ResourceType{
				iotago.MustAddressFromHex("0xabc"), "Coin", "Xxxx",
				&iotago.ResourceType{
					iotago.MustAddressFromHex("0x789"), "AAA", "ppp",
					&iotago.ResourceType{iotago.MustAddressFromHex("0x111"), "mod3", "func3", nil, nil},
					nil,
				},
				nil,
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
			str:     "0x1::m1::f1<<0x3::m3::f3>0x2::m2::f2>",
			wantErr: true,
		},
		{
			name:    "error format2",
			str:     "0x1::m1::f1<<0x3::m3::f3>0x2::m2::f2>",
			wantErr: true,
		},
		{
			name:    "error format3",
			str:     "<0x3::m3::f3>0x1::m1::f1<0x2::m2::f2>",
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
			name:   "successful, two levels",
			str:    "0xe87e::swap::Pool<0x2f63::testcoin::TESTCOIN>",
			target: &iotago.ResourceType{Module: "swap", ObjectName: "Pool"},
			want:   true,
		},
		{
			name:   "successful, two levels, inner",
			str:    "0xe87e::swap::Pool<0x2f63::testcoin::TESTCOIN>",
			target: &iotago.ResourceType{Module: "testcoin", ObjectName: "TESTCOIN"},
			want:   true,
		},
		{
			name:   "successful, dynamic field 1",
			str:    "0x2::dynamic_field::Field<0x1::ascii::String, 0x2::balance::Balance<0x2::iota::IOTA>>",
			target: &iotago.ResourceType{Module: "ascii", ObjectName: "String"},
			want:   true,
		},
		{
			name:   "successful, dynamic field 2",
			str:    "0x2::dynamic_field::Field<0x1::ascii::String, 0x2::balance::Balance<0x2::iota::IOTA>>",
			target: &iotago.ResourceType{Module: "balance", ObjectName: "Balance"},
			want:   true,
		},
		{
			name:   "successful, dynamic field inner",
			str:    "0x2::dynamic_field::Field<0x1::ascii::String, 0x2::balance::Balance<0x2::iota::IOTA>>",
			target: &iotago.ResourceType{Module: "iota", ObjectName: "IOTA"},
			want:   true,
		},
		{
			name:   "failed, two levels",
			str:    "0xe87e::swap::Pool<0x2f63::testcoin::TESTCOIN>",
			target: &iotago.ResourceType{Module: "name", ObjectName: "Pool"},
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
			arg:  &iotago.ResourceType{iotago.MustAddressFromHex("0x1"), "m1", "f1", nil, nil},
			want: "0x1::m1::f1",
		},
		{
			arg: &iotago.ResourceType{
				iotago.MustAddressFromHex("0x1"), "m1", "f1",
				&iotago.ResourceType{
					iotago.MustAddressFromHex("2"), "m2", "f2",
					&iotago.ResourceType{iotago.MustAddressFromHex("0x123abcdef"), "m3", "f3", nil, nil},
					nil,
				},
				nil,
			},
			want: "0x1::m1::f1<0x2::m2::f2<0x123abcdef::m3::f3>>",
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
