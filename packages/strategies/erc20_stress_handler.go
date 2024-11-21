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
	stressErc20Bin = "608060405234801561000f575f80fd5b50336040518060400160405280601381526020017f5265706c61796f72537472657373546f6b656e000000000000000000000000008152506040518060400160405280600481526020017f5253535400000000000000000000000000000000000000000000000000000000815250816003908161008c91906106e6565b50806004908161009c91906106e6565b5050505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff160361010f575f6040517f1e4fbdf700000000000000000000000000000000000000000000000000000000815260040161010691906107f4565b60405180910390fd5b61011e8161014b60201b60201c565b5061014633678ac7230489e80000633b9aca0061013b919061083a565b61020e60201b60201c565b61090b565b5f60055f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1690508160055f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361027e575f6040517fec442f0500000000000000000000000000000000000000000000000000000000815260040161027591906107f4565b60405180910390fd5b61028f5f838361029360201b60201c565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036102e3578060025f8282546102d7919061087b565b925050819055506103b1565b5f805f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205490508181101561036c578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401610363939291906108bd565b60405180910390fd5b8181035f808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036103f8578060025f8282540392505081905550610442565b805f808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8360405161049f91906108f2565b60405180910390a3505050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061052757607f821691505b60208210810361053a576105396104e3565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f6008830261059c7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610561565b6105a68683610561565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f6105ea6105e56105e0846105be565b6105c7565b6105be565b9050919050565b5f819050919050565b610603836105d0565b61061761060f826105f1565b84845461056d565b825550505050565b5f90565b61062b61061f565b6106368184846105fa565b505050565b5b818110156106595761064e5f82610623565b60018101905061063c565b5050565b601f82111561069e5761066f81610540565b61067884610552565b81016020851015610687578190505b61069b61069385610552565b83018261063b565b50505b505050565b5f82821c905092915050565b5f6106be5f19846008026106a3565b1980831691505092915050565b5f6106d683836106af565b9150826002028217905092915050565b6106ef826104ac565b67ffffffffffffffff811115610708576107076104b6565b5b6107128254610510565b61071d82828561065d565b5f60209050601f83116001811461074e575f841561073c578287015190505b61074685826106cb565b8655506107ad565b601f19841661075c86610540565b5f5b828110156107835784890151825560018201915060208501945060208101905061075e565b868310156107a0578489015161079c601f8916826106af565b8355505b6001600288020188555050505b505050505050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6107de826107b5565b9050919050565b6107ee816107d4565b82525050565b5f6020820190506108075f8301846107e5565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f610844826105be565b915061084f836105be565b925082820261085d816105be565b915082820484148315176108745761087361080d565b5b5092915050565b5f610885826105be565b9150610890836105be565b92508282019050808211156108a8576108a761080d565b5b92915050565b6108b7816105be565b82525050565b5f6060820190506108d05f8301866107e5565b6108dd60208301856108ae565b6108ea60408301846108ae565b949350505050565b5f6020820190506109055f8301846108ae565b92915050565b61112f806109185f395ff3fe6080604052600436106100c5575f3560e01c806370a082311161007e57806395d89b411161005857806395d89b411461026c578063a9059cbb14610296578063dd62ed3e146102d2578063f2fde38b1461030e576100cc565b806370a08231146101f0578063715018a61461022c5780638da5cb5b14610242576100cc565b806306fdde03146100d0578063095ea7b3146100fa57806318160ddd1461013657806323b872dd14610160578063313ce5671461019c578063664e9704146101c6576100cc565b366100cc57005b5f80fd5b3480156100db575f80fd5b506100e4610336565b6040516100f19190610da8565b60405180910390f35b348015610105575f80fd5b50610120600480360381019061011b9190610e59565b6103c6565b60405161012d9190610eb1565b60405180910390f35b348015610141575f80fd5b5061014a6103e8565b6040516101579190610ed9565b60405180910390f35b34801561016b575f80fd5b5061018660048036038101906101819190610ef2565b6103f1565b6040516101939190610eb1565b60405180910390f35b3480156101a7575f80fd5b506101b061041f565b6040516101bd9190610f5d565b60405180910390f35b3480156101d1575f80fd5b506101da610427565b6040516101e79190610ed9565b60405180910390f35b3480156101fb575f80fd5b5061021660048036038101906102119190610f76565b610433565b6040516102239190610ed9565b60405180910390f35b348015610237575f80fd5b50610240610478565b005b34801561024d575f80fd5b5061025661048b565b6040516102639190610fb0565b60405180910390f35b348015610277575f80fd5b506102806104b3565b60405161028d9190610da8565b60405180910390f35b3480156102a1575f80fd5b506102bc60048036038101906102b79190610e59565b610543565b6040516102c99190610eb1565b60405180910390f35b3480156102dd575f80fd5b506102f860048036038101906102f39190610fc9565b610565565b6040516103059190610ed9565b60405180910390f35b348015610319575f80fd5b50610334600480360381019061032f9190610f76565b6105e7565b005b60606003805461034590611034565b80601f016020809104026020016040519081016040528092919081815260200182805461037190611034565b80156103bc5780601f10610393576101008083540402835291602001916103bc565b820191905f5260205f20905b81548152906001019060200180831161039f57829003601f168201915b5050505050905090565b5f806103d061066b565b90506103dd818585610672565b600191505092915050565b5f600254905090565b5f806103fb61066b565b9050610408858285610684565b610413858585610716565b60019150509392505050565b5f6012905090565b678ac7230489e8000081565b5f805f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b610480610806565b6104895f61088d565b565b5f60055f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6060600480546104c290611034565b80601f01602080910402602001604051908101604052809291908181526020018280546104ee90611034565b80156105395780601f1061051057610100808354040283529160200191610539565b820191905f5260205f20905b81548152906001019060200180831161051c57829003601f168201915b5050505050905090565b5f8061054d61066b565b905061055a818585610716565b600191505092915050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b6105ef610806565b5f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff160361065f575f6040517f1e4fbdf70000000000000000000000000000000000000000000000000000000081526004016106569190610fb0565b60405180910390fd5b6106688161088d565b50565b5f33905090565b61067f8383836001610950565b505050565b5f61068f8484610565565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81146107105781811015610701578281836040517ffb8f41b20000000000000000000000000000000000000000000000000000000081526004016106f893929190611064565b60405180910390fd5b61070f84848484035f610950565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610786575f6040517f96c6fd1e00000000000000000000000000000000000000000000000000000000815260040161077d9190610fb0565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036107f6575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016107ed9190610fb0565b60405180910390fd5b610801838383610b1f565b505050565b61080e61066b565b73ffffffffffffffffffffffffffffffffffffffff1661082c61048b565b73ffffffffffffffffffffffffffffffffffffffff161461088b5761084f61066b565b6040517f118cdaa70000000000000000000000000000000000000000000000000000000081526004016108829190610fb0565b60405180910390fd5b565b5f60055f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1690508160055f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16036109c0575f6040517fe602df050000000000000000000000000000000000000000000000000000000081526004016109b79190610fb0565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610a30575f6040517f94280d62000000000000000000000000000000000000000000000000000000008152600401610a279190610fb0565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015610b19578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92584604051610b109190610ed9565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610b6f578060025f828254610b6391906110c6565b92505081905550610c3d565b5f805f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015610bf8578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401610bef93929190611064565b60405180910390fd5b8181035f808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610c84578060025f8282540392505081905550610cce565b805f808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051610d2b9190610ed9565b60405180910390a3505050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f610d7a82610d38565b610d848185610d42565b9350610d94818560208601610d52565b610d9d81610d60565b840191505092915050565b5f6020820190508181035f830152610dc08184610d70565b905092915050565b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610df582610dcc565b9050919050565b610e0581610deb565b8114610e0f575f80fd5b50565b5f81359050610e2081610dfc565b92915050565b5f819050919050565b610e3881610e26565b8114610e42575f80fd5b50565b5f81359050610e5381610e2f565b92915050565b5f8060408385031215610e6f57610e6e610dc8565b5b5f610e7c85828601610e12565b9250506020610e8d85828601610e45565b9150509250929050565b5f8115159050919050565b610eab81610e97565b82525050565b5f602082019050610ec45f830184610ea2565b92915050565b610ed381610e26565b82525050565b5f602082019050610eec5f830184610eca565b92915050565b5f805f60608486031215610f0957610f08610dc8565b5b5f610f1686828701610e12565b9350506020610f2786828701610e12565b9250506040610f3886828701610e45565b9150509250925092565b5f60ff82169050919050565b610f5781610f42565b82525050565b5f602082019050610f705f830184610f4e565b92915050565b5f60208284031215610f8b57610f8a610dc8565b5b5f610f9884828501610e12565b91505092915050565b610faa81610deb565b82525050565b5f602082019050610fc35f830184610fa1565b92915050565b5f8060408385031215610fdf57610fde610dc8565b5b5f610fec85828601610e12565b9250506020610ffd85828601610e12565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061104b57607f821691505b60208210810361105e5761105d611007565b5b50919050565b5f6060820190506110775f830186610fa1565b6110846020830185610eca565b6110916040830184610eca565b949350505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f6110d082610e26565b91506110db83610e26565b92508282019050808211156110f3576110f2611099565b5b9291505056fea2646970667358221220b92fe02a783a60fedf867af6496dcd356baaf4f56b4efde6814602337a2eb6c864736f6c634300081a0033"
	erc20GasUsed   = int64(23_000)
)

