package clients

import (
	"context"
	"time"

	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	gn "github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
)

type Clients struct {
	SourceNode *ethclient.Client
	DestNode   *ethclient.Client
	EngineApi  *sources.EngineAPIClient
}

func SetupClients(cfg config.ReplayorConfig, logger log.Logger, ctx context.Context) (Clients, error) {
	sourceNode, err := ethclient.Dial(cfg.SourceNodeUrl)
	if err != nil {
		return Clients{}, err
	}

	// Add some retries around connecting to the dest node as it starts up at the same time as the replayor
	destNode, err := retry.Do(ctx, 120, retry.Fixed(time.Second), func() (*ethclient.Client, error) {
		d, e := ethclient.Dial(cfg.ExecutionUrl)
		if e != nil {
			logger.Info("waiting for geth (exec) to start")
		}

		return d, e
	})

	if err != nil {
		return Clients{}, err
	}

	auth := rpc.WithHTTPAuth(gn.NewJWTAuth(cfg.EngineApiSecret))
	opts := []client.RPCOption{
		client.WithGethRPCOptions(auth),
		client.WithBatchCallTimeout(10 * time.Minute),
		client.WithCallTimeout(10 * time.Minute),
	}

	// Add some retries around connecting to the dest node as it starts up at the same time as the replayor
	engineApi, err := retry.Do(ctx, 120, retry.Fixed(time.Second), func() (*sources.EngineAPIClient, error) {
		l2Node, err := client.NewRPC(ctx, logger, cfg.EngineApiUrl, opts...)
		if err != nil {
			logger.Info("waiting for geth (rpc) to start")
			return nil, err
		}

		engineApi := sources.NewEngineAPIClientWithTimeout(l2Node, logger, cfg.RollupConfig, 10*time.Minute)

		return engineApi, nil
	})

	if err != nil {
		return Clients{}, err
	}

	return Clients{
		SourceNode: sourceNode,
		DestNode:   destNode,
		EngineApi:  engineApi,
	}, nil
}
