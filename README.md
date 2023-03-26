# Interchain Accounts

Developers integrating Interchain Accounts may choose to firstly enable host chain functionality, and add authentication modules later as desired. This is the updated and modified version of the original demo maintained [here](https://github.com/cosmos/interchain-accounts-demo).

## Overview 

The Interchain Accounts module is now maintained within the `ibc-go` repository [here](https://github.com/cosmos/ibc-go/tree/main/modules/apps/27-interchain-accounts). 
Interchain Accounts is now available in the [`v3.0.0`](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0) release of `ibc-go`.

### Developer Documentation

Interchain Accounts developer docs can be found on the IBC documentation website.

https://ibc.cosmos.network/main/apps/interchain-accounts/overview.html

## Setup

1. Clone this repository and build the application binary

```bash
git clone https://github.com/pharshid/ica-demo.git
cd ica-demo

make install 
```

2. Download and install an IBC relayer. ([go relayer](https://github.com/cosmos/relayer))
```bash

git clone https://github.com/cosmos/relayer.git
cd relayer
make install
```

3. Bootstrap two chains, configure the relayer and create an IBC connection (on top of clients that are created as well)
```bash
# go relayer
make init-golang-rely
```
4. Start the relayer
```bash
#go relayer
make start-golang-rly
```

## Demo

**NOTE:** For the purposes of this demo the setup scripts have been provided with a set of hardcoded mnemonics that generate deterministic wallet addresses used below.

```bash
# Store the following account addresses within the current shell env
export WALLET_1=$(icad keys show wallet1 -a --keyring-backend test --home ./data/test-1) && echo $WALLET_1;
export WALLET_2=$(icad keys show wallet2 -a --keyring-backend test --home ./data/test-1) && echo $WALLET_2;
export WALLET_3=$(icad keys show wallet3 -a --keyring-backend test --home ./data/test-2) && echo $WALLET_3;
export WALLET_4=$(icad keys show wallet4 -a --keyring-backend test --home ./data/test-2) && echo $WALLET_4;
```

### Registering an Interchain Account via IBC

Register an Interchain Account using the `intertx register` cmd. 
Here the message signer is used as the account owner.

```bash
# Register an interchain account on behalf of WALLET_1 where chain test-2 is the interchain accounts host
icad tx intertx register --from $WALLET_1 --connection-id connection-0 --chain-id test-1 --home ./data/test-1 --node tcp://localhost:16657 --keyring-backend test -y

# Query the address of the interchain account
icad query intertx interchainaccounts connection-0 $WALLET_1 --home ./data/test-1 --node tcp://localhost:16657

# Store the interchain account address by parsing the query result: cosmos1hd0f4u7zgptymmrn55h3hy20jv2u0ctdpq23cpe8m9pas8kzd87smtf8al
export ICA_ADDR=$(icad query intertx interchainaccounts connection-0 $WALLET_1 --home ./data/test-1 --node tcp://localhost:16657 -o json | jq -r '.interchain_account_address') && echo $ICA_ADDR
```

#### Funding the Interchain Account wallet

Allocate funds to the new Interchain Account wallet by using the `bank send` cmd.
Note this is executed on the host chain to provide the account with an initial balance to execute transactions.

```bash
# Query the interchain account balance on the host chain. It should be empty.
icad q bank balances $ICA_ADDR --chain-id test-2 --node tcp://localhost:26657

# Send funds to the interchain account.
icad tx bank send $WALLET_3 $ICA_ADDR 10000stake --chain-id test-2 --home ./data/test-2 --node tcp://localhost:26657 --keyring-backend test -y

# Query the balance once again and observe the changes
icad q bank balances $ICA_ADDR --chain-id test-2 --node tcp://localhost:26657
```

#### Sending Interchain Account transactions

Send Interchain Accounts transactions using the `intertx submit` cmd. 
This command accepts a generic `sdk.Msg` JSON payload or path to JSON file as an arg.

- **Example:** Staking Delegation

```bash
# Output the host chain validator operator address: cosmosvaloper1qnk2n4nlkpw9xfqntladh74w6ujtulwnmxnh3k
cat ./data/test-2/config/genesis.json | jq -r '.app_state.genutil.gen_txs[0].body.messages[0].validator_address'

# Submit a staking delegation tx using the interchain account via ibc
icad tx intertx submit \
'{
    "@type":"/cosmos.staking.v1beta1.MsgDelegate",
    "delegator_address":"cosmos15ccshhmp0gsx29qpqq6g4zmltnnvgmyu9ueuadh9y2nc5zj0szls5gtddz",
    "validator_address":"cosmosvaloper1qnk2n4nlkpw9xfqntladh74w6ujtulwnmxnh3k",
    "amount": {
        "denom": "stake",
        "amount": "1000"
    }
}' --connection-id connection-0 --from $WALLET_1 --chain-id test-1 --home ./data/test-1 --node tcp://localhost:16657 --keyring-backend test -y

# Wait until the relayer has relayed the packet

# Inspect the staking delegations on the host chain
icad q staking delegations-to cosmosvaloper1qnk2n4nlkpw9xfqntladh74w6ujtulwnmxnh3k --home ./data/test-2 --node tcp://localhost:26657
```