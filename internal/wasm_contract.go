// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"log"

	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"

	"google.golang.org/protobuf/proto"
)

// WasmContract provides the Init and Invoke functions required by Fabric and
// represents a smart contract in Wasm.
type WasmContract struct {
	contextStore     *ContextStore
	wasmGuestInvoker WasmGuestInvoker
}

// NewWasmContract returns a new smart contract to invoke Wasm transactions
func NewWasmContract(contextStore *ContextStore, invoker WasmGuestInvoker) *WasmContract {
	contract := WasmContract{}
	contract.contextStore = contextStore
	contract.wasmGuestInvoker = invoker

	return &contract
}

// Init does nothing
func (wc *WasmContract) Init(APIstub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke calls a Wasm transaction
func (wc *WasmContract) Invoke(APIstub shim.ChaincodeStubInterface) pb.Response {
	result, err := wc.callTransaction(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(result)
}

func (wc *WasmContract) callTransaction(APIstub shim.ChaincodeStubInterface) ([]byte, error) {
	txID := APIstub.GetTxID()
	channelID := APIstub.GetChannelID()

	err := wc.contextStore.Put(channelID, txID, APIstub)
	if err != nil {
		log.Printf("[host] error putting stub for context chid %s txid %s: %s\n", channelID, txID, err)
		return nil, err
	}
	defer func() {
		err := wc.contextStore.Remove(channelID, txID)
		if err != nil {
			log.Printf("[host] error removing stub for context chid %s txid %s: %s\n", channelID, txID, err)
		}
	}()

	function, params := APIstub.GetFunctionAndParameters()

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		log.Printf("[host] error creating invoke transaction request message: %s\n", err)
		return nil, err
	}

	log.Printf("[host] calling %s with context chid %s txid %s\n", function, channelID, txID)

	args, err := createInvokeTransactionArgs(channelID, txID, function, params, transientMap)
	if err != nil {
		log.Printf("[host] error creating invoke transaction request message: %s\n", err)
		return nil, err
	}

	result, err := wc.wasmGuestInvoker.InvokeWasmOperation("InvokeTransaction", args)
	if err != nil {
		log.Printf("[host] error invoking transaction: %s\n", err)
		return nil, err
	}

	response := &contract.InvokeTransactionResponse{}
	err = proto.Unmarshal(result, response)
	responsePayload := response.GetPayload()

	log.Printf("[host] success result=%s\n", string(responsePayload))
	return responsePayload, nil
}

func createInvokeTransactionArgs(channelID string, txID string, fnname string, params []string, transientMap map[string][]byte) ([]byte, error) {
	args := make([][]byte, len(params))
	for i, p := range params {
		args[i] = []byte(p)
	}

	context := &contract.TransactionContext{
		ChannelId:     channelID,
		TransactionId: txID,
	}
	msg := &contract.InvokeTransactionRequest{
		Context:         context,
		TransactionName: fnname,
		Args:            args,
		TransientArgs:   transientMap,
	}

	argsBuffer, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return argsBuffer, nil
}
