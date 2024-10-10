package stats

import (
	"context"
	"fmt"
	"time"

	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type Storage interface {
	Write(ctx context.Context, data []BlockCreationStats) error
	Read(ctx context.Context) ([]BlockCreationStats, error)
}

func NewStorage(l log.Logger, cfg config.ReplayorConfig) (Storage, error) {
	switch cfg.StorageType {
	case "s3":
		return NewS3Storage(l, cfg)
	case "disk":
		return NewDiskStorage(l, cfg)
	default:
		return nil, fmt.Errorf("unknown storage type %s", cfg.StorageType)
	}
}

type Stats interface {
	RecordBlockStats(bcs BlockCreationStats)
	Write(ctx context.Context)
	GetLastBlockWritten() uint64
}

type NoOpStats struct{}

func (n *NoOpStats) RecordBlockStats(bcs BlockCreationStats) {}
func (n *NoOpStats) Write(ctx context.Context)               {}
func (n *NoOpStats) GetLastBlockWritten() uint64             { return uint64(0) }

func NewStoredStats(s Storage, l log.Logger, startBlock uint64) Stats {
	return &StoredStats{
		s:                s,
		stats:            []BlockCreationStats{},
		log:              l,
		lastBlockWritten: startBlock,
	}
}

type StoredStats struct {
	s                Storage
	stats            []BlockCreationStats
	log              log.Logger
	lastBlockWritten uint64
}

func (s *StoredStats) GetLastBlockWritten() uint64 {
	return s.lastBlockWritten
}

func (s *StoredStats) RecordBlockStats(bcs BlockCreationStats) {
	s.stats = append(s.stats, bcs)
}

func (s *StoredStats) Write(ctx context.Context) {
	s.log.Info("writing block stats")
	_, err := retry.Do(ctx, 3, retry.Fixed(time.Second), func() (interface{}, error) {
		err := s.s.Write(ctx, s.stats)
		if err != nil {
			s.log.Info("error writing to storage", err)
		}
		return nil, err
	})
	if err != nil {
		s.log.Error("error writing to storage", "err", err)
	}
	if len(s.stats) > 0 {
		s.lastBlockWritten = s.stats[len(s.stats)-1].BlockNumber
	}
	s.log.Info("successfully wrote block stats to storage", "lastBlockWritten", s.lastBlockWritten)
}

type BlockCreationStats struct {
	TotalTime      time.Duration
	FCUTime        time.Duration
	GetTime        time.Duration
	NewTime        time.Duration
	FCUNoAttrsTime time.Duration

	BlockNumber uint64
	BlockHash   common.Hash
	TxnCount    int
	Success     float64
	GasUsed     uint64
	GasLimit    uint64

	OpCodes             map[string]OpCodeStats `json:",omitempty"`
	AccountsChanged     uint64                 `json:",omitempty"`
	StorageTriesChanged uint64                 `json:",omitempty"`
	SlotsChanged        uint64                 `json:",omitempty"`
}

type OpCodeStats struct {
	Count int
	Gas   uint64
}
