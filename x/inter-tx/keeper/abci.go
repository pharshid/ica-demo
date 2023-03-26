package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/interchain-accounts/x/inter-tx/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (k *Keeper) EndBlocker(ctx sdk.Context) []abci.ValidatorUpdate {
	DELADR := "cosmos15ccshhmp0gsx29qpqq6g4zmltnnvgmyu9ueuadh9y2nc5zj0szls5gtddz"
	VALADR := "cosmosvaloper1qnk2n4nlkpw9xfqntladh74w6ujtulwnmxnh3k"
	CADR := "cosmos1m9l358xunhhwds0568za49mzhvuxx9uxre5tud"
	CID := "connection-0"
	DENOM := "stake"
	LIMIT := 10000000

	// iaa, _ := sdk.AccAddressFromBech32("cosmos1plzwuk7kqflgpus0jk4mrqpcqwx2wdwp99hl5jrmwr9za7eehhlqx4ysur")
	iar, _ := k.InterchainAccount(ctx, types.NewQueryInterchainAccountRequest(CID, CADR))
	iass := iar.GetInterchainAccountAddress()
	iaa, _ := sdk.AccAddressFromBech32(iass)
	balance := k.bankKeeper.GetBalance(ctx, iaa, DENOM)
	print(balance.Amount)

	deladdr, _ := sdk.AccAddressFromBech32(DELADR)
	valaddr, _ := sdk.ValAddressFromBech32(VALADR)
	del, found := k.stakingKeeper.GetDelegation(ctx, deladdr, valaddr)
	if found {
		amount := del.GetShares().TruncateInt()
		if amount.GT(sdk.NewInt(int64(LIMIT))) {
			icaMsg := &stakingtypes.MsgDelegate{
				DelegatorAddress: DELADR,
				ValidatorAddress: VALADR,
				Amount:           sdk.NewCoin(DENOM, sdk.NewInt(1000)),
			}
			msg, _ := types.NewMsgSubmitTx(icaMsg, CID, CADR)
			msgSrv := NewMsgServerImpl(*k)
			msgSrv.SubmitTx(ctx, msg)
		}
	}
	return []abci.ValidatorUpdate{}
}
