package consensus

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/plugins/runvm"
	"time"
)

type runCalculationsParams struct {
	requests        []*request
	leaderPeerIndex uint16
	balances        map[valuetransaction.ID][]*balance.Balance
	rewardAddress   address.Address
	timestamp       time.Time
}

// runs the VM for the request and posts result to committee's queue
func (op *operator) runCalculationsAsync(par runCalculationsParams) {
	reqRefs, _ := takeRefs(par.requests)
	ctx := &vm.VMTask{
		LeaderPeerIndex: par.leaderPeerIndex,
		ProgramHash:     op.committee.MetaData().ProgramHash,
		Address:         op.committee.Address(),
		Color:           op.committee.Color(),
		Balances:        par.balances,
		RewardAddress:   address.Address{},
		Requests:        reqRefs,
		Timestamp:       time.Time{},
		VariableState:   op.variableState,
		Log:             op.log,
	}
	ctx.OnFinish = func() {
		op.committee.ReceiveMessage(ctx)
	}
	if err := runvm.RunComputationsAsync(ctx); err != nil {
		op.log.Errorf("RunComputationsAsync: %v", err)
	}
}

func (op *operator) sendResultToTheLeader(result *vm.VMTask) {
	op.log.Debugw("sendResultToTheLeader")

	sigShare, err := op.dkshare.SignShare(result.ResultTransaction.EssenceBytes())
	if err != nil {
		op.log.Errorf("error while signing transaction %v", err)
		return
	}

	reqids := make([]sctransaction.RequestId, len(result.Requests))
	for i := range reqids {
		reqids[i] = *result.Requests[i].RequestId()
	}

	msgData := hashing.MustBytes(&committee.SignedHashMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: op.stateIndex(),
		},
		BatchHash:     vm.BatchHash(reqids, result.Timestamp),
		OrigTimestamp: result.Timestamp,
		EssenceHash:   *hashing.HashData(result.ResultTransaction.EssenceBytes()),
		SigShare:      sigShare,
	})

	if err := op.committee.SendMsg(result.LeaderPeerIndex, committee.MsgSignedHash, msgData); err != nil {
		op.log.Error(err)
	}
}

func (op *operator) saveOwnResult(result *vm.VMTask) {
	sigShare, err := op.dkshare.SignShare(result.ResultTransaction.EssenceBytes())
	if err != nil {
		op.log.Errorf("error while signing transaction %v", err)
		return
	}

	reqids := make([]sctransaction.RequestId, len(result.Requests))
	for i := range reqids {
		reqids[i] = *result.Requests[i].RequestId()
	}

	op.leaderStatus.resultTx = result.ResultTransaction
	op.leaderStatus.batch = result.ResultBatch
	op.leaderStatus.batchHash = vm.BatchHash(reqids, result.Timestamp)
	op.leaderStatus.signedResults[op.committee.OwnPeerIndex()] = &signedResult{
		essenceHash: *hashing.HashData(result.ResultTransaction.EssenceBytes()),
		sigShare:    sigShare,
	}
}

func (op *operator) aggregateSigShares(sigShares [][]byte) error {
	resTx := op.leaderStatus.resultTx

	finalSignature, err := op.dkshare.RecoverFullSignature(sigShares, resTx.EssenceBytes())
	if err != nil {
		return err
	}
	finalSignature = finalSignature
	// if err := resTx.PutSignature(finalSignature); err != nil{
	// 		return fmt.Errorf("something wrong while aggregating final signature: %v", err)
	// }
	return nil
}
