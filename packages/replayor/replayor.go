package replayor

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/danyalprout/replayor/packages/clients"
	"github.com/danyalprout/replayor/packages/config"
	"github.com/danyalprout/replayor/packages/storage"
	"github.com/danyalprout/replayor/packages/strategies"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.design/x/chann"
)

var ErrAlreadyStopped = errors.New("already stopped")

type Exec struct {
	stopped atomic.Bool
	clients clients.Clients

	log log.Logger
	cfg config.ReplayorConfig

	startBlock *types.Block

	incomingBlocks chan *types.Block
	processBlocks  chan strategies.BlockCreationParams
	recordStats    *chann.Chann[storage.BlockCreationStats]

	previousReplayedBlockHash common.Hash
	strategy                  strategies.Strategy
	storage                   storage.Storage

	stats    []storage.BlockCreationStats
	sm       sync.Mutex
	addCount int
}

func (r *Exec) getBlockFromSourceNode(ctx context.Context, blockNum uint64) (*types.Block, error) {
	return retry.Do(ctx, 10, retry.Exponential(), func() (*types.Block, error) {
		return r.clients.SourceNode.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
	})
}

func (r *Exec) loadBlocks(ctx context.Context) {
	startNum := r.startBlock.NumberU64()

	concurrency := uint64(25)

	endBlock := r.startBlock.NumberU64() + uint64(r.cfg.BlockCount)
	for blockStartRange := startNum + 1; blockStartRange <= endBlock; blockStartRange += concurrency {
		results := make([]*types.Block, concurrency)

		var wg sync.WaitGroup
		wg.Add(int(concurrency))

		var m sync.Mutex

		for i := uint64(0); i < concurrency; i++ {
			blockNum := blockStartRange + i

			go func(index uint64) {
				defer wg.Done()

				block, err := r.getBlockFromSourceNode(ctx, blockNum)
				if err != nil {
					panic(err)
				}

				if block == nil {
					panic(err)
				}

				m.Lock()
				results[index] = block
				m.Unlock()
			}(i)
		}

		wg.Wait()

		for _, block := range results {
			r.incomingBlocks <- block
		}
	}

	r.log.Info("finished loading blocks, closing channel")
	close(r.incomingBlocks)
}

func (r *Exec) addBlock(ctx context.Context, currentBlock strategies.BlockCreationParams) {
	l := r.log.New("source", "add-block", "block", currentBlock.Number)

	stats := storage.BlockCreationStats{}

	txns := currentBlock.Transactions

	stats.TxnCount = len(txns)

	state := &eth.ForkchoiceState{
		HeadBlockHash:      r.previousReplayedBlockHash,
		SafeBlockHash:      r.previousReplayedBlockHash,
		FinalizedBlockHash: r.previousReplayedBlockHash,
	}

	txnData := make([]eth.Data, len(txns))
	for i, txn := range txns {
		data, err := txn.MarshalBinary()
		if err != nil {
			panic(err)
		}
		txnData[i] = data
		stats.GasLimit += txn.Gas()
	}

	attrs := &eth.PayloadAttributes{
		Timestamp:             currentBlock.Time,
		NoTxPool:              true,
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
		Transactions:          txnData,
		GasLimit:              currentBlock.GasLimit,
		PrevRandao:            currentBlock.MixDigest,
		ParentBeaconBlockRoot: currentBlock.BeaconRoot,
	}

	if r.cfg.RollupConfig.IsCanyon(uint64(currentBlock.Time)) {
		attrs.Withdrawals = &types.Withdrawals{}
	}

	startTime := time.Now()

	result, err := r.clients.EngineApi.ForkchoiceUpdate(ctx, state, attrs)

	fcuEnd := time.Now()

	if err != nil {
		panic(err)
	}

	if result.PayloadStatus.Status != eth.ExecutionValid {
		panic(err)
	}

	stats.FCUTime = fcuEnd.Sub(startTime)

	envelope, err := r.clients.EngineApi.GetPayload(ctx, eth.PayloadInfo{
		ID:        *result.PayloadID,
		Timestamp: uint64(currentBlock.Time),
	})

	if err != nil {
		panic(err)
	}

	getTime := time.Now()

	stats.GetTime = getTime.Sub(fcuEnd)

	err = r.strategy.ValidateExecution(ctx, envelope, currentBlock)
	if err != nil {
		l.Error("validation failed", "err", err)
		panic(err)
	}

	status, err := r.clients.EngineApi.NewPayload(ctx, envelope.ExecutionPayload, envelope.ParentBeaconBlockRoot)
	if err != nil {
		panic(err)
	}

	newEnd := time.Now()
	stats.NewTime = newEnd.Sub(getTime)

	if status.Status != eth.ExecutionValid {
		panic(err)
	}

	state = &eth.ForkchoiceState{
		HeadBlockHash:      envelope.ExecutionPayload.BlockHash,
		SafeBlockHash:      envelope.ExecutionPayload.BlockHash,
		FinalizedBlockHash: envelope.ExecutionPayload.BlockHash,
	}

	status2, err := r.clients.EngineApi.ForkchoiceUpdate(ctx, state, nil)
	if err != nil {
		panic(err)
	}

	if status2.PayloadStatus.Status != eth.ExecutionValid {
		panic(err)
	}

	err = r.strategy.ValidateBlock(ctx, envelope, currentBlock)
	if err != nil {
		l.Error("validation failed", "err", err)
		panic(err)
	}

	fcu2Time := time.Now()

	stats.FCUNoAttrsTime = fcu2Time.Sub(newEnd)
	stats.TotalTime = fcu2Time.Sub(startTime)
	stats.BlockNumber = uint64(envelope.ExecutionPayload.BlockNumber)
	stats.BlockHash = envelope.ExecutionPayload.BlockHash

	r.previousReplayedBlockHash = envelope.ExecutionPayload.BlockHash

	r.recordStats.In() <- stats
}

