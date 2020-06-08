// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// Stash is a big hack! (Not thread safe and just an all round bad idea)
type Stash struct {
	stub shim.ChaincodeStubInterface
}

var stash *Stash

// GetStub gets the stub (eeek!)
func (s *Stash) GetStub() shim.ChaincodeStubInterface {
	return s.stub
}

// SetStub sets the stub (yikes!)
func (s *Stash) SetStub(newStub shim.ChaincodeStubInterface) {
	s.stub = newStub
}

// GetStash get the stash
func GetStash() *Stash {
	if stash == nil {
		stash = &Stash{}
	}

	return stash
}
