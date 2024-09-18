package isc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

type contractIdentityKind byte

type ContractIdentity struct {
	// can either be an Hname or a solidity contract
	kind contractIdentityKind

	// only 1 or the other will be filled
	evmAddr common.Address
	hname   Hname
}

const (
	contractIdentityKindEmpty contractIdentityKind = iota
	contractIdentityKindHname
	contractIdentityKindEthereum
)

func EmptyContractIdentity() ContractIdentity {
	return ContractIdentity{kind: contractIdentityKindEmpty}
}

func ContractIdentityFromHname(hn Hname) ContractIdentity {
	return ContractIdentity{hname: hn, kind: contractIdentityKindHname}
}

func ContractIdentityFromEVMAddress(addr common.Address) ContractIdentity {
	return ContractIdentity{evmAddr: addr, kind: contractIdentityKindEthereum}
}

func (c *ContractIdentity) String() string {
	switch c.kind {
	case contractIdentityKindHname:
		return c.hname.String()
	case contractIdentityKindEthereum:
		return c.evmAddr.String()
	}
	return ""
}

func (c *ContractIdentity) MarshalBCS(e *bcs.Encoder) error {
	e.WriteEnumIdx(int(c.kind))

	switch c.kind {
	case contractIdentityKindHname:
		return e.WriteUint32(uint32(c.hname))
	case contractIdentityKindEthereum:
		return e.Encode(c.evmAddr)
	}

	return nil
}

func (c *ContractIdentity) UnmarshalBCS(d *bcs.Decoder) error {
	c.kind = contractIdentityKind(d.ReadEnumIdx())

	switch c.kind {
	case contractIdentityKindHname:
		c.hname = Hname(d.ReadUint32())
	case contractIdentityKindEthereum:
		return d.Decode(&c.evmAddr)
	}

	return d.Err()
}

func (c *ContractIdentity) AgentID(chainID ChainID) AgentID {
	switch c.kind {
	case contractIdentityKindHname:
		return NewContractAgentID(chainID, c.hname)
	case contractIdentityKindEthereum:
		return NewEthereumAddressAgentID(chainID, c.evmAddr)
	}
	return &NilAgentID{}
}

func (c *ContractIdentity) Hname() (Hname, error) {
	if c.kind == contractIdentityKindHname {
		return c.hname, nil
	}
	return 0, fmt.Errorf("not an Hname contract")
}

func (c *ContractIdentity) EVMAddress() (common.Address, error) {
	if c.kind == contractIdentityKindEthereum {
		return c.evmAddr, nil
	}
	return common.Address{}, fmt.Errorf("not an EVM contract")
}

func (c *ContractIdentity) Empty() bool {
	return c.kind == contractIdentityKindEmpty
}
