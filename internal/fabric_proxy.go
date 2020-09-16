// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

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
func (proxy *FabricProxy) FabricCall(ctx context.Context, binding, namespace, operation string, payload []byte) (result []byte, err error) {
	// Route the payload to any custom functionality accordingly.
	// You can even route to other waPC modules!!!
	log.Printf("[host] bd %s ns %s op %s payload length %d\n", binding, namespace, operation, len(payload))

	// Need to recover from any panics in FabricCall otherwise the chaincode
	// exits and, since this is being called by the Wasm guest code which was
	// itself called by the Wasm host, it's difficult to work out why
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[host] Recovering from panic in FabricCall: %v \nStack: %s \n", r, string(debug.Stack()))
			err = fmt.Errorf("Operation panicked: %s %s %s", binding, namespace, operation)
		}
	}()

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
		case "GetHash":
			log.Printf("[host] Processing GetHash...\n")
			return proxy.getHash(payload)
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
	stateKey := state.GetKey()
	log.Printf("[host] CreateState txid %s chid %s key %s value length %d\n", context.TransactionId, context.ChannelId, stateKey, len(state.Value))

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("CreateState failed: %s", err.Error())
	}

	collection := request.GetCollection()
	if collection != nil && collection.GetName() != "" {
		collectionName := collection.GetName()

		stateBytes, err := stub.GetPrivateData(collectionName, stateKey)
		if err != nil {
			return nil, fmt.Errorf("CreateState failed for collection %s: %s", collectionName, err.Error())
		}

		if stateBytes != nil {
			return nil, fmt.Errorf("CreateState failed for collection %s: State already exists for key %s", collectionName, stateKey)
		}

		err = stub.PutPrivateData(collectionName, stateKey, state.GetValue())
		if err != nil {
			return nil, fmt.Errorf("CreateState failed for collection %s: %s", collectionName, err.Error())
		}
	} else {
		stateBytes, err := stub.GetState(stateKey)
		if err != nil {
			return nil, fmt.Errorf("CreateState failed: %s", err.Error())
		}

		if stateBytes != nil {
			return nil, fmt.Errorf("CreateState failed: State already exists for key %s", stateKey)
		}

		err = stub.PutState(stateKey, state.GetValue())
		if err != nil {
			return nil, fmt.Errorf("CreateState failed: %s", err.Error())
		}
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
	stateKey := state.GetKey()
	log.Printf("[host] UpdateState txid %s chid %s key %s value length %d\n", context.TransactionId, context.ChannelId, stateKey, len(state.Value))

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("UpdateState failed: %s", err.Error())
	}

	collection := request.GetCollection()
	if collection != nil && collection.GetName() != "" {
		collectionName := collection.GetName()

		stateBytes, err := stub.GetPrivateData(collectionName, stateKey)
		if err != nil {
			return nil, fmt.Errorf("UpdateState failed for collection %s: %s", collectionName, err.Error())
		}

		if stateBytes == nil {
			return nil, fmt.Errorf("UpdateState failed for collection %s: No state exists for key %s", collectionName, stateKey)
		}

		err = stub.PutPrivateData(collectionName, stateKey, state.GetValue())
		if err != nil {
			return nil, fmt.Errorf("UpdateState failed for collection %s: %s", collectionName, err.Error())
		}
	} else {
		stateBytes, err := stub.GetState(stateKey)
		if err != nil {
			return nil, fmt.Errorf("UpdateState failed: %s", err.Error())
		}

		if stateBytes == nil {
			return nil, fmt.Errorf("UpdateState failed: No state exists for key %s", stateKey)
		}

		err = stub.PutState(stateKey, state.GetValue())
		if err != nil {
			return nil, fmt.Errorf("UpdateState failed: %s", err.Error())
		}
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
		return nil, fmt.Errorf("ReadState failed: %s", err.Error())
	}

	response := &contract.ReadStateResponse{}
	state := &contract.State{}

	var stateBytes []byte
	collection := request.GetCollection()
	if collection != nil && collection.GetName() != "" {
		collectionName := collection.GetName()

		stateBytes, err = stub.GetPrivateData(collectionName, stateKey)
		if err != nil {
			return nil, fmt.Errorf("ReadState failed for collection %s: %s", collectionName, err.Error())
		}

		if stateBytes == nil {
			return nil, fmt.Errorf("ReadState failed for collection %s: State %s does not exist", collectionName, stateKey)
		}
	} else {
		stateBytes, err = stub.GetState(stateKey)
		if err != nil {
			return nil, fmt.Errorf("ReadState failed: %s", err.Error())
		}

		if stateBytes == nil {
			return nil, fmt.Errorf("ReadState failed: State %s does not exist", stateKey)
		}
	}

	state.Key = stateKey
	state.Value = stateBytes
	response.State = state

	log.Printf("[host] Read State done\n")
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
		return nil, fmt.Errorf("ExistsState failed: %s", err.Error())
	}

	var stateBytes []byte
	collection := request.GetCollection()
	if collection != nil && collection.GetName() != "" {
		collectionName := collection.GetName()

		stateBytes, err = stub.GetPrivateData(collectionName, stateKey)
		if err != nil {
			return nil, fmt.Errorf("ExistsState failed for collection %s: %s", collectionName, err.Error())
		}
	} else {
		stateBytes, err = stub.GetState(stateKey)
		if err != nil {
			return nil, fmt.Errorf("ExistsState failed: %s", err.Error())
		}
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

func (proxy *FabricProxy) getHash(payload []byte) ([]byte, error) {
	request := &contract.GetHashRequest{}
	err := proto.Unmarshal(payload, request)
	if err != nil {
		return nil, err
	}

	context := request.GetContext()
	stateKey := request.GetStateKey()
	log.Printf("[host] GetHash txid %s chid %s key %s\n", context.TransactionId, context.ChannelId, request.StateKey)

	stub, err := proxy.contextStore.Get(context)
	if err != nil {
		return nil, fmt.Errorf("GetHash failed: %s", err.Error())
	}

	response := &contract.GetHashResponse{}

	var hashBytes []byte
	collection := request.GetCollection()
	if collection != nil && collection.GetName() != "" {
		collectionName := collection.GetName()

		hashBytes, err = stub.GetPrivateDataHash(collectionName, stateKey)
		if err != nil {
			return nil, fmt.Errorf("GetHash failed for collection %s: %s", collectionName, err.Error())
		}

		if hashBytes == nil {
			return nil, fmt.Errorf("GetHash failed for collection %s: State %s does not exist", collectionName, stateKey)
		}
	} else {
		return nil, fmt.Errorf("GetHash failed: Operation not supported for world state")
	}

	response.Hash = hashBytes

	log.Printf("[host] GetHash done\n")
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
