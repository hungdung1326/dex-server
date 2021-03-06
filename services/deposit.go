package services

import (
	"math/big"
	"strings"

	"github.com/tomochain/dex-server/errors"
	"github.com/tomochain/dex-server/ws"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tomochain/dex-server/ethereum"
	"github.com/tomochain/dex-server/interfaces"
	"github.com/tomochain/dex-server/rabbitmq"
	"github.com/tomochain/dex-server/swap"
	swapConfig "github.com/tomochain/dex-server/swap/config"
	"github.com/tomochain/dex-server/types"
	"github.com/tomochain/dex-server/utils/binance"
	"github.com/tomochain/dex-server/utils/math"
	"gopkg.in/mgo.v2/bson"
)

// need to refractor using interface.SwappEngine and only expose neccessary methods
type DepositService struct {
	configDao      interfaces.ConfigDao
	associationDao interfaces.AssociationDao
	pairDao        interfaces.PairDao
	orderDao       interfaces.OrderDao
	swapEngine     *swap.Engine
	engine         interfaces.Engine
	broker         *rabbitmq.Connection
}

// NewAddressService returns a new instance of accountService
func NewDepositService(
	configDao interfaces.ConfigDao,
	associationDao interfaces.AssociationDao,
	pairDao interfaces.PairDao,
	orderDao interfaces.OrderDao,
	swapEngine *swap.Engine,
	engine interfaces.Engine,
	broker *rabbitmq.Connection,
) *DepositService {

	depositService := &DepositService{configDao, associationDao, pairDao, orderDao, swapEngine, engine, broker}

	// set storage engine to this service
	swapEngine.SetStorage(depositService)

	swapEngine.SetQueue(depositService)

	// run watching
	swapEngine.Start()

	return depositService
}

func (s *DepositService) EthereumClient() interfaces.EthereumClient {
	provider := s.engine.Provider().(*ethereum.EthereumProvider)
	return provider.Client
}

func (s *DepositService) WethAddress() common.Address {
	provider := s.engine.Provider().(*ethereum.EthereumProvider)
	return provider.Config.WethAddress()
}

func (s *DepositService) SetDelegate(handler interfaces.SwapEngineHandler) {
	// set event handler delegate to this service
	s.swapEngine.SetDelegate(handler)
}

func (s *DepositService) GenerateAddress(chain types.Chain) (common.Address, uint64, error) {

	err := s.configDao.IncrementAddressIndex(chain)
	if err != nil {
		return swapConfig.EmptyAddress, 0, err
	}
	index, err := s.configDao.GetAddressIndex(chain)
	if err != nil {
		return swapConfig.EmptyAddress, 0, err
	}
	logger.Infof("Current index: %d", index)
	address, err := s.swapEngine.EthereumAddressGenerator().Generate(index)
	return address, index, err
}

func (s *DepositService) SignerPublicKey() common.Address {
	return s.swapEngine.SignerPublicKey()
}

func (s *DepositService) GetSchemaVersion() uint64 {
	return s.configDao.GetSchemaVersion()
}

func (s *DepositService) RecoveryTransaction(chain types.Chain, address common.Address) error {
	return nil
}

/***** implement Storage interface ***/
func (s *DepositService) GetBlockToProcess(chain types.Chain) (uint64, error) {
	return s.configDao.GetBlockToProcess(chain)
}

func (s *DepositService) SaveLastProcessedBlock(chain types.Chain, block uint64) error {
	return s.configDao.SaveLastProcessedBlock(chain, block)
}

func (s *DepositService) SaveDepositTransaction(chain types.Chain, sourceAccount common.Address, txEnvelope string) error {
	return s.associationDao.SaveDepositTransaction(chain, sourceAccount, txEnvelope)
}

