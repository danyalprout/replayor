package strategies

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type OneForOne struct{}

func blockToCreationParams(input *types.Block) BlockCreationParams {
	gl := eth.Uint64Quantity(input.GasLimit())

	return BlockCreationParams{
		Number:       input.NumberU64(),
		Transactions: input.Transactions(),
		GasLimit:     &gl,
		Time:         eth.Uint64Quantity(input.Time()),
		MixDigest:    eth.Bytes32(input.MixDigest()),
		BeaconRoot:   input.BeaconRoot(),
		validateInfo: input.Hash(),
	}
}

func (o *OneForOne) BlockReceived(ctx context.Context, input *types.Block) *BlockCreationParams {
	res := blockToCreationParams(input)
	return &res
}

func (o *OneForOne) ValidateExecution(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error {
	expectedHash := a.validateInfo.(common.Hash)
	if e.ExecutionPayload.BlockHash != expectedHash {
		return fmt.Errorf("expected block hash %s, got %s", expectedHash.Hex(), e.ExecutionPayload.BlockHash.Hex())
	}

	return nil
}

func (o *OneForOne) ValidateBlock(ctx context.Context, e *eth.ExecutionPayloadEnvelope, a BlockCreationParams) error {
	return o.ValidateExecution(ctx, e, a)
}
