package strategies

import (
	"math/big"

	"github.com/danyalprout/replayor/packages/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

const (
	stressOpcodeBin = "608060405234801561000f575f80fd5b506102f48061001d5f395ff3fe608060405234801561000f575f80fd5b506004361061004a575f3560e01c806318e86fa41461004e5780635387694b1461006a57806373f080781461009a57806382f161ca146100b6575b5f80fd5b610068600480360381019061006391906101b5565b6100d4565b005b610084600480360381019061007f91906101b5565b610110565b60405161009191906101ef565b60405180910390f35b6100b460048036038101906100af9190610208565b610124565b005b6100be610178565b6040516100cb91906101ef565b60405180910390f35b6001545f5b82811015610101576040518181525f60208201526040812080815550506001810190506100d9565b50506001548181016001555050565b5f602052805f5260405f205f915090505481565b5f820361013957610134816100d4565b610174565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161016b906102a0565b60405180910390fd5b5050565b60015481565b5f80fd5b5f819050919050565b61019481610182565b811461019e575f80fd5b50565b5f813590506101af8161018b565b92915050565b5f602082840312156101ca576101c961017e565b5b5f6101d7848285016101a1565b91505092915050565b6101e981610182565b82525050565b5f6020820190506102025f8301846101e0565b92915050565b5f806040838503121561021e5761021d61017e565b5b5f61022b858286016101a1565b925050602061023c858286016101a1565b9150509250929050565b5f82825260208201905092915050565b7f556e737570706f72746564206f70636f646500000000000000000000000000005f82015250565b5f61028a601283610246565b915061029582610256565b602082019050919050565b5f6020820190508181035f8301526102b78161027e565b905091905056fea26469706673582212202aed1a24447171ab9259f96fc475f73c9b246b148fb5790d9b966193cb93728d64736f6c63430008140033"
	SstoreDesc      = "SSTORE"
	BaseGasUsed     = uint64(26_800)
)

const (
	SstoreCode OpCodeType = iota
)

var (
	opcodeAddr      = common.Address{}
	opcodeSignature = crypto.Keccak256([]byte("executeOpcode(uint256,uint256)"))[:4]
)

type OpCodeType int

func opCodeTypeFrom(desc string) OpCodeType {
	switch desc {
	case SstoreDesc:
		return SstoreCode
	default:
		return SstoreCode
	}
}

type OpcodeStressPacker struct {
	cfg        *config.ReplayorConfig
	signer     types.Signer
	opCodeType OpCodeType
	gasUsed    uint64
	logger     log.Logger
}

func NewOpcodeStressPacker(cfg *config.ReplayorConfig, logger log.Logger) *OpcodeStressPacker {
	return &OpcodeStressPacker{
		cfg:        cfg,
		signer:     types.NewLondonSigner(cfg.ChainId),
		logger:     logger,
		opCodeType: opCodeTypeFrom(cfg.StressConfig.OpCodeType),
		gasUsed:    BaseGasUsed * cfg.StressConfig.OpCodeExecNum,
	}
}

func (h *OpcodeStressPacker) PrepareDeployTx(input *types.Block) (*types.Transaction, error) {
	nonceMu.Lock()
	nonce := nonces[0]
	nonceMu.Unlock()
	txn := types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		Value:     big.NewInt(0),
		Gas:       uint64(2_000_000),
		GasTipCap: input.Transactions()[1].GasTipCap(),
		GasFeeCap: big.NewInt(999_999_999_999),
		Data:      common.Hex2Bytes(stressOpcodeBin),
	})

	if signedTx, err := types.SignTx(txn, h.signer, privateKeys[0]); err != nil {
		h.logger.Error("failed to sign tx", "err", err)
		return nil, err
	} else {
		opcodeAddr = crypto.CreateAddress(addresses[0], nonce)
		nonceMu.Lock()
		nonces[0] += 1
		nonceMu.Unlock()
		return signedTx, nil
	}
}

func (h *OpcodeStressPacker) PackUp(input *types.Block) types.Transactions {
	gasInfo := input.Transactions()[len(input.Transactions())-1]

	fillUpRemaining := h.cfg.GasTarget
	h.logger.Info("gas used", "blockNum", input.NumberU64(), "original", input.GasUsed(), "target", h.cfg.GasTarget)

	result := types.Transactions{}
	for fillUpRemaining >= h.gasUsed {
		txn := h.createOpCodeTx(gasInfo, h.gasUsed)
		if signedTx, err := types.SignTx(txn, h.signer, privateKeys[0]); err != nil {
			h.logger.Error("failed to sign tx", "err", err)
			continue
		} else {
			result = append(result, signedTx)
			nonceMu.Lock()
			nonces[0] += 1
			nonceMu.Unlock()
			fillUpRemaining -= h.gasUsed
		}
	}
	h.logger.Info("completed packing block", "blockNum", input.NumberU64(), "original", input.GasUsed(), "target", h.cfg.GasTarget)
	return result
}

func (h *OpcodeStressPacker) createOpCodeTx(gasInfo *types.Transaction, gasUsed uint64) *types.Transaction {
	nonceMu.Lock()
	defer nonceMu.Unlock()
	data := append(opcodeSignature, common.LeftPadBytes(big.NewInt(int64(h.opCodeType)).Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(big.NewInt(int64(h.cfg.StressConfig.OpCodeExecNum)).Bytes(), 32)...)
	txn := types.NewTx(&types.DynamicFeeTx{
		To:        &opcodeAddr,
		Nonce:     nonces[0],
		Value:     ZeroInt,
		Gas:       gasUsed,
		GasTipCap: gasInfo.GasTipCap(),
		GasFeeCap: big.NewInt(999_999_999_999),
		Data:      data,
	})
	return txn
}