var (
	erc20Address = common.Address{}
	// Hash("transfer(address,uint256)")[:4]
	transferSignature = common.Hex2Bytes("0xa9059cbb")
	// 1 wei
	transferAmount = common.LeftPadBytes(big.NewInt(1).Bytes(), 32)
)

type Erc20StressHandler struct {
	cfg    *config.ReplayorConfig
	signer types.Signer
	logger log.Logger
}

func NewErc20StressHandler(cfg *config.ReplayorConfig, logger log.Logger) *Erc20StressHandler {
	return &Erc20StressHandler{
		cfg:    cfg,
		signer: types.NewLondonSigner(cfg.ChainId),
		logger: logger,
	}
}

func (h *Erc20StressHandler) PrepareDeployTx(input *types.Block) (*types.Transaction, error) {
	nonceMu.Lock()
	nonce := nonces[0]
	nonceMu.Unlock()
	txn := types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonces[0],
		Value:     big.NewInt(0),
		Gas:       uint64(2_000_000),
		GasTipCap: input.Transactions()[1].GasTipCap(),
		GasFeeCap: big.NewInt(999_999_999_999),
		Data:      common.Hex2Bytes(stressErc20Bin),
	})

	if signedTx, err := types.SignTx(txn, h.signer, privateKeys[0]); err != nil {
		h.logger.Error("failed to sign tx", "err", err)
		return nil, err
	} else {
		erc20Address = crypto.CreateAddress(addresses[0], nonce)
		nonceMu.Lock()
		nonces[0] += 1
		nonceMu.Unlock()
		return signedTx, nil
	}
}

