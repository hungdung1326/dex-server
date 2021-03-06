# DEX backend

Official decentralized exchange backend, forked from the Proof project.  
Support deposit scenarios: coin-coin, token-coin, token-token.  
The matching-engine will be soon moved to blockchain services.

# Getting Started

## Requirements

- **mongoDB** version 3.6 or newer
- **rabbitmq** version 3.7.7 or newer
- **dep** latest

## Deployment guide step by step

https://github.com/tomochain/dex-client/blob/master/Deployment_step_by_step.md

## Booting up the server

**Install the dependencies**

You need to run `./install-requirements.sh` to install all required libraries

```bash
cd dex-server
export BACKEND=$GOPATH/src/github.com/tomochain/dex-server
mkdir -p $BACKEND
ln -sF $PWD $BACKEND
cd $BACKEND
dep ensure
```

**Start the development**

_If you need to generate genesis block, take a look at [seed-data](./utils/seed-data/README.md)_

```bash
# start dockers including mongo, rabbitmq
yarn start-env
# generate seeds data
yarn seeds
# If you need to reset
yarn restart-env
# If you need to run mongod outside docker-compose
mongod --dbpath utils/datadir
# this will start the server in hot-reload mode
yarn start
# check rabbitmq
docker-compose exec rabbitmq rabbitmq-plugins enable rabbitmq_management
```

=======

# REST API

Download [tomo-dex.postman_collection.json](tomo-dex.postman_collection.json)

See [REST_API.md](REST_API.md)

# Websocket API

See [WEBSOCKET_API.md](WEBSOCKET_API.md)

# Types

## Orders

Orders contain the information that is required to register an order in the orderbook as a "Maker".

- **id** is the primary ID of the order (possibly deprecated)
- **orderType** is either BUY or SELL. It is currently not parsed by the server and compute directly from buyToken, sellToken, buyAmount, sellAmount
- **exchangeAddress** is the exchange smart contract address
- **maker** is the maker (usually sender) ethereum account address
- **buyToken** is the BUY token ethereum address
- **sellToken** is the SELL token ethereum address
- **buyAmount** is the BUY amount (in BUY_TOKEN units)
- **sellAmount** is the SELL amount (in SELL_TOKEN units)
- **expires** is the order expiration timestamp
- **nonce** is the nonce that corresponds to
- **feeMake** is the maker fee (not implemented yet)
- **feeTake** is the taker fee (not implemented yet)
- **pairID** is a hash of the corresponding
- **hash** is a hash of the order details (see details below)
- **signature** is a signature of the order hash. The signer must equal to the maker address for the order to be valid.
- **price** corresponds to the pricepoint computed by the matching engine (not parsed)
- **amount** corresponds to the amount computed by the matching engine (not parsed)

**Order Price and Amount**

There are two ways to describe the amount of tokens being bought/sold. The smart-contract requires (buyToken, sellToken, buyAmount, sellAmount) while the
orderbook requires (pairID, amount, price).

The conversion between both systems can be found in the engine.ComputeOrderPrice
function

**Order Hash**

The order hash is a sha-256 hash of the following elements:

- Exchange address
- Token Buy address
- Amount Buy
- Token Sell Address
- Amount Sell
- Expires
- Nonce
- Maker Address

## Trades

When an order matches another order in the orderbook, the "taker" is required
to sign a trade object that matches an order.

- **orderHash** is the hash of the matching order
- **amount** is the amount of tokens that will be traded
- **trade nonce** is a unique integer to distinguish successive but identical orders (note: can probably be renamed to nonce)
- **taker** is the taker ethereum account address
- **pairID** is a hash identifying the token pair that will be traded
- **hash** is a unique identifier hash of the trade details (see details below)
- **signature** is a signature of the trade hash

Trade Hash:

The trade hash is a sha-256 hash of the following elements:

- Order Hash
- Amount
- Taker Address
- Trade Nonce

The (Order, Trade) tuple can then be used to perform an on-chain transaction for this trade.

## Quote Tokens and Token Pairs

In the same way as traditional exchanges function with the idea of base
currencies and quote currencies, the decentralized exchange works with
base tokens and quote tokens under the following principles:

- Only the exchange operator can register a quote token
- Anybody can register a token pair (but the quote token needs to be registered)

Token pairs are identified by an ID (a hash of both token addresses)
