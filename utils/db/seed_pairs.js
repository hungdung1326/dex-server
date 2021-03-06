const utils = require('ethers').utils;
const faker = require('faker');
const argv = require('yargs').argv;
const MongoClient = require('mongodb').MongoClient;
const { getNetworkID, getPriceMultiplier } = require('./utils/helpers');
const { DB_NAME, mongoUrl, network } = require('./utils/config');
const networkID = getNetworkID(network);

let client, db;

const seed = async () => {
  try {
    client = await MongoClient.connect(
      mongoUrl,
      { useNewUrlParser: true }
    );
    db = client.db(DB_NAME);

    let pairs = [];

    const tokens = await db
      .collection('tokens')
      .find({ quote: false }, { symbol: 1, contractAddress: 1, decimals: 1 })
      .toArray();

    const quotes = await db
      .collection('tokens')
      .find(
        { quote: true },
        { symbol: 1, contractAddress: 1, decimals: 1, makeFee: 1, takeFee: 1 }
      )
      .toArray();

    quotes.forEach(quote => {
      tokens.forEach(token => {
        pairs.push({
          baseTokenSymbol: token.symbol,
          baseTokenAddress: utils.getAddress(token.contractAddress),
          baseTokenDecimals: token.decimals,
          quoteTokenSymbol: quote.symbol,
          quoteTokenAddress: utils.getAddress(quote.contractAddress),
          quoteTokenDecimals: quote.decimals,
          priceMultiplier: getPriceMultiplier(
            token.decimals,
            quote.decimals
          ).toString(),
          active: true,
          makeFee: quote.makeFee,
          takeFee: quote.takeFee,
          createdAt: new Date(faker.fake('{{date.recent}}'))
        });
      });
    });

    console.log(pairs);

    const response = await db.collection('pairs').insertMany(pairs);
    console.log(response);
  } catch (e) {
    console.log(e.message);
  } finally {
    client.close();
  }
};

seed();
