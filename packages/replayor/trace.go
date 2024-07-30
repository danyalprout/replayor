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

func (r *Benchmark) computeTraceStats(ctx context.Context, s *stats.BlockCreationStats, receipts []*types.Receipt) {
	s.OpCodes = make(map[string]stats.OpCodeStats)

	for _, receipt := range receipts {
		r.traceReceipt(ctx, receipt, s.OpCodes)
	}
}

func (r *Benchmark) traceReceipt(ctx context.Context, receipt *types.Receipt, opCodes map[string]stats.OpCodeStats) {
	txTrace, err := retry.Do(ctx, 10, retry.Exponential(), func() (*TxTrace, error) {
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

	var gasUsage uint64
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

		gasUsage += opGas
		gasRefund = log.Refund
	}

	opCodes["REFUND"] = stats.OpCodeStats{
		Count: opCodes["REFUND"].Count + 1,
		Gas:   opCodes["REFUND"].Gas + gasRefund,
	}

	gasUsage -= gasRefund

	opCodes["TRANSACTION"] = stats.OpCodeStats{
		Count: opCodes["TRANSACTION"].Count + 1,
		Gas:   opCodes["TRANSACTION"].Gas + receipt.GasUsed - gasUsage,
	}
}
