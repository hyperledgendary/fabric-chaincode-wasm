// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"fmt"
	"log"

	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"
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
			request := &contract.CreateStateRequest{}
			err := proto.Unmarshal(payload, request)
			if err != nil {
				return nil, err
			}

			return proxy.createState(request)
		case "ReadState":
			log.Printf("[host] Processing ReadStateRequest...\n")
			request := &contract.ReadStateRequest{}
			err := proto.Unmarshal(payload, request)
			if err != nil {
				return nil, err
			}

			return proxy.readState(request)
		case "ExistsState":
			log.Printf("[host] Processing ExistsStateRequest...\n")
			request := &contract.ExistsStateRequest{}
			err := proto.Unmarshal(payload, request)
			if err != nil {
				return nil, err
			}

			return proxy.existsState(request)
		case "UpdateState":
			log.Printf("[host] Processing UpdateStateRequest...\n")
			request := &contract.UpdateStateRequest{}
			err := proto.Unmarshal(payload, request)
			if err != nil {
				return nil, err
			}

			return proxy.updateState(request)
		}
	}

	return nil, fmt.Errorf("Operation not supported: %s %s %s", binding, namespace, operation)
}

func (proxy *FabricProxy) createState(request *contract.CreateStateRequest) ([]byte, error) {
	context := request.Context
	state := request.State
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
	return nil, err
}

func (proxy *FabricProxy) updateState(request *contract.UpdateStateRequest) ([]byte, error) {
	context := request.Context
	state := request.State
	log.Printf("[host] UpdateState txid %s chid %s key %s value length %d\n", context.TransactionId, context.ChannelId, state.Key, len(state.Value))

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("Failed to write to world state. %s", err.Error())
	}

	err = stub.PutState(state.Key, state.Value)
	if err != nil {
		return nil, fmt.Errorf("Failed to write to world state. %s", err.Error())
	}

	log.Printf("[host] UpdateState done")
	return nil, err
}

func (proxy *FabricProxy) readState(request *contract.ReadStateRequest) ([]byte, error) {
	context := request.Context
	log.Printf("[host] ReadState txid %s chid %s key %s\n", context.TransactionId, context.ChannelId, request.StateKey)

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

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

	log.Printf("[host] Read State done")
	return proto.Marshal(response)
}

func (proxy *FabricProxy) existsState(request *contract.ExistsStateRequest) ([]byte, error) {
	context := request.Context
	log.Printf("[host] ExistsState txid %s chid %s key %s\n", context.TransactionId, context.ChannelId, request.StateKey)

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

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

	log.Printf("[host] Exists State done")
	return proto.Marshal(response)
}
