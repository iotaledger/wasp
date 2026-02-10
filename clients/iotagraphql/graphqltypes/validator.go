package graphqltypes

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type StakeStatus = Status

type Status struct {
	Pending *struct{} `json:"Pending,omitempty"`
	Active  *struct {
		EstimatedReward *BigInt `json:"estimatedReward"`
	} `json:"Active,omitempty"`
	Unstaked *struct{} `json:"Unstaked,omitempty"`
}

func (s Status) Tag() string {
	return "status"
}

func (s Status) Content() string {
	return ""
}

const (
	StakeStatusActive   = "Active"
	StakeStatusPending  = "Pending"
	StakeStatusUnstaked = "Unstaked"
)

type Stake struct {
	StakedIotaID      iotago.ObjectID `json:"stakedIotaId"`
	StakeRequestEpoch *BigInt         `json:"stakeRequestEpoch"`
	StakeActiveEpoch  *BigInt         `json:"stakeActiveEpoch"`
	Principal         *BigInt         `json:"principal"`
	StakeStatus       *StakeStatus    `json:"status"`
}

func (s *Stake) IsActive() bool {
	return s.StakeStatus.Active != nil
}

type DelegatedStake struct {
	ValidatorAddress iotago.Address  `json:"validatorAddress"`
	StakingPool      iotago.ObjectID `json:"stakingPool"`
	Stakes           []Stake         `json:"stakes"`
}

type IotaValidatorSummary struct {
	IotaAddress            iotago.Address    `json:"iotago.Address"`
	ProtocolPubkeyBytes    iotago.Base64Data `json:"protocolPubkeyBytes"`
	NetworkPubkeyBytes     iotago.Base64Data `json:"networkPubkeyBytes"`
	WorkerPubkeyBytes      iotago.Base64Data `json:"workerPubkeyBytes"`
	ProofOfPossessionBytes iotago.Base64Data `json:"proofOfPossessionBytes"`
	OperationCapID         iotago.ObjectID   `json:"operationCapId"`
	Name                   string            `json:"name"`
	Description            string            `json:"description"`
	ImageURL               string            `json:"imageUrl"`
	ProjectURL             string            `json:"projectUrl"`
	P2pAddress             string            `json:"p2pAddress"`
	NetAddress             string            `json:"netAddress"`
	PrimaryAddress         string            `json:"primaryAddress"`
	WorkerAddress          string            `json:"workerAddress"`

	NextEpochProtocolPubkeyBytes iotago.Base64Data `json:"nextEpochProtocolPubkeyBytes"`
	NextEpochProofOfPossession   iotago.Base64Data `json:"nextEpochProofOfPossession"`
	NextEpochNetworkPubkeyBytes  iotago.Base64Data `json:"nextEpochNetworkPubkeyBytes"`
	NextEpochWorkerPubkeyBytes   iotago.Base64Data `json:"nextEpochWorkerPubkeyBytes"`
	NextEpochNetAddress          string            `json:"nextEpochNetAddress"`
	NextEpochP2pAddress          string            `json:"nextEpochP2pAddress"`
	NextEpochPrimaryAddress      string            `json:"nextEpochPrimaryAddress"`
	NextEpochWorkerAddress       string            `json:"nextEpochWorkerAddress"`

	VotingPower             *BigInt         `json:"votingPower"`
	GasPrice                *BigInt         `json:"gasPrice"`
	CommissionRate          *BigInt         `json:"commissionRate"`
	NextEpochStake          *BigInt         `json:"nextEpochStake"`
	NextEpochGasPrice       *BigInt         `json:"nextEpochGasPrice"`
	NextEpochCommissionRate *BigInt         `json:"nextEpochCommissionRate"`
	StakingPoolID           iotago.ObjectID `json:"stakingPoolId"`

	StakingPoolActivationEpoch   *BigInt `json:"stakingPoolActivationEpoch"`
	StakingPoolDeactivationEpoch *BigInt `json:"stakingPoolDeactivationEpoch"`

	StakingPoolIotaBalance   *BigInt         `json:"stakingPoolIotaBalance"`
	RewardsPool              *BigInt         `json:"rewardsPool"`
	PoolTokenBalance         *BigInt         `json:"poolTokenBalance"`
	PendingStake             *BigInt         `json:"pendingStake"`
	PendingPoolTokenWithdraw *BigInt         `json:"pendingPoolTokenWithdraw"`
	PendingTotalIotaWithdraw *BigInt         `json:"pendingTotalIotaWithdraw"`
	ExchangeRatesID          iotago.ObjectID `json:"exchangeRatesId"`
	ExchangeRatesSize        *BigInt         `json:"exchangeRatesSize"`
}

