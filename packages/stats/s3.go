package stats

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum/go-ethereum/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	defaultEndpoint = "s3.amazonaws.com"
)

type S3Storage struct {
	s3       *minio.Client
	bucket   string
	log      log.Logger
	fileName string
}

func NewS3Storage(l log.Logger, cfg config.ReplayorConfig) (*S3Storage, error) {
	c := credentials.NewIAM("")

	client, err := minio.New(defaultEndpoint, &minio.Options{
		Creds:  c,
		Secure: true,
	})

	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("%s/%s/%d", cfg.TestName, cfg.TestDescription(), time.Now().Unix())

	return &S3Storage{
		s3:       client,
		bucket:   cfg.Bucket,
		log:      l,
		fileName: fileName,
	}, nil
}

func (s *S3Storage) Write(ctx context.Context, data []BlockCreationStats) error {
	b, err := json.Marshal(data)
	if err != nil {
		s.log.Warn("error encoding stats", "err", err)
		return errors.New("error encoding data")
	}

	options := minio.PutObjectOptions{
		ContentType: "application/json",
	}

	reader := bytes.NewReader(b)

	_, err = s.s3.PutObject(ctx, s.bucket, s.fileName, reader, int64(len(b)), options)

	if err != nil {
		s.log.Warn("error writing results", "err", err)
		return err
	}

	s.log.Info("wrote results", "data", s.fileName)
	return nil
}

func (s *S3Storage) Read(ctx context.Context) ([]BlockCreationStats, error) {
	res, err := s.s3.GetObject(ctx, s.bucket, s.fileName, minio.GetObjectOptions{})
	if err != nil {
		s.log.Info("unexpected error fetching results", "name", s.fileName, "err", err)
		return nil, err
	}

	defer res.Close()
	var reader io.ReadCloser = res
	defer reader.Close()

	var data []BlockCreationStats
	err = json.NewDecoder(reader).Decode(&data)
	if err != nil {
		s.log.Warn("error decoding results", "name", s.fileName, "err", err)
	}

	return data, nil
}