//func (h *Erc20StressHandler) PackUp(input *types.Block) types.Transactions {
//	gasInfo := input.Transactions()[len(input.Transactions())-1]
//
//	originalGasUsed := input.GasUsed()
//	targetUsage := h.cfg.GasTarget
//
//	fillUpRemaining := int64(targetUsage)
//	h.logger.Info("gas used", "blockNum", input.NumberU64(), "original", originalGasUsed, "target", targetUsage)
//
//	result := types.Transactions{}
//
//	if fillUpRemaining <= 0 {
//		return result
//	}
//
//	for {
//		gasUsed := int64(23_000)
//		if fillUpRemaining < gasUsed {
//			break
//		}
//
//		fillUpRemaining -= gasUsed
//
//		maxFeePerGas := gasInfo.GasFeeCap()
//		oneHundredTen := big.NewInt(150)
//		maxFeePerGas.Mul(maxFeePerGas, oneHundredTen)
//		maxFeePerGas.Div(maxFeePerGas, big.NewInt(100))
//
//		nonceMu.Lock()
//		erc20ReceiveAddr := crypto.Keccak256(big.NewInt(int64(nonces[0])).Bytes())[:20]
//
//		var data []byte
//		data = append(data, transferSignature...)
//		data = append(data, common.LeftPadBytes(erc20ReceiveAddr, 32)...)
//		data = append(data, transferAmount...)
//		txn := types.NewTx(&types.DynamicFeeTx{
//			To:        &erc20Address,
//			Nonce:     nonces[0],
//			Value:     big.NewInt(0),
//			Gas:       uint64(gasUsed),
//			GasTipCap: gasInfo.GasTipCap(),
//			GasFeeCap: big.NewInt(999_999_999_999),
//			Data:      data,
//		})
//		nonceMu.Unlock()
//
//		signer := types.NewLondonSigner(h.cfg.ChainId)
//		signedTx, err := types.SignTx(txn, signer, privateKeys[0])
//		if err != nil {
//			h.logger.Error("failed to sign tx", "err", err)
//			continue
//		}
//
//		result = append(result, signedTx)
//		nonceMu.Lock()
//		nonces[0] += 1
//		nonceMu.Unlock()
//	}
//	h.logger.Info("completed packing block", "blockNum", input.NumberU64(), "original", originalGasUsed, "target", targetUsage)
//
//	return result
//}