type (
	TypeName []iotago.Address
	// IotaSystemStateSummary represents the IOTA system state summary.
	// TODO: This struct may be outdated and needs to be updated.
	IotaSystemStateSummary struct {
		Epoch                                 *BigInt                `json:"epoch"`
		ProtocolVersion                       *BigInt                `json:"protocolVersion"`
		SystemStateVersion                    *BigInt                `json:"systemStateVersion"`
		IotaTotalSupply                       *BigInt                `json:"iotaTotalSupply"`
		StorageFundTotalObjectStorageRebates  *BigInt                `json:"storageFundTotalObjectStorageRebates"`
		StorageFundNonRefundableBalance       *BigInt                `json:"storageFundNonRefundableBalance"`
		ReferenceGasPrice                     *BigInt                `json:"referenceGasPrice"`
		SafeMode                              bool                   `json:"safeMode"`
		SafeModeStorageCharges                *BigInt                `json:"safeModeStorageCharges"`
		SafeModeStorageRewards                *BigInt                `json:"safeModeStorageRewards"`
		SafeModeComputationRewards            *BigInt                `json:"safeModeComputationRewards"`
		SafeModeStorageRebates                *BigInt                `json:"safeModeStorageRebates"`
		SafeModeNonRefundableStorageFee       *BigInt                `json:"safeModeNonRefundableStorageFee"`
		EpochStartTimestampMs                 *BigInt                `json:"epochStartTimestampMs"`
		EpochDurationMs                       *BigInt                `json:"epochDurationMs"`
		MinValidatorCount                     *BigInt                `json:"minValidatorCount"`
		StakeSubsidyStartEpoch                *BigInt                `json:"stakeSubsidyStartEpoch"`
		MaxValidatorCount                     *BigInt                `json:"maxValidatorCount"`
		MinValidatorJoiningStake              *BigInt                `json:"minValidatorJoiningStake"`
		ValidatorLowStakeThreshold            *BigInt                `json:"validatorLowStakeThreshold"`
		ValidatorVeryLowStakeThreshold        *BigInt                `json:"validatorVeryLowStakeThreshold"`
		ValidatorLowStakeGracePeriod          *BigInt                `json:"validatorLowStakeGracePeriod"`
		StakeSubsidyBalance                   *BigInt                `json:"stakeSubsidyBalance"`
		StakeSubsidyDistributionCounter       *BigInt                `json:"stakeSubsidyDistributionCounter"`
		StakeSubsidyCurrentDistributionAmount *BigInt                `json:"stakeSubsidyCurrentDistributionAmount"`
		StakeSubsidyPeriodLength              *BigInt                `json:"stakeSubsidyPeriodLength"`
		StakeSubsidyDecreaseRate              uint16                 `json:"stakeSubsidyDecreaseRate"`
		TotalStake                            *BigInt                `json:"totalStake"`
		ActiveValidators                      []IotaValidatorSummary `json:"activeValidators"`
		PendingActiveValidatorsID             iotago.ObjectID        `json:"pendingActiveValidatorsId"`
		PendingActiveValidatorsSize           *BigInt                `json:"pendingActiveValidatorsSize"`
		PendingRemovals                       []*BigInt              `json:"pendingRemovals"`
		StakingPoolMappingsID                 iotago.ObjectID        `json:"stakingPoolMappingsId"`
		StakingPoolMappingsSize               *BigInt                `json:"stakingPoolMappingsSize"`
		InactivePoolsID                       iotago.ObjectID        `json:"inactivePoolsId"`
		InactivePoolsSize                     *BigInt                `json:"inactivePoolsSize"`
		ValidatorCandidatesID                 iotago.ObjectID        `json:"validatorCandidatesId"`
		ValidatorCandidatesSize               *BigInt                `json:"validatorCandidatesSize"`
		AtRiskValidators                      interface{}            `json:"atRiskValidators"`
		ValidatorReportRecords                interface{}            `json:"validatorReportRecords"`
	}
)

type ValidatorsApy struct {
	Epoch *BigInt `json:"epoch"`
	Apys  []struct {
		Address string  `json:"address"`
		Apy     float64 `json:"apy"`
	} `json:"apys"`
}

func (apys *ValidatorsApy) ApyMap() map[string]float64 {
	res := make(map[string]float64)
	for _, apy := range apys.Apys {
		res[apy.Address] = apy.Apy
	}
	return res
}
