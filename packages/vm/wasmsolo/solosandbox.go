package wasmsolo

import (
	"bytes"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/sandbox/sandbox_utils"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

// NOTE: These functions correspond to the Sandbox fnXxx constants in WasmLib
var sandboxFunctions = []func(*SoloSandbox, []byte) []byte{
	nil,
	(*SoloSandbox).fnAccountID,
	(*SoloSandbox).fnBalance,
	(*SoloSandbox).fnBalances,
	(*SoloSandbox).fnBlockContext,
	(*SoloSandbox).fnCall,
	(*SoloSandbox).fnCaller,
	(*SoloSandbox).fnChainID,
	(*SoloSandbox).fnChainOwnerID,
	(*SoloSandbox).fnContract,
	(*SoloSandbox).fnContractCreator,
	(*SoloSandbox).fnDeployContract,
	(*SoloSandbox).fnEntropy,
	(*SoloSandbox).fnEvent,
	(*SoloSandbox).fnIncomingTransfer,
	(*SoloSandbox).fnLog,
	(*SoloSandbox).fnMinted,
	(*SoloSandbox).fnPanic,
	(*SoloSandbox).fnParams,
	(*SoloSandbox).fnPost,
	(*SoloSandbox).fnRequest,
	(*SoloSandbox).fnRequestID,
	(*SoloSandbox).fnResults,
	(*SoloSandbox).fnSend,
	(*SoloSandbox).fnStateAnchor,
	(*SoloSandbox).fnTimestamp,
	(*SoloSandbox).fnTrace,
	(*SoloSandbox).fnUtilsBase58Decode,
	(*SoloSandbox).fnUtilsBase58Encode,
	(*SoloSandbox).fnUtilsBlsAddress,
	(*SoloSandbox).fnUtilsBlsAggregate,
	(*SoloSandbox).fnUtilsBlsValid,
	(*SoloSandbox).fnUtilsEd25519Address,
	(*SoloSandbox).fnUtilsEd25519Valid,
	(*SoloSandbox).fnUtilsHashBlake2b,
	(*SoloSandbox).fnUtilsHashName,
	(*SoloSandbox).fnUtilsHashSha3,
}

type SoloSandbox struct {
	ctx   *SoloContext
	utils iscp.Utils
}

var _ wasmhost.ISandbox = new(SoloSandbox)

func NewSoloSandbox(ctx *SoloContext) *SoloSandbox {
	return &SoloSandbox{ctx: ctx, utils: sandbox_utils.NewUtils()}
}

func (s *SoloSandbox) Call(funcNr int32, params []byte) []byte {
	return sandboxFunctions[-funcNr](s, params)
}

func (s *SoloSandbox) checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func (s *SoloSandbox) Panicf(format string, args ...interface{}) {
	s.ctx.Chain.Log.Panicf(format, args...)
}

func (s *SoloSandbox) Tracef(format string, args ...interface{}) {
	s.ctx.Chain.Log.Debugf(format, args...)
}

func (s *SoloSandbox) postSync(contract, function string, params dict.Dict, transfer colored.Balances) []byte {
	req := solo.NewCallParamsFromDic(contract, function, params)
	req.WithTransfers(transfer)
	ctx := s.ctx
	_ = wasmhost.Connect(ctx.wasmHostOld)
	var res dict.Dict
	if ctx.offLedger {
		ctx.offLedger = false
		res, ctx.Err = ctx.Chain.PostRequestOffLedger(req, ctx.keyPair)
	} else if !ctx.isRequest {
		ctx.Tx, res, ctx.Err = ctx.Chain.PostRequestSyncTx(req, ctx.keyPair)
	} else {
		ctx.isRequest = false
		ctx.Tx, _, ctx.Err = ctx.Chain.RequestFromParamsToLedger(req, nil)
		if ctx.Err == nil {
			ctx.Chain.Env.EnqueueRequests(ctx.Tx)
		}
	}
	_ = wasmhost.Connect(ctx.wc)
	if ctx.Err != nil {
		return nil
	}
	return res.Bytes()
}

//////////////////// sandbox functions \\\\\\\\\\\\\\\\\\\\

func (s *SoloSandbox) fnAccountID(args []byte) []byte {
	return s.ctx.AccountID().Bytes()
}

func (s *SoloSandbox) fnBalance(args []byte) []byte {
	color := wasmtypes.ColorFromBytes(args)
	return codec.EncodeUint64(s.ctx.Balance(s.ctx.Account(), color))
}

func (s *SoloSandbox) fnBalances(args []byte) []byte {
	agent := s.ctx.Account()
	account := iscp.NewAgentID(agent.address, agent.hname)
	balances := s.ctx.Chain.GetAccountBalance(account)
	return balances.Bytes()
}

func (s *SoloSandbox) fnBlockContext(args []byte) []byte {
	panic("implement me")
}

func (s *SoloSandbox) fnCall(args []byte) []byte {
	req := wasmrequests.NewCallRequestFromBytes(args)
	contract, err := iscp.HnameFromBytes(req.Contract.Bytes())
	s.checkErr(err)
	if contract != iscp.Hn(s.ctx.scName) {
		s.Panicf("unknown contract: %s", contract.String())
	}
	function, err := iscp.HnameFromBytes(req.Function.Bytes())
	s.checkErr(err)
	funcName := s.ctx.wc.Host().FunctionFromCode(uint32(function))
	if funcName == "" {
		s.Panicf("unknown function: %s", function.String())
	}
	s.Tracef("CALL c'%s' f'%s'", s.ctx.scName, funcName)
	params, err := dict.FromBytes(req.Params)
	s.checkErr(err)
	transfer, err := colored.BalancesFromBytes(req.Transfer)
	s.checkErr(err)

	if len(transfer) != 0 {
		return s.postSync(s.ctx.scName, funcName, params, transfer)
	}

	res, err := s.ctx.Chain.CallView(s.ctx.scName, funcName, params)
	s.ctx.Err = err
	if err != nil {
		return nil
	}
	return res.Bytes()
}

func (s *SoloSandbox) fnCaller(args []byte) []byte {
	return s.ctx.Chain.OriginatorAgentID.Bytes()
}

func (s *SoloSandbox) fnChainID(args []byte) []byte {
	return s.ctx.ChainID().Bytes()
}

func (s *SoloSandbox) fnChainOwnerID(args []byte) []byte {
	return s.ctx.ChainOwnerID().Bytes()
}

func (s *SoloSandbox) fnContract(args []byte) []byte {
	return s.ctx.Account().hname.Bytes()
}

func (s *SoloSandbox) fnContractCreator(args []byte) []byte {
	return s.ctx.ContractCreator().Bytes()
}

func (s *SoloSandbox) fnDeployContract(args []byte) []byte {
	panic("implement me")
}

func (s *SoloSandbox) fnEntropy(args []byte) []byte {
	return s.ctx.Chain.ChainID.Bytes()[1:]
}

func (s *SoloSandbox) fnEvent(args []byte) []byte {
	s.Panicf("solo cannot send events")
	return nil
}

func (s *SoloSandbox) fnIncomingTransfer(args []byte) []byte {
	// zero incoming balance
	return colored.NewBalances().Bytes()
}

func (s *SoloSandbox) fnLog(args []byte) []byte {
	s.ctx.Chain.Log.Infof(string(args))
	return nil
}

func (s *SoloSandbox) fnMinted(args []byte) []byte {
	panic("implement me")
}

func (s *SoloSandbox) fnPanic(args []byte) []byte {
	s.ctx.Chain.Log.Panicf(string(args))
	return nil
}

func (s *SoloSandbox) fnParams(args []byte) []byte {
	return make(dict.Dict).Bytes()
}

func (s *SoloSandbox) fnPost(args []byte) []byte {
	req := wasmrequests.NewPostRequestFromBytes(args)
	if !bytes.Equal(req.ChainID.Bytes(), s.fnChainID(nil)) {
		s.Panicf("unknown chain id: %s", req.ChainID.String())
	}
	contract, err := iscp.HnameFromBytes(req.Contract.Bytes())
	s.checkErr(err)
	if contract != iscp.Hn(s.ctx.scName) {
		s.Panicf("unknown contract: %s", contract.String())
	}
	function, err := iscp.HnameFromBytes(req.Function.Bytes())
	s.checkErr(err)
	funcName := s.ctx.wc.Host().FunctionFromCode(uint32(function))
	if funcName == "" {
		s.Panicf("unknown function: %s", function.String())
	}
	s.Tracef("POST c'%s' f'%s'", s.ctx.scName, funcName)
	params, err := dict.FromBytes(req.Params)
	s.checkErr(err)
	transfer, err := colored.BalancesFromBytes(req.Transfer)
	s.checkErr(err)
	if len(transfer) == 0 && !s.ctx.offLedger {
		s.Panicf("transfer is required for post")
	}
	if req.Delay != 0 {
		s.Panicf("cannot delay solo post")
	}
	return s.postSync(s.ctx.scName, funcName, params, transfer)
}

func (s *SoloSandbox) fnRequest(args []byte) []byte {
	panic("implement me")
}

func (s *SoloSandbox) fnRequestID(args []byte) []byte {
	return append(s.ctx.Chain.ChainID.Bytes()[1:], 0, 0)
}

func (s *SoloSandbox) fnResults(args []byte) []byte {
	panic("implement me")
}

// transfer tokens to address
func (s *SoloSandbox) fnSend(args []byte) []byte {
	panic("implement me")
}

func (s *SoloSandbox) fnStateAnchor(args []byte) []byte {
	panic("implement me")
}

func (s *SoloSandbox) fnTimestamp(args []byte) []byte {
	return codec.EncodeInt64(time.Now().UnixNano())
}

func (s *SoloSandbox) fnTrace(args []byte) []byte {
	s.ctx.Chain.Log.Debugf(string(args))
	return nil
}
