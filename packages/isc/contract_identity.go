package isc

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type contractIdentityKind rwutil.Kind

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

func (c *ContractIdentity) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	c.kind = contractIdentityKind(rr.ReadKind())
	switch c.kind {
	case contractIdentityKindHname:
		rr.Read(&c.hname)
	case contractIdentityKindEthereum:
		rr.ReadN(c.evmAddr[:])
	}
	return rr.Err
}

func (c *ContractIdentity) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(c.kind))
	switch c.kind {
	case contractIdentityKindHname:
		ww.Write(&c.hname)
	case contractIdentityKindEthereum:
		ww.WriteN(c.evmAddr[:])
	}
	return ww.Err
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
