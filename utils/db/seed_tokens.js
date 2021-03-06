const fs = require('fs');
const path = require('path');
const faker = require('faker');
const process = require('process');
const argv = require('yargs').argv;
const utils = require('ethers').utils;
const MongoClient = require('mongodb').MongoClient;
const { getNetworkID, queryToken } = require('./utils/helpers');
const { DB_NAME, mongoUrl, network } = require('./utils/config');
const networkID = getNetworkID(network);

const truffleBuildPath = path.join(
  `${process.env.TOMO_DEX_PATH}`,
  `/build/contracts`
);
const {
  quoteTokens,
  baseTokens,
  contractAddresses,
  decimals
} = require('./utils/config');

// console.log(quoteTokens, baseTokens, decimals);

let documents = [];
let addresses = contractAddresses[networkID];
let client, db, response;

const seed = async () => {
  try {
    client = await MongoClient.connect(
      mongoUrl,
      { useNewUrlParser: true }
    );
    console.log('Seeding tokens');
    db = client.db(DB_NAME);

    documents = baseTokens.map(symbol => ({
      symbol: symbol,
      contractAddress: utils.getAddress(addresses[symbol]),
      decimals: decimals[symbol],
      quote: false,
      createdAt: new Date(faker.fake('{{date.recent}}'))
    }));

    response = await db.collection('tokens').insertMany(documents);
    console.log(response);
    client.close();
  } catch (e) {
    throw new Error(e.message);
  } finally {
    client.close();
  }
};

seed();
