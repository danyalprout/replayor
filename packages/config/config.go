package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	opnode "github.com/ethereum-optimism/optimism/op-node"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

type ReplayorConfig struct {
	EngineApiSecret common.Hash
	SourceNodeUrl   string
	ChainId         string
	RollupConfig    *rollup.Config
	EngineApiUrl    string
	ExecutionUrl    string
	Strategy        string
	BlockCount      int
	GasTarget       int
	GasLimit        int
	TestName        string
	Bucket          string
	StorageType     string
	DiskPath        string
}

func (r ReplayorConfig) TestDescription() string {
	return fmt.Sprintf("%s-%d", r.Strategy, r.BlockCount)
}

func LoadReplayorConfig(cliCtx *cli.Context, l log.Logger) (ReplayorConfig, error) {
	jwtFile := cliCtx.String(EngineApiSecret.Name)
	jwtBytes, err := ioutil.ReadFile(jwtFile)
	if err != nil {
		return ReplayorConfig{}, err
	}

	secret := common.HexToHash(strings.TrimSpace(string(jwtBytes)))

	chainId := cliCtx.String(ChainId.Name)

	rollupCfg, err := opnode.NewRollupConfig(l, chainId, "")
	if err != nil {
		return ReplayorConfig{}, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return ReplayorConfig{}, err
	}

	return ReplayorConfig{
		EngineApiSecret: secret,
		SourceNodeUrl:   cliCtx.String(SourceNodeUrl.Name),
		ChainId:         chainId,
		RollupConfig:    rollupCfg,
		EngineApiUrl:    cliCtx.String(EngineApiUrl.Name),
		ExecutionUrl:    cliCtx.String(ExecutionUrl.Name),
		Strategy:        cliCtx.String(Strategy.Name),
		BlockCount:      cliCtx.Int(BlockCount.Name),
		GasTarget:       cliCtx.Int(GasTarget.Name),
		GasLimit:        cliCtx.Int(GasLimit.Name),
		TestName:        hostname,
		Bucket:          cliCtx.String(S3Bucket.Name),
		StorageType:     cliCtx.String(StorageType.Name),
		DiskPath:        cliCtx.String(DiskPath.Name),
	}, nil
}
