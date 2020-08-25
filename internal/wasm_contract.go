// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	wapc "github.com/wapc/wapc-go"
	"google.golang.org/protobuf/proto"
)

// WasmContract holds a pool of waPC instances to run smart contract transactions in
type WasmContract struct {
	contextStore *ContextStore
	wapcPool     *wapc.Pool
}

// NewWasmContract returns a new smart contract to invoke transactions in a waPC instance
func NewWasmContract(contextStore *ContextStore, wapcPool *wapc.Pool) *WasmContract {
	contract := WasmContract{}
	contract.contextStore = contextStore
	contract.wapcPool = wapcPool

	return &contract
}

// Init does nothing
func (wc *WasmContract) Init(APIstub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke calls a transaction in the waPC instance
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

	log.Printf("[host] calling %s with context chid %s txid %s\n", function, channelID, txID)

	args, err := createInvokeTransactionArgs(channelID, txID, function, params)
	if err != nil {
		log.Printf("[host] error creating invoke transaction request message: %s\n", err)
		return nil, err
	}

	wapcInstance, err := wc.wapcPool.Get(10 * time.Millisecond)
	if err != nil {
		log.Printf("[host] error getting waPC instance: %s\n", err)
		return nil, err
	}
	defer func() {
		err := wc.wapcPool.Return(wapcInstance)
		if err != nil {
			log.Printf("[host] error returning waPC instance: %s\n", err)
		}
	}()

	ctx := context.Background()
	result, err := wapcInstance.Invoke(ctx, "InvokeTransaction", args)
	if err != nil {
		log.Printf("[host] error invoking transaction: %s\n", err)
		return nil, err
	}

	log.Printf("[host] success result=%s\n", string(result))
	fmt.Println(string(result))
	return result, nil
}

func createInvokeTransactionArgs(channelID string, txID string, fnname string, params []string) ([]byte, error) {
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
		Args:            args}

	argsBuffer, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return argsBuffer, nil
}
