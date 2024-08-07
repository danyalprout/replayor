package replayor

import (
	"context"

	"github.com/danyalprout/replayor/packages/stats"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum/core/types"
)

type TxTrace struct {
	Gas        uint64      `json:"gas"`
	Failed     bool        `json:"failed"`
	StructLogs []StructLog `json:"structLogs"`
}

type StructLog struct {
	Op      string `json:"op"`
	Gas     uint64 `json:"gas"`
	GasCost uint64 `json:"gasCost"`
	Refund  uint64 `json:"refund"`
	Depth   uint64 `json:"depth"`
}

var tracerOptions = map[string]any{
	"disableStack":   true,
	"disableStorage": true,
}

type StorageTrace struct {
	TxHash string `json:"txHash"`
	Result struct {
		Pre  map[string]StorageAccountState `json:"pre"`
		Post map[string]StorageAccountState `json:"post"`
	} `json:"result"`
}

type StorageAccountState struct {
	Balance string            `json:"balance"`
	Nonce   uint64            `json:"nonce"`
	Storage map[string]string `json:"storage"`
}

var storageTracerOptions = map[string]any{
	"tracer": "prestateTracer",
	"tracerConfig": map[string]any{
		"diffMode": true,
	},
}

func (r *Benchmark) computeTraceStats(ctx context.Context, s *stats.BlockCreationStats, receipts []*types.Receipt) {
	if r.benchmarkOpcodes {
		s.OpCodes = make(map[string]stats.OpCodeStats)

		for _, receipt := range receipts {
			r.traceReceipt(ctx, receipt, s.OpCodes)
		}
	}
	r.recordStorageChanges(ctx, s)
}

func (r *Benchmark) traceReceipt(ctx context.Context, receipt *types.Receipt, opCodes map[string]stats.OpCodeStats) {
	tx, err := retry.Do(ctx, 3, retry.Exponential(), func() (*types.Transaction, error) {
		tx, _, err := r.clients.DestNode.TransactionByHash(ctx, receipt.TxHash)
		return tx, err
	})
	if err != nil {
		r.log.Warn("unable to load tx", "err", err)
		opCodes["UNKNOWN"] = stats.OpCodeStats{
			Count: opCodes["UNKNOWN"].Count + 1,
			Gas:   opCodes["UNKNOWN"].Gas + receipt.GasUsed,
		}
		return
	}

	gasLimit := tx.Gas()

	txTrace, err := retry.Do(ctx, 5, retry.Exponential(), func() (*TxTrace, error) {
		var txTrace TxTrace
		err := r.clients.DestNode.Client().Call(&txTrace, "debug_traceTransaction", receipt.TxHash, tracerOptions)
		if err != nil {
			return nil, err
		}
		return &txTrace, nil
	})
	if err != nil {
		r.log.Warn("unable to load tx trace", "err", err)
		opCodes["UNKNOWN"] = stats.OpCodeStats{
			Count: opCodes["UNKNOWN"].Count + 1,
			Gas:   opCodes["UNKNOWN"].Gas + receipt.GasUsed,
		}
		return
	}

	gasUsage := txTrace.Gas

	if len(txTrace.StructLogs) > 0 {
		gasUsage = gasLimit - txTrace.StructLogs[0].Gas
	}

	opCodes["TRANSACTION"] = stats.OpCodeStats{
		Count: opCodes["TRANSACTION"].Count + 1,
		Gas:   opCodes["TRANSACTION"].Gas + gasUsage,
	}

	var gasRefund uint64

	for idx, log := range txTrace.StructLogs {
		opGas := log.GasCost

		var nextLog *StructLog
		if idx < len(txTrace.StructLogs)-1 {
			nextLog = &txTrace.StructLogs[idx+1]
		}
		if nextLog != nil {
			if log.Depth < nextLog.Depth {
				opGas = log.GasCost - nextLog.Gas
			} else if log.Depth == nextLog.Depth {
				opGas = log.Gas - nextLog.Gas
			}
		}

		opCodes[log.Op] = stats.OpCodeStats{
			Count: opCodes[log.Op].Count + 1,
			Gas:   opCodes[log.Op].Gas + opGas,
		}

		gasRefund = log.Refund
	}

	opCodes["REFUND"] = stats.OpCodeStats{
		Count: opCodes["REFUND"].Count + 1,
		Gas:   opCodes["REFUND"].Gas + gasRefund,
	}
}

func (r *Benchmark) recordStorageChanges(ctx context.Context, s *stats.BlockCreationStats) {
	storageTraces, err := retry.Do(ctx, 3, retry.Exponential(), func() ([]StorageTrace, error) {
		var storageTraces []StorageTrace
		err := r.clients.DestNode.Client().Call(&storageTraces, "debug_traceBlockByHash", s.BlockHash, storageTracerOptions)
		if err != nil {
			return nil, err
		}
		return storageTraces, nil
	})
	if err != nil {
		r.log.Warn("unable to load block storage trace", "err", err)
		return
	}

	var mergedTrace StorageTrace
	mergedTrace.Result.Pre = make(map[string]StorageAccountState)
	mergedTrace.Result.Post = make(map[string]StorageAccountState)

	for _, storageTrace := range storageTraces {
		for account, storage := range storageTrace.Result.Pre {
			if _, ok := mergedTrace.Result.Pre[account]; !ok {
				mergedTrace.Result.Pre[account] = StorageAccountState{
					Storage: make(map[string]string),
				}
			}
			for slot, value := range storage.Storage {
				mergedTrace.Result.Pre[account].Storage[slot] = value
			}
		}
		for account, storage := range storageTrace.Result.Post {
			if _, ok := mergedTrace.Result.Post[account]; !ok {
				mergedTrace.Result.Post[account] = StorageAccountState{
					Storage: make(map[string]string),
				}
			}
			for slot, value := range storage.Storage {
				mergedTrace.Result.Post[account].Storage[slot] = value
			}
		}
	}

	s.AccountsChanged = uint64(len(mergedTrace.Result.Post))
	s.StorageTriesChanged = 0
	s.SlotsChanged = 0

	changedStorageTries := make(map[string]bool)

	for address, storage := range mergedTrace.Result.Pre {
		for slot := range storage.Storage {
			// count deleted storage slots
			if _, ok := mergedTrace.Result.Post[address].Storage[slot]; !ok {
				s.SlotsChanged++
				changedStorageTries[address] = true
			}
		}
	}
	for address, storage := range mergedTrace.Result.Post {
		// count added/updated storage slots
		s.SlotsChanged += uint64(len(storage.Storage))
		if len(storage.Storage) > 0 {
			changedStorageTries[address] = true
		}
	}

	s.StorageTriesChanged = uint64(len(changedStorageTries))
}
