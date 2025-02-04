package precompiles

import (
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const PrecompileTargeterBytecode = "6080604052348015600e575f80fd5b506109188061001c5f395ff3fe608060405234801561000f575f80fd5b5060043610610091575f3560e01c8063496ebc8511610064578063496ebc85146100bd578063556d2a77146100c757806370a15753146100d1578063755ef941146100db578063d4dbf9f3146100e557610091565b8063075e86f51461009557806309d7f0491461009f578063244bc209146100a957806345794e80146100b3575b5f80fd5b61009d6100ef565b005b6100a76101ef565b005b6100b161027b565b005b6100bb6102fc565b005b6100c56103e3565b005b6100cf610475565b005b6100d961050d565b005b6100e3610635565b005b6100ed6106c7565b005b5f620493e090505f70010000000000000000000000000000000090505f70020000000000000000000000000000000090505f7f200000000000000000000000000000000000000000000000000000000000000090505b835a11156101e957610155610759565b83815f600381106101695761016861079e565b5b60200201818152505082816001600381106101875761018661079e565b5b60200201818152505081816002600381106101a5576101a461079e565b5b6020020181815250505f8060608360055f19fa5083806101c490610801565b94505082806101d290610801565b93505081806101e090610801565b92505050610145565b50505050565b5f620493e090505f600190505f600290505f600390505f600490505b845a1115610274575f60405180608001604052808681526020018581526020018481526020018381525090505f8060808360065f19fa5060018561024f9190610848565b945060028461025e9190610848565b9350600183901b9250600282901b91505061020b565b5050505050565b5f620493e090505f600190505f600290505f600390505b835a11156102f6575f60405180606001604052808581526020018481526020018381525090505f8060608360075f19fa506001846102d09190610848565b93506002836102df9190610848565b92506003826102ee9190610848565b915050610292565b50505050565b5f620493e090505f60015f1b90505f601b90505f60025f1b90505f60035f1b90505b845a11156103dc575f608067ffffffffffffffff8111156103425761034161087b565b5b6040519080825280601f01601f1916602001820160405280156103745781602001600182028036833780820191505090505b5090508460208201528360408201528260608201528160808201525f8060806020840160015f19fa506001855f1c6103ac9190610848565b5f1b94506001835f1c6103bf9190610848565b5f1b92506001825f1c6103d29190610848565b5f1b91505061031e565b5050505050565b5f620493e090505f604067ffffffffffffffff8111156104065761040561087b565b5b6040519080825280601f01601f1916602001820160405280156104385781602001600182028036833780820191505090505b5090505f600190505b825a1115610470578060208301525f8060406020850160025f19fa50808061046890610801565b915050610441565b505050565b5f620493e090505f60d567ffffffffffffffff8111156104985761049761087b565b5b6040519080825280601f01601f1916602001820160405280156104ca5781602001600182028036833780820191505090505b5090505f600c90508060208301525b825a1115610508575f8060d56020850160095f19fa5080806104fa906108b7565b9150508060208301526104d9565b505050565b5f620493e0905061051c61077b565b6001815f600c81106105315761053061079e565b5b6020020181815250506002816001600c81106105505761054f61079e565b5b6020020181815250506003816002600c811061056f5761056e61079e565b5b6020020181815250506004816003600c811061058e5761058d61079e565b5b6020020181815250506005816004600c81106105ad576105ac61079e565b5b6020020181815250506006816005600c81106105cc576105cb61079e565b5b6020020181815250505b815a1115610631575f806101808360085f19fa505f5b600c81101561062b578181600c81106106085761060761079e565b5b60200201805180919061061a90610801565b8152505080806001019150506105ec565b506105d6565b5050565b5f620493e090505f604067ffffffffffffffff8111156106585761065761087b565b5b6040519080825280601f01601f19166020018201604052801561068a5781602001600182028036833780820191505090505b5090505f600190505b825a11156106c2578060208301525f8060406020850160045f19fa5080806106ba90610801565b915050610693565b505050565b5f620493e090505f604067ffffffffffffffff8111156106ea576106e961087b565b5b6040519080825280601f01601f19166020018201604052801561071c5781602001600182028036833780820191505090505b5090505f600190505b825a1115610754578060208301525f8060406020850160035f19fa50808061074c90610801565b915050610725565b505050565b6040518060600160405280600390602082028036833780820191505090505090565b604051806101800160405280600c90602082028036833780820191505090505090565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f819050919050565b5f61080b826107f8565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff820361083d5761083c6107cb565b5b600182019050919050565b5f610852826107f8565b915061085d836107f8565b9250828201905080821115610875576108746107cb565b5b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b5f63ffffffff82169050919050565b5f6108c1826108a8565b915063ffffffff82036108d7576108d66107cb565b5b60018201905091905056fea264697066735822122042fc323ea3bc5368f3694708d107b1a98f187837b7b9ea4919a5a1f9af96d2b964736f6c634300081a0033"

var PrecompileSignatures = map[string][]byte{
	"modexp":    common.Hex2Bytes("075e86f5"), // cast sig "modexpPrecompile()"
	"ecadd":     common.Hex2Bytes("09d7f049"), // cast sig "ecaddPrecompile()"
	"ecmul":     common.Hex2Bytes("244bc209"), // cast sig "ecmulPrecompile()"
	"ecpairing": common.Hex2Bytes("70a15753"), // cast sig "ecpairingPrecompile()"
	"sha256":    common.Hex2Bytes("496ebc85"), // cast sig "sha256Precompile()"
	"ripemd160": common.Hex2Bytes("d4dbf9f3"), // cast sig "ripemd160Precompile()"
	"identity":  common.Hex2Bytes("755ef941"), // cast sig "identityPrecompile()"
	"ecrecover": common.Hex2Bytes("45794e80"), // cast sig "ecrecoverPrecompile()"
	"blake2f":   common.Hex2Bytes("556d2a77"), // cast sig "blake2fPrecompile()"
}

func GetValidPrecompileNames() string {
	names := make([]string, 0, len(PrecompileSignatures))
	for name := range PrecompileSignatures {
		names = append(names, name)
	}
	sort.Strings(names) // Sort for consistent output
	return strings.Join(names, ", ")
}
