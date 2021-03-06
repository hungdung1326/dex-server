# REST API

There are 7 different resources on the matching engine REST API:

* accounts
* pairs
* tokens
* trades
* orderbook
* orders
* ohlcv

Moreover, there is one resource for getting general information:

* info


# Account resource

### GET /account/{userAddress}

Retrieve the account information for a certain Ethereum address (mainly token balances)

### GET /account/{userAddress}/{tokenAddress}

Retrieve the token balance of a certain Ethereum address

* {userAddress} is the Ethereum address of a user/client wallet
* {tokenAddress} is the Ethereum address of a token (base or quote)

### POST /account/create?address={newAddress}

* {newAddress} is the Ethereum address of a user/client wallet


# Pairs resource

### GET /pair?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve the pair information corresponding to a baseToken and a quoteToken where:

* {baseToken} is the Ethereum address of a base token
* {quoteToken} is the Ethereum address of a quote token

### GET /pairs

Retrieve all pairs currently registered on the exchange

### GET /pairs/data?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve pair data corresponding to a baseToken and quoteToken where

* {baseToken} is the Ethereum address of a base token
* {quoteToken} is the Ethereum address of a quote token

This endpoints returns the Open, High, Low, Close, Volume and Change for the last 24 hours
as well as the last price.

### POST /pairs

Create/Insert pair in DB.

* Sample request parameters:
```json
{
    "BaseTokenAddress":"0x4f696e8A1A3fB3AEA9f72EB100eA8d97c5130B32",
    "BaseTokenSymbol": "KCS",
    "QuoteTokenAddress":"0x1888a8db0b7db59413ce07150b3373972bf818d4", // Must be registered as quote when creating new token
    "QuoteTokenSymbol":"HTC",
    "active":true
}
```

# Tokens resource

### GET /tokens

Retrieve all tokens currently registered on the exchange

### GET /tokens/base

Retrieve all base tokens currently registered on the exchange

### GET /tokens/quote

Retrieve all quote tokens currently registered on the exchange

### GET /tokens/{address}

Retrieve token information for a token at a certain address

* {address} is an Ethereum address

### POST /tokens

Create/Insert token in DB.

* Sample request parameters:
```json
{
	"name":"HotPotCoin",
	"symbol":"HPC",
	"decimal":18,
	"contractAddress":"0x1888a8db0b7db59413ce07150b3373972bf818d3",
	"active":true,
	"quote":true // This is required when creating pair
}
```

# Orderbook resource

### GET /orderbook?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve the orderbook (amount and pricepoint) corresponding to a a baseToken and a quoteToken where:

* {baseToken} is the Ethereum address of a base token
* {quoteToken} is the Ethereum address of a quote token

### GET /orderbook/raw?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve the orderbook (full raw orders, including fields such as hashes, maker, taker addresses, signatures, etc.)
corresponding to a baseToken and a quoteToken.

* {baseToken} is the Ethereum address of a base token
* {quoteToken} is the Ethereum address of a quote token


# Trade resource

### GET /trades?address={address}&limit={limit}

Retrieve the sorted list of trades for an Ethereum address in which the given address is either maker or taker

* {address} is an Ethereum address
* {limit} is the number of records returned

### GET /trades/pair?baseToken={baseToken}&quoteToken={quoteToken}&limit={limit}

Retrieve all trades corresponding to a baseToken and a quoteToken

* {baseToken} is the Ethereum address of a base token
* {quoteToken} is the Ethereum address of a quote token
* {limit} is the number of records returned


# Order resource

### GET /orders?address={address}

Retrieve the sorted list of orders for an Ethereum address

### GET /orders/positions?address={address}

Retrieve the list of positions for an Ethereum address. Positions are order that have been sent
to the matching engine and that are waiting to be matched

* {address} is an Ethereum address

### GET /orders/history?address={address}

Retrieve the list of filled order for an Ethereum address.

* {address} is an Ethereum address

### GET /orders/feeds/{address}?tokenAddress={tokenAddress}

Retrieve the list of filled order for an Ethereum address from Swarm Feed.

* {address} is an Ethereum address
* {tokenAddress} is the token address


# OHLCV resource

### GET /ohlcv?baseToken={baseToken}&quoteToken={quoteToken}&pairName={pairName}&unit={unit}&duration={duration}&from={from}&to={to}

Retrieve OHLCV data corresponding to a baseToken and a quoteToken.

* {baseToken} is the Ethereum address of a baseToken
* {quoteToken} is the Ethereum address of a quoteToken
* {pairName} is the pair name under the format {baseTokenSymbol}/{quoteTokenSymbol}(eg. "ZRX/WETH"). I believe this parameter is currently required but it's planned to be optional. The idea is for this parameter to be used for verifications purposes and the API to send back an eror if it does not correspond to a baseToken/quoteToken parameters
* {duration} is the duration (in units, see param below) of each candlestick
* {units} is the unit used to represent the above duration: "minute", "hour", "day", "week", "month"
* {from} is the beginning timestamp from which ohlcv data has to be queried
* {to} is the ending timestamp until which ohlcv data has to be queried


# Info resource

### GET /info

Get general information (exchange address, fees and operators)

### GET /info/exchange

Get exchange address

### GET /info/operators

Get operators information

### GET /info/fees

Get fees information
