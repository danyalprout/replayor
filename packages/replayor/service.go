package replayor

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/danyalprout/replayor/packages/clients"
	"github.com/danyalprout/replayor/packages/config"
	"github.com/danyalprout/replayor/packages/stats"
	"github.com/danyalprout/replayor/packages/strategies"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum/log"
)

var ErrAlreadyStopped = errors.New("already stopped")

type Service struct {
	stopped         atomic.Bool
	benchmarkCancel context.CancelFunc
	clients         clients.Clients
	stats           stats.Stats
	cfg             config.ReplayorConfig
	log             log.Logger
	close           context.CancelCauseFunc
}

func NewService(c clients.Clients, s stats.Stats, cfg config.ReplayorConfig, l log.Logger, close context.CancelCauseFunc) *Service {
	return &Service{
		clients: c,
		stats:   s,
		cfg:     cfg,
		log:     l,
		close:   close,
	}
}

func (r *Service) Start(ctx context.Context) error {
	r.log.Info("starting replayor service")
	cCtx, cancelFunc := context.WithCancel(ctx)
	r.benchmarkCancel = cancelFunc

	currentBlock, err := r.clients.DestNode.BlockByNumber(ctx, nil)
	if err != nil {
		panic(err)
	}
	r.log.Info("retrieved current block number", "blockNum", currentBlock.Number())

	retry.Do(ctx, 720, retry.Fixed(10*time.Second), func() (bool, error) {
		result, err := r.clients.EngineApi.ForkchoiceUpdate(ctx, &eth.ForkchoiceState{
			HeadBlockHash:      currentBlock.Hash(),
			SafeBlockHash:      currentBlock.Hash(),
			FinalizedBlockHash: currentBlock.Hash(),
		}, nil)
		if err != nil {
			r.log.Info("waiting for engine API to stop syncing", "err", err)
			return false, err
		} else if result.PayloadStatus.Status != eth.ExecutionValid {
			r.log.Info("waiting for execution API to stop syncing", "status", result.PayloadStatus.Status)
			return false, errors.New("syncing")
		}
		return true, nil
	})

	if r.cfg.BenchmarkStartBlock != 0 {
		if currentBlock.NumberU64() < r.cfg.BenchmarkStartBlock {
			walkUpToBlock := NewBenchmark(
				r.clients,
				r.cfg.RollupConfig,
				r.log,
				&strategies.OneForOne{},
				&stats.NoOpStats{},
				currentBlock,
				r.cfg.BenchmarkStartBlock-currentBlock.NumberU64(),
				false,
				false)

			walkUpToBlock.Run(cCtx)
		} else {
			panic("current block is greater than benchmark start block")
		}
	}

	currentBlock, err = r.clients.DestNode.BlockByNumber(ctx, nil)
	if err != nil {
		panic(err)
	}

	if r.cfg.BenchmarkStartBlock != 0 && r.cfg.BenchmarkStartBlock != currentBlock.NumberU64() {
		r.log.Error("current block is not equal to benchmark start block", "current_block", currentBlock.NumberU64(), "benchmark_start_block", r.cfg.BenchmarkStartBlock)
		return errors.New("current block is not equal to benchmark start block")
	}

	r.log.Info("Starting benchmark", "start_block", currentBlock.NumberU64())
	strategy := strategies.LoadStrategy(r.cfg, r.log, r.clients, currentBlock)
	if strategy == nil {
		panic("failed to load strategy")
	}

	benchmark := NewBenchmark(r.clients, r.cfg.RollupConfig, r.log, strategy, r.stats, currentBlock, uint64(r.cfg.BlockCount), r.cfg.BenchmarkOpcodes, r.cfg.ComputeStorageDiffs)
	benchmark.Run(cCtx)

	// shutdown the whole application
	r.close(nil)
	return nil
}

func (r *Service) Stop(ctx context.Context) error {
	if r.stopped.Load() {
		return ErrAlreadyStopped
	}

	r.stopped.Store(true)
	r.benchmarkCancel()

	return nil
}

func (r *Service) Stopped() bool {
	return r.stopped.Load()
}