func (s *DepositService) QueueAdd(transaction *types.DepositTransaction) error {
	err := s.broker.PublishDepositTransaction(transaction)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// QueuePool receives and removes the head of this queue. Returns nil if no elements found.
func (s *DepositService) QueuePool() (<-chan *types.DepositTransaction, error) {
	return s.broker.QueuePoolDepositTransactions()
}

func (s *DepositService) MinimumValueWei() *big.Int {
	return s.swapEngine.MinimumValueWei()
}

func (s *DepositService) MinimumValueSat() int64 {
	return s.swapEngine.MinimumValueSat()
}

func (s *DepositService) GetAssociationByChainAddress(chain types.Chain, userAddress common.Address) (*types.AddressAssociationRecord, error) {
	return s.associationDao.GetAssociationByChainAddress(chain, userAddress)
}

func (s *DepositService) GetAssociationByChainAssociatedAddress(chain types.Chain, associatedAddress common.Address) (*types.AddressAssociationRecord, error) {
	return s.associationDao.GetAssociationByChainAssociatedAddress(chain, associatedAddress)
}

func (s *DepositService) SaveAssociationByChainAddress(chain types.Chain, address common.Address, index uint64, associatedAddress common.Address, pairAddresses *types.PairAddresses) error {

	association := &types.AddressAssociationRecord{
		ID:                bson.NewObjectId(),
		Chain:             chain,
		Address:           address.Hex(),
		AddressIndex:      index,
		Status:            types.PENDING,
		AssociatedAddress: associatedAddress.Hex(),
		PairName:          pairAddresses.Name,
		BaseTokenAddress:  pairAddresses.BaseToken.Hex(),
		QuoteTokenAddress: pairAddresses.QuoteToken.Hex(),
	}

	return s.associationDao.SaveAssociation(association)
}

func (s *DepositService) SaveAssociationStatusByChainAddress(addressAssociation *types.AddressAssociationRecord, status string) error {

	if addressAssociation == nil {
		return errors.New("AddressAssociationRecord is nil")
	}

	userAddress := common.HexToAddress(addressAssociation.AssociatedAddress)
	address := common.HexToAddress(addressAssociation.Address)

	// send message to channel deposit to noti the status, should limit the txEnvelope < 100
	// if not it would be very slow
	if status == types.SUCCESS {
		ws.SendDepositMessage(types.SUCCESS_EVENT, userAddress, addressAssociation)
	} else if status != types.PENDING {
		// just pending and return the status
		ws.SendDepositMessage(types.PENDING, userAddress, status)
	}

	return s.associationDao.SaveAssociationStatus(addressAssociation.Chain, address, status)
}

func (s *DepositService) getTokenAmountFromOracle(baseTokenSymbol, quoteTokenSymbol string, quoteAmount *big.Int) (*big.Int, error) {
	lastPrice, err := binance.GetLastPrice(baseTokenSymbol, quoteTokenSymbol)
	if err != nil {
		return quoteAmount, nil
	}
	exchangeRate, ok := new(big.Int).SetString(lastPrice, 10)
	if !ok {
		return quoteAmount, nil
	}
	// last price is the price in quoteToken for a baseToken, also means baseToken/quoteToken exchange rate
	baseAmount := math.Mul(quoteAmount, exchangeRate)
	return baseAmount, nil
}

func (s *DepositService) GetBaseTokenAmount(pairName string, quoteAmount *big.Int) (*big.Int, error) {

	tokenSymbols := strings.Split(pairName, "/")
	if len(tokenSymbols) != 2 {
		return nil, errors.Errorf("Pair name is wrong format: %s", pairName)
	}
	baseTokenSymbol := strings.ToUpper(tokenSymbols[0])
	quoteTokenSymbol := strings.ToUpper(tokenSymbols[1])

	// this is 1:1 exchange
	if baseTokenSymbol == quoteTokenSymbol {
		return quoteAmount, nil
	}

	pair, err := s.pairDao.GetByTokenSymbols(baseTokenSymbol, quoteTokenSymbol)
	if err != nil {
		return nil, err
	}

	if pair == nil {
		// there is no exchange rate yet
		return s.getTokenAmountFromOracle(baseTokenSymbol, quoteTokenSymbol, quoteAmount)
	}

	logger.Debugf("Got pair :%v", pair)

	// get best Bid, the highest bid available
	bids, err := s.orderDao.GetSideOrderBook(pair, types.BUY, -1, 1)
	if err != nil {
		return nil, err
	}

	// if there is no exchange rate, should return one from oracle service like coin market cap
	if len(bids) < 1 {
		return s.getTokenAmountFromOracle(baseTokenSymbol, quoteTokenSymbol, quoteAmount)
	}

	pricepoint := math.ToBigInt(bids[0]["pricepoint"])
	if math.IsZero(pricepoint) {
		return nil, errors.New("Pricepoint is zero")
	}

	// calculate the tokenAmount
	tokenAmount := math.Div(math.Mul(quoteAmount, pair.PricepointMultiplier()), pricepoint)

	return tokenAmount, nil
}

// Create function performs the DB insertion task for Balance collection
func (s *DepositService) GetAssociationByUserAddress(chain types.Chain, userAddress common.Address) (*types.AddressAssociation, error) {
	// get from feed
	var addressAssociationFeed types.AddressAssociationFeed
	err := s.engine.GetFeed(userAddress, chain.Bytes(), &addressAssociationFeed)

	logger.Infof("feed :%v", addressAssociationFeed)

	if err == nil {
		return addressAssociationFeed.GetJSON()
	}
	return nil, err
}
