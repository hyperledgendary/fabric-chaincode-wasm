package proxychaincode

import (
	"github.com/hyperledgendary/fabric-chaincode-wasm/wasmruntime"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

// ProxyContract is the structure to give to the chaincode library to appear as a chaincode
// to get the invoke/init methods
type ProxyChaincode struct {
	runtime *wasmruntime.WasmPcRuntime
}

// Init is called for chaincode initialization
func (c *ProxyChaincode) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke is called for chaindcode innvocations. t is called for chaincode initialization
func (c *ProxyChaincode) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	txid := APIstub.GetTxID()
	channelid := APIstub.GetChannelID()
	c.runtime.Call(function, args, txid, channelid)
	return shim.Success(nil)
}

// NewChaincode creates a proxy chanincode
func NewChaincode(runtime *wasmruntime.WasmPcRuntime) *ProxyChaincode {
	cc := ProxyChaincode{}
	cc.runtime = runtime
	return &cc
}
