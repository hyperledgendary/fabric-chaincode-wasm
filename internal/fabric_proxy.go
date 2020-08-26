// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"fmt"
	"log"

	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"google.golang.org/protobuf/proto"
)

// FabricProxy routes calls from Wasm contract to the correct Fabric stub
type FabricProxy struct {
	contextStore *ContextStore
}

// NewFabricProxy returns a new proxy to handle calls to the Fabric contract API
func NewFabricProxy(contextStore *ContextStore) *FabricProxy {
	proxy := FabricProxy{}
	proxy.contextStore = contextStore

	return &proxy
}

// FabricCall is the waPC HostCall function for interacting with the ledger
func (proxy *FabricProxy) FabricCall(ctx context.Context, binding, namespace, operation string, payload []byte) ([]byte, error) {
	// Route the payload to any custom functionality accordingly.
	// You can even route to other waPC modules!!!
	log.Printf("[host] bd %s ns %s op %s payload length %d\n", binding, namespace, operation, len(payload))

	if binding == "wapc" && namespace == "LedgerService" {
		switch operation {
		case "CreateState":
			log.Printf("[host] Processing CreateStateRequest...\n")
			return proxy.createState(payload)
		case "ReadState":
			log.Printf("[host] Processing ReadStateRequest...\n")
			return proxy.readState(payload)
		case "ExistsState":
			log.Printf("[host] Processing ExistsStateRequest...\n")
			return proxy.existsState(payload)
		case "UpdateState":
			log.Printf("[host] Processing UpdateStateRequest...\n")
			return proxy.updateState(payload)
		case "GetStates":
			log.Printf("[host] Processing GetStatesRequest...\n")
			return proxy.getStates(payload)
		}
	}

	return nil, fmt.Errorf("Operation not supported: %s %s %s", binding, namespace, operation)
}

func (proxy *FabricProxy) createState(payload []byte) ([]byte, error) {
	request := &contract.CreateStateRequest{}
	err := proto.Unmarshal(payload, request)
	if err != nil {
		return nil, err
	}

	context := request.GetContext()
	state := request.GetState()
	log.Printf("[host] CreateState txid %s chid %s key %s value length %d\n", context.TransactionId, context.ChannelId, state.Key, len(state.Value))

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("CreateState failed: %s", err.Error())
	}

	stateBytes, err := stub.GetState(state.Key)
	if err != nil {
		return nil, fmt.Errorf("CreateState failed: %s", err.Error())
	}

	if stateBytes != nil {
		return nil, fmt.Errorf("CreateState failed: State already exists for key %s", state.Key)
	}

	err = stub.PutState(state.Key, state.Value)
	if err != nil {
		return nil, fmt.Errorf("CreateState failed: %s", err.Error())
	}

	log.Printf("[host] CreateState done")
	return nil, nil
}

func (proxy *FabricProxy) updateState(payload []byte) ([]byte, error) {
	request := &contract.UpdateStateRequest{}
	err := proto.Unmarshal(payload, request)
	if err != nil {
		return nil, err
	}

	context := request.GetContext()
	state := request.GetState()
	log.Printf("[host] UpdateState txid %s chid %s key %s value length %d\n", context.TransactionId, context.ChannelId, state.Key, len(state.Value))

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("UpdateState failed: %s", err.Error())
	}

	stateBytes, err := stub.GetState(state.Key)
	if err != nil {
		return nil, fmt.Errorf("UpdateState failed: %s", err.Error())
	}

	if stateBytes == nil {
		return nil, fmt.Errorf("UpdateState failed: No state exists for key %s", state.Key)
	}

	err = stub.PutState(state.Key, state.Value)
	if err != nil {
		return nil, fmt.Errorf("UpdateState failed: %s", err.Error())
	}

	log.Printf("[host] UpdateState done")
	return nil, nil
}

func (proxy *FabricProxy) readState(payload []byte) ([]byte, error) {
	request := &contract.ReadStateRequest{}
	err := proto.Unmarshal(payload, request)
	if err != nil {
		return nil, err
	}

	context := request.GetContext()
	stateKey := request.GetStateKey()
	log.Printf("[host] ReadState txid %s chid %s key %s\n", context.TransactionId, context.ChannelId, request.StateKey)

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	response := &contract.ReadStateResponse{}
	state := &contract.State{}
	log.Printf("ReadState done")
	stateBytes, err := stub.GetState(stateKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	state.Key = request.StateKey
	state.Value = stateBytes
	response.State = state

	log.Printf("[host] Read State done")
	return proto.Marshal(response)
}

func (proxy *FabricProxy) existsState(payload []byte) ([]byte, error) {
	request := &contract.ExistsStateRequest{}
	err := proto.Unmarshal(payload, request)
	if err != nil {
		return nil, err
	}

	context := request.GetContext()
	stateKey := request.GetStateKey()
	log.Printf("[host] ExistsState txid %s chid %s key %s\n", context.TransactionId, context.ChannelId, request.StateKey)

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	stateBytes, err := stub.GetState(stateKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	response := &contract.ExistsStateResponse{}

	if stateBytes == nil {
		response.Exists = false
	} else {
		response.Exists = true
	}

	log.Printf("[host] Exists State done")
	return proto.Marshal(response)
}

func (proxy *FabricProxy) getStates(payload []byte) ([]byte, error) {
	request := &contract.GetStatesRequest{}
	err := proto.Unmarshal(payload, request)
	if err != nil {
		return nil, err
	}

	context := request.GetContext()
	log.Printf("[host] GetStates txid %s chid %s\n", context.TransactionId, context.ChannelId)

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("GetStates failed: %s", err.Error())
	}

	switch qt := request.Query.(type) {
	case *contract.GetStatesRequest_ByKeyRange:
		keyRangeQuery := request.GetByKeyRange()
		return proxy.getStatesByKeyRange(stub, keyRangeQuery)
	default:
		return nil, fmt.Errorf("GetStates failed: unsupported query type %T", qt)
	}
}

func (proxy *FabricProxy) getStatesByKeyRange(stub shim.ChaincodeStubInterface, query *contract.KeyRangeQuery) ([]byte, error) {

	resultsIterator, err := stub.GetStateByRange(query.StartKey, query.EndKey)
	if err != nil {
		return nil, fmt.Errorf("GetStates (ByKeyRange) failed: %s", err.Error())
	}
	defer resultsIterator.Close()

	response := &contract.GetStatesResponse{}
	states := []*contract.State{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("GetStates (ByKeyRange) failed: %s", err.Error())
		}

		state := &contract.State{}
		state.Key = queryResponse.Key
		state.Value = queryResponse.Value

		states = append(states, state)
	}
	response.States = states

	log.Printf("[host] Get States (ByKeyRange) done")
	return proto.Marshal(response)
}
