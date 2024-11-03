package keeper

import (
	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DistributeFromFeePool distributes funds from the distribution module account to
// a receiver address while updating the community pool
func (k Keeper) DistributeFromFeePool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	// NOTE the community pool isn't a module account, however its coins
	// are held in the distribution module account. Thus the community pool
	// must be reduced separately from the SendCoinsFromModuleToAccount call
	newPool, negative := feePool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(amount...))
	if negative {
		return types.ErrBadDistribution
	}

	feePool.CommunityPool = newPool

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
	if err != nil {
		return err
	}

	return k.FeePool.Set(ctx, feePool)
}

func (k Keeper) MintCoins(ctx sdk.Context, coin sdk.Coin) error {
	coins := sdk.NewCoins(coin)

	// skip as no coins need to be minted
	if coins.Empty() {
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
}

func (k Keeper) AllocateCommunityRewards(ctx sdk.Context, coin sdk.Coin) error {

	//Calculate Rewards for block duration

	lastHalvingAmount, _ := k.GetLastHalvingAmount(ctx)
	previousBlockTime, _ := k.GetPreviousBlockTime(ctx)
	lastHalvingTime, err := k.GetLastHalvingTime(ctx)

	if err != nil {
		lastHalvingTime = time.Now()
	}

	tmNow := time.Now()

	if lastHalvingTime.Add(time.Hour * 365 * 24).Before(tmNow) {
		lastHalvingAmount = lastHalvingAmount.Quo(math.NewInt(2))
		k.SetLastHalvingAmount(ctx, lastHalvingAmount)
	}

	denom, _ := sdk.GetBaseDenom()
	duration := tmNow.Sub(previousBlockTime)
	amount := lastHalvingAmount.MulRaw(duration.Milliseconds()).QuoRaw(365 * 24 * 3600 * 1000)

	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))

	// skip as no coins need to be minted
	if coin.IsZero() {
		return nil
	}

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}

	k.FundCommunityPool(ctx, coins, sdk.AccAddress(types.ModuleName))

	return nil
}
