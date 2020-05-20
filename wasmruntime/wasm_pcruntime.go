// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package wasmruntime

import (
	"context"
	"fmt"
	"log"

	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"

	wapc "github.com/wapc/wapc-go"
)

// WasmPcRuntime is an abstraction of the instance of the Wasm engine
type WasmPcRuntime struct {
	ctx          context.Context
	wapcInstance wapc.Instance
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// CallContract makes the requried tx call on the wasm contract
func (wr *WasmPcRuntime) CallContract(fnname string, args [][]byte, txid string, channelid string) {

	msg := &contract.InvokeTransactionRequest{
		ChannelId:       channelid,
		TransactionId:   txid,
		TransactionName: fnname,
		Args:            args}

	argsBuffer, err := proto.Marshal(msg)
	check(err)

	log.Printf("[host] calling %s txid=%s", fnname, txid)
	// result, err := wr.vm.Run(entryID, int64(len(fnBuffer)), int64(len(argsBuffer)))
	// log.Printf("%d %s", result, wr.callctx.finalResult.Data)

	result, err := wr.wapcInstance.Invoke(wr.ctx, "InvokeTransaction", []byte(argsBuffer))
	if err != nil {
		panic(err)
	}

	fmt.Println(string(result))
}

// NewRuntime Get the runtime
func NewRuntime(ctx context.Context, wapcInstance wapc.Instance) *WasmPcRuntime {
	wr := WasmPcRuntime{}
	wr.ctx = ctx
	wr.wapcInstance = wapcInstance

	return &wr
}

// Init is called for chaincode initialization
func (wr *WasmPcRuntime) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke is called for chaindcode innvocations. t is called for chaincode initialization
func (wr *WasmPcRuntime) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	txid := APIstub.GetTxID()
	channelid := APIstub.GetChannelID()

	bargs := make([][]byte, len(args))
	for i, a := range args {
		bargs[i] = []byte(a)
	}

	wr.CallContract(function, bargs, txid, channelid)
	return shim.Success(nil)
}
