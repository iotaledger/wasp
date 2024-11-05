// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testchain

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

////////////////////////////////////////////////////////////////////////////////
// TestChainLedger

type TestChainLedger struct {
	t           *testing.T
	l1client    clients.L1Client
	iscPackage  *iotago.PackageID
	governor    *cryptolib.KeyPair
	chainID     isc.ChainID
	fetchedReqs map[cryptolib.AddressKey]map[iotago.ObjectID]bool
}

func NewTestChainLedger(
	t *testing.T,
	originator *cryptolib.KeyPair,
	iscPackage *iotago.PackageID,
	l1client clients.L1Client,
) *TestChainLedger {
	return &TestChainLedger{
		t:           t,
		governor:    originator,
		l1client:    l1client,
		iscPackage:  iscPackage,
		fetchedReqs: map[cryptolib.AddressKey]map[iotago.ObjectID]bool{},
	}
}

// Only set after MakeTxChainOrigin.
func (tcl *TestChainLedger) ChainID() isc.ChainID {
	return tcl.chainID
}

func (tcl *TestChainLedger) MakeTxChainOrigin(committeeAddress *cryptolib.Address) *isc.StateAnchor {
	coinType := iotajsonrpc.IotaCoinType
	resGetCoins, err := tcl.l1client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: tcl.governor.Address().AsIotaAddress(), CoinType: &coinType})
	require.NoError(tcl.t, err)
	originDeposit := resGetCoins.Data[2]
	schemaVersion := allmigrations.DefaultScheme.LatestSchemaVersion()
	initParams := isc.NewCallArguments([]byte{1, 2, 3})

	// FIXME failed to add origin deposit
	l1commitment := origin.L1Commitment(schemaVersion, initParams, 0, isc.BaseTokenCoinInfo)
	stateMetadata := transaction.NewStateMetadata(
		schemaVersion,
		l1commitment,
		&gas.FeePolicy{
			GasPerToken: util.Ratio32{
				A: 1,
				B: 2,
			},
			EVMGasRatio: util.Ratio32{
				A: 3,
				B: 4,
			},
			ValidatorFeeShare: 5,
		},
		initParams,
		"https://iota.org",
	)

	// FIXME this may refer to the ObjectRef with older version, and trigger panic
	anchorRef, err := tcl.l1client.L2().StartNewChain(
		context.Background(),
		tcl.governor,
		*tcl.iscPackage,
		stateMetadata.Bytes(),
		originDeposit.Ref(),
		[]*iotago.ObjectRef{resGetCoins.Data[0].Ref()},
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(tcl.t, err)
	stateAnchor := isc.NewStateAnchor(anchorRef, tcl.governor.Address(), *tcl.iscPackage)
	require.NotNil(tcl.t, stateAnchor)
	tcl.chainID = stateAnchor.ChainID()

	return &stateAnchor
}

func (tcl *TestChainLedger) MakeTxAccountsDeposit(account *cryptolib.KeyPair) (isc.Request, error) {
	resp, err := tcl.l1client.L2().CreateAndSendRequestWithAssets(
		context.Background(),
		account,
		*tcl.iscPackage,
		tcl.chainID.AsAddress().AsIotaAddress(),
		iscmove.NewAssets(100_000_00),
		&iscmove.Message{
			Contract: uint32(isc.Hn("accounts")),
			Function: uint32(isc.Hn("deposit")),
		},
		iscmove.NewAssets(100_000_000),
		1000,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	if err != nil {
		return nil, err
	}
	reqRef, err := resp.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	if err != nil {
		return nil, err
	}
	req, err := tcl.l1client.L2().GetRequestFromObjectID(context.Background(), reqRef.ObjectID)
	if err != nil {
		return nil, err
	}
	return isc.OnLedgerFromRequest(req, tcl.chainID.AsAddress())
}

func (tcl *TestChainLedger) RunOnChainStateTransition(anchor *isc.StateAnchor, pt iotago.ProgrammableTransaction) (*isc.StateAnchor, error) {
	signer := cryptolib.SignerToIotaSigner(tcl.governor)

	coinPage, err := tcl.l1client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address()})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
	}
	var gasPayments []*iotago.ObjectRef
	for _, coin := range coinPage.Data {
		if !pt.IsInInputObjects(coin.CoinObjectID) {
			gasPayments = []*iotago.ObjectRef{coin.Ref()}
			break
		}
	}
	tx := iotago.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TransactionData: %w", err)
	}
	_, err = tcl.l1client.SignAndExecuteTransaction(
		context.Background(),
		signer,
		txBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to SignAndExecuteTransaction: %w", err)
	}
	return tcl.UpdateAnchor(anchor)
}

func (tcl *TestChainLedger) UpdateAnchor(anchor *isc.StateAnchor) (*isc.StateAnchor, error) {
	anchorRef, err := tcl.l1client.UpdateObjectRef(context.Background(), anchor.GetObjectRef())
	if err != nil {
		return nil, err
	}
	anchorWithRef, err := tcl.l1client.L2().GetAnchorFromObjectID(context.Background(), anchorRef.ObjectID)
	if err != nil {
		return nil, err
	}
	anchor.Anchor = anchorWithRef
	return anchor, nil
}

func (tcl *TestChainLedger) FakeRotationTX(anchor *isc.StateAnchor, nextCommitteeAddr *cryptolib.Address) *isc.StateAnchor {
	// FIXME a temp impl before the decision of Rotation
	signer := cryptolib.SignerToIotaSigner(tcl.governor)
	ptb := iotago.NewProgrammableTransactionBuilder()

	ptb.Command(iotago.Command{
		TransferObjects: &iotago.ProgrammableTransferObjects{
			Objects: []iotago.Argument{ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: anchor.GetObjectRef()})},
			Address: ptb.MustPure(nextCommitteeAddr),
		},
	},
	)

	pt := ptb.Finish()
	coins, err := tcl.l1client.GetCoinObjsForTargetAmount(context.Background(), signer.Address(), iotaclient.DefaultGasBudget)
	require.NoError(tcl.t, err)
	gasPayments := coins.CoinRefs()

	tx := iotago.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)

	txnBytes, err := bcs.Marshal(&tx)
	require.NoError(tcl.t, err)
	txnResponse, err := tcl.l1client.SignAndExecuteTransaction(
		context.Background(),
		signer,
		txnBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(tcl.t, err)
	require.True(tcl.t, txnResponse.Effects.Data.IsSuccess())

	anchorRef, err := tcl.l1client.L2().GetAnchorFromObjectID(context.Background(), anchor.GetObjectID())
	require.NoError(tcl.t, err)
	tmp := isc.NewStateAnchor(anchorRef, tcl.governor.Address(), *tcl.iscPackage)
	return &tmp
}

func BuildAndDeployISCContracts(t *testing.T, client clients.L1Client, signer cryptolib.Signer) iotago.PackageID {
	iotaSigner := cryptolib.SignerToIotaSigner(signer)
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(context.Background(), iotaclient.PublishRequest{
		Sender:          iotaSigner.Address(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		iotaSigner,
		txnBytes.TxBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	return *packageID
}
