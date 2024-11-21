package strategies

import (
	"crypto/rand"
	"math/big"

	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

const transferGasUsed = int64(21_000)

type TransferStressHandler struct {
	cfg    *config.ReplayorConfig
	signer types.Signer
	logger log.Logger
}

func NewTransferStressHandler(cfg *config.ReplayorConfig, logger log.Logger) *TransferStressHandler {
	return &TransferStressHandler{
		cfg:    cfg,
		signer: types.NewLondonSigner(cfg.ChainId),
		logger: logger,
	}
}

func (h *TransferStressHandler) PrepareDeployTx(input *types.Block) (*types.Transaction, error) {
	// do nothing
	return nil, nil
}

//func (h *TransferStressHandler) PackUp(input *types.Block) types.Transactions {
//	gasInfo := input.Transactions()[len(input.Transactions())-1]
//
//	originalGasUsed := input.GasUsed()
//	targetUsage := h.cfg.GasTarget
//
//	fillUpRemaining := int64(targetUsage)
//	h.logger.Info("gas used", "blockNum", input.NumberU64(), "original", originalGasUsed, "target", targetUsage)
//
//	result := types.Transactions{}
//
//	if fillUpRemaining <= 0 {
//		return result
//	}
//
//	for {
//		if fillUpRemaining < transferGasUsed {
//			break
//		}
//
//		from, err := rand.Int(rand.Reader, length)
//		if err != nil {
//			h.logger.Error("failed to find random int", "err", err)
//			continue
//		}
//
//		to, err := rand.Int(rand.Reader, length)
//		if err != nil {
//			h.logger.Error("failed to find random int", "err", err)
//			continue
//		}
//
//		if from.Cmp(to) == 0 {
//			continue
//		}
//
//		fillUpRemaining -= transferGasUsed
//
//		maxFeePerGas := gasInfo.GasFeeCap()
//		oneHundredTen := big.NewInt(150)
//		maxFeePerGas.Mul(maxFeePerGas, oneHundredTen)
//		maxFeePerGas.Div(maxFeePerGas, big.NewInt(100))
//
//		nonceMu.Lock()
//		txn := types.NewTx(&types.DynamicFeeTx{
//			To:        &addresses[to.Int64()],
//			Nonce:     nonces[from.Int64()],
//			Value:     big.NewInt(1),
//			Gas:       uint64(transferGasUsed),
//			GasTipCap: gasInfo.GasTipCap(),
//			GasFeeCap: big.NewInt(999_999_999_999),
//		})
//		nonceMu.Unlock()
//
//		signer := types.NewLondonSigner(h.cfg.ChainId)
//		signedTx, err := types.SignTx(txn, signer, privateKeys[from.Int64()])
//		if err != nil {
//			h.logger.Error("failed to sign tx", "err", err)
//			continue
//		}
//
//		result = append(result, signedTx)
//		nonceMu.Lock()
//		nonces[from.Int64()]++
//		nonceMu.Unlock()
//	}
//	h.logger.Info("completed packing block", "blockNum", input.NumberU64(), "original", originalGasUsed, "target", targetUsage)
//
//	return result
//}

func (h *TransferStressHandler) PackUp(input *types.Block) types.Transactions {
	gasInfo := input.Transactions()[len(input.Transactions())-1]
	fillUpRemaining := int64(h.cfg.GasTarget) - int64(input.GasUsed())
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
			nonces[0]++
			nonceMu.Unlock()
		} else {
			h.logger.Error("failed to sign tx", "err", err)
		}
	}

	h.logger.Info("completed packing block", "blockNum", input.NumberU64(), "original", input.GasUsed(), "target", h.cfg.GasTarget)
	return result
}

func (h *TransferStressHandler) getRandomAddresses() (int64, int64) {
	from, _ := rand.Int(rand.Reader, length)
	to, _ := rand.Int(rand.Reader, length)
	return from.Int64(), to.Int64()
}

func (h *TransferStressHandler) createTransaction(from, to int64, gasInfo *types.Transaction) *types.Transaction {
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
