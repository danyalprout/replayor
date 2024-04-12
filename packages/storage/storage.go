package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/danyalprout/replayor/packages/config"
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
}
