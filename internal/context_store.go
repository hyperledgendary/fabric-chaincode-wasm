// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"log"
	"sync"

	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type stubKey struct {
	channelID, txID string
}

// ContextStore keeps track of which stub belongs to which channel ID + transaction ID context
type ContextStore struct {
	sync.RWMutex
	stubs map[stubKey]shim.ChaincodeStubInterface
}

// NewContextStore returns a new store for keeping track of transaction context stubs
func NewContextStore() *ContextStore {
	store := ContextStore{}
	store.stubs = make(map[stubKey]shim.ChaincodeStubInterface)

	return &store
}

// Get returns the specified stub from the context store
func (store *ContextStore) Get(context *contract.TransactionContext) (shim.ChaincodeStubInterface, error) {
	key := stubKey{
		channelID: context.ChannelId,
		txID:      context.TransactionId,
	}

	log.Printf("[host] Getting stub for context chid %s txid %s\n", key.channelID, key.txID)

	store.RLock()
	defer store.RUnlock()

	if _, ok := store.stubs[key]; !ok {
		return nil, fmt.Errorf("No stub found for transaction context %s %s", key.channelID, key.txID)
	}

	stub := store.stubs[key]

	return stub, nil
}

// Put stores the passed stub in the context store
func (store *ContextStore) Put(channelID string, txID string, stub shim.ChaincodeStubInterface) error {
	key := stubKey{
		channelID,
		txID,
	}

	log.Printf("[host] Putting stub for context chid %s txid %s\n", key.channelID, key.txID)

	store.Lock()
	defer store.Unlock()

	if _, ok := store.stubs[key]; ok {
		return fmt.Errorf("Stub already exists for transaction context %s %s", key.channelID, key.txID)
	}

	store.stubs[key] = stub

	return nil
}

// Remove removes the specified stub from the context store
func (store *ContextStore) Remove(channelID string, txID string) error {
	key := stubKey{
		channelID,
		txID,
	}

	log.Printf("[host] Removing stub for context chid %s txid %s\n", key.channelID, key.txID)

	store.Lock()
	defer store.Unlock()

	if _, ok := store.stubs[key]; !ok {
		return fmt.Errorf("No stub found for transaction context %s %s", key.channelID, key.txID)
	}

	delete(store.stubs, key)

	return nil
}
