package votepower

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/zennittians/intelchain/crypto/bls"

	shardingconfig "github.com/zennittians/intelchain/internal/configs/sharding"

	"github.com/ethereum/go-ethereum/common"
	bls_core "github.com/zennittians/bls/ffi/go/bls"
	"github.com/zennittians/intelchain/numeric"
	"github.com/zennittians/intelchain/shard"
)

var (
	slotList        shard.SlotList
	totalStake      numeric.Dec
	intelchainNodes = 10
	stakedNodes     = 10
	maxAccountGen   = int64(98765654323123134)
	accountGen      = rand.New(rand.NewSource(1337))
	maxKeyGen       = int64(98765654323123134)
	keyGen          = rand.New(rand.NewSource(42))
	maxStakeGen     = int64(200)
	stakeGen        = rand.New(rand.NewSource(541))
)

func init() {
	shard.Schedule = shardingconfig.LocalnetSchedule
	for i := 0; i < intelchainNodes; i++ {
		newSlot := generateRandomSlot()
		newSlot.EffectiveStake = nil
		slotList = append(slotList, newSlot)
	}

	totalStake = numeric.ZeroDec()
	for j := 0; j < stakedNodes; j++ {
		newSlot := generateRandomSlot()
		totalStake = totalStake.Add(*newSlot.EffectiveStake)
		slotList = append(slotList, newSlot)
	}
}

func generateRandomSlot() shard.Slot {
	addr := common.Address{}
	addr.SetBytes(big.NewInt(int64(accountGen.Int63n(maxAccountGen))).Bytes())
	secretKey := bls_core.SecretKey{}
	secretKey.Deserialize(big.NewInt(int64(keyGen.Int63n(maxKeyGen))).Bytes())
	key := bls.SerializedPublicKey{}
	key.FromLibBLSPublicKey(secretKey.GetPublicKey())
	stake := numeric.NewDecFromBigInt(big.NewInt(int64(stakeGen.Int63n(maxStakeGen))))
	return shard.Slot{EcdsaAddress: addr, BLSPublicKey: key, EffectiveStake: &stake}
}

func TestCompute(t *testing.T) {
	expectedRoster := NewRoster(shard.BeaconChainShardID)
	// Calculated when generated
	expectedRoster.TotalEffectiveStake = totalStake
	expectedRoster.ITCSlotCount = int64(intelchainNodes)

	asDecITCSlotCount := numeric.NewDec(expectedRoster.ITCSlotCount)
	ourPercentage := numeric.ZeroDec()
	theirPercentage := numeric.ZeroDec()

	staked := slotList
	for i := range staked {
		member := AccommodateIntelchainVote{
			PureStakedVote: PureStakedVote{
				EarningAccount: staked[i].EcdsaAddress,
				Identity:       staked[i].BLSPublicKey,
				GroupPercent:   numeric.ZeroDec(),
				EffectiveStake: numeric.ZeroDec(),
			},
			OverallPercent:   numeric.ZeroDec(),
			IsIntelchainNode: false,
		}

		// Real Staker
		intelchainPercent := shard.Schedule.InstanceForEpoch(big.NewInt(3)).IntelchainVotePercent()
		externalPercent := shard.Schedule.InstanceForEpoch(big.NewInt(3)).ExternalVotePercent()
		if e := staked[i].EffectiveStake; e != nil {
			member.EffectiveStake = member.EffectiveStake.Add(*e)
			member.GroupPercent = e.Quo(expectedRoster.TotalEffectiveStake)
			member.OverallPercent = member.GroupPercent.Mul(externalPercent)
			theirPercentage = theirPercentage.Add(member.OverallPercent)
		} else { // Our node
			member.IsIntelchainNode = true
			member.OverallPercent = intelchainPercent.Quo(asDecITCSlotCount)
			member.GroupPercent = member.OverallPercent.Quo(intelchainPercent)
			ourPercentage = ourPercentage.Add(member.OverallPercent)
		}

		expectedRoster.Voters[staked[i].BLSPublicKey] = &member
	}

	expectedRoster.OurVotingPowerTotalPercentage = ourPercentage
	expectedRoster.TheirVotingPowerTotalPercentage = theirPercentage

	computedRoster, err := Compute(&shard.Committee{
		ShardID: shard.BeaconChainShardID, Slots: slotList,
	}, big.NewInt(3))
	if err != nil {
		t.Error("Computed Roster failed on vote summation to one")
	}

	if !compareRosters(expectedRoster, computedRoster, t) {
		t.Errorf("Compute Roster mismatch with expected Roster")
	}
	// Check that voting percents sum to 100
	if !computedRoster.OurVotingPowerTotalPercentage.Add(
		computedRoster.TheirVotingPowerTotalPercentage,
	).Equal(numeric.OneDec()) {
		t.Errorf(
			"Total voting power does not equal 1. Intelchain voting power: %s, Staked voting power: %s",
			computedRoster.OurVotingPowerTotalPercentage,
			computedRoster.TheirVotingPowerTotalPercentage,
		)
	}
}

func compareRosters(a, b *Roster, t *testing.T) bool {
	voterMatch := true
	for k, voter := range a.Voters {
		if other, exists := b.Voters[k]; exists {
			if !compareStakedVoter(voter, other) {
				t.Error("voter slot not match")
				voterMatch = false
			}
		} else {
			t.Error("computed roster missing")
			voterMatch = false
		}
	}
	return a.OurVotingPowerTotalPercentage.Equal(b.OurVotingPowerTotalPercentage) &&
		a.TheirVotingPowerTotalPercentage.Equal(b.TheirVotingPowerTotalPercentage) &&
		a.TotalEffectiveStake.Equal(b.TotalEffectiveStake) &&
		a.ITCSlotCount == b.ITCSlotCount && voterMatch
}

func compareStakedVoter(a, b *AccommodateIntelchainVote) bool {
	return a.IsIntelchainNode == b.IsIntelchainNode &&
		a.EarningAccount == b.EarningAccount &&
		a.OverallPercent.Equal(b.OverallPercent) &&
		a.EffectiveStake.Equal(b.EffectiveStake)
}