func (h *Erc20StressHandler) PackUp(input *types.Block) types.Transactions {
	gasInfo := input.Transactions()[len(input.Transactions())-1]
	targetUsage := h.cfg.GasTarget
	fillUpRemaining := int64(targetUsage) - int64(input.GasUsed())

	h.logger.Info("gas used", "blockNum", input.NumberU64(), "original", input.GasUsed(), "target", targetUsage)

	if fillUpRemaining <= 0 {
		return types.Transactions{}
	}

	result := make(types.Transactions, 0, fillUpRemaining/erc20GasUsed)

	for fillUpRemaining >= erc20GasUsed {
		txn := h.createERC20TransferTx(gasInfo)
		if signedTx, err := types.SignTx(txn, h.signer, privateKeys[0]); err == nil {
			result = append(result, signedTx)
			fillUpRemaining -= erc20GasUsed
			nonceMu.Lock()
			nonces[0]++
			nonceMu.Unlock()
		} else {
			h.logger.Error("failed to sign tx", "err", err)
		}
	}

	h.logger.Info("completed packing block", "blockNum", input.NumberU64(), "txCount", len(result))
	return result
}

func (h *Erc20StressHandler) createERC20TransferTx(gasInfo *types.Transaction) *types.Transaction {
	nonceMu.Lock()
	defer nonceMu.Unlock()

	erc20ReceiveAddr := crypto.Keccak256(big.NewInt(int64(nonces[0])).Bytes())[:20]
	data := append(transferSignature, common.LeftPadBytes(erc20ReceiveAddr, 32)...)
	data = append(data, transferAmount...)

	return types.NewTx(&types.DynamicFeeTx{
		To:        &erc20Address,
		Nonce:     nonces[0],
		Value:     big.NewInt(0),
		Gas:       uint64(erc20GasUsed),
		GasTipCap: gasInfo.GasTipCap(),
		GasFeeCap: big.NewInt(999_999_999_999),
		Data:      data,
	})
}
