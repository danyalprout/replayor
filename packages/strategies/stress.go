package strategies

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"

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
	length = big.NewInt(10) // Generate number between 0 and 9

	// Test private keys to make some fake txns from
	privateKeyRaw = []string{
		"80c67f24f604d6dfd5c473781ef9b8680bdc2fbfa73a3370a7ade44a9225e965", // 0x1282BEd9E883442a0119c575D0Dfef608929fB40
		"097a1ba66d00d218dc09d9f23d9b41a5a7400c032472b78246aa47485ed7add8", // 0x24e998fF764D4938010FdC63D2c7F44b1B3B69d1
		"8df8c356a846c7bd2d93d9364f05d984aec76823a46b8addf5b783a498625855", // 0xDEa3a7Ac8a26da5dd464Fb87d845062b05881648
		"2f8e7f4f18044a956103dc285f1ac5691f9b487f9847e493d290b303c8cf105d", // 0x43463b7579113C90e32b0781Eba9A8C1983c7D0B
		"fa99ac6eea2366efe47557eaea12ea5c813050dd4b4f7761e59db2b33a692132", // 0x3879e21DBf8cd4F408D337893335b4C7f093d326
		"a0d9d7f16cd95808963ade84c95962aeddae45cfc89f0dce68d1e21a3334c74c", // 0xCc76B0A858779510F65a51170Fc693c0F8965244
		"91df0afc2cf88d349e5acd7d20da3472e31480e74bb16c80f0019b3a0ef1d5bf", // 0x4A020A989f5057651ab0EE28c9b132e586234be6
		"36440f03f2f8378a416c154b23971cf87d129d3032b2d8584389e42ef5017991", // 0xEC89f08d40e34C876AF7DF398a0389a2e0De3294
		"8ddc54aa97c6c6c40bc0829855751ad7ac5c2082d18f9c9a61ec7d463dae37d3", // 0x3D0e99b64E7a9CCA57eec8829670Af4310eE313E
		"cd721e1ea2d6394a76be87454d617af095583b6858b5b211f587a7f557db2b00", // 0xF49C6A2F225DcbDe512F4C78c0bC40D24947966D
	}

	privateKey = []*ecdsa.PrivateKey{}

	addresses = []common.Address{
		common.HexToAddress("0x1282BEd9E883442a0119c575D0Dfef608929fB40"),
		common.HexToAddress("0x24e998fF764D4938010FdC63D2c7F44b1B3B69d1"),
		common.HexToAddress("0xDEa3a7Ac8a26da5dd464Fb87d845062b05881648"),
		common.HexToAddress("0x43463b7579113C90e32b0781Eba9A8C1983c7D0B"),
		common.HexToAddress("0x3879e21DBf8cd4F408D337893335b4C7f093d326"),
		common.HexToAddress("0xCc76B0A858779510F65a51170Fc693c0F8965244"),
		common.HexToAddress("0x4A020A989f5057651ab0EE28c9b132e586234be6"),
		common.HexToAddress("0xEC89f08d40e34C876AF7DF398a0389a2e0De3294"),
		common.HexToAddress("0x3D0e99b64E7a9CCA57eec8829670Af4310eE313E"),
		common.HexToAddress("0xF49C6A2F225DcbDe512F4C78c0bC40D24947966D"),
	}

	nonces = []uint64{}
)

func init() {
	if len(privateKeyRaw) != len(addresses) {
		panic("mismatched privateKeyRaw and addresses")
	}

	for i, key := range privateKeyRaw {
		privKey, err := crypto.HexToECDSA(key)
		if err != nil {
			panic(err)
		}

		privateKey = append(privateKey, privKey)

		addr := crypto.PubkeyToAddress(privKey.PublicKey)
		if addr != addresses[i] {
			panic("mismatched address")
		}
	}

	nonces = make([]uint64, len(addresses))
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

func (s *StressTest) BlockReceived(ctx context.Context, input *types.Block) *BlockCreationParams {
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

func (s *StressTest) ValidateExecution(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error {
	// Skip: Validate block covers everything
	return nil
}

func (s *StressTest) ValidateBlock(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error {
	// Todo: Validate enough of the txns were included
	return nil
}

func (s *StressTest) packItUp(input *types.Block) types.Transactions {
	gasInfo := input.Transactions()[len(input.Transactions())-1]

	originTransactionUse := input.GasUsed()
	targetUsage := originTransactionUse * 2

	fillUp := targetUsage - originTransactionUse

	result := types.Transactions{}

	if fillUp <= 0 {
		return result
	}

	fillUp = uint64(float64(fillUp) * 0.75)

	for {
		gasUsed := uint64(21_000)
		if fillUp < gasUsed {
			break
		}

		from, err := rand.Int(rand.Reader, length)
		if err != nil {
			panic(err)
		}

		to, err := rand.Int(rand.Reader, length)
		if err != nil {
			panic(err)
		}

		if from.Cmp(to) == 0 {
			continue
		}

		fillUp -= gasUsed

		txn := types.NewTx(&types.DynamicFeeTx{
			To:        &addresses[to.Int64()],
			Nonce:     nonces[from.Int64()],
			Value:     big.NewInt(100),
			Gas:       gasUsed,
			GasTipCap: gasInfo.GasTipCap(),
			GasFeeCap: gasInfo.GasFeeCap(),
		})

		signer := types.NewLondonSigner(big.NewInt(8453))
		signedTx, err := types.SignTx(txn, signer, privateKey[from.Int64()])
		if err != nil {
			panic(err)
		}

		result = append(result, signedTx)
		nonces[from.Int64()]++
	}

	return result
}
