package strategies

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
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
	startBlock *types.Block
	buffer     []*types.Block
	logger     log.Logger
	cfg        config.ReplayorConfig
	clients    clients.Clients
}

func NewStressTest(startBlock *types.Block, logger log.Logger, cfg config.ReplayorConfig, c clients.Clients) Strategy {
	return &StressTest{
		startBlock: startBlock,
		buffer:     []*types.Block{},
		logger:     logger,
		cfg:        cfg,
		clients:    c,
	}
}

func (s *StressTest) modifyTransactions(input *types.Block, transactions types.Transactions) types.Transactions {
	depositTxn := transactions[0]
	l1Info, err := derive.L1BlockInfoFromBytes(s.cfg.RollupConfig, input.Time(), depositTxn.Data())
	if err != nil {
		panic(err)
	}

	originalTxNum := len(transactions)

	result := types.Transactions{}

	depositTxns := types.Transactions{}
	userTxns := types.Transactions{}

	for _, txn := range transactions {
		if txn.IsDepositTx() {
			depositTxns = append(depositTxns, txn)
		} else {
			userTxns = append(userTxns, txn)
		}
	}

	currentBlockNum := input.NumberU64()

	startBlockNum := s.startBlock.NumberU64()
	// If first 10 blocks: Add one deposit transactions to send ETH to accounts
	// Maybe if seqNum=0 (?)
	if currentBlockNum < startBlockNum+11 {
		i := currentBlockNum - startBlockNum - 1
		addr := addresses[i]

		// Add deposit transaction
		var dep types.DepositTx

		source := derive.UserDepositSource{
			L1BlockHash: l1Info.BlockHash, // l1 block info
			LogIndex:    10_000,           // hardcode to a very large number
		}

		val, ok := new(big.Int).SetString("100000000000000000000", 10)
		if !ok {
			panic("big.Int SetString failed")
		}

		dep.SourceHash = source.SourceHash()
		dep.From = addr
		dep.To = &addr
		dep.Mint = val
		dep.Value = val
		dep.Gas = 21_000
		dep.IsSystemTransaction = false

		depositTxns = append(depositTxns, types.NewTx(&dep))

		nonceMu.Lock()
		nonces[i] += 1
		nonceMu.Unlock()
	} else {
		userTxns = s.packItUp(input)
	}

	result = append(result, depositTxns...)
	result = append(result, userTxns...)
	finalTxNum := len(result)

	s.logger.Info("tx count", "blockNum", input.NumberU64(), "original", originalTxNum, "final", finalTxNum)

	return result
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

// Adds new eth transfer txs to the block until the target gas usage is met
func (s *StressTest) packItUp(input *types.Block) types.Transactions {
	gasInfo := input.Transactions()[len(input.Transactions())-1]

	originalGasUsed := input.GasUsed()
	targetUsage := s.cfg.GasTarget

	fillUpRemaining := int64(targetUsage)
	s.logger.Info("gas used", "blockNum", input.NumberU64(), "original", originalGasUsed, "target", targetUsage)

	result := types.Transactions{}

	if fillUpRemaining <= 0 {
		return result
	}

	for {
		gasUsed := int64(21_000)
		if fillUpRemaining < gasUsed {
			break
		}

		from, err := rand.Int(rand.Reader, length)
		if err != nil {
			s.logger.Error("failed to find random int", "err", err)
			continue
		}

		to, err := rand.Int(rand.Reader, length)
		if err != nil {
			s.logger.Error("failed to find random int", "err", err)
			continue
		}

		if from.Cmp(to) == 0 {
			continue
		}

		fillUpRemaining -= gasUsed

		maxFeePerGas := gasInfo.GasFeeCap()
		oneHundredTen := big.NewInt(150)
		maxFeePerGas.Mul(maxFeePerGas, oneHundredTen)
		maxFeePerGas.Div(maxFeePerGas, big.NewInt(100))

		nonceMu.Lock()
		txn := types.NewTx(&types.DynamicFeeTx{
			To:        &addresses[to.Int64()],
			Nonce:     nonces[from.Int64()],
			Value:     big.NewInt(1),
			Gas:       uint64(gasUsed),
			GasTipCap: gasInfo.GasTipCap(),
			GasFeeCap: big.NewInt(999_999_999_999),
		})
		nonceMu.Unlock()

		signer := types.NewLondonSigner(s.cfg.ChainId)
		signedTx, err := types.SignTx(txn, signer, privateKeys[from.Int64()])
		if err != nil {
			s.logger.Error("failed to sign tx", "err", err)
			continue
		}

		result = append(result, signedTx)
		nonceMu.Lock()
		nonces[from.Int64()]++
		nonceMu.Unlock()
	}
	s.logger.Info("completed packing block", "blockNum", input.NumberU64(), "original", originalGasUsed, "target", targetUsage)

	return result
}
