package iotago

import (
	"errors"
	"fmt"
	"strings"
)

type ResourceType struct {
	Address    *Address
	Module     Identifier
	ObjectName Identifier // it can be function name or struct name, etc.

	SubType1 *ResourceType `bcs:"optional"`
	SubType2 *ResourceType `bcs:"optional"`
}

func IsSameResource(a, b string) (bool, error) {
	ra, err := NewResourceType(a)
	if err != nil {
		return false, fmt.Errorf("failed to parse resource type: %w", err)
	}
	rb, err := NewResourceType(b)
	if err != nil {
		return false, fmt.Errorf("failed to parse resource type: %w", err)
	}
	return ra.String() == rb.String(), nil
}

func MustNewResourceType(s string) *ResourceType {
	resource, err := NewResourceType(s)
	if err != nil {
		panic(err)
	}
	return resource
}

func NewResourceType(str string) (*ResourceType, error) {
	var err error

	ltIdx := strings.Index(str, "<")
	var subType1, subType2 *ResourceType
	if ltIdx != -1 {
		gtIdx := strings.LastIndex(str, ">")
		if gtIdx != len(str)-1 {
			return nil, errors.New("invalid type string literal")
		}
		commaIdx := strings.Index(str, ",")
		if commaIdx == -1 {
			subType1, err = NewResourceType(str[ltIdx+1 : gtIdx])
			if err != nil {
				return nil, err
			}
		} else {
			subType1, err = NewResourceType(str[ltIdx+1 : commaIdx])
			if err != nil {
				return nil, err
			}
			subType2, err = NewResourceType(strings.TrimSpace(str[commaIdx+1 : gtIdx]))
			if err != nil {
				return nil, err
			}
		}
	}

	parts := strings.Split(str, "::")
	addr, err := AddressFromHex(parts[0])
	if err != nil {
		return nil, err
	}
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid resource type string: %q", str)
	}
	module := parts[1]
	var objectName string
	if strings.Contains(parts[2], "<") {
		objectName = parts[2][:strings.Index(parts[2], "<")]
	} else {
		objectName = parts[2]
	}

	return &ResourceType{
		Address:    addr,
		Module:     module,
		ObjectName: objectName,
		SubType1:   subType1,
		SubType2:   subType2,
	}, nil
}

func (t *ResourceType) UnmarshalJSON(data []byte) error {
	resource, err := NewResourceType(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*t = *resource
	return nil
}

func (t *ResourceType) Contains(address *Address, moduleName string, funcName string) bool {
	if t == nil {
		return false
	}
	if t.Module == moduleName && t.ObjectName == funcName {
		if address == nil {
			return true
		}
		if t.Address.String() == address.String() {
			return true
		}
	}
	if t.SubType1 == nil {
		return false
	}
	return t.SubType1.Contains(address, moduleName, funcName) || t.SubType2.Contains(address, moduleName, funcName)
}

func (t *ResourceType) String() string {
	if t.SubType2 != nil {
		return fmt.Sprintf(
			"%v::%v::%v<%v, %v>",
			t.Address.String(),
			t.Module,
			t.ObjectName,
			t.SubType1.String(),
			t.SubType1.String(),
		)
	} else if t.SubType1 != nil {
		return fmt.Sprintf("%v::%v::%v<%v>", t.Address.String(), t.Module, t.ObjectName, t.SubType1.String())
	} else {
		return fmt.Sprintf("%v::%v::%v", t.Address.String(), t.Module, t.ObjectName)
	}
}

func (t *ResourceType) ShortString() string {
	if t.SubType2 != nil {
		return fmt.Sprintf(
			"%v::%v::%v<%v, %v>",
			t.Address.ShortString(),
			t.Module,
			t.ObjectName,
			t.SubType1.ShortString(),
			t.SubType1.ShortString(),
		)
	} else if t.SubType1 != nil {
		return fmt.Sprintf("%v::%v::%v<%v>", t.Address.ShortString(), t.Module, t.ObjectName, t.SubType1.ShortString())
	} else {
		return fmt.Sprintf("%v::%v::%v", t.Address.ShortString(), t.Module, t.ObjectName)
	}
}
