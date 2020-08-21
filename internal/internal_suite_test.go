// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes/stub.go --fake-name ChaincodeStubInterface github.com/hyperledger/fabric-chaincode-go/shim.ChaincodeStubInterface
//counterfeiter:generate -o fakes/state_query_iterator.go --fake-name StateQueryIteratorInterface github.com/hyperledger/fabric-chaincode-go/shim.StateQueryIteratorInterface

package internal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Internal Suite")
}
