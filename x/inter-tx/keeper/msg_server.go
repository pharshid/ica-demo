package keeper

import (
	"context"
	"fmt"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/cosmos/interchain-accounts/x/inter-tx/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl creates and returns a new types.MsgServer, fulfilling the intertx Msg service interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterAccount implements the Msg/RegisterAccount interface
func (k msgServer) RegisterAccount(goCtx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.Owner, msg.Version); err != nil {
		return nil, err
	}

	return &types.MsgRegisterAccountResponse{}, nil
}

// SubmitTx implements the Msg/SubmitTx interface
func (k msgServer) SubmitTx(goCtx context.Context, msg *types.MsgSubmitTx) (*types.MsgSubmitTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	portID, err := icatypes.NewControllerPortID(msg.Owner)
	if err != nil {
		return nil, err
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, []proto.Message{msg.GetTxMsg()})
	if err != nil {
		return nil, err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to sastisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	_, err = k.icaControllerKeeper.SendTx(ctx, nil, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitTxResponse{}, nil
}

// SubmitTx implements the Msg/SubmitTx interface
func (k msgServer) IBCDelegate(goCtx context.Context, msg *types.MsgIBCDelegate) (*types.MsgIBCDelegateResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)
	res, _ := k.bankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{Address: msg.Address, Denom: msg.Denom})
	if res.GetBalance().Amount.IsPositive() {
		k.SubmitTx(goCtx, &types.MsgSubmitTx{Owner: msg.Owner,
			ConnectionId: msg.ConnectionId,
			Msg: &codectypes.Any{Value: []byte(fmt.Sprintf(string(`"@type":"/cosmos.staking.v1beta1.MsgDelegate",
    "delegator_address":"%s",
    "validator_address":"%s",
    "amount": {
        "denom": "%s",
        "amount": "%d"
    }`), msg.DelAddr, msg.Validator, msg.Denom, msg.BondAmt))}})
	}
	return &types.MsgIBCDelegateResponse{}, nil
}
