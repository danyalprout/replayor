package config

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/danyalprout/replayor/packages/precompiles"
	"github.com/danyalprout/replayor/packages/storage_stressor"
	opnode "github.com/ethereum-optimism/optimism/op-node"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

type ReplayorConfig struct {
	EngineApiSecret     common.Hash
	SourceNodeUrl       string
	ChainId             *big.Int
	RollupConfig        *rollup.Config
	EngineApiUrl        string
	ExecutionUrl        string
	Strategy            string
	BlockCount          int
	GasTarget           uint64
	GasLimit            uint64
	BenchmarkStartBlock uint64
	BenchmarkOpcodes    bool
	ComputeStorageDiffs bool
	TestName            string
	Bucket              string
	StorageType         string
	DiskPath            string
	StressTestType      string
	StressTestFunction  string
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

	testName := cliCtx.String(TestName.Name)
	if testName == "" {
		testName, err = os.Hostname()
		if err != nil {
			return ReplayorConfig{}, err
		}
	}

	gasLimit := cliCtx.Uint64(GasLimit.Name)
	gasTarget := cliCtx.Uint64(GasTarget.Name)
	if gasTarget > gasLimit {
		return ReplayorConfig{}, fmt.Errorf("cannot set gasTarget greater than gasLimit")
	}

	strategy := cliCtx.String(Strategy.Name)
	stressTestType := cliCtx.String(StressTestType.Name)
	stressTestFunction := cliCtx.String(StressTestFunction.Name)

	// Validate stress test configuration
	if strategy == "stress" {
		if stressTestType == "" {
			return ReplayorConfig{}, fmt.Errorf("must provide stress test type when strategy is 'stress'")
		}

		switch stressTestType {
		case "precompile":
			if stressTestFunction != "" {
				if _, ok := precompiles.PrecompileSignatures[stressTestFunction]; !ok {
					return ReplayorConfig{},
						fmt.Errorf("invalid precompile function: %s. Valid functions: %s",
							stressTestFunction,
							precompiles.GetValidPrecompileNames(),
						)
				}
			}
		case "storage":
			if stressTestFunction != "" {
				if _, ok := storage_stressor.StorageStressorSignatures[stressTestFunction]; !ok {
					return ReplayorConfig{},
						fmt.Errorf("invalid storage stressor function: %s. Valid functions: %s",
							stressTestFunction,
							storage_stressor.GetValidStorageStressorNames(),
						)
				}
			}
		case "erc20":
			// No specific function needed for ERC20
		default:
			return ReplayorConfig{}, fmt.Errorf("invalid stress test type: %s. Valid types: precompile, storage, erc20", stressTestType)
		}
	} else if stressTestType != "" || stressTestFunction != "" {
		return ReplayorConfig{}, fmt.Errorf("stress test type and function can only be set when strategy is 'stress'")
	}

	// Replace the old fields with the new ones in the config
	return ReplayorConfig{
		EngineApiSecret:     secretHash,
		SourceNodeUrl:       cliCtx.String(SourceNodeUrl.Name),
		ChainId:             rollupCfg.L2ChainID,
		RollupConfig:        rollupCfg,
		EngineApiUrl:        cliCtx.String(EngineApiUrl.Name),
		ExecutionUrl:        cliCtx.String(ExecutionUrl.Name),
		Strategy:            strategy,
		BlockCount:          cliCtx.Int(BlockCount.Name),
		GasTarget:           gasTarget,
		GasLimit:            gasLimit,
		BenchmarkStartBlock: cliCtx.Uint64(BenchmarkStartBlock.Name),
		BenchmarkOpcodes:    cliCtx.Bool(BenchmarkOpcodes.Name),
		ComputeStorageDiffs: cliCtx.Bool(ComputeStorageDiffs.Name),
		TestName:            testName + "_",
		Bucket:              cliCtx.String(S3Bucket.Name),
		StorageType:         cliCtx.String(StorageType.Name),
		DiskPath:            cliCtx.String(DiskPath.Name),
		StressTestType:      stressTestType,
		StressTestFunction:  stressTestFunction,
	}, nil
}
