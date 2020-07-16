// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"
)

// CreateState is great
func CreateState(request *contract.CreateStateRequest) ([]byte, error) {
	context := request.Context
	state := request.State
	log.Printf("CreateState txid %s chid %s key %s value length %d\n", context.TransactionId, context.ChannelId, state.Key, len(state.Value))
	// stub := &shim.ChaincodeStub{
	// 	TxID:      context.TransactionId,
	// 	ChannelID: context.ChannelId,
	// }
	stash := GetStash()
	stub := stash.GetStub()

	err := stub.PutState(state.Key, state.Value)

	log.Printf("CreateState done")
	return nil, err
}

// ReadState is great
func ReadState(request *contract.ReadStateRequest) ([]byte, error) {
	context := request.Context
	log.Printf("ReadState txid %s chid %s key %s\n", context.TransactionId, context.ChannelId, request.StateKey)

	// stub := &shim.ChaincodeStub{
	// 	TxID:      context.TransactionId,
	// 	ChannelID: context.ChannelId,
	// }
	stash := GetStash()
	stub := stash.GetStub()

	response := &contract.ReadStateResponse{}
	state := &contract.State{}
	log.Printf("ReadState done")
	stateBytes, err := stub.GetState(request.StateKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	state.Key = request.StateKey
	state.Value = stateBytes
	response.State = state

	log.Printf("Read State done")
	return proto.Marshal(response)
}

// ExistsState is great
func ExistsState(request *contract.ExistsStateRequest) ([]byte, error) {
	context := request.Context
	log.Printf("ExistsState txid %s chid %s key %s\n", context.TransactionId, context.ChannelId, request.StateKey)

	// stub := &shim.ChaincodeStub{
	// 	TxID:      context.TransactionId,
	// 	ChannelID: context.ChannelId,
	// }
	stash := GetStash()
	stub := stash.GetStub()

	stateBytes, err := stub.GetState(request.StateKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	response := &contract.ExistsStateResponse{}

	if stateBytes == nil {
		response.Exists = false
	} else {
		response.Exists = true
	}

	return proto.Marshal(response)
}