func (r *Exec) enrich(ctx context.Context, stats *storage.BlockCreationStats) {
	receipts, err := retry.Do(ctx, 10, retry.Exponential(), func() ([]*types.Receipt, error) {
		return r.clients.DestNode.BlockReceipts(ctx, rpc.BlockNumberOrHash{BlockHash: &stats.BlockHash})
	})

	if err != nil {
		r.log.Warn("unable to load receipts", "err", err)
	}

	success := 0

	for _, receipt := range receipts {
		stats.GasUsed += receipt.GasUsed
		if receipt.Status == types.ReceiptStatusSuccessful {
			success += 1
		}
	}

	stats.Success = float64(success) / float64(len(receipts))
}

func (r *Exec) enrichAndRecordStats(ctx context.Context) {
	for i := 0; i < 5; i++ {
		go func() {
			for {
				select {
				case stats := <-r.recordStats.Out():
					r.enrich(ctx, &stats)

					r.sm.Lock()
					r.stats = append(r.stats, stats)
					r.sm.Unlock()

					r.log.Debug("block stats", "BlockNumber", stats.BlockNumber, "BlockHash", stats.BlockHash, "TxnCount", stats.TxnCount, "TotalTime", stats.TotalTime, "FCUTime", stats.FCUTime, "GetTime", stats.GetTime, "NewTime", stats.NewTime, "FCUNoAttrsTime", stats.FCUNoAttrsTime, "Success", stats.Success, "GasUsed", stats.GasUsed, "GasLimit", stats.GasLimit)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (r *Exec) submitBlocks(ctx context.Context) {
	for {
		select {
		case block := <-r.processBlocks:
			r.addBlock(ctx, block)
		case <-ctx.Done():
			return
		}
	}
}

func (r *Exec) RecordStats(ctx context.Context) {
	r.sm.Lock()
	stats := r.stats
	r.sm.Unlock()

	_, err := retry.Do(ctx, 3, retry.Fixed(time.Second), func() (interface{}, error) {
		err := r.storage.Write(ctx, stats)
		if err != nil {
			r.log.Info("error writing to storage", err)
		}
		return nil, err
	})

	if err != nil {
		r.log.Error("error writing to storage", "err", err)
	}
}

func (r *Exec) mapBlocks(ctx context.Context) {
	for {
		select {
		case b := <-r.incomingBlocks:
			if b == nil {
				r.log.Debug("stopping block processing")
				continue
			}

			params := r.strategy.BlockReceived(ctx, b)
			if params == nil {
				continue
			}

			r.processBlocks <- *params
			r.addCount -= 1

			if r.addCount == 0 {
				r.log.Info("finished processing blocks")
				return
			}
		case <-ctx.Done():
			return
		}
	}

}

func (r *Exec) Run(ctx context.Context) {
	go r.loadBlocks(ctx)
	go r.mapBlocks(ctx)
	go r.submitBlocks(ctx)
	r.enrichAndRecordStats(ctx)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	l := r.log.New("source", "monitor")

	lastBlockNum := r.startBlock.NumberU64()

	for {
		select {
		case <-ticker.C:
			currentBlock, err := r.clients.DestNode.BlockByNumber(ctx, nil)
			if err != nil {
				l.Error("unable to load current block", "err", err)
			}

			l.Info("replay progress", "blocks", currentBlock.NumberU64()-lastBlockNum, "incomingBlocks", len(r.incomingBlocks), "processBlocks", len(r.processBlocks), "currentBlock", currentBlock.NumberU64(), "statProgress", r.recordStats.Len())

			lastBlockNum = currentBlock.NumberU64()

			if r.addCount == 0 {
				r.RecordStats(ctx)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func NewExec(c clients.Clients, cfg config.ReplayorConfig, logger log.Logger, strategy strategies.Strategy, s storage.Storage, block *types.Block) *Exec {
	r := &Exec{
		clients:                   c,
		cfg:                       cfg,
		log:                       logger,
		incomingBlocks:            make(chan *types.Block, 25),
		processBlocks:             make(chan strategies.BlockCreationParams, 25),
		recordStats:               chann.New[storage.BlockCreationStats](),
		strategy:                  strategy,
		storage:                   s,
		startBlock:                block,
		previousReplayedBlockHash: block.Hash(),
		addCount:                  cfg.BlockCount,
	}

	return r
}

func (r *Exec) Start(ctx context.Context) error {
	r.Run(ctx)
	return nil
}

func (r *Exec) Stop(ctx context.Context) error {
	if r.stopped.Load() {
		return ErrAlreadyStopped
	}

	r.stopped.Store(true)
	// todo actually stop it

	return nil
}

func (r *Exec) Stopped() bool {
	return r.stopped.Load()
}
