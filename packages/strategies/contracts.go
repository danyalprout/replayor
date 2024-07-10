package strategies

import (
	"context"
	"math/big"
	"math/rand"

	"github.com/danyalprout/replayor/packages/clients"
	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type ContractStressTest struct {
	startBlock *types.Block
	buffer     []*types.Block
	logger     log.Logger
	cfg        config.ReplayorConfig
	clients    clients.Clients
}

type ContractSegment struct {
	bytes []byte
	gas   uint64
}

var (
	headSegment = ContractSegment{
		bytes: []byte{
			0x60, 0x80, // push1 0x80
		},
		gas: 3,
	}

	tailSegment = ContractSegment{
		bytes: []byte{
			0x50, // pop
		},
		gas: 2,
	}

	addSegment = ContractSegment{
		bytes: []byte{
			0x60, 0x40, // push1 0x40
			0x01, // add
		},
		gas: 6,
	}

	subSegment = ContractSegment{
		bytes: []byte{
			0x60, 0x40, // push1 0x40
			0x03, // sub
		},
		gas: 6,
	}

	mulSegment = ContractSegment{
		bytes: []byte{
			0x60, 0x02, // push1 0x40
			0x02, // mul
		},
		gas: 8,
	}

	divSegment = ContractSegment{
		bytes: []byte{
			0x60, 0x02, // push1 0x40
			0x04, // div
		},
		gas: 8,
	}

	sloadSegment = ContractSegment{
		bytes: []byte{
			0x54, // sload
		},
		gas: 2100,
	}

	sstoreSegment = ContractSegment{
		bytes: []byte{
			0x55, // sstore
		},
		gas: 5000,
	}

	bodySegments = []ContractSegment{
		addSegment,
		subSegment,
		mulSegment,
		divSegment,
		sloadSegment,
		sstoreSegment,
	}
)

func NewContractStressTest(startBlock *types.Block, logger log.Logger, cfg config.ReplayorConfig, c clients.Clients) Strategy {
	return &ContractStressTest{
		startBlock: startBlock,
		buffer:     []*types.Block{},
		logger:     logger,
		cfg:        cfg,
		clients:    c,
	}
}

func (s *ContractStressTest) modifyTransactions(input *types.Block, transactions types.Transactions) types.Transactions {
	depositTxn := transactions[0]
	l1Info, err := derive.L1BlockInfoFromBytes(s.cfg.RollupConfig, input.Time(), depositTxn.Data())
	if err != nil {
		panic(err)
	}

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
			panic(err)
		}

		dep.SourceHash = source.SourceHash()
		dep.From = addr
		dep.To = &addr
		dep.Mint = val
		dep.Value = val
		dep.Gas = 21_000
		dep.IsSystemTransaction = false

		depositTxns = append(depositTxns, types.NewTx(&dep))

		nonces[i] += 1
	} else {
		userTxns = append(userTxns, s.packItUp(input)...)
	}

	result = append(result, depositTxns...)
	result = append(result, userTxns...)
	return result
}

func (s *ContractStressTest) BlockReceived(ctx context.Context, input *types.Block) *BlockCreationParams {
	gl := eth.Uint64Quantity(s.cfg.GasLimit)

	txns := s.modifyTransactions(input, input.Transactions())

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

func (s *ContractStressTest) ValidateExecution(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error {
	// Skip: Validate block covers everything
	return nil
}

func (s *ContractStressTest) ValidateBlock(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error {
	// Todo: Validate enough of the txns were included
	return nil
}

func (s *ContractStressTest) packItUp(input *types.Block) types.Transactions {
	gasInfo := input.Transactions()[len(input.Transactions())-1]

	originTransactionUse := input.GasUsed()
	targetUsage := input.GasLimit()

	fillUp := targetUsage - originTransactionUse

	result := types.Transactions{}

	if fillUp <= 0 {
		return result
	}

	fillUp = uint64(float64(fillUp) * 0.75)

	for {
		data, gasUsed := generateContractData()
		if fillUp < gasUsed {
			break
		}

		from := rand.Intn(length)

		fillUp -= gasUsed

		txn := types.NewTx(&types.DynamicFeeTx{
			Nonce:     nonces[from],
			Value:     big.NewInt(100),
			Gas:       gasUsed,
			GasTipCap: gasInfo.GasTipCap(),
			GasFeeCap: gasInfo.GasFeeCap(),
			Data:      data,
		})

		signer := types.NewLondonSigner(big.NewInt(8453))
		signedTx, err := types.SignTx(txn, signer, privateKey[from])
		if err != nil {
			panic(err)
		}

		result = append(result, signedTx)
		nonces[from]++
	}

	return result
}

func generateContractData() ([]byte, uint64) {
	gasEstimate := uint64(21_000)
	data := append([]byte{}, headSegment.bytes...)
	gasEstimate += headSegment.gas + uint64(16*len(headSegment.bytes))

	numBodySegments := rand.Intn(500)
	for i := 0; i < numBodySegments; i++ {
		segment := bodySegments[rand.Intn(len(bodySegments))]
		data = append(data, segment.bytes...)
		gasEstimate += segment.gas + uint64(16*len(segment.bytes))
	}

	data = append(data, tailSegment.bytes...)
	gasEstimate += tailSegment.gas + uint64(16*len(tailSegment.bytes))

	return data, gasEstimate
}
