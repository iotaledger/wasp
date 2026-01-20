package graphqltypes

import "github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

type ValidatorSummary struct {
	IotaAddress            iotago.Address
	ProtocolPubkeyBytes    []byte
	NetworkPubkeyBytes     []byte
	WorkerPubkeyBytes      []byte
	ProofOfPossessionBytes []byte
	OperationCapID         iotago.ObjectID
	Name                   string
	Description            string
	ImageURL               string
	ProjectURL             string
	P2pAddress             string
	NetAddress             string
	PrimaryAddress         string
	WorkerAddress          string

	NextEpochProtocolPubkeyBytes []byte
	NextEpochProofOfPossession   []byte
	NextEpochNetworkPubkeyBytes  []byte
	NextEpochWorkerPubkeyBytes   []byte
	NextEpochNetAddress          string
	NextEpochP2pAddress          string
	NextEpochPrimaryAddress      string
	NextEpochWorkerAddress       string

	VotingPower             uint64
	GasPrice                uint64
	CommissionRate          uint64
	NextEpochStake          uint64
	NextEpochGasPrice       uint64
	NextEpochCommissionRate uint64
	StakingPoolID           iotago.ObjectID

	StakingPoolActivationEpoch   uint64
	StakingPoolDeactivationEpoch uint64

	StakingPoolIotaBalance   uint64
	RewardsPool              uint64
	PoolTokenBalance         uint64
	PendingStake             uint64
	PendingPoolTokenWithdraw uint64
	PendingTotalIotaWithdraw uint64
	ExchangeRatesID          iotago.ObjectID
	ExchangeRatesSize        uint64
}

type SystemStateSummary struct {
	Epoch                                 uint64
	ProtocolVersion                       uint64
	SystemStateVersion                    uint64
	IotaTotalSupply                       uint64
	StorageFundTotalObjectStorageRebates  uint64
	StorageFundNonRefundableBalance       uint64
	ReferenceGasPrice                     uint64
	SafeMode                              bool
	SafeModeStorageCharges                uint64
	SafeModeStorageRewards                uint64
	SafeModeComputationRewards            uint64
	SafeModeStorageRebates                uint64
	SafeModeNonRefundableStorageFee       uint64
	EpochStartTimestampMs                 uint64
	EpochDurationMs                       uint64
	MinValidatorCount                     uint64
	StakeSubsidyStartEpoch                uint64
	MaxValidatorCount                     uint64
	MinValidatorJoiningStake              uint64
	ValidatorLowStakeThreshold            uint64
	ValidatorVeryLowStakeThreshold        uint64
	ValidatorLowStakeGracePeriod          uint64
	StakeSubsidyBalance                   uint64
	StakeSubsidyDistributionCounter       uint64
	StakeSubsidyCurrentDistributionAmount uint64
	StakeSubsidyPeriodLength              uint64
	StakeSubsidyDecreaseRate              uint16
	TotalStake                            uint64
	ActiveValidators                      []ValidatorSummary
	PendingActiveValidatorsID             iotago.ObjectID
	PendingActiveValidatorsSize           uint64
	PendingRemovals                       []uint64
	StakingPoolMappingsID                 iotago.ObjectID
	StakingPoolMappingsSize               uint64
	InactivePoolsID                       iotago.ObjectID
	InactivePoolsSize                     uint64
	ValidatorCandidatesID                 iotago.ObjectID
	ValidatorCandidatesSize               uint64
	AtRiskValidators                      interface{}
	ValidatorReportRecords                interface{}
}
