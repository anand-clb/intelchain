package staking

import (
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	isValidatorKeyStr     = "Intelchain/IsValidator/Key/v1"
	isValidatorStr        = "Intelchain/IsValidator/Value/v1"
	collectRewardsStr     = "Intelchain/CollectRewards"
	delegateStr           = "Intelchain/Delegate"
	unDelegateStr         = "Intelchain/UnDelegate"
	firstElectionEpochStr = "Intelchain/FirstElectionEpoch/Key/v1"
)

// keys used to retrieve staking related informatio
var (
	IsValidatorKey        = crypto.Keccak256Hash([]byte(isValidatorKeyStr))
	IsValidator           = crypto.Keccak256Hash([]byte(isValidatorStr))
	CollectRewardsTopic   = crypto.Keccak256Hash([]byte(collectRewardsStr))
	DelegateTopic         = crypto.Keccak256Hash([]byte(delegateStr))
	UnDelegateTopic       = crypto.Keccak256Hash([]byte(unDelegateStr))
	FirstElectionEpochKey = crypto.Keccak256Hash([]byte(firstElectionEpochStr))
)
