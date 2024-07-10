package replayor

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/danyalprout/replayor/packages/clients"
	"github.com/danyalprout/replayor/packages/config"
	"github.com/danyalprout/replayor/packages/stats"
	"github.com/danyalprout/replayor/packages/strategies"
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
}

func NewService(c clients.Clients, s stats.Stats, cfg config.ReplayorConfig, l log.Logger) *Service {
	return &Service{
		clients: c,
		stats:   s,
		cfg:     cfg,
		log:     l,
	}
}

func (r *Service) Start(ctx context.Context) error {
	cCtx, cancelFunc := context.WithCancel(ctx)
	r.benchmarkCancel = cancelFunc

	currentBlock, err := r.clients.DestNode.BlockByNumber(ctx, nil)
	if err != nil {
		return err
	}

	if r.cfg.BenchmarkStartBlock != 0 {
		if currentBlock.NumberU64() < r.cfg.BenchmarkStartBlock {
			walkUpToBlock := NewBenchmark(
				r.clients,
				r.cfg.RollupConfig,
				r.log,
				&strategies.OneForOne{},
				&stats.NoOpStats{},
				currentBlock,
				r.cfg.BenchmarkStartBlock-currentBlock.NumberU64()-1)

			walkUpToBlock.Run(cCtx)
		} else {
			return errors.New("current block is greater than benchmark start block")
		}
	}

	currentBlock, err = r.clients.DestNode.BlockByNumber(ctx, nil)
	if err != nil {
		return err
	}

	if r.cfg.BenchmarkStartBlock != 0 && r.cfg.BenchmarkStartBlock != currentBlock.NumberU64() {
		r.log.Error("current block is not equal to benchmark start block", "current_block", currentBlock.NumberU64(), "benchmark_start_block", r.cfg.BenchmarkStartBlock)
		return errors.New("current block is not equal to benchmark start block")
	}

	r.log.Info("Starting benchmark", "start_block", currentBlock.NumberU64())
	strategy := strategies.LoadStrategy(r.cfg, r.log, r.clients, currentBlock)
	if strategy == nil {
		return errors.New("invalid strategy")
	}

	benchmark := NewBenchmark(r.clients, r.cfg.RollupConfig, r.log, strategy, r.stats, currentBlock, uint64(r.cfg.BlockCount))
	benchmark.Run(cCtx)

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
