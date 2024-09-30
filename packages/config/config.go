package config

import (
	"fmt"
	"os"
	"strings"

	opnode "github.com/ethereum-optimism/optimism/op-node"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

type ReplayorConfig struct {
	EngineApiSecret     common.Hash
	SourceNodeUrl       string
	ChainId             string
	RollupConfig        *rollup.Config
	EngineApiUrl        string
	ExecutionUrl        string
	Strategy            string
	BlockCount          int
	GasTarget           int
	GasLimit            int
	BenchmarkStartBlock uint64
	BenchmarkOpcodes    bool
	ComputeStorageDiffs bool
	TestName            string
	Bucket              string
	StorageType         string
	DiskPath            string
}

func (r ReplayorConfig) TestDescription() string {
	return fmt.Sprintf("%s-%d", r.Strategy, r.BlockCount)
}

func valueOrNil(i *uint64) string {
	if i == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *i)
}

func LoadReplayorConfig(cliCtx *cli.Context, l log.Logger) (ReplayorConfig, error) {
	secret := cliCtx.String(EngineApiSecret.Name)
	if secret == "" {
		return ReplayorConfig{}, fmt.Errorf("must provide REPLAYOR_ENGINE_API_SECRET env var")
	}

	secretHash := common.HexToHash(strings.TrimSpace(secret))

	chainId := cliCtx.String(ChainId.Name)
	rollupCfgPath := cliCtx.String(RollupConfigPath.Name)

	if chainId == "" && rollupCfgPath == "" {
		return ReplayorConfig{}, fmt.Errorf("must provide either chain id or rollup config path")
	}

	rollupCfg, err := opnode.NewRollupConfig(l, chainId, rollupCfgPath)
	if err != nil {
		return ReplayorConfig{}, err
	}

	l.Info("activation", "canyon", valueOrNil(rollupCfg.CanyonTime), "delta", valueOrNil(rollupCfg.DeltaTime), "ecotone", valueOrNil(rollupCfg.EcotoneTime), "fjord", valueOrNil(rollupCfg.FjordTime))

	hostname, err := os.Hostname()
	if err != nil {
		return ReplayorConfig{}, err
	}

	return ReplayorConfig{
		EngineApiSecret:     secretHash,
		SourceNodeUrl:       cliCtx.String(SourceNodeUrl.Name),
		ChainId:             chainId,
		RollupConfig:        rollupCfg,
		EngineApiUrl:        cliCtx.String(EngineApiUrl.Name),
		ExecutionUrl:        cliCtx.String(ExecutionUrl.Name),
		Strategy:            cliCtx.String(Strategy.Name),
		BlockCount:          cliCtx.Int(BlockCount.Name),
		GasTarget:           cliCtx.Int(GasTarget.Name),
		GasLimit:            cliCtx.Int(GasLimit.Name),
		BenchmarkStartBlock: cliCtx.Uint64(BenchmarkStartBlock.Name),
		BenchmarkOpcodes:    cliCtx.Bool(BenchmarkOpcodes.Name),
		ComputeStorageDiffs: cliCtx.Bool(ComputeStorageDiffs.Name),
		TestName:            hostname,
		Bucket:              cliCtx.String(S3Bucket.Name),
		StorageType:         cliCtx.String(StorageType.Name),
		DiskPath:            cliCtx.String(DiskPath.Name),
	}, nil
}
