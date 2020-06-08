// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"log"

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

	log.Printf("ReadState done")
	return stub.GetState(request.StateKey)
}
