package keeper

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// InitGenesis new mint genesis
func (keeper Keeper) InitGenesis(ctx sdk.Context, ak types.AccountKeeper, data *types.GenesisState) {
	if err := keeper.Minter.Set(ctx, data.Minter); err != nil {
		panic(err)
	}

	if err := keeper.Params.Set(ctx, data.Params); err != nil {
		panic(err)
	}

	ak.GetModuleAccount(ctx, types.ModuleName)
	bigIntAmount := new(big.Int)
	bigIntAmount.SetString("23668256824195824000000000", 10)
	amount := sdk.NewCoins((sdk.NewCoin("aepix", math.NewIntFromBigInt(bigIntAmount))))

	if err := keeper.MintCoins(ctx, amount); err != nil {
		panic(fmt.Sprintf("mint coin to %s module", types.ModuleName))
	}

	if err := keeper.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, "distribution", amount); err != nil {
		panic(fmt.Sprintf("mint coin to distribution module from %s module", types.ModuleName))
	}

}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minter, err := keeper.Minter.Get(ctx)
	if err != nil {
		panic(err)
	}

	params, err := keeper.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(minter, params)
}
