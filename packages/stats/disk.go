package stats

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum/go-ethereum/log"
)

type DiskStorage struct {
	log  log.Logger
	path string
}

func NewDiskStorage(l log.Logger, cfg config.ReplayorConfig) (*DiskStorage, error) {
	dir := fmt.Sprintf("%s/%s/%s", cfg.DiskPath, cfg.TestName, cfg.TestDescription())
	path := fmt.Sprintf("%s/%d", dir, time.Now().Unix())

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	return &DiskStorage{
		log:  l,
		path: path,
	}, nil
}

func (s *DiskStorage) Write(ctx context.Context, data []BlockCreationStats) error {
	b, err := json.Marshal(data)
	if err != nil {
		s.log.Warn("error encoding stats", "err", err)
		return errors.New("error encoding data")
	}

	err = os.WriteFile(s.path, b, 0644)
	if err != nil {
		s.log.Warn("error writing results", "name", s.path, "err", err)
		return errors.New("error writing data")
	}

	s.log.Info("wrote results", "data", s.path)
	return nil
}

func (s *DiskStorage) Read(ctx context.Context) ([]BlockCreationStats, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		s.log.Warn("error reading results", "name", s.path, "err", err)
		return nil, errors.New("error reading data")
	}

	r := bytes.NewReader(b)

	var data []BlockCreationStats
	err = json.NewDecoder(r).Decode(&data)
	if err != nil {
		s.log.Warn("error decoding results", "name", s.path, "err", err)
	}

	return data, nil
}
