package config

import (
	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/urfave/cli/v2"
)

const EnvVarPrefix = "REPLAYOR"

var (
	EngineApiSecret = &cli.StringFlag{
		Name:     "engine-api-secret",
		Usage:    "Engine api secret",
		Required: true,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "ENGINE_API_SECRET"),
	}
	SourceNodeUrl = &cli.StringFlag{
		Name:     "source-node-url",
		Usage:    "The URL of the source node to fetch transactions from",
		Required: true,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "SOURCE_NODE_URL"),
	}
	ChainId = &cli.StringFlag{
		Name:     "chain-id",
		Usage:    "The chain id for the node being benchmarked",
		Required: false,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "CHAIN_ID"),
	}
	RollupConfigPath = &cli.StringFlag{
		Name:     "rollup-config-path",
		Usage:    "The path to the rollup config",
		Required: false,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "ROLLUP_CFG_PATH"),
	}
	EngineApiUrl = &cli.StringFlag{
		Name:    "engine-api-url",
		Usage:   "The URL of the engine api",
		Value:   "ws://0.0.0.0:8551",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "ENGINE_API_URL"),
	}
	ExecutionUrl = &cli.StringFlag{
		Name:    "execution-url",
		Usage:   "The URL of the execution node",
		Value:   "http://0.0.0.0:8545",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "EXECUTION_URL"),
	}
	Strategy = &cli.StringFlag{
		Name:     "strategy",
		Usage:    "The strategy to use for replaying transactions",
		Required: true,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "STRATEGY"),
	}
	BlockCount = &cli.IntFlag{
		Name:     "block-count",
		Usage:    "How many blocks to replay",
		Required: true,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "BLOCK_COUNT"),
	}
	GasTarget = &cli.IntFlag{
		Name:    "gas-target",
		Usage:   "desired gas target",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "GAS_TARGET"),
	}
	GasLimit = &cli.IntFlag{
		Name:    "gas-limit",
		Usage:   "desired gas limit",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "GAS_LIMIT"),
	}
	BenchmarkStartBlock = &cli.Uint64Flag{
		Name:    "benchmark-start-block",
		Usage:   "start block for the benchmarking",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "BENCHMARK_START_BLOCK"),
	}
	BenchmarkOpcodes = &cli.BoolFlag{
		Name:    "benchmark-opcodes",
		Usage:   "whether to include opcode metrics in the benchmark results",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "BENCHMARK_OPCODES"),
	}
	ComputeStorageDiffs = &cli.BoolFlag{
		Name:    "compute-storage-diffs",
		Usage:   "whether to include storage diff metrics in the benchmark results",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "COMPUTE_STORAGE_DIFFS"),
	}
	TestName = &cli.StringFlag{
		Name:     "test-name",
		Usage:    "test name used as prefix for output file",
		Required: false,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "TEST_NAME"),
	}
	S3Bucket = &cli.StringFlag{
		Name:     "s3-bucket",
		Usage:    "The S3 bucket to store results in",
		Required: false,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "S3_BUCKET"),
	}
	StorageType = &cli.StringFlag{
		Name:     "storage-type",
		Usage:    "where to store the results either s3 or disk",
		Required: true,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "STORAGE_TYPE"),
	}
	DiskPath = &cli.StringFlag{
		Name:     "disk-path",
		Usage:    "The path to store results in",
		Required: false,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "DISK_PATH"),
	}
	InjectErc20 = &cli.BoolFlag{
		Name:     "inject-erc20-txs",
		Usage:    "whether to inject erc20 txs",
		Required: false,
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "INJECT_ERC20_TXS"),
	}
	PrecompileTarget = &cli.StringFlag{
		Name:     "precompile-target",
		Usage:    "precompile opcode indicating which nject PrecompileTargeter txs",
		Required: false,
		Value:    "",
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "PRECOMPILE_TARGET"),
	}
)

func init() {
	Flags = append(Flags, oplog.CLIFlags(EnvVarPrefix)...)
	Flags = append(Flags,
		EngineApiSecret,
		SourceNodeUrl,
		ChainId,
		EngineApiUrl,
		ExecutionUrl,
		Strategy,
		BlockCount,
		GasTarget,
		GasLimit,
		S3Bucket,
		StorageType,
		DiskPath,
		BenchmarkStartBlock,
		BenchmarkOpcodes,
		ComputeStorageDiffs,
		TestName,
		RollupConfigPath,
		InjectErc20,
		PrecompileTarget,
	)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag
