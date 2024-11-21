package strategies

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/danyalprout/replayor/packages/clients"
	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

const (
	STRESS_TRANSFER = "transfer"
	STRESS_ERC20    = "erc20"
	STRESS_OPCODE   = "opcode"
)

var (
	numAddresses = 10
	nonces       = []uint64{}
	nonceMu      = sync.Mutex{}
	privateKeys  = []*ecdsa.PrivateKey{}
	addresses    = []common.Address{}
	length       = big.NewInt(int64(numAddresses)) // Generate number between 0 and 9
)

func init() {
	privateKeys = make([]*ecdsa.PrivateKey, numAddresses)
	addresses = make([]common.Address, numAddresses)
	nonces = make([]uint64, numAddresses)

	for i := 0; i < numAddresses; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			panic("failed to generate private key")
		}
		privateKeys[i] = privateKey
		addresses[i] = crypto.PubkeyToAddress(privateKey.PublicKey)
	}
}

type StressTest struct {
	startBlock  *types.Block
	buffer      []*types.Block
	logger      log.Logger
	cfg         config.ReplayorConfig
	clients     clients.Clients
	handler     StressHandler
	prepareOnce sync.Once
}

type StressHandler interface {
	PrepareDeployTx(input *types.Block) (*types.Transaction, error)
	PackUp(input *types.Block) types.Transactions
}

func NewStressTest(startBlock *types.Block, logger log.Logger, cfg config.ReplayorConfig, c clients.Clients) Strategy {
	return &StressTest{
		startBlock: startBlock,
		buffer:     []*types.Block{},
		logger:     logger,
		cfg:        cfg,
		clients:    c,
		handler:    createStressHandler(&cfg, logger),
	}
}

func createStressHandler(cfg *config.ReplayorConfig, logger log.Logger) StressHandler {
	switch cfg.StressConfig.Type {
	case STRESS_TRANSFER:
		return NewTransferStressHandler(cfg, logger)
	case STRESS_ERC20:
		return NewErc20StressHandler(cfg, logger)
	default:
		panic(fmt.Errorf("unknown stress test type: %s=%s", "type", cfg.StressConfig.Type))
		return nil
	}
}

func (s *StressTest) modifyTransactions(input *types.Block, transactions types.Transactions) types.Transactions {
	depositTxn := transactions[0]
	l1Info, err := derive.L1BlockInfoFromBytes(s.cfg.RollupConfig, input.Time(), depositTxn.Data())
	if err != nil {
		panic(err)
	}

	currentBlockNum := input.NumberU64()
	startBlockNum := s.startBlock.NumberU64()

	depositTxns, userTxns := s.splitTxs(transactions)

	if currentBlockNum < startBlockNum+11 {
		depositTxns = append(depositTxns, s.prepareDeposit(l1Info, currentBlockNum))

		s.prepareOnce.Do(func() {
			if prepareTx, err := s.handler.PrepareDeployTx(input); err == nil && prepareTx != nil {
				userTxns = append(types.Transactions{prepareTx}, userTxns...)
			}
		})
	} else {
		userTxns = s.handler.PackUp(input)
	}

	result := append(depositTxns, userTxns...)
	s.logger.Info("tx count", "blockNum", currentBlockNum, "original", len(transactions), "final", len(result))

	return result
}

// Add deposit transaction
func (s *StressTest) prepareDeposit(l1Info *derive.L1BlockInfo, currentBlockNum uint64) *types.Transaction {
	i := currentBlockNum - s.startBlock.NumberU64() - 1
	addr := addresses[i]

	source := derive.UserDepositSource{
		L1BlockHash: l1Info.BlockHash, // l1 block info
		LogIndex:    10_000,           // hardcode to a very large number
	}

	val, ok := new(big.Int).SetString("100000000000000000000", 10)
	if !ok {
		panic("big.Int SetString failed")
	}
	dep := types.DepositTx{
		SourceHash:          source.SourceHash(),
		From:                addr,
		To:                  &addr,
		Mint:                val,
		Value:               val,
		Gas:                 21_000,
		IsSystemTransaction: false,
	}

	nonceMu.Lock()
	nonces[i] += 1
	nonceMu.Unlock()

	return types.NewTx(&dep)
}

func (s *StressTest) randAddrIndex() (from *big.Int, to *big.Int, err error) {
	for {
		from, err = rand.Int(rand.Reader, length)
		if err != nil {
			s.logger.Error("failed to find random int", "err", err)
			return
		}

		to, err = rand.Int(rand.Reader, length)
		if err != nil {
			s.logger.Error("failed to find random int", "err", err)
			return
		}

		if from.Cmp(to) == 0 {
			continue
		} else {
			return
		}
	}
}

func (s *StressTest) splitTxs(txs types.Transactions) (depositTxns types.Transactions, userTxns types.Transactions) {
	for _, txn := range txs {
		if txn.IsDepositTx() {
			depositTxns = append(depositTxns, txn)
		} else {
			userTxns = append(userTxns, txn)
		}
	}
	return
}

func (s *StressTest) BlockReceived(ctx context.Context, input *types.Block) *BlockCreationParams {
	txns := s.modifyTransactions(input, input.Transactions())

	gl := eth.Uint64Quantity(s.cfg.GasLimit)

	return &BlockCreationParams{
		Number:       input.NumberU64(),
		Transactions: txns,
		GasLimit:     &gl,
		Time:         eth.Uint64Quantity(input.Time()),
		MixDigest:    eth.Bytes32(input.MixDigest()),
		BeaconRoot:   input.BeaconRoot(),
		validateInfo: input,
		Withdrawals:  input.Withdrawals(),
	}
}

func (s *StressTest) ValidateExecution(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error {
	// Skip: Validate block covers everything
	return nil
}

func (s *StressTest) ValidateBlock(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error {
	// Todo: Validate enough of the txns were included
	return nil
}
