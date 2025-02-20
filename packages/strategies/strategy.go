package strategies

import (
	"context"

	"github.com/danyalprout/replayor/packages/clients"
	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

const (
	strategyTypeStress = "stress"
	strategyTypeReplay = "replay"
)

type BlockCreationParams struct {
	Number       uint64
	Transactions types.Transactions
	GasLimit     *eth.Uint64Quantity
	Time         eth.Uint64Quantity
	MixDigest    eth.Bytes32
	BeaconRoot   *common.Hash
	FeeRecipient common.Address
	Withdrawals  types.Withdrawals
	Extra        []byte
	validateInfo interface{}
}

type Strategy interface {
	BlockReceived(ctx context.Context, input *types.Block) *BlockCreationParams
	ValidateExecution(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error
	ValidateBlock(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error
}

func LoadStrategy(cfg config.ReplayorConfig, logger log.Logger, c clients.Clients, startBlock *types.Block) Strategy {
	switch cfg.Strategy {
	case strategyTypeStress:
		return NewStressTest(startBlock, logger, cfg, c)
	case strategyTypeReplay:
		return &OneForOne{}
	default:
		return nil
	}
}
