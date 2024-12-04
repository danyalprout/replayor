package strategies

import (
	"crypto/rand"
	"math/big"

	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

const transferGasUsed = uint64(21_000)

type TransferStressPacker struct {
	cfg    *config.ReplayorConfig
	signer types.Signer
	logger log.Logger
}

func NewTransferStressPacker(cfg *config.ReplayorConfig, logger log.Logger) *TransferStressPacker {
	return &TransferStressPacker{
		cfg:    cfg,
		signer: types.NewLondonSigner(cfg.ChainId),
		logger: logger,
	}
}

func (h *TransferStressPacker) PrepareDeployTx(input *types.Block) (*types.Transaction, error) {
	// do nothing
	return nil, nil
}

func (h *TransferStressPacker) PackUp(input *types.Block) types.Transactions {
	gasInfo := input.Transactions()[len(input.Transactions())-1]
	fillUpRemaining := h.cfg.GasTarget
	result := types.Transactions{}

	h.logger.Info("gas used", "blockNum", input.NumberU64(), "original", input.GasUsed(), "target", h.cfg.GasTarget)

	for fillUpRemaining >= transferGasUsed {
		from, to := h.getRandomAddresses()
		if from == to {
			continue
		}

		txn := h.createTransaction(from, to, gasInfo)
		if signedTx, err := types.SignTx(txn, h.signer, privateKeys[0]); err == nil {
			result = append(result, signedTx)
			fillUpRemaining -= transferGasUsed
			nonceMu.Lock()
			nonces[from] += 1
			nonceMu.Unlock()
		} else {
			h.logger.Error("failed to sign tx", "err", err)
		}
	}

	h.logger.Info("completed packing block", "blockNum", input.NumberU64(), "original", input.GasUsed(), "target", h.cfg.GasTarget)
	return result
}

func (h *TransferStressPacker) getRandomAddresses() (int64, int64) {
	from, _ := rand.Int(rand.Reader, length)
	to, _ := rand.Int(rand.Reader, length)
	return from.Int64(), to.Int64()
}

func (h *TransferStressPacker) createTransaction(from, to int64, gasInfo *types.Transaction) *types.Transaction {
	nonceMu.Lock()
	defer nonceMu.Unlock()
	return types.NewTx(&types.DynamicFeeTx{
		To:        &addresses[to],
		Nonce:     nonces[from],
		Value:     big.NewInt(1),
		Gas:       uint64(transferGasUsed),
		GasTipCap: gasInfo.GasTipCap(),
		GasFeeCap: big.NewInt(999_999_999_999),
	})
}
